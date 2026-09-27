package authentities

import (
	"time"

	"github.com/google/uuid"
)

// UserAuth es la entidad que devuelve el repositorio: el crudo que botó la BD
// ya mapeado por el mapper de infra, descartando la data que no importa.
type UserAuth struct {
	ID        uuid.UUID
	CreatedAt time.Time
	UserName  string
	Password  string

	AccountValid string
	Role         string

	SessionID        uuid.UUID
	RefreshToken     string
	RefreshCode      string
	SessionCreatedAt time.Time
}
