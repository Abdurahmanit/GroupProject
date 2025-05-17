package main

import (
	"context"
	"log"
	"net"

	grpcAdapter "github.com/Abdurahmanit/GroupProject/listing-service/internal/adapter/grpc"
	"github.com/Abdurahmanit/GroupProject/listing-service/internal/adapter/messaging/nats"
	"github.com/Abdurahmanit/GroupProject/listing-service/internal/adapter/repository/mongodb"
	"github.com/Abdurahmanit/GroupProject/listing-service/internal/adapter/storage/s3"
	"github.com/Abdurahmanit/GroupProject/listing-service/internal/config"
	pb "github.com/Abdurahmanit/GroupProject/listing-service/genproto/listing_service"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Connect to MongoDB
	mongoClient, err := mongo.Connect(context.Background(), options.Client().ApplyURI(cfg.MongoURI))
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer mongoClient.Disconnect(context.Background())
	db := mongoClient.Database("bike_store")

	// Initialize repositories
	listingRepo := mongodb.NewListingRepository(db)
	favoriteRepo := mongodb.NewFavoriteRepository(db)

	// Initialize storage (MinIO/S3)
	storageClient, err := s3.NewS3Storage(cfg.MinIOEndpoint, cfg.MinIOAccessKey, cfg.MinIOSecretKey, cfg.MinIOBucket)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}

	// Initialize NATS publisher
	natsPublisher, err := nats.NewPublisher(cfg.NATSURL)
	if err != nil {
		log.Fatalf("Failed to initialize NATS: %v", err)
	}
	defer natsPublisher.Close()

	// Set up gRPC server
	lis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer, cleanup := grpcAdapter.NewGRPCServer()
	defer cleanup()

	handler := grpcAdapter.NewHandler(listingRepo, favoriteRepo, storageClient, natsPublisher)
	pb.RegisterListingServiceServer(grpcServer, handler)

	log.Printf("Starting gRPC server on port %s", cfg.GRPCPort)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
