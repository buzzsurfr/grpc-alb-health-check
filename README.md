# grpc-alb-health-check
ALB Health Check gRPC Implementation

This is a simple package that wraps the [gRPC Health Check Protocol](https://github.com/grpc/grpc/blob/master/doc/health-checking.md) for use with the [Application Load Balancer](https://aws.amazon.com/elasticloadbalancing/application-load-balancer/)'s default gRPC health check.

The key difference from the gRPC Health Check Protocol is this service will return an error _unless_ the specified service (or the server if no service is specified) returns a `SERVING` status.

## Example

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
