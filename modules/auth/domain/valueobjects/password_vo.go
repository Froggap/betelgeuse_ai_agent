package authvalueobjects

import (
	commondomainerrors "scrapper-ai/modules/shared/domain/errors"
)

const minPasswordLength = 8

type PasswordVO struct {
	value string
}

func NewPasswordVO(raw string) (*PasswordVO, error) {
	if raw == "" {
		return nil, commondomainerrors.NewValidationError(
			"password",
			"password is required",
		)
	}

	if len(raw) < minPasswordLength {
		return nil, commondomainerrors.NewValidationError(
			"password",
			"password must be at least 8 characters",
		)
	}

	return &PasswordVO{value: raw}, nil
}

func (p *PasswordVO) Value() string {
	return p.value
}
