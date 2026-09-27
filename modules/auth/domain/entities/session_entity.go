package authentities

import (
	"time"

	"github.com/google/uuid"
)

type Session struct {
	SessionData SessionData
	UserData    User
}

func CreateSession(
	id uuid.UUID,
	createdAt time.Time,
	userName string,
	role string,
	accessToken string,
	refreshToken string,
	refreshCode string,
	sessionCreatedAt time.Time,
	isValid bool,
) *Session {
	return &Session{
		UserData: User{
			ID:        id,
			CreatedAt: createdAt,
			UserName:  userName,
			Role:      role,
		},
		SessionData: SessionData{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			RefreshCode:  refreshCode,
			CreatedAt:    sessionCreatedAt,
			IsValid:      isValid,
		},
	}
}
