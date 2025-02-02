package grpc

import (
	"context"
	"fmt"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"log/slog"
	"math"
	"net"
)

var maxSendMsgSize = grpc.MaxSendMsgSize(math.MaxInt32)

type ServerRegistrationFunc func(registrar grpc.ServiceRegistrar, mux *runtime.ServeMux) error

type Server struct {
	logger    *slog.Logger
	config    PublicGrpcConfig
	grpcSever *grpc.Server
}

// NewGRPC new grpc server
func NewGRPC(logger *slog.Logger, config PublicGrpcConfig) *Server {
	opts := []grpc.ServerOption{
		maxSendMsgSize,
	}

	return &Server{
		logger:    logger,
		config:    config,
		grpcSever: grpc.NewServer(opts...),
	}
}

// AddServerImplementation adds server implementation
func (s *Server) AddServerImplementation(regFunc func(registrar grpc.ServiceRegistrar)) *Server {
	regFunc(s.grpcSever)
	return s
}

// AddGrpcHealthCheck adds grpc health check
func (s *Server) AddGrpcHealthCheck() *Server {
	hs := health.NewServer()
	hs.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
	healthpb.RegisterHealthServer(s.grpcSever, hs)
	return s
}

// ListenAndServe start grpc-gateway
func (s *Server) ListenAndServe() error {
	network := s.config.GrpcProtocol
	address := s.config.GrpcAddress

	s.logger.Info("Starting gRPC server", slog.String("grpc", fmt.Sprintf("%v:%v", network, address)))

	l, err := net.Listen(network, address)
	if err != nil {
		return err
	}

	return s.grpcSever.Serve(l)
}

// Shutdown to run on app shutdown
func (s *Server) Shutdown(_ context.Context) error {
	s.grpcSever.GracefulStop()
	return nil
}
