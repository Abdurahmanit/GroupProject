package integration

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"testing"

	pb "github.com/Abdurahmanit/GroupProject/review-service"
	grpcAdapter "github.com/Abdurahmanit/GroupProject/review-service/internal/adapter/grpc"
	natsAdapter "github.com/Abdurahmanit/GroupProject/review-service/internal/adapter/messaging/nats"
	mongoRepo "github.com/Abdurahmanit/GroupProject/review-service/internal/adapter/repository/mongodb"
	"github.com/Abdurahmanit/GroupProject/review-service/internal/config"
	"github.com/Abdurahmanit/GroupProject/review-service/internal/domain"
	"github.com/Abdurahmanit/GroupProject/review-service/internal/middleware" // For context keys
	platformLogger "github.com/Abdurahmanit/GroupProject/review-service/internal/platform/logger"
	"github.com/Abdurahmanit/GroupProject/review-service/internal/usecase"

	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	gogrpc "google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

var (
	testDBClient   *mongo.Client
	testReviewRepo *mongoRepo.ReviewRepository
	testNatsURL    string
	testNatsPub    *natsAdapter.Publisher
	reviewClient   pb.ReviewServiceClient
	testLogger     *platformLogger.Logger
	testCfg        *config.Config
)

const (
	testProductID        = "product123"
	testAnotherProductID = "product789"
	testUserID           = "user456"
	testAnotherUserID    = "userABC"
	testAdminID          = "adminXYZ"
	adminRole            = "admin"
	customerRole         = "customer"
)

// TestMain sets up the test environment (MongoDB, NATS, gRPC server).
func TestMain(m *testing.M) {
	var err error
	testLogger = platformLogger.NewLogger()

	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not construct pool: %s", err)
	}
	err = pool.Client.Ping()
	if err != nil {
		log.Fatalf("Could not connect to Docker: %s", err)
	}

	mongoResource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "mongo",
		Tag:        "5.0",
		Env: []string{
			"MONGO_INITDB_ROOT_USERNAME=root",
			"MONGO_INITDB_ROOT_PASSWORD=password",
		},
	}, func(config *docker.HostConfig) {
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	if err != nil {
		log.Fatalf("Could not start MongoDB resource: %s", err)
	}
	mongoURI := fmt.Sprintf("mongodb://root:password@%s/test_reviews_db?authSource=admin", mongoResource.GetHostPort("27017/tcp"))

	natsResource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "nats",
		Tag:        "2.9",
		Cmd:        []string{"-js"},
	}, func(config *docker.HostConfig) {
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	if err != nil {
		log.Fatalf("Could not start NATS resource: %s", err)
	}
	testNatsURL = fmt.Sprintf("nats://%s", natsResource.GetHostPort("4222/tcp"))

	if err := pool.Retry(func() error {
		var errRetry error
		testDBClient, errRetry = mongo.Connect(context.Background(), options.Client().ApplyURI(mongoURI))
		if errRetry != nil {
			return errRetry
		}
		return testDBClient.Ping(context.Background(), nil)
	}); err != nil {
		log.Fatalf("Could not connect to MongoDB: %s", err)
	}

	if err := pool.Retry(func() error {
		var errRetry error
		testNatsPub, errRetry = natsAdapter.NewPublisher(testNatsURL, testLogger, "test-review-service-integration")
		if errRetry != nil {
			testLogger.Error("NATS connection attempt failed in TestMain", zap.Error(errRetry))
			return errRetry
		}
		return nil
	}); err != nil {
		log.Fatalf("Could not connect to NATS: %s", err)
	}

	db := testDBClient.Database("test_reviews_db")
	testReviewRepo, err = mongoRepo.NewReviewRepository(db, testLogger)
	if err != nil {
		log.Fatalf("Could not create test review repository: %s", err)
	}
	reviewUsecase := usecase.NewReviewUsecase(testReviewRepo, testNatsPub, testLogger)

	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		log.Fatalf("Failed to listen on a port: %v", err)
	}
	grpcTestServerAddr := listener.Addr().String()
	testLogger.Info("Test gRPC server listening", zap.String("address", grpcTestServerAddr))

	testCfg = &config.Config{
		JWTSecret: "test-secret-for-integration",
	}

	publicMethods := map[string]bool{
		"/review.ReviewService/GetReview":               true,
		"/review.ReviewService/ListReviewsByProduct":    true,
		"/review.ReviewService/GetProductAverageRating": true,
	}
	requiredRoles := map[string][]string{
		"/review.ReviewService/ModerateReview": {adminRole},
	}

	grpcServer := grpcAdapter.NewGRPCServerWithInterceptors(testLogger, testCfg.JWTSecret, nil, publicMethods, requiredRoles)
	pb.RegisterReviewServiceServer(grpcServer, grpcAdapter.NewReviewHandler(reviewUsecase, testLogger))

	go func() {
		if err := grpcServer.Serve(listener); err != nil {
			testLogger.Error("Test gRPC server failed to serve", zap.Error(err))
		}
	}()
	defer grpcServer.Stop()

	conn, err := gogrpc.Dial(grpcTestServerAddr, gogrpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to test gRPC server: %v", err)
	}
	defer conn.Close()
	reviewClient = pb.NewReviewServiceClient(conn)

	code := m.Run()

	if err := pool.Purge(mongoResource); err != nil {
		log.Fatalf("Could not purge MongoDB resource: %s", err)
	}
	if err := pool.Purge(natsResource); err != nil {
		log.Fatalf("Could not purge NATS resource: %s", err)
	}
	testNatsPub.Close()
	os.Exit(code)
}

