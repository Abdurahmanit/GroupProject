package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/Abdurahmanit/GroupProject/order-service/internal/app/config"
	"github.com/Abdurahmanit/GroupProject/order-service/internal/platform/logger"
	grpcserver "github.com/Abdurahmanit/GroupProject/order-service/internal/port/grpc"
	orderservicepb "github.com/Abdurahmanit/GroupProject/order-service/proto/service"
)

type App struct {
	cfg    *config.Config
	log    logger.Logger
	server *grpcserver.Server
}

func New(cfg *config.Config) (*App, error) {
	logCfg := logger.ZapLoggerConfig{
		Level:      cfg.Logger.Level,
		Encoding:   cfg.Logger.Encoding,
		TimeFormat: cfg.Logger.TimeFormat,
	}
	appLogger, err := logger.NewZapLogger(logCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}

	appLogger.Info("Logger initialized")
	appLogger.Infof("Configuration loaded: Env=%s", cfg.Env)

	var orderServiceImplementation orderservicepb.OrderServiceServer

	appLogger.Warn("OrderServiceServer implementation is not yet initialized (using nil)")

	grpcSrv := grpcserver.NewServer(
		appLogger,
		cfg.GRPCServer.Port,
		cfg.GRPCServer.TimeoutGraceful,
		cfg.GRPCServer.MaxConnectionIdle,
		orderServiceImplementation,
	)

	return &App{
		cfg:    cfg,
		log:    appLogger,
		server: grpcSrv,
	}, nil
}

func (a *App) Run() {
	a.log.Info("Starting application")

	go func() {
		if err := a.server.Start(); err != nil {
			a.log.Fatalf("Failed to start gRPC server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	a.log.Info("Shutting down application")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), a.cfg.GRPCServer.TimeoutGraceful)
	defer cancel()

	if err := a.server.Stop(shutdownCtx); err != nil {
		a.log.Errorf("Error during gRPC server graceful shutdown: %v", err)
	}

	a.log.Info("Application shut down successfully")
}
