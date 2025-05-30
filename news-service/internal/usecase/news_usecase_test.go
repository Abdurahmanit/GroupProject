package usecase

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/Abdurahmanit/GroupProject/news-service/internal/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

type MockNewsRepository struct{ mock.Mock }

func (m *MockNewsRepository) Create(ctx context.Context, news *entity.News) (string, error) {
	args := m.Called(ctx, news)
	return args.String(0), args.Error(1)
}
func (m *MockNewsRepository) GetByID(ctx context.Context, id string) (*entity.News, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.News), args.Error(1)
}
func (m *MockNewsRepository) Update(ctx context.Context, news *entity.News) error {
	args := m.Called(ctx, news)
	return args.Error(0)
}
func (m *MockNewsRepository) Delete(ctx context.Context, id string, sc mongo.SessionContext) error {
	args := m.Called(ctx, id, sc)
	return args.Error(0)
}
func (m *MockNewsRepository) List(ctx context.Context, page, pageSize int, filter map[string]interface{}) ([]*entity.News, int, error) {
	args := m.Called(ctx, page, pageSize, filter)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*entity.News), args.Int(1), args.Error(2)
}

type MockCommentRepository struct{ mock.Mock }

func (m *MockCommentRepository) Create(ctx context.Context, comment *entity.Comment) (string, error) {
	args := m.Called(ctx, comment)
	return args.String(0), args.Error(1)
}
func (m *MockCommentRepository) GetByID(ctx context.Context, id string) (*entity.Comment, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Comment), args.Error(1)
}
func (m *MockCommentRepository) GetByNewsID(ctx context.Context, newsID string, page, pageSize int) ([]*entity.Comment, int, error) {
	args := m.Called(ctx, newsID, page, pageSize)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*entity.Comment), args.Int(1), args.Error(2)
}
func (m *MockCommentRepository) Update(ctx context.Context, comment *entity.Comment) error {
	args := m.Called(ctx, comment)
	return args.Error(0)
}
func (m *MockCommentRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
func (m *MockCommentRepository) DeleteByNewsID(ctx context.Context, newsID string, sessionContext mongo.SessionContext) (int64, error) {
	args := m.Called(ctx, newsID, sessionContext)
	return args.Get(0).(int64), args.Error(1)
}

type MockLikeRepository struct{ mock.Mock }

func (m *MockLikeRepository) AddLike(ctx context.Context, contentType string, contentID string, userID string) error {
	args := m.Called(ctx, contentType, contentID, userID)
	return args.Error(0)
}
func (m *MockLikeRepository) RemoveLike(ctx context.Context, contentType string, contentID string, userID string) error {
	args := m.Called(ctx, contentType, contentID, userID)
	return args.Error(0)
}
func (m *MockLikeRepository) GetLikesCount(ctx context.Context, contentType string, contentID string) (int64, error) {
	args := m.Called(ctx, contentType, contentID)
	return args.Get(0).(int64), args.Error(1)
}
func (m *MockLikeRepository) HasLiked(ctx context.Context, contentType string, contentID string, userID string) (bool, error) {
	args := m.Called(ctx, contentType, contentID, userID)
	return args.Bool(0), args.Error(1)
}
func (m *MockLikeRepository) DeleteByContentID(ctx context.Context, contentType string, contentID string, sessionContext mongo.SessionContext) (int64, error) {
	args := m.Called(ctx, contentType, contentID, sessionContext)
	return args.Get(0).(int64), args.Error(1)
}

type MockNATSPublisher struct{ mock.Mock }

func (m *MockNATSPublisher) PublishNewsCreated(ctx context.Context, news *entity.News) error {
	args := m.Called(ctx, news)
	return args.Error(0)
}
func (m *MockNATSPublisher) PublishNewsUpdated(ctx context.Context, news *entity.News) error {
	args := m.Called(ctx, news)
	return args.Error(0)
}
func (m *MockNATSPublisher) PublishNewsDeleted(ctx context.Context, newsID string) error {
	args := m.Called(ctx, newsID)
	return args.Error(0)
}

type MockCacheRepository struct{ mock.Mock }

func (m *MockCacheRepository) Get(ctx context.Context, key string) ([]byte, error) {
	args := m.Called(ctx, key)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}
func (m *MockCacheRepository) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	args := m.Called(ctx, key, value, ttl)
	return args.Error(0)
}
func (m *MockCacheRepository) Delete(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

type MockEmailSender struct{ mock.Mock }

func (m *MockEmailSender) SendEmail(to []string, subject, body string) error {
	args := m.Called(to, subject, body)
	return args.Error(0)
}

type MockUserServiceClient struct{ mock.Mock }

func (m *MockUserServiceClient) GetAuthorEmail(ctx context.Context, authorID string) (string, error) {
	args := m.Called(ctx, authorID)
	return args.String(0), args.Error(1)
}
func (m *MockUserServiceClient) Close() error {
	args := m.Called()
	return args.Error(0)
}

func TestNewsUseCase_CreateNews_EmailFlow(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	mockNewsRepo := new(MockNewsRepository)
	mockCommentRepo := new(MockCommentRepository)
	mockLikeRepo := new(MockLikeRepository)
	mockNatsPub := new(MockNATSPublisher)
	mockCache := new(MockCacheRepository)
	mockEmail := new(MockEmailSender)
	mockUserSvc := new(MockUserServiceClient)

	uc := NewNewsUseCase(
		nil,
		mockNewsRepo,
		mockCommentRepo,
		mockLikeRepo,
		mockNatsPub,
		mockCache,
		mockEmail,
		mockUserSvc,
		logger,
	)

	ctx := context.Background()
	input := CreateNewsInput{
		Title:    "Email Test News",
		Content:  "Content for email test",
		AuthorID: "author123",
		ImageURL: "url",
		Category: "cat",
	}
	mockNewsID := "mockNewsID"
	authorEmail := "author@example.com"

	t.Run("EmailSentSuccessfully", func(t *testing.T) {
		mockNewsRepo.On("Create", ctx, mock.AnythingOfType("*entity.News")).Return(mockNewsID, nil).Once()
		mockCache.On("Set", ctx, newsCacheKey(mockNewsID), mock.Anything, newsCacheTTL).Return(nil).Maybe().Once()
		mockNatsPub.On("PublishNewsCreated", ctx, mock.AnythingOfType("*entity.News")).Return(nil).Once()
		mockUserSvc.On("GetAuthorEmail", ctx, input.AuthorID).Return(authorEmail, nil).Once()
		expectedSubject := fmt.Sprintf("Ваша новость опубликована: %s", input.Title)
		expectedBody := fmt.Sprintf("Поздравляем!\n\nВаша новость '%s' была успешно опубликована на нашем портале.\n\nID новости: %s", input.Title, mockNewsID)
		mockEmail.On("SendEmail", []string{authorEmail}, expectedSubject, expectedBody).Return(nil).Once()

		createdNews, err := uc.CreateNews(ctx, input)

		assert.NoError(t, err)
		assert.NotNil(t, createdNews)
		assert.Equal(t, mockNewsID, createdNews.ID)

		mockNewsRepo.AssertExpectations(t)
		mockCache.AssertExpectations(t)
		mockNatsPub.AssertExpectations(t)
		mockUserSvc.AssertExpectations(t)
		mockEmail.AssertExpectations(t)

		mockNewsRepo.Mock = mock.Mock{}
		mockCache.Mock = mock.Mock{}
		mockNatsPub.Mock = mock.Mock{}
		mockUserSvc.Mock = mock.Mock{}
		mockEmail.Mock = mock.Mock{}
	})

	t.Run("UserServiceReturnsError_EmailNotSent", func(t *testing.T) {
		mockNewsRepo.On("Create", ctx, mock.AnythingOfType("*entity.News")).Return(mockNewsID, nil).Once()
		mockCache.On("Set", ctx, newsCacheKey(mockNewsID), mock.Anything, newsCacheTTL).Return(nil).Maybe().Once()
		mockNatsPub.On("PublishNewsCreated", ctx, mock.AnythingOfType("*entity.News")).Return(nil).Once()
		mockUserSvc.On("GetAuthorEmail", ctx, input.AuthorID).Return("", errors.New("user service error")).Once()

		createdNews, err := uc.CreateNews(ctx, input)

		assert.NoError(t, err)
		assert.NotNil(t, createdNews)

		mockNewsRepo.AssertExpectations(t)
		mockCache.AssertExpectations(t)
		mockNatsPub.AssertExpectations(t)
		mockUserSvc.AssertExpectations(t)
		mockEmail.AssertNotCalled(t, "SendEmail", mock.Anything, mock.Anything, mock.Anything)

		mockNewsRepo.Mock = mock.Mock{}
		mockCache.Mock = mock.Mock{}
		mockNatsPub.Mock = mock.Mock{}
		mockUserSvc.Mock = mock.Mock{}
		mockEmail.Mock = mock.Mock{}
	})

	t.Run("UserServiceReturnsEmptyEmail_EmailNotSent", func(t *testing.T) {
		mockNewsRepo.On("Create", ctx, mock.AnythingOfType("*entity.News")).Return(mockNewsID, nil).Once()
		mockCache.On("Set", ctx, newsCacheKey(mockNewsID), mock.Anything, newsCacheTTL).Return(nil).Maybe().Once()
		mockNatsPub.On("PublishNewsCreated", ctx, mock.AnythingOfType("*entity.News")).Return(nil).Once()
		mockUserSvc.On("GetAuthorEmail", ctx, input.AuthorID).Return("", nil).Once()

		createdNews, err := uc.CreateNews(ctx, input)

		assert.NoError(t, err)
		assert.NotNil(t, createdNews)

		mockNewsRepo.AssertExpectations(t)
		mockCache.AssertExpectations(t)
		mockNatsPub.AssertExpectations(t)
		mockUserSvc.AssertExpectations(t)
		mockEmail.AssertNotCalled(t, "SendEmail", mock.Anything, mock.Anything, mock.Anything)

		mockNewsRepo.Mock = mock.Mock{}
		mockCache.Mock = mock.Mock{}
		mockNatsPub.Mock = mock.Mock{}
		mockUserSvc.Mock = mock.Mock{}
		mockEmail.Mock = mock.Mock{}
	})
}
