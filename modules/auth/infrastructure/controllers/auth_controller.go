package authcontrollers

import (
	"encoding/json"
	"net/http"

	authrequestsdtos "scrapper-ai/modules/auth/applicastion/dtos/requests"
	authresponsesdtos "scrapper-ai/modules/auth/applicastion/dtos/responses"
	authusecases "scrapper-ai/modules/auth/applicastion/usecases"
	authentities "scrapper-ai/modules/auth/domain/entities"
	authinframappers "scrapper-ai/modules/auth/infrastructure/mappers"
	authmiddlewares "scrapper-ai/modules/auth/infrastructure/middlewares"
	shareddtovalidators "scrapper-ai/modules/shared/infrastructure/dtovalidator"
	commonoutmappers "scrapper-ai/modules/shared/infrastructure/mappers"

	"github.com/go-chi/chi/v5"
)

type AuthController struct {
	v        *shareddtovalidators.DTOValidator
	am       *authmiddlewares.AuthMiddleware
	loginUC  *authusecases.LoginAccountUC
	reloadUC *authusecases.CheckRefreshCodeUC
	regUC    *authusecases.RegisterAccountUC
}

func NewAuthController(
	v *shareddtovalidators.DTOValidator,
	am *authmiddlewares.AuthMiddleware,
	loginUC *authusecases.LoginAccountUC,
	reloadUC *authusecases.CheckRefreshCodeUC,
	regUC *authusecases.RegisterAccountUC,
) *AuthController {
	return &AuthController{
		v:        v,
		am:       am,
		loginUC:  loginUC,
		reloadUC: reloadUC,
		regUC:    regUC,
	}
}

func (ac *AuthController) Login(w http.ResponseWriter, r *http.Request) {
	var dto authrequestsdtos.LoginRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		commonoutmappers.RespondWithError(w, err)
		return
	}

	if err := ac.v.ValidateStruct(&dto); err != nil {
		commonoutmappers.RespondWithError(w, err)
		return
	}

	session, err := ac.loginUC.Execute(r.Context(), &dto)
	if err != nil {
		commonoutmappers.RespondWithError(w, err)
		return
	}

	respondSession(w, http.StatusOK, session)
}

func (ac *AuthController) Register(w http.ResponseWriter, r *http.Request) {
	var dto authrequestsdtos.RegisterRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		commonoutmappers.RespondWithError(w, err)
		return
	}

	if err := ac.v.ValidateStruct(&dto); err != nil {
		commonoutmappers.RespondWithError(w, err)
		return
	}

	session, err := ac.regUC.Execute(r.Context(), &dto)
	if err != nil {
		commonoutmappers.RespondWithError(w, err)
		return
	}

	respondSession(w, http.StatusCreated, session)
}

func (ac *AuthController) ReloadSession(w http.ResponseWriter, r *http.Request) {
	refreshToken := r.Header.Get("X-Refresh-Token")
	refreshCode := r.Header.Get("X-Refresh-Code")

	// Si el refresh token expiró, no llega userID al contexto y el
	// use-case recuperará al usuario a partir del refresh code.
	userID := ""
	if id, ok := authmiddlewares.UserIDFromContext(r.Context()); ok {
		userID = id.String()
	}

	session, err := ac.reloadUC.Execute(r.Context(), userID, refreshToken, refreshCode)
	if err != nil {
		commonoutmappers.RespondWithError(w, err)
		return
	}

	respondSession(w, http.StatusOK, session)
}

func respondSession(w http.ResponseWriter, status int, session *authentities.Session) {
	response := authresponsesdtos.SessionResponseDTO{
		UserData: authresponsesdtos.UserDataDTO{
			ID:        session.UserData.ID,
			CreatedAt: session.UserData.CreatedAt,
			UserName:  session.UserData.UserName,
			Role:      session.UserData.Role,
		},
		SessionData: authresponsesdtos.SessionDataDTO{
			AccessToken: session.SessionData.AccessToken,
			CreatedAt:   session.SessionData.CreatedAt,
			IsValid:     session.SessionData.IsValid,
		},
	}

	authinframappers.AuthHeadersMapper(w, session.SessionData.RefreshToken, session.SessionData.RefreshCode)
	commonoutmappers.RespondWithJSON(w, status, response)
}

func AuthMapRoutes(ac *AuthController) chi.Router {
	r := chi.NewRouter()
	// rutas públicas
	r.Post("/login", ac.Login)
	r.Post("/register", ac.Register)

	// ruta con middleware
	r.Group(func(r chi.Router) {
		r.Use(ac.am.RefreshToken)
		r.Get("/refresh-token", ac.ReloadSession)
	})

	return r
}
