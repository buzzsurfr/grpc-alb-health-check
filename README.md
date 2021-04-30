# grpc-alb-health-check
ALB Health Check gRPC Implementation

This is a simple package that wraps the [gRPC Health Check Protocol](https://github.com/grpc/grpc/blob/master/doc/health-checking.md) for use with the [Application Load Balancer](https://aws.amazon.com/elasticloadbalancing/application-load-balancer/)'s default gRPC health check.

The key difference from the gRPC Health Check Protocol is this service will return an error _unless_ the specified service (or the server if no service is specified) returns a `SERVING` status.
