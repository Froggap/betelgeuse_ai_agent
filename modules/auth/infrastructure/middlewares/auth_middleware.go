package authmiddlewares

import (
	"context"
	"net/http"
	"strings"

	authrequestsdtos "scrapper-ai/modules/auth/applicastion/dtos/requests"
	authports "scrapper-ai/modules/auth/applicastion/ports"
	globalerrors "scrapper-ai/modules/shared/infrastructure/errors"
	sharedinframappers "scrapper-ai/modules/shared/infrastructure/mappers"

	"github.com/google/uuid"
)

type contextKey string

const SessionContextKey contextKey = "user_session"
const NeedsReloadContextKey contextKey = "needs_reload"

type AuthMiddleware struct {
	JwtService authports.JwtPort
}

func NewAuthMiddleware(jwtService authports.JwtPort) *AuthMiddleware {
	return &AuthMiddleware{JwtService: jwtService}
}

func (a *AuthMiddleware) AccessToken(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			sharedinframappers.RespondWithError(w, globalerrors.NewAppError(401, "Unauthorized", "Missing or invalid token", nil))
			return
		}
		token := strings.Split(authHeader, " ")[1]
		userData, err := a.JwtService.VerifyToken(token)
		if err != nil {
			sharedinframappers.RespondWithError(w, globalerrors.NewAppError(401, "Unauthorized", "token expired", nil))
			return
		}
		ctx := context.WithValue(r.Context(), SessionContextKey, userData)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (a *AuthMiddleware) RefreshToken(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		refreshCode := r.Header.Get("X-Refresh-Code")
		refreshToken := r.Header.Get("X-Refresh-Token")

		if strings.TrimSpace(refreshCode) == "" || strings.TrimSpace(refreshToken) == "" {
			sharedinframappers.RespondWithError(w, globalerrors.NewAppError(401, "Unauthorized", "Missing or invalid refresh token", nil))
			return
		}

		// Intentar verificar el refresh token
		userData, err := a.JwtService.VerifyToken(refreshToken)
		if err != nil {
			// Token expirado: marcar que necesita reload y continuar.
			// El use-case validará el refresh code.
			ctx := context.WithValue(r.Context(), NeedsReloadContextKey, true)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		// Token válido: pasar user data al contexto
		ctx := context.WithValue(r.Context(), SessionContextKey, userData)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// UserIDFromContext extrae el id del usuario autenticado del contexto.
// El middleware guarda los claims del JWT (map) o un TokenSessionDto.
func UserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	raw := ctx.Value(SessionContextKey)

	switch v := raw.(type) {
	case authrequestsdtos.TokenSessionDto:
		return v.ID, true
	case map[string]interface{}:
		if idStr, ok := v["ID"].(string); ok {
			if id, err := uuid.Parse(idStr); err == nil {
				return id, true
			}
		}
	}

	return uuid.Nil, false
}
