package main

import "testing"

func TestCorsConfigDefaultsInDevelopment(t *testing.T) {
	t.Setenv("APP_ENV", "")
	t.Setenv("RAILWAY_ENVIRONMENT", "")
	t.Setenv("CORS_ALLOWED_ORIGINS", "")

	config, err := corsConfigFromEnv()
	if err != nil {
		t.Fatalf("expected dev config, got error %v", err)
	}
	if config.AllowOrigins != defaultDevOrigins {
		t.Fatalf("expected default origins %q, got %q", defaultDevOrigins, config.AllowOrigins)
	}
}

func TestCorsConfigRequiresOriginsInProduction(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("RAILWAY_ENVIRONMENT", "")
	t.Setenv("CORS_ALLOWED_ORIGINS", "")

	if _, err := corsConfigFromEnv(); err == nil {
		t.Fatal("expected production CORS config error")
	}
}
