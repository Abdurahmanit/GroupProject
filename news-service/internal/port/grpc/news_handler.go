package grpc

import (
	"context"
	"errors"

	"github.com/Abdurahmanit/GroupProject/news-service/internal/entity"
	"github.com/Abdurahmanit/GroupProject/news-service/internal/port/repository"
	"github.com/Abdurahmanit/GroupProject/news-service/internal/usecase"
	newspb "github.com/Abdurahmanit/GroupProject/news-service/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type NewsHandler struct {
	newspb.UnimplementedNewsServiceServer
	newsUseCase    *usecase.NewsUseCase
	commentUseCase *usecase.CommentUseCase
	likeUseCase    *usecase.LikeUseCase
}

func NewNewsHandler(newsUC *usecase.NewsUseCase, commentUC *usecase.CommentUseCase, likeUC *usecase.LikeUseCase) *NewsHandler {
	return &NewsHandler{
		newsUseCase:    newsUC,
		commentUseCase: commentUC,
		likeUseCase:    likeUC,
	}
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
		ImageUrl:  n.ImageURL,
		CreatedAt: timestamppb.New(n.CreatedAt),
		UpdatedAt: timestamppb.New(n.UpdatedAt),
	}
}

func commentEntityToProto(c *entity.Comment) *newspb.Comment {
	if c == nil {
		return nil
	}
	return &newspb.Comment{
		Id:        c.ID,
		NewsId:    c.NewsID,
		UserId:    c.UserID,
		Content:   c.Content,
		CreatedAt: timestamppb.New(c.CreatedAt),
		UpdatedAt: timestamppb.New(c.UpdatedAt),
	}
}

func (h *NewsHandler) CreateNews(ctx context.Context, req *newspb.CreateNewsRequest) (*newspb.CreateNewsResponse, error) {
	input := usecase.CreateNewsInput{
		Title:    req.GetTitle(),
		Content:  req.GetContent(),
		AuthorID: req.GetAuthorId(),
		ImageURL: req.GetImageUrl(),
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
		if errors.Is(err, repository.ErrNotFound) {
			return nil, status.Errorf(codes.NotFound, "news with id %s not found", req.GetId())
		}
		return nil, status.Errorf(codes.Internal, "failed to get news: %v", err)
	}
	return &newspb.GetNewsResponse{News: newsEntityToProto(newsEntity)}, nil
}

func (h *NewsHandler) ListNews(ctx context.Context, req *newspb.ListNewsRequest) (*newspb.ListNewsResponse, error) {
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
	input := usecase.UpdateNewsInput{ID: req.GetId()}
	if req.Title != nil {
		input.Title = req.Title
	}
	if req.Content != nil {
		input.Content = req.Content
	}
	if req.ImageUrl != nil {
		input.ImageURL = req.ImageUrl
	}
	updatedNews, err := h.newsUseCase.UpdateNews(ctx, input)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, status.Errorf(codes.NotFound, "news with id %s not found for update", req.GetId())
		}
		return nil, status.Errorf(codes.Internal, "failed to update news: %v", err)
	}
	return &newspb.UpdateNewsResponse{News: newsEntityToProto(updatedNews)}, nil
}

func (h *NewsHandler) DeleteNews(ctx context.Context, req *newspb.DeleteNewsRequest) (*newspb.DeleteNewsResponse, error) {
	err := h.newsUseCase.DeleteNews(ctx, req.GetId())
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, status.Errorf(codes.NotFound, "news with id %s not found for delete", req.GetId())
		}
		return nil, status.Errorf(codes.Internal, "failed to delete news: %v", err)
	}
	return &newspb.DeleteNewsResponse{Success: true}, nil
}

func (h *NewsHandler) CreateComment(ctx context.Context, req *newspb.CreateCommentRequest) (*newspb.CreateCommentResponse, error) {
	input := usecase.CreateCommentInput{
		NewsID:  req.GetNewsId(),
		UserID:  req.GetUserId(),
		Content: req.GetContent(),
	}
	createdComment, err := h.commentUseCase.CreateComment(ctx, input)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create comment: %v", err)
	}
	return &newspb.CreateCommentResponse{Id: createdComment.ID}, nil
}

