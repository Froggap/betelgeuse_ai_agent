package sharedinframappers

import (
	"encoding/json"
	"log"
	"net/http"
	commondomainerrors "scrapper-ai/modules/shared/domain/errors"
	globalerrors "scrapper-ai/modules/shared/infrastructure/errors"
)

func RespondWithJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(payload)
}

func RespondWithError(w http.ResponseWriter, err error) {
	red := "\033[31m"
	orange := "\033[38;5;208m"
	reset := "\033[0m"

	if appErr, ok := err.(*globalerrors.AppError); ok {
		log.Printf("%s[infrastructure]%s %s%s | %s", red, reset, appErr.Title, reset, appErr.Message)
		if appErr.Err != nil {
			log.Printf("  %scause:%s %v", red, reset, appErr.Err)
		}
		RespondWithJSON(w, appErr.Status, map[string]any{
			"title":   appErr.Title,
			"message": appErr.Message,
			"status":  appErr.Status,
		})
		return
	}

	if domainErr, ok := err.(*commondomainerrors.DomainError); ok {
		appErr := MapDomainError(domainErr)
		log.Printf("%s[domain]%s %s%s | %s", orange, reset, appErr.Title, reset, appErr.Message)
		RespondWithJSON(w, appErr.Status, map[string]any{
			"title":   appErr.Title,
			"message": appErr.Message,
			"status":  appErr.Status,
			"field":   domainErr.Field,
		})
		return
	}

	log.Printf("%s[unexpected]%s %v", red, reset, err)
	RespondWithJSON(w, http.StatusInternalServerError, map[string]any{
		"title":   "Internal Server Error",
		"message": "An unexpected error occurred",
		"status":  http.StatusInternalServerError,
	})
}
