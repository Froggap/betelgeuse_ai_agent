package authentities

import (
	"time"

	"github.com/google/uuid"
)

type SessionCreated struct {
	ID        uuid.UUID
	CreatedAt time.Time
}
