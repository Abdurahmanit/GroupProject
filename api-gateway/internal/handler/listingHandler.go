package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"io"
	"github.com/Abdurahmanit/GroupProject/listing-service/genproto/listing_service"
	"github.com/go-chi/chi/v5" // Возвращаем импорт chi
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// ListingHandler обрабатывает запросы к Listing Service
type ListingHandler struct {
	client *grpc.ClientConn
	logger *zap.Logger
}

// NewListingHandler создает новый обработчик для Listing Service
func NewListingHandler(client *grpc.ClientConn, logger *zap.Logger) *ListingHandler {
	return &ListingHandler{client: client, logger: logger}
}

// HandleCreateListing обрабатывает создание нового объявления
func (h *ListingHandler) HandleCreateListing(w http.ResponseWriter, r *http.Request) { // Сигнатура для chi
	var req listing_service.CreateListingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Invalid request body for CreateListing", zap.Error(err))
		http.Error(w, status.Errorf(codes.InvalidArgument, "Invalid request body: %v", err).Error(), http.StatusBadRequest)
		return
	}

	// Получаем userID из контекста, если он был добавлен middleware (например, JWTAuth)
	// userID, ok := r.Context().Value("user_id").(string)
	// if !ok {
	// 	// Обработка случая, если user_id не найден (если этот эндпоинт требует авторизации)
	// }
	// req.CreatorId = userID // Если нужно установить создателя

	ctx := withAuth(r.Context(), r)
	client := listing_service.NewListingServiceClient(h.client)
	resp, err := client.CreateListing(ctx, &req)
	if err != nil {
		h.logger.Error("Failed to create listing via gRPC", zap.Error(err))
		st, ok := status.FromError(err)
		if ok {
			http.Error(w, st.Message(), http.StatusInternalServerError)
		} else {
			http.Error(w, "Failed to create listing: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Error("Failed to encode CreateListing response", zap.Error(err))
		http.Error(w, status.Errorf(codes.Internal, "Failed to encode response: %v", err).Error(), http.StatusInternalServerError)
	}
}

// HandleUpdateListing обрабатывает обновление объявления
func (h *ListingHandler) HandleUpdateListing(w http.ResponseWriter, r *http.Request) { // Сигнатура для chi
	id := chi.URLParam(r, "id") // Используем chi.URLParam
	var req listing_service.UpdateListingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Invalid request body for UpdateListing", zap.String("id", id), zap.Error(err))
		http.Error(w, status.Errorf(codes.InvalidArgument, "Invalid request body: %v", err).Error(), http.StatusBadRequest)
		return
	}
	req.Id = id

	ctx := withAuth(r.Context(), r)
	client := listing_service.NewListingServiceClient(h.client)
	resp, err := client.UpdateListing(ctx, &req)
	if err != nil {
		h.logger.Error("Failed to update listing via gRPC", zap.String("id", id), zap.Error(err))
		st, ok := status.FromError(err)
		if ok {
			http.Error(w, st.Message(), http.StatusInternalServerError)
		} else {
			http.Error(w, "Failed to update listing: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Error("Failed to encode UpdateListing response", zap.String("id", id), zap.Error(err))
		http.Error(w, status.Errorf(codes.Internal, "Failed to encode response: %v", err).Error(), http.StatusInternalServerError)
	}
}

// HandleDeleteListing обрабатывает удаление объявления
func (h *ListingHandler) HandleDeleteListing(w http.ResponseWriter, r *http.Request) { // Сигнатура для chi
	id := chi.URLParam(r, "id") // Используем chi.URLParam
	if id == "" {
		h.logger.Error("Missing id parameter for DeleteListing")
		http.Error(w, status.Errorf(codes.InvalidArgument, "Missing id parameter").Error(), http.StatusBadRequest)
		return
	}

	ctx := withAuth(r.Context(), r)
	client := listing_service.NewListingServiceClient(h.client)
	_, err := client.DeleteListing(ctx, &listing_service.DeleteListingRequest{Id: id})
	if err != nil {
		h.logger.Error("Failed to delete listing via gRPC", zap.String("id", id), zap.Error(err))
		st, ok := status.FromError(err)
		if ok {
			http.Error(w, st.Message(), http.StatusInternalServerError)
		} else {
			http.Error(w, "Failed to delete listing: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// HandleGetListingByID обрабатывает получение объявления по ID
func (h *ListingHandler) HandleGetListingByID(w http.ResponseWriter, r *http.Request) { // Сигнатура для chi
	id := chi.URLParam(r, "id") // Используем chi.URLParam
	if id == "" {
		h.logger.Error("Missing id parameter for GetListingByID")
		http.Error(w, status.Errorf(codes.InvalidArgument, "Missing id parameter").Error(), http.StatusBadRequest)
		return
	}

	ctx := withAuth(r.Context(), r)
	client := listing_service.NewListingServiceClient(h.client)
	resp, err := client.GetListingByID(ctx, &listing_service.GetListingRequest{Id: id})
	if err != nil {
		h.logger.Error("Failed to get listing by ID via gRPC", zap.String("id", id), zap.Error(err))
		st, ok := status.FromError(err)
		if ok {
			http.Error(w, st.Message(), http.StatusInternalServerError)
		} else {
			http.Error(w, "Failed to get listing: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Error("Failed to encode GetListingByID response", zap.String("id", id), zap.Error(err))
		http.Error(w, status.Errorf(codes.Internal, "Failed to encode response: %v", err).Error(), http.StatusInternalServerError)
	}
}

// HandleSearchListings обрабатывает поиск объявлений
func (h *ListingHandler) HandleSearchListings(w http.ResponseWriter, r *http.Request) { // Сигнатура для chi
	var req listing_service.SearchListingsRequest
	// Для GET запросов с параметрами поиска, лучше парсить r.URL.Query()
	// Если это POST, то json.NewDecoder остается.
	// Пример для GET:
	// query := r.URL.Query().Get("query")
	// req.Query = query
	// ... и т.д. для других параметров
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil { // Оставляю для POST варианта
		h.logger.Error("Invalid request body for SearchListings", zap.Error(err))
		http.Error(w, status.Errorf(codes.InvalidArgument, "Invalid request body: %v", err).Error(), http.StatusBadRequest)
		return
	}

	ctx := withAuth(r.Context(), r)
	client := listing_service.NewListingServiceClient(h.client)
	resp, err := client.SearchListings(ctx, &req)
	if err != nil {
		h.logger.Error("Failed to search listings via gRPC", zap.Error(err))
		st, ok := status.FromError(err)
		if ok {
			http.Error(w, st.Message(), http.StatusInternalServerError)
		} else {
			http.Error(w, "Failed to search listings: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Error("Failed to encode SearchListings response", zap.Error(err))
		http.Error(w, status.Errorf(codes.Internal, "Failed to encode response: %v", err).Error(), http.StatusInternalServerError)
	}
}

// HandleUploadPhoto обрабатывает загрузку фотографии
func (h *ListingHandler) HandleUploadPhoto(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	// Ограничение размера файла (например, 10 МБ)
	r.ParseMultipartForm(10 << 20) // 10MB

	file, handler, err := r.FormFile("photo_file") // ключ — "photo_file"
	if err != nil {
		http.Error(w, "Failed to get uploaded file: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Прочитаем содержимое файла
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Failed to read file: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Соберем gRPC-запрос
	req := &listing_service.UploadPhotoRequest{
		ListingId: id,
		FileName:  handler.Filename,
		Data:  fileBytes,
	}

	ctx := withAuth(r.Context(), r)
	client := listing_service.NewListingServiceClient(h.client)
	resp, err := client.UploadPhoto(ctx, req)
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			http.Error(w, st.Message(), http.StatusInternalServerError)
		} else {
			http.Error(w, "Failed to upload photo: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}


// HandleGetListingStatus обрабатывает получение статуса объявления
func (h *ListingHandler) HandleGetListingStatus(w http.ResponseWriter, r *http.Request) { // Сигнатура для chi
	id := chi.URLParam(r, "id") // Используем chi.URLParam
	if id == "" {
		h.logger.Error("Missing id parameter for GetListingStatus")
		http.Error(w, status.Errorf(codes.InvalidArgument, "Missing id parameter").Error(), http.StatusBadRequest)
		return
	}

	ctx := withAuth(r.Context(), r)
	client := listing_service.NewListingServiceClient(h.client)
	resp, err := client.GetListingStatus(ctx, &listing_service.GetListingRequest{Id: id})
	if err != nil {
		h.logger.Error("Failed to get listing status via gRPC", zap.String("id", id), zap.Error(err))
		st, ok := status.FromError(err)
		if ok {
			http.Error(w, st.Message(), http.StatusInternalServerError)
		} else {
			http.Error(w, "Failed to get listing status: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Error("Failed to encode GetListingStatus response", zap.String("id", id), zap.Error(err))
		http.Error(w, status.Errorf(codes.Internal, "Failed to encode response: %v", err).Error(), http.StatusInternalServerError)
	}
}

// HandleAddFavorite обрабатывает добавление в избранное
func (h *ListingHandler) HandleAddFavorite(w http.ResponseWriter, r *http.Request) { // Сигнатура для chi
	userID, ok := r.Context().Value("user_id").(string)
	if !ok || userID == "" {
		h.logger.Error("User ID not found in context for AddFavorite")
		http.Error(w, "Unauthorized: User ID missing", http.StatusUnauthorized)
		return
	}

	var req listing_service.AddFavoriteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Invalid request body for AddFavorite", zap.String("user_id", userID), zap.Error(err))
		http.Error(w, status.Errorf(codes.InvalidArgument, "Invalid request body: %v", err).Error(), http.StatusBadRequest)
		return
	}
	req.UserId = userID // Устанавливаем userID из контекста

	ctx := withAuth(r.Context(), r)
	client := listing_service.NewListingServiceClient(h.client)
	_, err := client.AddFavorite(ctx, &req)
	if err != nil {
		h.logger.Error("Failed to add favorite via gRPC", zap.String("user_id", userID), zap.Error(err))
		st, sOk := status.FromError(err)
		if sOk {
			http.Error(w, st.Message(), http.StatusInternalServerError)
		} else {
			http.Error(w, "Failed to add favorite: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// HandleRemoveFavorite обрабатывает удаление из избранного
func (h *ListingHandler) HandleRemoveFavorite(w http.ResponseWriter, r *http.Request) { // Сигнатура для chi
	userID, ok := r.Context().Value("user_id").(string)
	if !ok || userID == "" {
		h.logger.Error("User ID not found in context for RemoveFavorite")
		http.Error(w, "Unauthorized: User ID missing", http.StatusUnauthorized)
		return
	}

	var req listing_service.RemoveFavoriteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Invalid request body for RemoveFavorite", zap.String("user_id", userID), zap.Error(err))
		http.Error(w, status.Errorf(codes.InvalidArgument, "Invalid request body: %v", err).Error(), http.StatusBadRequest)
		return
	}
	req.UserId = userID // Устанавливаем userID из контекста

	ctx := withAuth(r.Context(), r)
	client := listing_service.NewListingServiceClient(h.client)
	_, err := client.RemoveFavorite(ctx, &req)
	if err != nil {
		h.logger.Error("Failed to remove favorite via gRPC", zap.String("user_id", userID), zap.Error(err))
		st, sOk := status.FromError(err)
		if sOk {
			http.Error(w, st.Message(), http.StatusInternalServerError)
		} else {
			http.Error(w, "Failed to remove favorite: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// HandleGetFavorites обрабатывает получение списка избранного
func (h *ListingHandler) HandleGetFavorites(w http.ResponseWriter, r *http.Request) { // Сигнатура для chi
	userID, ok := r.Context().Value("user_id").(string)
	if !ok || userID == "" {
		h.logger.Error("User ID not found in context for GetFavorites")
		http.Error(w, "Unauthorized: User ID missing", http.StatusUnauthorized)
		return
	}

	// Теперь GetFavoritesRequest должен содержать UserID.
	// Если он передавался через query params, то так:
	// req := listing_service.GetFavoritesRequest{UserId: userID}
	// Если он был в теле JSON, но теперь мы берем из контекста:
	var req listing_service.GetFavoritesRequest // Оставляем, если есть другие поля в GetFavoritesRequest
	// Убираем декодирование, если UserID - единственное поле или если другие поля из query.
	// if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
	// 	h.logger.Error("Invalid request body for GetFavorites", zap.String("user_id", userID), zap.Error(err))
	// 	http.Error(w, status.Errorf(codes.InvalidArgument, "Invalid request body: %v", err).Error(), http.StatusBadRequest)
	// 	return
	// }
	req.UserId = userID // Устанавливаем userID из контекста

	ctx := withAuth(r.Context(), r)
	client := listing_service.NewListingServiceClient(h.client)
	resp, err := client.GetFavorites(ctx, &req)
	if err != nil {
		h.logger.Error("Failed to get favorites via gRPC", zap.String("user_id", userID), zap.Error(err))
		st, sOk := status.FromError(err)
		if sOk {
			http.Error(w, st.Message(), http.StatusInternalServerError)
		} else {
			http.Error(w, "Failed to get favorites: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Error("Failed to encode GetFavorites response", zap.String("user_id", userID), zap.Error(err))
		http.Error(w, status.Errorf(codes.Internal, "Failed to encode response: %v", err).Error(), http.StatusInternalServerError)
	}
}

// HandleGetPhotoURLs обрабатывает получение URL фотографий
func (h *ListingHandler) HandleGetPhotoURLs(w http.ResponseWriter, r *http.Request) { // Сигнатура для chi
	id := chi.URLParam(r, "id") // Используем chi.URLParam
	if id == "" {
		h.logger.Error("Missing id parameter for GetPhotoURLs")
		http.Error(w, status.Errorf(codes.InvalidArgument, "Missing id parameter").Error(), http.StatusBadRequest)
		return
	}

	ctx := withAuth(r.Context(), r)
	client := listing_service.NewListingServiceClient(h.client)
	resp, err := client.GetPhotoURLs(ctx, &listing_service.GetListingRequest{Id: id})
	if err != nil {
		h.logger.Error("Failed to get photo URLs via gRPC", zap.String("listing_id", id), zap.Error(err))
		st, ok := status.FromError(err)
		if ok {
			http.Error(w, st.Message(), http.StatusInternalServerError)
		} else {
			http.Error(w, "Failed to get photo URLs: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Error("Failed to encode GetPhotoURLs response", zap.String("listing_id", id), zap.Error(err))
		http.Error(w, status.Errorf(codes.Internal, "Failed to encode response: %v", err).Error(), http.StatusInternalServerError)
	}
}

// HandleUpdateListingStatus обрабатывает обновление статуса объявления
func (h *ListingHandler) HandleUpdateListingStatus(w http.ResponseWriter, r *http.Request) { // Сигнатура для chi
	id := chi.URLParam(r, "id") // Используем chi.URLParam
	var req listing_service.UpdateListingStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Invalid request body for UpdateListingStatus", zap.String("id", id), zap.Error(err))
		http.Error(w, status.Errorf(codes.InvalidArgument, "Invalid request body: %v", err).Error(), http.StatusBadRequest)
		return
	}
	req.Id = id

	ctx := withAuth(r.Context(), r)
	client := listing_service.NewListingServiceClient(h.client)
	resp, err := client.UpdateListingStatus(ctx, &req)
	if err != nil {
		h.logger.Error("Failed to update listing status via gRPC", zap.String("id", id), zap.Error(err))
		st, ok := status.FromError(err)
		if ok {
			http.Error(w, st.Message(), http.StatusInternalServerError)
		} else {
			http.Error(w, "Failed to update listing status: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Error("Failed to encode UpdateListingStatus response", zap.String("id", id), zap.Error(err))
		http.Error(w, status.Errorf(codes.Internal, "Failed to encode response: %v", err).Error(), http.StatusInternalServerError)
	}
}

// withAuth добавляет JWT-токен в метаданные контекста для gRPC вызовов
func withAuth(ctx context.Context, r *http.Request) context.Context {
	token := r.Header.Get("Authorization") // Это оригинальный Bearer токен
	if token != "" {
		// Для gRPC нам нужен сам токен, без "Bearer "
		// Эта логика может быть более сложной, если ваше middleware
		// уже извлекло чистый токен и положило его в контекст.
		// Здесь предполагаем, что gRPC сервис ожидает полный Bearer токен
		// или вы это обрабатываете на стороне gRPC сервиса.
		// Если gRPC сервис ожидает только сам токен, то:
		// if strings.HasPrefix(token, "Bearer ") {
		// 	token = strings.TrimPrefix(token, "Bearer ")
		// }
		return metadata.AppendToOutgoingContext(ctx, "authorization", token)
	}
	// Если middleware JWTAuth уже положило 'user_id' в контекст,
	// то этот 'user_id' будет доступен в `ctx`.
	// Функция `withAuth` сейчас просто передает оригинальный токен дальше,
	// если он есть, для gRPC вызова.
	return ctx
}