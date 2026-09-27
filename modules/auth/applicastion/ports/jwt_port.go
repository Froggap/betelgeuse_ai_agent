package authports

type JwtPort interface {
	GenerateToken(data any, durationStr string) (string, error)
	VerifyToken(tokenString string) (any, error)
	GenerateRefreshCode(accessToken string, refreshToken string) (string, error)
}
