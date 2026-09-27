package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	dbconnection "scrapper-ai/config/db/connections"
	enviromentsloader "scrapper-ai/config/enviroments"
	"scrapper-ai/modules/auth"
	shareddtovalidators "scrapper-ai/modules/shared/infrastructure/dtovalidator"
	commonoutmappers "scrapper-ai/modules/shared/infrastructure/mappers"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/cors"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "healthcheck" {
		os.Exit(runHealthcheck())
	}

	enviromentsloader.LoadEnvs()
	if enviromentsloader.Envs.Enviroment == "production" &&
		(enviromentsloader.Envs.SecretJwt == "" || enviromentsloader.Envs.SecretJwt == "default_secret_key") {
		log.Fatal("JWT_SECRET es obligatorio (y no puede ser el default) cuando ENVIROMENT=production")
	}

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

	// ─── instancias globales compartidas ───
	dtoValidator := shareddtovalidators.NewDTOValidator()

	// ─── rutas raíz ───
	r.Get("/health", func(w http.ResponseWriter, req *http.Request) {
		sqlDB, err := db.DB()
		if err == nil {
			err = sqlDB.PingContext(req.Context())
		}
		if err != nil {
			commonoutmappers.RespondWithJSON(w, http.StatusServiceUnavailable, map[string]string{
				"status": "unavailable",
			})
			return
		}
		commonoutmappers.RespondWithJSON(w, http.StatusOK, map[string]string{
			"status": "ok",
		})
	})

	// ─── módulos ───
	r.Mount(
		"/auth",
		auth.AuthBootstrap(
			db,
			dtoValidator,
		),
	)

	srv := &http.Server{Addr: ":3000", Handler: r}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()
	log.Println("Server running on :3000")

	// Graceful shutdown: SIGTERM (docker stop / compose down) drena las
	// peticiones en curso antes de salir.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Error en el shutdown: %v", err)
	}
}

// runHealthcheck se ejecuta como `server healthcheck` (lo usa el HEALTHCHECK
// del Dockerfile; scratch no tiene shell ni curl).
func runHealthcheck() int {
	client := http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get("http://127.0.0.1:3000/health")
	if err != nil {
		return 1
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return 1
	}
	return 0
}

func allowedCORSOrigins(defaults []string) []string {
	if raw := os.Getenv("CORS_ORIGINS"); raw != "" {
		return strings.Split(raw, ",")
	}
	return defaults
}
