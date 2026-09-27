package authinframappers

import (
	authentities "scrapper-ai/modules/auth/domain/entities"
	authraws "scrapper-ai/modules/auth/infrastructure/raws"
)

// UserRawToEntity mapea el crudo que botó la BD a la entidad de dominio,
// eligiendo solamente la data que le importa a la aplicación
// (se descarta, por ejemplo, created_by).
func UserRawToEntity(raw *authraws.UserFoundRaw) *authentities.UserAuth {
	return &authentities.UserAuth{
		ID:        raw.ID,
		CreatedAt: raw.CreatedAt,
		UserName:  raw.UserName,
		Password:  raw.Password,

		AccountValid: raw.AccountValid,
		Role:         raw.RoleID,

		SessionID:        raw.SessionID,
		RefreshToken:     raw.RefreshToken,
		RefreshCode:      raw.RefreshCode,
		SessionCreatedAt: raw.SessionCreatedAt,
	}
}
