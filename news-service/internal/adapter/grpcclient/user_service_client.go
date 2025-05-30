package grpcclient

import (
	"context"
	"fmt"
	"time"

	usergrpc "github.com/Abdurahmanit/GroupProject/news-service/internal/clients/usergrpc"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

type UserServiceClient interface {
	GetAuthorEmail(ctx context.Context, authorID string) (string, error)
	Close() error
}

type userServiceGRPCClient struct {
	conn   *grpc.ClientConn
	client usergrpc.UserServiceClient
	logger *zap.Logger
}

func NewUserServiceGRPCClient(targetAddress string, logger *zap.Logger) (UserServiceClient, error) {
	logger.Info("Attempting to connect to User Service via gRPC", zap.String("address", targetAddress))

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	}

	dialCtx, dialCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer dialCancel()

	conn, err := grpc.DialContext(dialCtx, targetAddress, opts...)
	if err != nil {
		logger.Error("Failed to dial User Service", zap.String("address", targetAddress), zap.Error(err))
		return nil, fmt.Errorf("failed to dial user service at %s: %w", targetAddress, err)
	}
	logger.Info("Successfully connected to User Service via gRPC", zap.String("address", targetAddress))

	client := usergrpc.NewUserServiceClient(conn)
	return &userServiceGRPCClient{
		conn:   conn,
		client: client,
		logger: logger.Named("UserServiceClient"),
	}, nil
}

func (c *userServiceGRPCClient) GetAuthorEmail(ctx context.Context, authorID string) (string, error) {
	c.logger.Debug("Requesting profile from User Service", zap.String("author_id", authorID))
	req := &usergrpc.GetProfileRequest{UserId: authorID}

	callCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	resp, err := c.client.GetProfile(callCtx, req)
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			c.logger.Error("User Service GetProfile RPC failed",
				zap.String("author_id", authorID),
				zap.String("grpc_code", st.Code().String()),
				zap.String("grpc_message", st.Message()),
				zap.Error(err),
			)
			if st.Code() == codes.NotFound {
				return "", fmt.Errorf("author with id %s not found in user service: %w", authorID, err)
			}
		} else {
			c.logger.Error("User Service GetProfile call failed with non-gRPC error",
				zap.String("author_id", authorID),
				zap.Error(err),
			)
		}
		return "", fmt.Errorf("user service GetProfile failed for author %s: %w", authorID, err)
	}

	if resp == nil {
		c.logger.Warn("User Service returned nil profile", zap.String("author_id", authorID))
		return "", fmt.Errorf("user service returned nil profile for author %s", authorID)
	}

	userEmail := resp.GetEmail()
	if userEmail == "" {
		c.logger.Warn("User Service returned profile with empty email", zap.String("author_id", authorID))
		return "", fmt.Errorf("user service returned profile with empty email for author %s", authorID)
	}

	c.logger.Info("Successfully retrieved email from User Service", zap.String("author_id", authorID), zap.String("email", userEmail))
	return userEmail, nil
}

func (c *userServiceGRPCClient) Close() error {
	if c.conn != nil {
		c.logger.Info("Closing User Service gRPC client connection")
		return c.conn.Close()
	}
	return nil
}
