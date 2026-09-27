package authadapters

import (
	"context"
	"errors"

	dbmodels "scrapper-ai/config/db/dbmodels"
	authports "scrapper-ai/modules/auth/applicastion/ports"
	authentities "scrapper-ai/modules/auth/domain/entities"
	authinframappers "scrapper-ai/modules/auth/infrastructure/mappers"
	authraws "scrapper-ai/modules/auth/infrastructure/raws"
	globalerrors "scrapper-ai/modules/shared/infrastructure/errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AuthRepository struct {
	db *gorm.DB
}

func NewAuthRepository(db *gorm.DB) authports.AuthRepositoryPort {
	return &AuthRepository{db: db}
}

const userFoundSelect = `
	u.id,
	u.user_name,
	u.password,
	u.created_at,
	u.created_by,
	u.account_valid,
	s.id AS session_id,
	s.refresh_token,
	s.refresh_code,
	s.created_at AS session_created_at,
	r.id AS role_id
`

const userFoundJoins = `
	INNER JOIN auth.sessions AS s ON s.id = u.session_id
	INNER JOIN auth.roles AS r ON r.id = u.role_id
`

func (a *AuthRepository) findUserWhere(ctx context.Context, where string, args ...any) (*authentities.UserAuth, error) {
	var raw authraws.UserFoundRaw

	result := a.db.WithContext(ctx).
		Table("auth.users AS u").
		Select(userFoundSelect).
		Joins(userFoundJoins).
		Where(where, args...).
		Limit(1).
		Scan(&raw)

	if result.Error != nil {
		return nil, globalerrors.NewAppError(
			500,
			"Database Error",
			"An unexpected error occurred while querying the user",
			result.Error,
		)
	}

	if result.RowsAffected == 0 {
		return nil, nil
	}

	return authinframappers.UserRawToEntity(&raw), nil
}

func (a *AuthRepository) FindUserByUserName(ctx context.Context, username string) (*authentities.UserAuth, error) {
	user, err := a.findUserWhere(ctx, "u.user_name = ?", username)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, globalerrors.NewAppError(
			404,
			"User Not Found",
			"No user was found with the provided username",
			nil,
		)
	}
	return user, nil
}

func (a *AuthRepository) FindUserByID(ctx context.Context, id uuid.UUID) (*authentities.UserAuth, error) {
	user, err := a.findUserWhere(ctx, "u.id = ?", id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, globalerrors.NewAppError(
			404,
			"User Not Found",
			"No user was found with the provided id",
			nil,
		)
	}
	return user, nil
}

func (a *AuthRepository) FindUserByRefreshCode(ctx context.Context, refreshCode string) (*authentities.UserAuth, error) {
	return a.findUserWhere(ctx, "s.refresh_code = ?", refreshCode)
}

func (a *AuthRepository) CheckUsernameExists(ctx context.Context, username string) (bool, error) {
	var count int64

	result := a.db.WithContext(ctx).
		Model(&dbmodels.UserModel{}).
		Where("user_name = ?", username).
		Count(&count)

	if result.Error != nil {
		return false, globalerrors.NewAppError(
			500,
			"Database Error",
			"Failed to check username availability",
			result.Error,
		)
	}

	return count > 0, nil
}

func (a *AuthRepository) CreateSession(ctx context.Context) (*authentities.SessionCreated, error) {
	session := &dbmodels.SessionModel{ID: uuid.New()}

	if err := a.db.WithContext(ctx).Create(session).Error; err != nil {
		return nil, globalerrors.NewAppError(
			500,
			"Database Error",
			"Failed to create session",
			err,
		)
	}

	return &authentities.SessionCreated{
		ID:        session.ID,
		CreatedAt: session.CreatedAt,
	}, nil
}

func (a *AuthRepository) CreateUser(ctx context.Context, username string, hashedPassword string, roleID string, sessionID uuid.UUID) (*authentities.User, error) {
	user := &dbmodels.UserModel{
		ID:           uuid.New(),
		UserName:     username,
		Password:     hashedPassword,
		AccountValid: "VALID",
		SessionID:    sessionID,
		RoleID:       roleID,
	}

	if err := a.db.WithContext(ctx).Create(user).Error; err != nil {
		return nil, globalerrors.NewAppError(
			500,
			"Database Error",
			"Failed to create user",
			err,
		)
	}

	return &authentities.User{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UserName:  user.UserName,
		Role:      user.RoleID,
	}, nil
}

func (a *AuthRepository) UpdateSessionTokens(ctx context.Context, sessionID uuid.UUID, refreshToken string, refreshCode string) error {
	result := a.db.WithContext(ctx).
		Model(&dbmodels.SessionModel{}).
		Where("id = ?", sessionID).
		Updates(map[string]any{
			"refresh_token": refreshToken,
			"refresh_code":  refreshCode,
		})

	if result.Error != nil {
		return globalerrors.NewAppError(
			500,
			"Database Error",
			"Failed to update session tokens",
			result.Error,
		)
	}

	if result.RowsAffected == 0 {
		return globalerrors.NewAppError(
			404,
			"Session Not Found",
			"No session was found with the provided id",
			errors.New("session not found"),
		)
	}

	return nil
}
