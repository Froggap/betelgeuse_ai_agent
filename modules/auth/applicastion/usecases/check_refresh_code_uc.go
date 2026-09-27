package authusecases

import (
	"context"
	"log"

	authrequestsdtos "scrapper-ai/modules/auth/applicastion/dtos/requests"
	authports "scrapper-ai/modules/auth/applicastion/ports"
	authentities "scrapper-ai/modules/auth/domain/entities"
	authenums "scrapper-ai/modules/auth/domain/enums"
	commonvalueobjects "scrapper-ai/modules/shared/domain/valueobjects"
	globalerrors "scrapper-ai/modules/shared/infrastructure/errors"
)

type CheckRefreshCodeUC struct {
	authRepo   authports.AuthRepositoryPort
	jwtService authports.JwtPort
}

func NewCheckRefreshCodeUC(
	authRepo authports.AuthRepositoryPort,
	jwtService authports.JwtPort,
) *CheckRefreshCodeUC {
	return &CheckRefreshCodeUC{
		authRepo:   authRepo,
		jwtService: jwtService,
	}
}

func (c *CheckRefreshCodeUC) Execute(ctx context.Context, userID string, refreshToken string, refreshCode string) (*authentities.Session, error) {
	user, err := c.resolveUser(ctx, userID, refreshCode)
	if err != nil {
		return nil, err
	}

	tokenData := &authrequestsdtos.TokenSessionDto{
		ID:       user.ID,
		UserName: user.UserName,
	}

	accessToken, err := c.jwtService.GenerateToken(tokenData, "5m")
	if err != nil {
		return nil, err
	}

	newRefreshToken, err := c.jwtService.GenerateToken(tokenData, "72h")
	if err != nil {
		return nil, err
	}

	newRefreshCode, err := c.jwtService.GenerateRefreshCode(accessToken, newRefreshToken)
	if err != nil {
		return nil, err
	}

	if err := c.authRepo.UpdateSessionTokens(ctx, user.SessionID, newRefreshToken, newRefreshCode); err != nil {
		return nil, err
	}

	return authentities.CreateSession(
		user.ID,
		user.CreatedAt,
		user.UserName,
		user.Role,
		accessToken,
		newRefreshToken,
		newRefreshCode,
		user.SessionCreatedAt,
		user.AccountValid == authenums.AccountValid,
	), nil
}

// resolveUser obtiene al usuario asociado a la sesión.
//
// Nivel 1 (fuerte): si el middleware pudo verificar el refresh token, llega el
// userID del propio token. En ese caso confiamos en la firma/expiración del JWT:
// el cliente es legítimo y la sesión se "re-engancha" rotando los tokens.
//
// Nivel 2 (código de respaldo): si el refresh token expiró o no pudo verificarse,
// recuperamos al usuario a partir del refresh code almacenado en la BD, que es la
// fuente de verdad persistida del lado del servidor.
func (c *CheckRefreshCodeUC) resolveUser(ctx context.Context, userID string, refreshCode string) (*authentities.UserAuth, error) {
	if userID != "" {
		idParsed, err := commonvalueobjects.NewUUIDVO(userID)
		if err != nil {
			return nil, err
		}
		user, err := c.authRepo.FindUserByID(ctx, idParsed.Value())
		if err != nil {
			log.Printf("reload session: refresh token valido (userID=%s) pero FindUserByID fallo", userID)
			return nil, globalerrors.NewAppError(
				401,
				"Refresh code",
				"session not valid",
				err,
			)
		}
		return user, nil
	}

	log.Printf("reload session: refresh token no verificable, buscando por refresh code (len=%d)", len(refreshCode))
	user, err := c.authRepo.FindUserByRefreshCode(ctx, refreshCode)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, globalerrors.NewAppError(
			401,
			"Refresh code",
			"session not valid",
			nil,
		)
	}
	return user, nil
}
