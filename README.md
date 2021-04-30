# grpc-alb-health-check
ALB Health Check gRPC Implementation

This is a simple package that wraps the [gRPC Health Check Protocol](https://github.com/grpc/grpc/blob/master/doc/health-checking.md) for use with the [Application Load Balancer](https://aws.amazon.com/elasticloadbalancing/application-load-balancer/)'s default gRPC health check.

The key difference from the gRPC Health Check Protocol is this service will return an error _unless_ the specified service (or the server if no service is specified) returns a `SERVING` status.

## Examples

### Add to existing Go package

```go
package main

import (
	albHealth "github.com/buzzsurfr/grpc-alb-health-check/health"
	albpb "github.com/buzzsurfr/grpc-alb-health-check/health/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

func main() {
	...

	s := grpc.NewServer()                           // Standard gRPC server creation
	healthcheck := health.NewServer()               // Create a gRPC Health Check server
	healthpb.RegisterHealthServer(s, healthcheck)   // Register the Health Check server
	albhealthcheck := albHealth.NewServer(albHealth.WithHealthServer(healthcheck))
	albpb.RegisterALBServer(s, albhealthcheck)

	...
}

```

### Run as proxy (sidecar or separate binary)

#### Usage

```
A simple proxy which listens on AWS.ALB/Healthcheck and proxies
a request to grpc.health.v1.Health/Check.

The key difference from the gRPC Health Check Protocol is
the service will return an error unless the specified
service (or the server if no service is specified) returns
a SERVING status.

Usage:
  proxy [flags]

Flags:
  -a, --address string     address:port for the grpc.health.v1.Health service (default "localhost:50051")
  -h, --help               help for proxy
  -p, --port int           Listener port (default 50052)
      --timeout duration   health check timeout (default 1s)
```
