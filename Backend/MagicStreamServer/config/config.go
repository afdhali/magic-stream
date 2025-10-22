package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Port     string
	MongoURI string
	GinMode  string
	DatabaseName string
	BackendServerURI string
	JWTAccessSecret      string
	JWTRefreshSecret     string
	AccessTokenExpireMin int
	RefreshTokenExpireHr int
}

func LoadConfig() *Config {
	// Load env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	accessExp, _ := strconv.Atoi(getEnv("ACCESS_TOKEN_EXPIRE_MINUTES", "15"))
	refreshExp, _ := strconv.Atoi(getEnv("REFRESH_TOKEN_EXPIRE_HOURS", "168"))

	return &Config{
		Port: getEnv("PORT","5000"),
		MongoURI: getEnv("MONGO_URI",""),
		GinMode: getEnv("GIN_MODE","debug"),
		DatabaseName: getEnv("DATABASE_NAME",""),
		BackendServerURI: getEnv("BACKEND_URI",""),
		JWTAccessSecret: getEnv("JWT_ACCESS_SECRET",""),
		JWTRefreshSecret: getEnv("JWT_REFRESH_SECRET",""),
		AccessTokenExpireMin: accessExp,
		RefreshTokenExpireHr: refreshExp,
	}
}

func getEnv(key, defaulValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaulValue
	}
	return  value
}