func clearReviewsCollection(t *testing.T) {
	_, err := testDBClient.Database("test_reviews_db").Collection("reviews").DeleteMany(context.Background(), bson.M{})
	require.NoError(t, err, "Failed to clear reviews collection")
}

func createAuthContext(userID, userRole string) context.Context {
	md := metadata.New(map[string]string{
		string(middleware.UserIDKey):   userID,
		string(middleware.UserRoleKey): userRole,
		"authorization":                "Bearer mocktokenfor_" + userID + "_" + userRole,
	})
	return metadata.NewOutgoingContext(context.Background(), md)
}

// --- Test Cases ---

func TestCreateAndGetReview(t *testing.T) {
	clearReviewsCollection(t)
	ctx := createAuthContext(testUserID, customerRole)

	createReq := &pb.CreateReviewRequest{
		UserId:    testUserID,
		ProductId: testProductID,
		Rating:    5,
		Comment:   "Excellent product!",
	}

	createdReview, err := reviewClient.CreateReview(ctx, createReq)
	require.NoError(t, err)
	require.NotNil(t, createdReview)
	assert.Equal(t, testUserID, createdReview.UserId)
	assert.Equal(t, testProductID, createdReview.ProductId)
	assert.Equal(t, int32(5), createdReview.Rating)
	assert.Equal(t, "Excellent product!", createdReview.Comment)
	assert.NotEmpty(t, createdReview.Id)
	assert.Equal(t, string(domain.ReviewStatusPending), createdReview.Status)

	getReq := &pb.GetReviewRequest{ReviewId: createdReview.Id}
	fetchedReview, err := reviewClient.GetReview(context.Background(), getReq)
	require.NoError(t, err)
	require.NotNil(t, fetchedReview)
	assert.Equal(t, createdReview.Id, fetchedReview.Id)
}

func TestCreateReview_InvalidInput_Rating(t *testing.T) {
	clearReviewsCollection(t)
	ctx := createAuthContext(testUserID, customerRole)
	createReq := &pb.CreateReviewRequest{UserId: testUserID, ProductId: testProductID, Rating: 0, Comment: "Bad rating"}
	_, err := reviewClient.CreateReview(ctx, createReq)
	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.InvalidArgument, st.Code())
	assert.Contains(t, st.Message(), "rating must be between 1 and 5")
}

func TestCreateReview_Duplicate(t *testing.T) {
	clearReviewsCollection(t)
	ctx := createAuthContext(testUserID, customerRole)
	createReq := &pb.CreateReviewRequest{UserId: testUserID, ProductId: testProductID, Rating: 4, Comment: "First review"}
	_, err := reviewClient.CreateReview(ctx, createReq)
	require.NoError(t, err)

	_, err = reviewClient.CreateReview(ctx, createReq)
	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.AlreadyExists, st.Code())
	assert.Contains(t, st.Message(), domain.ErrReviewAlreadyExists.Error())
}

