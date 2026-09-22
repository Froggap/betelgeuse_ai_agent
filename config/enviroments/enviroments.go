package enviromentsloader

import (
	"os"

	"github.com/joho/godotenv"
)

type ConfigEnvs struct {
	DbHost     string
	DbUser     string
	DbPassword string
	DbPort     string
	DbName     string
	DbSSLMode  string

	SecretJwt string
	Enviroment	string

}

var Envs *ConfigEnvs

func LoadEnvs() {
	_ = godotenv.Load()

	Envs = &ConfigEnvs{
		DbHost:     getEnv("DB_HOST", "localhost"),
		DbUser:     getEnv("DB_USER", "postgres"),
		DbPassword: getEnv("DB_PASSWORD", ""),
		DbName:     getEnv("DB_NAME", "postgres"),
		DbPort:     getEnv("DB_PORT", "5432"),
		DbSSLMode:  getEnv("DB_SSLMODE", "disable"),
		SecretJwt: getEnv("JWT_SECRET", "default_secret_key"),
		Enviroment: getEnv("ENVIROMENT", "development"),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
