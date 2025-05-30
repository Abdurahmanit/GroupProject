package middleware

import (
	"context"
	"errors"
	"strings"

	"github.com/Abdurahmanit/GroupProject/review-service/internal/platform/logger"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap" // Import zap
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// UserIDKeyType is a custom type for the user ID context key to avoid collisions.
type UserIDKeyType string

// UserRoleKeyType is a custom type for the user role context key.
type UserRoleKeyType string

const (
	// UserIDKey is the key used to store and retrieve the authenticated UserID from the context.
	UserIDKey UserIDKeyType = "authenticatedUserID"
	// UserRoleKey is the key used to store and retrieve the authenticated user's role from the context.
	UserRoleKey UserRoleKeyType = "authenticatedUserRole"
)

// Claims defines the structure of the JWT claims expected from the token.
type Claims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// AuthInterceptor creates a gRPC unary server interceptor for authentication and basic authorization.
func AuthInterceptor(jwtSecret string, log *logger.Logger, publicMethods map[string]bool, requiredRoles map[string][]string) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		log.Debug("AuthInterceptor: processing request", zap.String("method", info.FullMethod))

		if publicMethods[info.FullMethod] {
			log.Debug("AuthInterceptor: public method, skipping authentication", zap.String("method", info.FullMethod))
			return handler(ctx, req)
		}
		log.Debug("AuthInterceptor: protected method, proceeding with authentication", zap.String("method", info.FullMethod))

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			log.Warn("AuthInterceptor: missing metadata from context", zap.String("method", info.FullMethod))
			return nil, status.Errorf(codes.Unauthenticated, "metadata is not provided")
		}

		authHeaders := md.Get("authorization")
		if len(authHeaders) == 0 {
			log.Warn("AuthInterceptor: 'authorization' header not found", zap.String("method", info.FullMethod))
			return nil, status.Errorf(codes.Unauthenticated, "authorization token is not provided")
		}

		authHeader := authHeaders[0]
		parts := strings.Fields(authHeader)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			log.Warn("AuthInterceptor: invalid 'authorization' header format", zap.String("method", info.FullMethod), zap.String("header_value", authHeader))
			return nil, status.Errorf(codes.Unauthenticated, "authorization token format is invalid, expected 'Bearer <token>'")
		}
		tokenString := parts[1]

		if tokenString == "" {
			log.Warn("AuthInterceptor: token string is empty", zap.String("method", info.FullMethod))
			return nil, status.Errorf(codes.Unauthenticated, "authorization token is empty")
		}

		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				log.Error("AuthInterceptor: unexpected signing method", zap.String("method", info.FullMethod), zap.Any("algorithm", token.Header["alg"]))
				return nil, status.Errorf(codes.Unauthenticated, "unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(jwtSecret), nil
		})

		if err != nil {
			log.Warn("AuthInterceptor: token parsing/validation failed", zap.String("method", info.FullMethod), zap.Error(err))
			if errors.Is(err, jwt.ErrTokenExpired) {
				return nil, status.Errorf(codes.Unauthenticated, "token has expired")
			}
			return nil, status.Errorf(codes.Unauthenticated, "token is invalid: %v", err)
		}

		if !token.Valid {
			log.Warn("AuthInterceptor: token is not valid", zap.String("method", info.FullMethod))
			return nil, status.Errorf(codes.Unauthenticated, "token is not valid")
		}

		if claims.UserID == "" {
			log.Error("AuthInterceptor: UserID not found in token claims", zap.String("method", info.FullMethod))
			return nil, status.Errorf(codes.Unauthenticated, "UserID not found in token claims")
		}
		// Role check is important for authorization logic that follows
		if claims.Role == "" {
			log.Warn("AuthInterceptor: Role not found in token claims, proceeding with caution", zap.String("method", info.FullMethod), zap.String("user_id", claims.UserID))
			// Potentially deny access if role is mandatory for all authenticated routes
			// return nil, status.Errorf(codes.PermissionDenied, "Role not found in token claims, access denied")
		}

		if roles, methodRequiresRoles := requiredRoles[info.FullMethod]; methodRequiresRoles {
			authorized := false
			for _, requiredRole := range roles {
				if claims.Role == requiredRole {
					authorized = true
					break
				}
			}
			if !authorized {
				log.Warn("AuthInterceptor: user does not have required role",
					zap.String("method", info.FullMethod),
					zap.String("user_id", claims.UserID),
					zap.String("user_role", claims.Role),
					zap.Strings("required_roles", roles))
				return nil, status.Errorf(codes.PermissionDenied, "user role '%s' not authorized for this action", claims.Role)
			}
			log.Debug("AuthInterceptor: user role authorized", zap.String("method", info.FullMethod), zap.String("user_role", claims.Role))
		}

		newCtx := context.WithValue(ctx, UserIDKey, claims.UserID)
		newCtx = context.WithValue(newCtx, UserRoleKey, claims.Role)

		log.Info("AuthInterceptor: user authenticated and authorized",
			zap.String("method", info.FullMethod),
			zap.String("user_id", claims.UserID),
			zap.String("role", claims.Role))

		return handler(newCtx, req)
	}
}
