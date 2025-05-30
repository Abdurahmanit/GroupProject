package middleware

// ContextKey - пользовательский тип для ключей контекста, чтобы избежать коллизий.
type ContextKey string

const (
	// UserIDCtxKey - ключ, используемый для хранения и извлечения аутентифицированного UserID из контекста.
	// Убедитесь, что ваш JWTAuth middleware использует именно этот ключ.
	UserIDCtxKey = ContextKey("user_id")

	// UserRoleCtxKey - ключ, используемый для хранения и извлечения роли аутентифицированного пользователя из контекста.
	// Убедитесь, что ваш JWTAuth middleware использует именно этот ключ, если вы извлекаете роль.
	UserRoleCtxKey = ContextKey("user_role")
)
