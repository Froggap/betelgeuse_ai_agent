package dbconnection

import (
	"fmt"
	"log"
	"net/url"
	"path/filepath"
	"strings"

	enviromentsloader "scrapper-ai/config/enviroments"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func LoadConnectionDB() *gorm.DB {
	enviromentsloader.LoadEnvs()
	e := enviromentsloader.Envs

	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s&search_path=dw_lubrisur",
		e.DbUser,
		e.DbPassword,
		e.DbHost,
		e.DbPort,
		e.DbName,
		e.DbSSLMode,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		log.Fatalf("Error creando conexión a la BD: %v\n", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Error obteniendo el pool de conexiones: %v\n", err)
	}

	if err := sqlDB.Ping(); err != nil {
		log.Fatalf("No se pudo hacer ping a la BD: %v\n", err)
	}

	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(10)

	fmt.Println("Conexión exitosa con GORM")
	return db
}

func RunMigrations() {
	enviromentsloader.LoadEnvs()
	e := enviromentsloader.Envs

	dbURL := url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(e.DbUser, e.DbPassword),
		Host:   fmt.Sprintf("%s:%s", e.DbHost, e.DbPort),
		Path:   e.DbName,
	}
	q := dbURL.Query()
	q.Set("sslmode", e.DbSSLMode)
	q.Set("search_path", "dw_lubrisur")
	dbURL.RawQuery = q.Encode()

	migrationsPath, err := filepath.Abs("config/db/migrations")
	if err != nil {
		log.Fatalf("Error resolviendo la ruta de migraciones: %v\n", err)
	}

	filePath := filepath.ToSlash(migrationsPath)
	if !strings.HasPrefix(filePath, "/") {
		filePath = "/" + filePath
	}
	migrationURL := (&url.URL{Scheme: "file", Path: filePath}).String()

	m, err := migrate.New(migrationURL, dbURL.String())
	if err != nil {
		log.Fatalf("Error inicializando migrate: %v\n", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("Error ejecutando migraciones: %v\n", err)
	}

	fmt.Println("Migraciones aplicadas correctamente")
}
