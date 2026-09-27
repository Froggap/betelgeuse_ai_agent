package authusecases

import (
	"context"

	authrequestsdtos "scrapper-ai/modules/auth/applicastion/dtos/requests"
	authports "scrapper-ai/modules/auth/applicastion/ports"
	authentities "scrapper-ai/modules/auth/domain/entities"
	authenums "scrapper-ai/modules/auth/domain/enums"
	authvalueobjects "scrapper-ai/modules/auth/domain/valueobjects"
	commondomainerrors "scrapper-ai/modules/shared/domain/errors"
)

type LoginAccountUC struct {
	authRepo   authports.AuthRepositoryPort
	encrypter  authports.EncryptorPort
	jwtService authports.JwtPort
}

func NewLoginAccountUC(
	authRepo authports.AuthRepositoryPort,
	encrypter authports.EncryptorPort,
	jwtService authports.JwtPort,
) *LoginAccountUC {
	return &LoginAccountUC{
		authRepo:   authRepo,
		encrypter:  encrypter,
		jwtService: jwtService,
	}
}

func (l *LoginAccountUC) Execute(ctx context.Context, dto *authrequestsdtos.LoginRequestDTO) (*authentities.Session, error) {
	usernameVO, err := authvalueobjects.NewUsernameVO(dto.Username)
	if err != nil {
		return nil, err
	}

	user, err := l.authRepo.FindUserByUserName(ctx, usernameVO.Value())
	if err != nil {
		return nil, err
	}

	match, err := l.encrypter.Compare(dto.Password, user.Password)
	if err != nil {
		return nil, err
	}

	if !match {
		return nil, commondomainerrors.NewValidationError(
			"password",
			"password invalid",
		)
	}

	tokenData := &authrequestsdtos.TokenSessionDto{
		ID:       user.ID,
		UserName: user.UserName,
	}

	accessToken, err := l.jwtService.GenerateToken(tokenData, "5m")
	if err != nil {
		return nil, err
	}

	refreshToken, err := l.jwtService.GenerateToken(tokenData, "72h")
	if err != nil {
		return nil, err
	}

	refreshCode, err := l.jwtService.GenerateRefreshCode(accessToken, refreshToken)
	if err != nil {
		return nil, err
	}

	if err := l.authRepo.UpdateSessionTokens(ctx, user.SessionID, refreshToken, refreshCode); err != nil {
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
		user.SessionCreatedAt,
		user.AccountValid == authenums.AccountValid,
	), nil
}
