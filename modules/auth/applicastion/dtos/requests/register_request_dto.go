package authrequestsdtos

type RegisterRequestDTO struct {
	Username string `json:"userName" validate:"required"`
	Password string `json:"password" validate:"required,min=8"`
}