func TestUpdateReview_ByAuthor_Success(t *testing.T) {
	clearReviewsCollection(t)
	authCtx := createAuthContext(testUserID, customerRole)

	created, _ := reviewClient.CreateReview(authCtx, &pb.CreateReviewRequest{UserId: testUserID, ProductId: testProductID, Rating: 3, Comment: "Initial comment"})
	require.NotNil(t, created)

	updateReq := &pb.UpdateReviewRequest{
		ReviewId: created.Id,
		UserId:   testUserID,
		Rating:   4,
		Comment:  "Updated comment",
	}
	updatedReview, err := reviewClient.UpdateReview(authCtx, updateReq)
	require.NoError(t, err)
	require.NotNil(t, updatedReview)
	assert.Equal(t, int32(4), updatedReview.Rating)
	assert.Equal(t, "Updated comment", updatedReview.Comment)
	assert.NotEqual(t, created.UpdatedAt, updatedReview.UpdatedAt)
}

func TestUpdateReview_ByNonAuthor_Forbidden(t *testing.T) {
	clearReviewsCollection(t)
	authorCtx := createAuthContext(testUserID, customerRole)
	nonAuthorCtx := createAuthContext(testAnotherUserID, customerRole)

	created, _ := reviewClient.CreateReview(authorCtx, &pb.CreateReviewRequest{UserId: testUserID, ProductId: testProductID, Rating: 3, Comment: "Initial"})

	updateReq := &pb.UpdateReviewRequest{ReviewId: created.Id, UserId: testAnotherUserID, Rating: 5}
	_, err := reviewClient.UpdateReview(nonAuthorCtx, updateReq)
	require.Error(t, err)
	st, _ := status.FromError(err)
	assert.Equal(t, codes.PermissionDenied, st.Code())
}

func TestDeleteReview_ByAuthor_Success(t *testing.T) {
	clearReviewsCollection(t)
	authCtx := createAuthContext(testUserID, customerRole)
	created, _ := reviewClient.CreateReview(authCtx, &pb.CreateReviewRequest{UserId: testUserID, ProductId: testProductID, Rating: 2, Comment: "To be deleted"})

	_, err := reviewClient.DeleteReview(authCtx, &pb.DeleteReviewRequest{ReviewId: created.Id, UserId: testUserID})
	require.NoError(t, err)

	_, err = reviewClient.GetReview(context.Background(), &pb.GetReviewRequest{ReviewId: created.Id})
	require.Error(t, err)
	st, _ := status.FromError(err)
	assert.Equal(t, codes.NotFound, st.Code())
}

func TestDeleteReview_ByNonAuthor_Forbidden(t *testing.T) {
	clearReviewsCollection(t)
	authorCtx := createAuthContext(testUserID, customerRole)
	nonAuthorCtx := createAuthContext(testAnotherUserID, customerRole)
	created, _ := reviewClient.CreateReview(authorCtx, &pb.CreateReviewRequest{UserId: testUserID, ProductId: testProductID, Rating: 1, Comment: "Protected"})

	_, err := reviewClient.DeleteReview(nonAuthorCtx, &pb.DeleteReviewRequest{ReviewId: created.Id, UserId: testAnotherUserID})
	require.Error(t, err)
	st, _ := status.FromError(err)
	assert.Equal(t, codes.PermissionDenied, st.Code())
}

func TestListReviewsByUser_Success(t *testing.T) {
	clearReviewsCollection(t)
	authCtxUser1 := createAuthContext(testUserID, customerRole)
	authCtxUser2 := createAuthContext(testAnotherUserID, customerRole)

	_, _ = reviewClient.CreateReview(authCtxUser1, &pb.CreateReviewRequest{UserId: testUserID, ProductId: testProductID, Rating: 5, Comment: "User1 Review1"})
	_, _ = reviewClient.CreateReview(authCtxUser1, &pb.CreateReviewRequest{UserId: testUserID, ProductId: testAnotherProductID, Rating: 4, Comment: "User1 Review2"})
	_, _ = reviewClient.CreateReview(authCtxUser2, &pb.CreateReviewRequest{UserId: testAnotherUserID, ProductId: testProductID, Rating: 3, Comment: "User2 Review1"})

	listReq := &pb.ListReviewsByUserRequest{UserId: testUserID, Page: 1, Limit: 10}
	resp, err := reviewClient.ListReviewsByUser(authCtxUser1, listReq)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Len(t, resp.Reviews, 2)
	assert.Equal(t, int64(2), resp.Total)
	for _, r := range resp.Reviews {
		assert.Equal(t, testUserID, r.UserId)
	}
}

