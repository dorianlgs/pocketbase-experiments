package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
)

// AppConfig holds application configuration
type AppConfig struct {
	IsDevEnv     bool
	TOTPIssuer   string
	Proto        string
	Host         string
	Port         string
	Origin       string
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*AppConfig, error) {
	config := &AppConfig{}
	
	// Determine if running in development
	config.IsDevEnv = strings.HasPrefix(os.Args[0], os.TempDir())

	// Load environment file
	if err := loadEnvFile(config.IsDevEnv); err != nil {
		return nil, fmt.Errorf("failed to load environment file: %w", err)
	}

	// Load required environment variables
	config.TOTPIssuer = os.Getenv("TOTP_ISSUER")
	if config.TOTPIssuer == "" {
		return nil, fmt.Errorf("env TOTP_ISSUER not found")
	}

	config.Proto = os.Getenv("PROTO")
	config.Host = os.Getenv("HOST")

	// Build origin URL
	if config.IsDevEnv {
		config.Port = os.Getenv("PORT")
		config.Origin = fmt.Sprintf("%s://%s%s", config.Proto, config.Host, config.Port)
	} else {
		config.Origin = fmt.Sprintf("%s://%s", config.Proto, config.Host)
	}

	return config, nil
}

// loadEnvFile loads the appropriate environment file
func loadEnvFile(isDevEnv bool) error {
	if isDevEnv {
		return godotenv.Load()
	}

	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return err
	}
	
	environmentPath := filepath.Join(dir, ".env.production")
	log.Printf("Loading production environment from: %s", environmentPath)
	
	return godotenv.Load(environmentPath)
}