package authrequestsdtos

import "github.com/google/uuid"

type TokenSessionDto struct {
	ID       uuid.UUID
	UserName string
}
