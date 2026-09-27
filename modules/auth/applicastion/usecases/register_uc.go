package authusecases

import (
	"context"

	authrequestsdtos "scrapper-ai/modules/auth/applicastion/dtos/requests"
	authports "scrapper-ai/modules/auth/applicastion/ports"
	authentities "scrapper-ai/modules/auth/domain/entities"
	authenums "scrapper-ai/modules/auth/domain/enums"
	authvalueobjects "scrapper-ai/modules/auth/domain/valueobjects"
	globalerrors "scrapper-ai/modules/shared/infrastructure/errors"
)

type RegisterAccountUC struct {
	authRepo   authports.AuthRepositoryPort
	encrypter  authports.EncryptorPort
	jwtService authports.JwtPort
}

func NewRegisterAccountUC(
	authRepo authports.AuthRepositoryPort,
	encrypter authports.EncryptorPort,
	jwtService authports.JwtPort,
) *RegisterAccountUC {
	return &RegisterAccountUC{
		authRepo:   authRepo,
		encrypter:  encrypter,
		jwtService: jwtService,
	}
}

func (r *RegisterAccountUC) Execute(ctx context.Context, dto *authrequestsdtos.RegisterRequestDTO) (*authentities.Session, error) {
	usernameVO, err := authvalueobjects.NewUsernameVO(dto.Username)
	if err != nil {
		return nil, err
	}

	passwordVO, err := authvalueobjects.NewPasswordVO(dto.Password)
	if err != nil {
		return nil, err
	}

	exists, err := r.authRepo.CheckUsernameExists(ctx, usernameVO.Value())
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, globalerrors.NewAppError(
			409,
			"Conflict",
			"username already exists",
			nil,
		)
	}

	hashedPassword, err := r.encrypter.Encode(passwordVO.Value())
	if err != nil {
		return nil, err
	}

	sessionCreated, err := r.authRepo.CreateSession(ctx)
	if err != nil {
		return nil, err
	}

	user, err := r.authRepo.CreateUser(ctx, usernameVO.Value(), hashedPassword, authenums.RoleAdmin, sessionCreated.ID)
	if err != nil {
		return nil, err
	}

	tokenData := &authrequestsdtos.TokenSessionDto{
		ID:       user.ID,
		UserName: user.UserName,
	}

	accessToken, err := r.jwtService.GenerateToken(tokenData, "5m")
	if err != nil {
		return nil, err
	}

	refreshToken, err := r.jwtService.GenerateToken(tokenData, "72h")
	if err != nil {
		return nil, err
	}

	refreshCode, err := r.jwtService.GenerateRefreshCode(accessToken, refreshToken)
	if err != nil {
		return nil, err
	}

	if err := r.authRepo.UpdateSessionTokens(ctx, sessionCreated.ID, refreshToken, refreshCode); err != nil {
		return nil, err
	}

	return authentities.CreateSession(
		user.ID,
		user.CreatedAt,
		user.UserName,
		user.Role,
		accessToken,
		refreshToken,
		refreshCode,
		sessionCreated.CreatedAt,
		true,
	), nil
}
