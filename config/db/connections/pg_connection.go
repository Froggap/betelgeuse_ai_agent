package dbconnection

import (
	"context"
	"fmt"
	"log"
	enviromentsloader "scrapper-ai/config/enviroments"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
)

func LoadConnectionDB() *pgxpool.Pool {
	enviromentsloader.LoadEnvs()

	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s&search_path=dw_lubrisur",
		enviromentsloader.Envs.DbUser,
		enviromentsloader.Envs.DbPassword,
		enviromentsloader.Envs.DbHost,
		enviromentsloader.Envs.DbPort,
		enviromentsloader.Envs.DbName,
		enviromentsloader.Envs.DbSSLMode,
	)

	dbpool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		log.Fatalf("Error creando pool de conexión: %v\n", err)
	}

	if err := dbpool.Ping(context.Background()); err != nil {
		log.Fatalf("No se pudo hacer ping a la BD: %v\n", err)
	}

	fmt.Println("Conexión exitosa con pgx/v5")
	return dbpool
}

func RunMigrations() {
	env := enviromentsloader.Envs

	migrationURL := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		env.DbUser, env.DbPassword, env.DbHost, env.DbPort, env.DbName, env.DbSSLMode,
	)

	m, err := migrate.New(
		"file://config/db/migrations",
		migrationURL,
	)
	if err != nil {
		log.Fatalf("Error inicializando migrate: %v\n", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("Error ejecutando migraciones: %v\n", err)
	}

	fmt.Println("Migraciones aplicadas correctamente")
}
