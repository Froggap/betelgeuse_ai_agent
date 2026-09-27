package authports

import (
	"context"

	authentities "scrapper-ai/modules/auth/domain/entities"

	"github.com/google/uuid"
)

// AuthRepositoryPort es el puerto de salida que el use-case usa para
// consultar la BD. Siempre devuelve entidades de dominio, nunca crudos.
type AuthRepositoryPort interface {
	FindUserByUserName(ctx context.Context, username string) (*authentities.UserAuth, error)
	FindUserByID(ctx context.Context, id uuid.UUID) (*authentities.UserAuth, error)
	FindUserByRefreshCode(ctx context.Context, refreshCode string) (*authentities.UserAuth, error)
	CheckUsernameExists(ctx context.Context, username string) (bool, error)

	CreateSession(ctx context.Context) (*authentities.SessionCreated, error)
	CreateUser(ctx context.Context, username string, hashedPassword string, roleID string, sessionID uuid.UUID) (*authentities.User, error)

	UpdateSessionTokens(ctx context.Context, sessionID uuid.UUID, refreshToken string, refreshCode string) error
}
