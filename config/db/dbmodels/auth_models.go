package dbmodels

import (
	"time"

	"github.com/google/uuid"
)

type RoleModel struct {
	ID       string `gorm:"column:id;primaryKey"`
	NameRole string `gorm:"column:name_role;uniqueIndex"`
}

func (RoleModel) TableName() string {
	return "auth.roles"
}

type SessionModel struct {
	ID           uuid.UUID `gorm:"column:id;primaryKey;type:uuid"`
	RefreshToken string    `gorm:"column:refresh_token"`
	RefreshCode  string    `gorm:"column:refresh_code"`
	CreatedAt    time.Time `gorm:"column:created_at"`
}

func (SessionModel) TableName() string {
	return "auth.sessions"
}

type UserModel struct {
	ID           uuid.UUID `gorm:"column:id;primaryKey;type:uuid"`
	UserName     string    `gorm:"column:user_name;uniqueIndex"`
	Password     string    `gorm:"column:password"`
	CreatedAt    time.Time `gorm:"column:created_at"`
	CreatedBy    uuid.UUID `gorm:"column:created_by;type:uuid"`
	AccountValid string    `gorm:"column:account_valid"`
	SessionID    uuid.UUID `gorm:"column:session_id;type:uuid"`
	RoleID       string    `gorm:"column:role_id"`
}

func (UserModel) TableName() string {
	return "auth.users"
}
