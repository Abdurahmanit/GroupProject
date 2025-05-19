package grpc

import (
	"context"

	"github.com/Abdurahmanit/GroupProject/news-service/internal/entity"
	"github.com/Abdurahmanit/GroupProject/news-service/internal/usecase"
	newspb "github.com/Abdurahmanit/GroupProject/news-service/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type NewsHandler struct {
	newspb.UnimplementedNewsServiceServer
	newsUseCase *usecase.NewsUseCase
}

func NewNewsHandler(uc *usecase.NewsUseCase) *NewsHandler {
	return &NewsHandler{newsUseCase: uc}
}

func newsEntityToProto(n *entity.News) *newspb.News {
	if n == nil {
		return nil
	}
	return &newspb.News{
		Id:        n.ID,
		Title:     n.Title,
		Content:   n.Content,
		AuthorId:  n.AuthorID,
		CreatedAt: timestamppb.New(n.CreatedAt),
		UpdatedAt: timestamppb.New(n.UpdatedAt),
	}
}

func (h *NewsHandler) CreateNews(ctx context.Context, req *newspb.CreateNewsRequest) (*newspb.CreateNewsResponse, error) {
	input := usecase.CreateNewsInput{
		Title:    req.GetTitle(),
		Content:  req.GetContent(),
		AuthorID: req.GetAuthorId(),
	}

	createdNews, err := h.newsUseCase.CreateNews(ctx, input)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create news: %v", err)
	}

	return &newspb.CreateNewsResponse{Id: createdNews.ID}, nil
}

func (h *NewsHandler) GetNews(ctx context.Context, req *newspb.GetNewsRequest) (*newspb.GetNewsResponse, error) {
	newsEntity, err := h.newsUseCase.GetNewsByID(ctx, req.GetId())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get news: %v", err)
	}
	if newsEntity == nil {
		return nil, status.Errorf(codes.NotFound, "news with id %s not found", req.GetId())
	}
	return &newspb.GetNewsResponse{News: newsEntityToProto(newsEntity)}, nil
}

func (h *NewsHandler) ListNews(ctx context.Context, req *newspb.ListNewsRequest) (*newspb.ListNewsResponse, error) {
	// TODO: Добавить передачу фильтров из req в use case, если они есть в proto
	input := usecase.ListNewsInput{
		Page:     int(req.GetPage()),
		PageSize: int(req.GetPageSize()),
		Filter:   nil,
	}
	output, err := h.newsUseCase.ListNews(ctx, input)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list news: %v", err)
	}

	pbNewsList := make([]*newspb.News, len(output.News))
	for i, n := range output.News {
		pbNewsList[i] = newsEntityToProto(n)
	}

	return &newspb.ListNewsResponse{News: pbNewsList, TotalCount: int32(output.TotalCount)}, nil
}

func (h *NewsHandler) UpdateNews(ctx context.Context, req *newspb.UpdateNewsRequest) (*newspb.UpdateNewsResponse, error) {
	input := usecase.UpdateNewsInput{
		ID: req.GetId(),
	}
	if req.Title != nil {
		input.Title = req.Title
	}
	if req.Content != nil {
		input.Content = req.Content
	}

	updatedNews, err := h.newsUseCase.UpdateNews(ctx, input)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update news: %v", err)
	}
	return &newspb.UpdateNewsResponse{News: newsEntityToProto(updatedNews)}, nil
}

func (h *NewsHandler) DeleteNews(ctx context.Context, req *newspb.DeleteNewsRequest) (*newspb.DeleteNewsResponse, error) {
	err := h.newsUseCase.DeleteNews(ctx, req.GetId())
	if err != nil {
		// TODO: Обработать ErrNotFound
		return nil, status.Errorf(codes.Internal, "failed to delete news: %v", err)
	}
	return &newspb.DeleteNewsResponse{Success: true}, nil
}