func TestListReviewsByUser_AttemptOtherUser_Forbidden(t *testing.T) {
	clearReviewsCollection(t)
	authCtxUser1 := createAuthContext(testUserID, customerRole)

	listReq := &pb.ListReviewsByUserRequest{UserId: testAnotherUserID, Page: 1, Limit: 10}
	_, err := reviewClient.ListReviewsByUser(authCtxUser1, listReq)
	require.Error(t, err)
	st, _ := status.FromError(err)
	assert.Equal(t, codes.PermissionDenied, st.Code())
}

func TestGetProductAverageRating_Success(t *testing.T) {
	clearReviewsCollection(t)
	ctx := context.Background()
	adminAuthCtx := createAuthContext(testAdminID, adminRole)

	r1, _ := reviewClient.CreateReview(createAuthContext("userA", customerRole), &pb.CreateReviewRequest{UserId: "userA", ProductId: testProductID, Rating: 5, Comment: "Excellent"})
	r2, _ := reviewClient.CreateReview(createAuthContext("userB", customerRole), &pb.CreateReviewRequest{UserId: "userB", ProductId: testProductID, Rating: 4, Comment: "Very Good"})
	// r3 is created but remains pending, so it's not used in the average calculation.
	_, _ = reviewClient.CreateReview(createAuthContext("userC", customerRole), &pb.CreateReviewRequest{UserId: "userC", ProductId: testProductID, Rating: 3, Comment: "Okay, but pending"})
	_, _ = reviewClient.CreateReview(createAuthContext("userD", customerRole), &pb.CreateReviewRequest{UserId: "userD", ProductId: testAnotherProductID, Rating: 5, Comment: "Different product"})

	_, err := reviewClient.ModerateReview(adminAuthCtx, &pb.ModerateReviewRequest{ReviewId: r1.Id, AdminId: testAdminID, NewStatus: string(domain.ReviewStatusApproved)})
	require.NoError(t, err)
	_, err = reviewClient.ModerateReview(adminAuthCtx, &pb.ModerateReviewRequest{ReviewId: r2.Id, AdminId: testAdminID, NewStatus: string(domain.ReviewStatusApproved)})
	require.NoError(t, err)

	avgReq := &pb.GetProductAverageRatingRequest{ProductId: testProductID}
	avgResp, err := reviewClient.GetProductAverageRating(ctx, avgReq)
	require.NoError(t, err)
	require.NotNil(t, avgResp)
	assert.Equal(t, testProductID, avgResp.ProductId)
	assert.InDelta(t, 4.5, avgResp.AverageRating, 0.01)
	assert.Equal(t, int32(2), avgResp.ReviewCount)
}

func TestGetProductAverageRating_NoApprovedReviews(t *testing.T) {
	clearReviewsCollection(t)
	ctx := context.Background()
	_, _ = reviewClient.CreateReview(createAuthContext("userA", customerRole), &pb.CreateReviewRequest{UserId: "userA", ProductId: testProductID, Rating: 5, Comment: "Pending"})
	r2, _ := reviewClient.CreateReview(createAuthContext("userB", customerRole), &pb.CreateReviewRequest{UserId: "userB", ProductId: testProductID, Rating: 4, Comment: "To be rejected"})

	_, err := reviewClient.ModerateReview(createAuthContext(testAdminID, adminRole), &pb.ModerateReviewRequest{ReviewId: r2.Id, AdminId: testAdminID, NewStatus: string(domain.ReviewStatusRejected)})
	require.NoError(t, err)

	avgReq := &pb.GetProductAverageRatingRequest{ProductId: testProductID}
	avgResp, err := reviewClient.GetProductAverageRating(ctx, avgReq)
	require.NoError(t, err)
	require.NotNil(t, avgResp)
	assert.Equal(t, 0.0, avgResp.AverageRating)
	assert.Equal(t, int32(0), avgResp.ReviewCount)
}

func TestModerateReview_AdminApprove_Success(t *testing.T) {
	clearReviewsCollection(t)
	adminAuthCtx := createAuthContext(testAdminID, adminRole)
	customerAuthCtx := createAuthContext(testUserID, customerRole)

	created, _ := reviewClient.CreateReview(customerAuthCtx, &pb.CreateReviewRequest{UserId: testUserID, ProductId: testProductID, Rating: 4, Comment: "Awaiting approval"})
	require.Equal(t, string(domain.ReviewStatusPending), created.Status)

	moderateReq := &pb.ModerateReviewRequest{
		ReviewId:          created.Id,
		AdminId:           testAdminID,
		NewStatus:         string(domain.ReviewStatusApproved),
		ModerationComment: "Looks good.",
	}
	moderatedReview, err := reviewClient.ModerateReview(adminAuthCtx, moderateReq)
	require.NoError(t, err)
	require.NotNil(t, moderatedReview)
	assert.Equal(t, string(domain.ReviewStatusApproved), moderatedReview.Status)
	assert.Equal(t, "Looks good.", moderatedReview.ModerationComment)

	fetched, _ := reviewClient.GetReview(context.Background(), &pb.GetReviewRequest{ReviewId: created.Id})
	assert.Equal(t, string(domain.ReviewStatusApproved), fetched.Status)
}

