package grpc

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/Abdurahmanit/GroupProject/order-service/internal/platform/logger"
	orderservicepb "github.com/Abdurahmanit/GroupProject/order-service/proto/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"
)

type Server struct {
	grpcServer      *grpc.Server
	log             logger.Logger
	port            string
	timeoutGraceful time.Duration
}

func NewServer(
	log logger.Logger,
	port string,
	timeoutGraceful time.Duration,
	maxConnectionIdle time.Duration,
	orderService orderservicepb.OrderServiceServer,
) *Server {

	serverOpts := []grpc.ServerOption{
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle:     maxConnectionIdle,
			Timeout:               20 * time.Second,
			MaxConnectionAge:      maxConnectionIdle,
			Time:                  maxConnectionIdle,
			MaxConnectionAgeGrace: 5 * time.Second,
		}),
	}

	grpcServer := grpc.NewServer(serverOpts...)

	if orderService != nil {
		orderservicepb.RegisterOrderServiceServer(grpcServer, orderService)
	}

	reflection.Register(grpcServer)

	return &Server{
		grpcServer:      grpcServer,
		log:             log,
		port:            port,
		timeoutGraceful: timeoutGraceful,
	}
}

func (s *Server) Start() error {
	s.log.Infof("gRPC server is starting on port %s", s.port)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", s.port))
	if err != nil {
		return fmt.Errorf("failed to listen on port %s: %w", s.port, err)
	}

	err = s.grpcServer.Serve(lis)
	if err != nil {
		return fmt.Errorf("gRPC server failed to serve: %w", err)
	}
	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	s.log.Info("gRPC server is stopping gracefully")

	stopped := make(chan struct{})
	go func() {
		s.grpcServer.GracefulStop()
		close(stopped)
	}()

	select {
	case <-ctx.Done():
		s.log.Warn("graceful shutdown timed out, forcing stop")
		s.grpcServer.Stop()
		return ctx.Err()
	case <-stopped:
		s.log.Info("gRPC server stopped gracefully")
		return nil
	}
}