func (h *NewsHandler) GetCommentsForNews(ctx context.Context, req *newspb.GetCommentsForNewsRequest) (*newspb.GetCommentsForNewsResponse, error) {
	input := usecase.ListCommentsInput{
		NewsID:   req.GetNewsId(),
		Page:     int(req.GetPage()),
		PageSize: int(req.GetPageSize()),
	}
	output, err := h.commentUseCase.GetCommentsByNewsID(ctx, input)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get comments: %v", err)
	}
	pbCommentList := make([]*newspb.Comment, len(output.Comments))
	for i, c := range output.Comments {
		pbCommentList[i] = commentEntityToProto(c)
	}
	return &newspb.GetCommentsForNewsResponse{Comments: pbCommentList, TotalCount: int32(output.TotalCount)}, nil
}

func (h *NewsHandler) DeleteComment(ctx context.Context, req *newspb.DeleteCommentRequest) (*newspb.DeleteCommentResponse, error) {
	input := usecase.DeleteCommentInput{
		CommentID: req.GetCommentId(),
		UserID:    req.GetUserId(),
	}
	err := h.commentUseCase.DeleteComment(ctx, input)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, status.Errorf(codes.NotFound, "comment with id %s not found for delete", req.GetCommentId())
		}
		return nil, status.Errorf(codes.Internal, "failed to delete comment: %v", err)
	}
	return &newspb.DeleteCommentResponse{Success: true}, nil
}

func (h *NewsHandler) LikeNews(ctx context.Context, req *newspb.LikeNewsRequest) (*newspb.LikeNewsResponse, error) {
	input := usecase.AddLikeInput{
		ContentType: usecase.ContentTypeNews,
		ContentID:   req.GetNewsId(),
		UserID:      req.GetUserId(),
	}
	err := h.likeUseCase.AddLike(ctx, input)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to like news: %v", err)
	}
	countInput := usecase.GetLikesCountInput{ContentType: usecase.ContentTypeNews, ContentID: req.GetNewsId()}
	count, err := h.likeUseCase.GetLikesCount(ctx, countInput)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get like count after liking: %v", err)
	}
	return &newspb.LikeNewsResponse{Success: true, LikeCount: count}, nil
}

func (h *NewsHandler) UnlikeNews(ctx context.Context, req *newspb.UnlikeNewsRequest) (*newspb.UnlikeNewsResponse, error) {
	input := usecase.RemoveLikeInput{
		ContentType: usecase.ContentTypeNews,
		ContentID:   req.GetNewsId(),
		UserID:      req.GetUserId(),
	}
	err := h.likeUseCase.RemoveLike(ctx, input)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
		} else {
			return nil, status.Errorf(codes.Internal, "failed to unlike news: %v", err)
		}
	}
	countInput := usecase.GetLikesCountInput{ContentType: usecase.ContentTypeNews, ContentID: req.GetNewsId()}
	count, err := h.likeUseCase.GetLikesCount(ctx, countInput)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get like count after unliking: %v", err)
	}
	return &newspb.UnlikeNewsResponse{Success: true, LikeCount: count}, nil
}

func (h *NewsHandler) GetLikesCount(ctx context.Context, req *newspb.GetLikesCountRequest) (*newspb.GetLikesCountResponse, error) {
	input := usecase.GetLikesCountInput{
		ContentType: usecase.ContentTypeNews,
		ContentID:   req.GetNewsId(),
	}
	count, err := h.likeUseCase.GetLikesCount(ctx, input)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, status.Errorf(codes.NotFound, "content not found: %s", req.GetNewsId())
		}
		return nil, status.Errorf(codes.Internal, "failed to get likes count: %v", err)
	}
	return &newspb.GetLikesCountResponse{LikeCount: count}, nil
}

func (h *NewsHandler) GetNewsByAuthor(ctx context.Context, req *newspb.GetNewsByAuthorRequest) (*newspb.ListNewsResponse, error) {
	input := usecase.ListNewsInput{
		Page:     int(req.GetPage()),
		PageSize: int(req.GetPageSize()),
		Filter:   map[string]interface{}{"author_id": req.GetAuthorId()},
	}
	output, err := h.newsUseCase.ListNews(ctx, input)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list news by author: %v", err)
	}
	pbNewsList := make([]*newspb.News, len(output.News))
	for i, n := range output.News {
		pbNewsList[i] = newsEntityToProto(n)
	}
	return &newspb.ListNewsResponse{News: pbNewsList, TotalCount: int32(output.TotalCount)}, nil
}
