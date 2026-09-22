package main

import (
	"log"
	"net/http"
	"os"
	dbconnection "scrapper-ai/config/db/connections"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/cors"
)

func main() {
	db := dbconnection.LoadConnectionDB()
	dbconnection.RunMigrations()

	r := chi.NewRouter()

	defaultOrigins := []string{"http://localhost:34115", "http://localhost:5173", "http://127.0.0.1:5173"}
	corsHandler := cors.New(cors.Options{
		AllowedOrigins:   allowedCORSOrigins(defaultOrigins),
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
	    AllowedHeaders:   []string{"Content-Type", "Authorization", "X-Refresh-Token", "X-Refresh-Code", "X-CSRF-Token"},
	    ExposedHeaders:   []string{"X-Refresh-Token", "X-Refresh-Code"},
	    AllowCredentials: true,
	})
	r.Use(corsHandler.Handler)
	r.Use(middleware.Logger)

	jwt := authinadapters.NewJwtAdapterService()
	dtoValidator := dtovalidators.NewDTOValidator()
	authmiddleware := authmiddlewares.NewAuthMiddleware(jwt, db)

	r.Mount("/auth",
		auth.AuthBootstrap(
			db,
			dtoValidator,
			authmiddleware,
		),
	)

	r.Mount("/docparser",
		docparser.DocParserBootStrap(
			db,
			dtoValidator,
			authmiddleware,
		),
	)

	r.Mount("/businessconfigurations",
		businessconfigurations.BusinessConfigurationsBootstrap(
			db,
			dtoValidator,
			authmiddleware,
		),
	)

	log.Println("Server running on :3000")
	http.ListenAndServe(":3000", r)
}

func allowedCORSOrigins(defaults []string) []string {
	if raw := os.Getenv("CORS_ORIGINS"); raw != "" {
		return strings.Split(raw, ",")
	}
	return defaults
}
