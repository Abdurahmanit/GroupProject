package middleware

import (
	"context"
	"strings"

	"github.com/Abdurahmanit/GroupProject/listing-service/internal/platform/logger" // Путь к твоему логгеру
	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// UserIDKeyType — кастомный тип для ключа контекста, чтобы избежать коллизий.
type UserIDKeyType string

// UserIDKey — ключ, используемый для хранения и извлечения UserID из контекста.
const UserIDKey UserIDKeyType = "authenticatedUserID"

// Claims определяет структуру claims в JWT, ожидаемую от user-service.
type Claims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

// AuthInterceptor создает gRPC унарный interceptor для аутентификации.
func AuthInterceptor(jwtSecret string, log *logger.Logger, publicMethods map[string]bool) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		log.Debug("AuthInterceptor: processing request", "method", info.FullMethod)

		// Проверяем, является ли метод публичным
		if publicMethods[info.FullMethod] {
			log.Debug("AuthInterceptor: public method, skipping authentication", "method", info.FullMethod)
			return handler(ctx, req)
		}
		log.Debug("AuthInterceptor: protected method, proceeding with authentication", "method", info.FullMethod)


		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			log.Warn("AuthInterceptor: missing metadata from context", "method", info.FullMethod)
			return nil, status.Errorf(codes.Unauthenticated, "metadata is not provided")
		}

		authHeaders := md.Get("authorization") // gRPC метаданные обычно в нижнем регистре
		if len(authHeaders) == 0 {
			log.Warn("AuthInterceptor: 'authorization' header not found in metadata", "method", info.FullMethod)
			return nil, status.Errorf(codes.Unauthenticated, "authorization token is not provided")
		}

		// Обычно токен передается как "Bearer <token>"
		authHeader := authHeaders[0]
		parts := strings.Fields(authHeader) // Разделяет по пробелам
		if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
			log.Warn("AuthInterceptor: invalid 'authorization' header format, expected 'Bearer <token>'", "method", info.FullMethod, "header_value", authHeader)
			return nil, status.Errorf(codes.Unauthenticated, "authorization token format is invalid, expected 'Bearer <token>'")
		}
		tokenString := parts[1]

		if tokenString == "" {
			log.Warn("AuthInterceptor: token string is empty after 'Bearer' prefix", "method", info.FullMethod)
			return nil, status.Errorf(codes.Unauthenticated, "authorization token is empty")
		}

		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			// Проверка алгоритма подписи
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				log.Error("AuthInterceptor: unexpected signing method", "method", info.FullMethod, "algorithm", token.Header["alg"])
				return nil, status.Errorf(codes.Unauthenticated, "unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(jwtSecret), nil
		})

		if err != nil {
			log.Warn("AuthInterceptor: token parsing or validation failed", "method", info.FullMethod, "error", err.Error())
			// Можно детализировать ошибки, например, для jwt.ErrTokenExpired
			if err == jwt.ErrTokenExpired {
				return nil, status.Errorf(codes.Unauthenticated, "token has expired")
			}
			return nil, status.Errorf(codes.Unauthenticated, "token is invalid: %v", err)
		}

		if !token.Valid {
			log.Warn("AuthInterceptor: token is not valid (claims validation failed or signature mismatch)", "method", info.FullMethod)
			return nil, status.Errorf(codes.Unauthenticated, "token is not valid")
		}

		if claims.UserID == "" {
			log.Error("AuthInterceptor: UserID not found in token claims after successful validation", "method", info.FullMethod)
			return nil, status.Errorf(codes.Unauthenticated, "UserID not found in token claims")
		}

		// Добавляем UserID в контекст
		newCtx := context.WithValue(ctx, UserIDKey, claims.UserID)
		log.Info("AuthInterceptor: user successfully authenticated", "method", info.FullMethod, "user_id", claims.UserID)

		// Передаем управление следующему обработчику или самому RPC методу
		return handler(newCtx, req)
	}
}