func TestModerateReview_NonAdmin_Forbidden(t *testing.T) {
	clearReviewsCollection(t)
	nonAdminAuthCtx := createAuthContext(testUserID, customerRole)
	customerAuthCtx := createAuthContext(testAnotherUserID, customerRole)

	created, _ := reviewClient.CreateReview(customerAuthCtx, &pb.CreateReviewRequest{UserId: testAnotherUserID, ProductId: testProductID, Rating: 3, Comment: "Some review"})

	moderateReq := &pb.ModerateReviewRequest{ReviewId: created.Id, AdminId: testUserID, NewStatus: string(domain.ReviewStatusApproved)}
	_, err := reviewClient.ModerateReview(nonAdminAuthCtx, moderateReq)
	require.Error(t, err)
	st, _ := status.FromError(err)
	assert.Equal(t, codes.PermissionDenied, st.Code())
}

func TestGetReview_NotFound(t *testing.T) {
	clearReviewsCollection(t)
	nonExistentID := primitive.NewObjectID().Hex()
	_, err := reviewClient.GetReview(context.Background(), &pb.GetReviewRequest{ReviewId: nonExistentID})
	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.NotFound, st.Code())
}

func TestListReviewsByProduct_Pagination(t *testing.T) {
	clearReviewsCollection(t)
	adminCtx := createAuthContext(testAdminID, adminRole)

	for i := 0; i < 5; i++ {
		userID := fmt.Sprintf("userTest%d", i)
		reviewCtx := createAuthContext(userID, customerRole)
		created, err := reviewClient.CreateReview(reviewCtx, &pb.CreateReviewRequest{
			UserId:    userID,
			ProductId: testProductID,
			Rating:    int32(i%5 + 1),
			Comment:   fmt.Sprintf("Review %d", i+1),
		})
		require.NoError(t, err)
		_, err = reviewClient.ModerateReview(adminCtx, &pb.ModerateReviewRequest{
			ReviewId:  created.Id,
			AdminId:   testAdminID,
			NewStatus: string(domain.ReviewStatusApproved),
		})
		require.NoError(t, err)
	}

	listReq1 := &pb.ListReviewsByProductRequest{ProductId: testProductID, Page: 1, Limit: 2, StatusFilter: string(domain.ReviewStatusApproved)}
	resp1, err := reviewClient.ListReviewsByProduct(context.Background(), listReq1)
	require.NoError(t, err)
	assert.Len(t, resp1.Reviews, 2)
	assert.Equal(t, int64(5), resp1.Total)
	assert.Equal(t, int32(1), resp1.Page)
	assert.Equal(t, int32(2), resp1.Limit)

	listReq2 := &pb.ListReviewsByProductRequest{ProductId: testProductID, Page: 2, Limit: 2, StatusFilter: string(domain.ReviewStatusApproved)}
	resp2, err := reviewClient.ListReviewsByProduct(context.Background(), listReq2)
	require.NoError(t, err)
	assert.Len(t, resp2.Reviews, 2)
	assert.Equal(t, int64(5), resp2.Total)

	listReq3 := &pb.ListReviewsByProductRequest{ProductId: testProductID, Page: 3, Limit: 2, StatusFilter: string(domain.ReviewStatusApproved)}
	resp3, err := reviewClient.ListReviewsByProduct(context.Background(), listReq3)
	require.NoError(t, err)
	assert.Len(t, resp3.Reviews, 1)
	assert.Equal(t, int64(5), resp3.Total)

	listReq4 := &pb.ListReviewsByProductRequest{ProductId: testProductID, Page: 4, Limit: 2, StatusFilter: string(domain.ReviewStatusApproved)}
	resp4, err := reviewClient.ListReviewsByProduct(context.Background(), listReq4)
	require.NoError(t, err)
	assert.Len(t, resp4.Reviews, 0)
	assert.Equal(t, int64(5), resp4.Total)
}
