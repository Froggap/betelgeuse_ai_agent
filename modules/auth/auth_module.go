package auth

import (
	authusecases "scrapper-ai/modules/auth/applicastion/usecases"
	authadapters "scrapper-ai/modules/auth/infrastructure/adapters"
	authcontrollers "scrapper-ai/modules/auth/infrastructure/controllers"
	authmiddlewares "scrapper-ai/modules/auth/infrastructure/middlewares"
	shareddtovalidators "scrapper-ai/modules/shared/infrastructure/dtovalidator"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

// AuthBootstrap inyecta todas las dependencias del módulo auth:
// repositorios -> use-cases -> controller.
func AuthBootstrap(
	db *gorm.DB,
	v *shareddtovalidators.DTOValidator,
) chi.Router {
	// adapters / servicios
	authRepository := authadapters.NewAuthRepository(db)
	jwtService := authadapters.NewJwtAdapterService()
	encrypterService := authadapters.NewEncryptorAdapter()

	// use-cases (orquestan adapters y servicios)
	loginUC := authusecases.NewLoginAccountUC(authRepository, encrypterService, jwtService)
	reloadUC := authusecases.NewCheckRefreshCodeUC(authRepository, jwtService)
	registerUC := authusecases.NewRegisterAccountUC(authRepository, encrypterService, jwtService)

	// middleware + controller
	authMiddleware := authmiddlewares.NewAuthMiddleware(jwtService)
	authController := authcontrollers.NewAuthController(
		v,
		authMiddleware,
		loginUC,
		reloadUC,
		registerUC,
	)

	r := chi.NewRouter()
	r.Mount("/", authcontrollers.AuthMapRoutes(authController))
	return r
}
