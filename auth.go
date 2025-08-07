package main

import (
	"fmt"

	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/pocketbase/pocketbase/core"
)

// AuthService handles authentication setup and operations
type AuthService struct {
	webAuthn  *webauthn.WebAuthn
	datastore PasskeyStore
	logger    Logger
	config    *AppConfig
}

// NewAuthService creates a new authentication service
func NewAuthService(config *AppConfig, logger Logger) (*AuthService, error) {
	// Configure WebAuthn
	wconfig := &webauthn.Config{
		RPDisplayName: "PB Experiments WebAuthn",
		RPID:          config.Host,
		RPOrigins:     []string{config.Origin},
	}

	webAuthn, err := webauthn.New(wconfig)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize WebAuthn: %w", err)
	}

	return &AuthService{
		webAuthn: webAuthn,
		logger:   logger,
		config:   config,
	}, nil
}

// SetDatastore sets the datastore for the auth service
func (a *AuthService) SetDatastore(datastore PasskeyStore) {
	a.datastore = datastore
}

// GetWebAuthn returns the WebAuthn instance
func (a *AuthService) GetWebAuthn() *webauthn.WebAuthn {
	return a.webAuthn
}

// GetDatastore returns the datastore instance
func (a *AuthService) GetDatastore() PasskeyStore {
	return a.datastore
}

// GetLogger returns the logger instance
func (a *AuthService) GetLogger() Logger {
	return a.logger
}

// GetTOTPIssuer returns the TOTP issuer from config
func (a *AuthService) GetTOTPIssuer() string {
	return a.config.TOTPIssuer
}

// getEmail extracts email from request body
func getEmail(e *core.RequestEvent) (string, error) {
	type User struct {
		Email string `json:"email"`
	}

	var u User
	if err := e.BindBody(&u); err != nil {
		return "", fmt.Errorf("failed to read request data: %w", err)
	}

	return u.Email, nil
}