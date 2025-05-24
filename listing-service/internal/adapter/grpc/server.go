package grpc

import (

	"google.golang.org/grpc"
	"github.com/Abdurahmanit/GroupProject/listing-service/internal/adapter/grpc/middleware"
	"github.com/Abdurahmanit/GroupProject/listing-service/internal/platform/logger" // Твой логгер
	// sdktrace "go.opentelemetry.io/otel/sdk/trace" // Если передаешь TracerProvider
)

// NewGRPCServer теперь принимает логгер и jwtSecret
func NewGRPCServer(
	appLogger *logger.Logger,
	jwtSecret string,
	// tracerProvider *sdktrace.TracerProvider, // Если трейсер инициализируется в main и передается
) (*grpc.Server, func()) { // cleanup для остановки сервера

	// Определяем публичные методы (полные пути, как их видит gRPC)
	// Пример: "/<package>.<Service>/<Method>"
	// Уточни имена пакета и сервиса из твоего listing.proto
	// package listing; service ListingService { ... } -> /listing.ListingService/MethodName
	publicMethods := map[string]bool{
		"/listing.ListingService/GetListingByID": true,
		"/listing.ListingService/SearchListings": true,
		// "/listing.ListingService/GetListingStatus": true, // Сделай публичным, если нужно
		// "/listing.ListingService/GetPhotoURLs":   true, // Сделай публичным, если нужно
		// Добавь сюда любые другие методы, которые должны быть доступны без токена.
	}

	unaryInterceptors := []grpc.UnaryServerInterceptor{
		middleware.TracingInterceptor(), // Предполагается, что он у тебя есть
		middleware.LoggingInterceptor(appLogger),
		middleware.AuthInterceptor(jwtSecret, appLogger, publicMethods), // Передаем карту публичных методов
	}

	server := grpc.NewServer(
		grpc.ChainUnaryInterceptor(unaryInterceptors...),
	)

	appLogger.Info("gRPC server configured with interceptors: Tracing, Logging, Auth")

	cleanup := func() {
		appLogger.Info("Calling gRPC server's GracefulStop...")
		server.GracefulStop()
		appLogger.Info("gRPC server GracefulStop completed.")
		// Если tracerProvider передавался и его shutdown нужно делать здесь:
		// if tracerProvider != nil {
		// 	if err := tracerProvider.Shutdown(context.Background()); err != nil {
		// 		appLogger.Error("Failed to shutdown tracer provider during gRPC server cleanup", "error", err.Error())
		// 	}
		// }
	}

	return server, cleanup
}