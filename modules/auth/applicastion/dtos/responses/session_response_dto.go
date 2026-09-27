package authresponsesdtos

import (
	"time"

	"github.com/google/uuid"
)

type UserDataDTO struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UserName  string    `json:"userName"`
	Role      string    `json:"role"`
}

type SessionDataDTO struct {
	AccessToken string    `json:"accessToken"`
	CreatedAt   time.Time `json:"createdAt"`
	IsValid     bool      `json:"isValid"`
}

type SessionResponseDTO struct {
	UserData    UserDataDTO    `json:"userData"`
	SessionData SessionDataDTO `json:"sessionData"`
}
