package grpc

import (
	"fmt"
	"net"

	"github.com/Abdurahmanit/GroupProject/news-service/internal/config"
	newspb "github.com/Abdurahmanit/GroupProject/news-service/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type Server struct {
	cfg         *config.GRPCConfig
	logger      *zap.Logger
	newsService newspb.NewsServiceServer
}

func NewServer(
	cfg *config.GRPCConfig,
	logger *zap.Logger,
	newsService newspb.NewsServiceServer,
) *Server {
	return &Server{
		cfg:         cfg,
		logger:      logger,
		newsService: newsService,
	}
}

func (s *Server) Run() error {
	addr := fmt.Sprintf(":%s", s.cfg.Port)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		s.logger.Fatal("Failed to listen gRPC port", zap.String("address", addr), zap.Error(err))
		return fmt.Errorf("failed to listen on %s: %w", addr, err)
	}

	grpcServer := grpc.NewServer(
		grpc.MaxRecvMsgSize(s.cfg.MaxRecvMsgSize),
		grpc.MaxSendMsgSize(s.cfg.MaxSendMsgSize),
	)

	newspb.RegisterNewsServiceServer(grpcServer, s.newsService)
	reflection.Register(grpcServer)

	s.logger.Info("gRPC server started", zap.String("address", addr))

	if err := grpcServer.Serve(lis); err != nil {
		s.logger.Error("Failed to serve gRPC server", zap.Error(err))
		return fmt.Errorf("failed to serve gRPC server: %w", err)
	}

	return nil
}
