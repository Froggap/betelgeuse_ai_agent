package authentities

import "time"

type SessionData struct {
	AccessToken  string
	RefreshToken string
	RefreshCode  string
	CreatedAt    time.Time
	IsValid      bool
}
