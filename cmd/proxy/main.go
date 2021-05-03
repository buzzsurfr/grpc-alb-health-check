package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"time"

	albpb "github.com/buzzsurfr/grpc-alb-health-check/health/v1"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var (
	port           = 50052
	address        = "localhost:50051"
	service        = ""
	connectTimeout = time.Second
	watchCmd       = false
	rootCmd        = &cobra.Command{
		Use:   "proxy",
		Short: "ALB to gRPC health check proxy",
		Long: `A simple proxy which listens on AWS.ALB/Healthcheck and proxies
a request to grpc.health.v1.Health/Check.

The key difference from the gRPC Health Check Protocol is
the service will return an error unless the specified
service (or the server if no service is specified) returns
a SERVING status.`,
		// Uncomment the following line if your bare application
		// has an action associated with it:
		RunE: func(cmd *cobra.Command, args []string) error {
			// Connect to backend
			clientCtx, clientCancel := context.WithTimeout(context.Background(), connectTimeout)
			defer clientCancel()

			conn, err := grpc.DialContext(clientCtx, address, grpc.WithInsecure(), grpc.WithBlock())
			if err != nil {
				if err == context.DeadlineExceeded {
					return status.Errorf(codes.Unavailable, "timeout: failed to connect to %v within %v", address, connectTimeout)
				} else {
					return status.Errorf(codes.Unavailable, "failed to connect to %v: %v", address, err)
				}
			}
			defer conn.Close()

			// Listen for frontend connections
			lis, err := net.Listen("tcp", validatePort(port))
			if err != nil {
				log.Fatalf("failed to listen: %v", err)
			}

			s := grpc.NewServer()
			albhealthcheck := &server{
				currentStatus: healthpb.HealthCheckResponse_SERVICE_UNKNOWN,
				healthClient:  healthpb.NewHealthClient(conn),
			}
			albpb.RegisterALBServer(s, albhealthcheck)

			if watchCmd {
				if err := albhealthcheck.watch(context.Background()); err != nil {
					return err
				}
			}

			if err := s.Serve(lis); err != nil {
				log.Fatalf("failed to serve: %v", err)
			}

			return nil
		},
	}
)

type server struct {
	albpb.UnimplementedALBServer
	currentStatus healthpb.HealthCheckResponse_ServingStatus
	healthClient  healthpb.HealthClient
}

func (s *server) watch(ctx context.Context) error {
	stream, err := s.healthClient.Watch(ctx, &healthpb.HealthCheckRequest{
		Service: service,
	})
	if err != nil {
		return err
	}

	go func() {
		for {
			res, err := stream.Recv()
			if err != nil {
				s.currentStatus = healthpb.HealthCheckResponse_SERVICE_UNKNOWN
				if err == io.EOF {
					log.Print("stream closed by server")
				} else {
					log.Printf("stream error: %v", err)
				}
				return
			}
			s.currentStatus = res.GetStatus()
		}
	}()

	return nil
}

func (s *server) Healthcheck(ctx context.Context, in *albpb.HealthCheckRequest) (*albpb.HealthCheckResponse, error) {
	blank := &albpb.HealthCheckResponse{}

	if watchCmd {
		return blank, parseStatus(s.currentStatus)
	}

	res, err := s.healthClient.Check(ctx, &healthpb.HealthCheckRequest{
		Service: service,
	})

	if err != nil {
		return blank, err
	}
	if res.GetStatus() != healthpb.HealthCheckResponse_SERVING {
		return blank, status.Errorf(codes.Unavailable, "service %s not available", service)
	}
	return blank, parseStatus(res.GetStatus())
}

func validatePort(port int) string {
	if port < 65536 {
		return fmt.Sprintf(":%d", port)
	}
	return ""
}

func parseStatus(res healthpb.HealthCheckResponse_ServingStatus) error {
	if res != healthpb.HealthCheckResponse_SERVING {
		return status.Errorf(codes.Unavailable, "service %s is not available", service)
	}
	return nil
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.Flags().IntVarP(&port, "port", "p", 50052, "Listener port")
	rootCmd.Flags().StringVarP(&address, "address", "a", "localhost:50051", "address:port for the grpc.health.v1.Health service")
	rootCmd.Flags().DurationVar(&connectTimeout, "timeout", time.Second, "backend connection timeout")
	rootCmd.Flags().BoolVarP(&watchCmd, "watch", "w", false, "use watch instead of check")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".proxy" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".proxy")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}

func main() {
	cobra.CheckErr(rootCmd.Execute())
}
