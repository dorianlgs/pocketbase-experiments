package main

import (
	"encoding/base64"
	"bytes"
	"reflect"

	"github.com/go-webauthn/webauthn/webauthn"
)

// UserTotp represents TOTP login request
type UserTotp struct {
	MfaId    string `json:"mfaId" form:"mfaId"`
	Passcode string `json:"passcode" form:"passcode"`
}

// LocalSession represents a WebAuthn session stored in memory
type LocalSession struct {
	SessionData webauthn.SessionData
	Email       string
}

// Logger interface for logging operations
type Logger interface {
	Printf(format string, v ...any)
}

// PasskeyUser extends webauthn.User with credential management
type PasskeyUser interface {
	webauthn.User
	AddCredential(*webauthn.Credential, string) error
	UpdateCredential(*webauthn.Credential) error
}

// PasskeyStore interface for managing users and sessions
type PasskeyStore interface {
	GetOrCreateUser(email string) (PasskeyUser, error)
	GenSessionID() (string, error)
	GetSession(token string) (LocalSession, bool)
	SaveSession(token string, data LocalSession)
	DeleteSession(token string)
}

// CredentialCreationResponse represents WebAuthn credential response
type CredentialCreationResponse struct {
	PublicKeyCredential
}

// PublicKeyCredential represents a WebAuthn public key credential
type PublicKeyCredential struct {
	RawID URLEncodedBase64 `json:"rawId"`
}

// URLEncodedBase64 handles base64 URL encoding for WebAuthn
type URLEncodedBase64 []byte

func (e URLEncodedBase64) String() string {
	return base64.RawURLEncoding.EncodeToString(e)
}

// UnmarshalJSON decodes base64 URL-encoded value
func (e *URLEncodedBase64) UnmarshalJSON(data []byte) error {
	if bytes.Equal(data, []byte("null")) {
		return nil
	}

	// Trim quotes from JSON string
	data = bytes.Trim(data, "\"")

	// Trim trailing equal characters
	data = bytes.TrimRight(data, "=")

	out := make([]byte, base64.RawURLEncoding.DecodedLen(len(data)))

	n, err := base64.RawURLEncoding.Decode(out, data)
	if err != nil {
		return err
	}

	v := reflect.ValueOf(e).Elem()
	v.SetBytes(out[:n])

	return nil
}

// MarshalJSON encodes value to base64 URL-encoded JSON
func (e URLEncodedBase64) MarshalJSON() ([]byte, error) {
	if e == nil {
		return []byte("null"), nil
	}

	return []byte(`"` + base64.RawURLEncoding.EncodeToString(e) + `"`), nil
}