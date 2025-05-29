package client

import (
	"fmt"
	"time"

	listingpb "github.com/Abdurahmanit/GroupProject/listing-service/genproto/listing_service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

const (
	listingServiceDialTimeout = 5 * time.Second
)

type ListingServiceClientConfig struct {
	Address string // Например, "localhost:50053" или "listing-service:50053" в Docker
}

func NewListingServiceClient(cfg ListingServiceClientConfig) (listingpb.ListingServiceClient, *grpc.ClientConn, error) {
	if cfg.Address == "" {
		return nil, nil, fmt.Errorf("listing service address is not configured")
	}

	dialOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                10 * time.Second,
			Timeout:             20 * time.Second,
			PermitWithoutStream: true,
		}),
	}

	conn, err := grpc.Dial(cfg.Address, dialOpts...)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to dial listing service at %s: %w", cfg.Address, err)
	}

	client := listingpb.NewListingServiceClient(conn)

	return client, conn, nil
}
