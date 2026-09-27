package authrequestsdtos

type LoginRequestDTO struct {
	Username string `json:"userName" validate:"required"`
	Password string `json:"password" validate:"required"`
}
