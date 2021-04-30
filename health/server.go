package health

import (
	"context"

	pb "github.com/buzzsurfr/grpc-alb-health-check/health/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"
)

type Server struct {
	pb.UnimplementedALBServer
	checkService string
	healthSrv    *health.Server
}

func (s *Server) Healthcheck(ctx context.Context, in *pb.HealthCheckRequest) (*pb.HealthCheckResponse, error) {
	blank := &pb.HealthCheckResponse{}
	res, err := s.healthSrv.Check(ctx, &grpc_health_v1.HealthCheckRequest{})
	if err != nil {
		return blank, err
	}
	if res.GetStatus() != grpc_health_v1.HealthCheckResponse_SERVING {
		return blank, status.Errorf(codes.Unavailable, "service %s not available", s.checkService)
	}
	return blank, nil
}

type ServerOption func(*Server)

func NewServer(options ...ServerOption) *Server {
	s := &Server{
		checkService: "",
		healthSrv:    health.NewServer(),
	}

	for _, option := range options {
		option(s)
	}

	return s
}

func WithService(name string) ServerOption {
	return func(s *Server) {
		s.checkService = name
	}
}

func WithHealthServer(srv *health.Server) ServerOption {
	return func(s *Server) {
		s.healthSrv = srv
	}
}
