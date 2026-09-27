package authraws

import (
	"time"

	"github.com/google/uuid"
)

// UserFoundRaw es el crudo que botá la BD (query con JOIN de users, sessions
// y roles). El mapper de infra lo transforma en la entidad de dominio.
type UserFoundRaw struct {
	ID               uuid.UUID  `gorm:"column:id"`
	UserName         string     `gorm:"column:user_name"`
	Password         string     `gorm:"column:password"`
	CreatedAt        time.Time  `gorm:"column:created_at"`
	CreatedBy        *uuid.UUID `gorm:"column:created_by"`
	AccountValid     string     `gorm:"column:account_valid"`
	SessionID        uuid.UUID  `gorm:"column:session_id"`
	RefreshToken     string     `gorm:"column:refresh_token"`
	RefreshCode      string     `gorm:"column:refresh_code"`
	SessionCreatedAt time.Time  `gorm:"column:session_created_at"`
	RoleID           string     `gorm:"column:role_id"`
}
