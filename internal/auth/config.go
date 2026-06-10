package auth

import (
	"errors"
	"os"
	"strings"
	"time"
)

const localDevelopmentSecret = "local-development-jwt-secret-change-me"

type Config struct {
	JWTSecret  string
	AccessTTL  time.Duration
	RefreshTTL time.Duration
	IsProd     bool
}

func LoadConfig() (Config, error) {
	isProd := isProduction()
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		if isProd {
			return Config{}, errors.New("JWT_SECRET is required in production")
		}
		secret = localDevelopmentSecret
	}

	return Config{
		JWTSecret:  secret,
		AccessTTL:   5 * time.Minute,
		RefreshTTL: 7 * 24 * time.Hour,
		IsProd:     isProd,
	}, nil
}

func isProduction() bool {
	appEnv := strings.ToLower(os.Getenv("APP_ENV"))
	railwayEnv := strings.ToLower(os.Getenv("RAILWAY_ENVIRONMENT"))
	return appEnv == "production" || railwayEnv == "production"
}
