package main

import (
	"net/http"
	"strings"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

// WebAuthnHandlers contains WebAuthn-related HTTP handlers
type WebAuthnHandlers struct {
	app  *pocketbase.PocketBase
	auth *AuthService
}

// NewWebAuthnHandlers creates new WebAuthn handlers
func NewWebAuthnHandlers(app *pocketbase.PocketBase, auth *AuthService) *WebAuthnHandlers {
	return &WebAuthnHandlers{
		app:  app,
		auth: auth,
	}
}

// HandleRegisterStart begins WebAuthn registration
func (h *WebAuthnHandlers) HandleRegisterStart(e *core.RequestEvent) error {
	email, err := getEmail(e)
	if err != nil {
		h.auth.GetLogger().Printf("[WARN] WebAuthn Register: Invalid email in request: %v", err)
		return e.BadRequestError("Invalid email address", nil)
	}

	// Basic email validation
	if len(email) < 3 || !strings.Contains(email, "@") {
		h.auth.GetLogger().Printf("[WARN] WebAuthn Register: Invalid email format: %s", email)
		return e.BadRequestError("Valid email address is required", nil)
	}

	user, err := h.auth.GetDatastore().GetOrCreateUser(email)
	if err != nil {
		h.auth.GetLogger().Printf("[ERROR] WebAuthn Register: Failed to get/create user for email: %s, error: %v", email, err)
		return e.InternalServerError("Failed to process user account", nil)
	}

	options, session, err := h.auth.GetWebAuthn().BeginRegistration(user)
	if err != nil {
		h.auth.GetLogger().Printf("[ERROR] WebAuthn Register: Failed to begin registration for email: %s, error: %v", email, err)
		return e.InternalServerError("Failed to initialize registration", nil)
	}

	sessionID, err := h.auth.GetDatastore().GenSessionID()
	if err != nil {
		h.auth.GetLogger().Printf("[ERROR] WebAuthn Register: Failed to generate session ID for email: %s, error: %v", email, err)
		return e.InternalServerError("Failed to create registration session", nil)
	}

	h.auth.GetLogger().Printf("[INFO] WebAuthn Register: Started registration for email: %s", email)

	h.auth.GetDatastore().SaveSession(sessionID, LocalSession{
		SessionData: *session,
		Email:       email,
	})

	e.Response.Header().Set("Session-Key", sessionID)

	return e.JSON(http.StatusOK, options)
}

// HandleRegisterFinish completes WebAuthn registration
func (h *WebAuthnHandlers) HandleRegisterFinish(e *core.RequestEvent) error {
	sessionID := e.Request.Header.Get("Session-Key")
	if sessionID == "" {
		h.auth.GetLogger().Printf("[WARN] WebAuthn Register Finish: Missing Session-Key header")
		return e.BadRequestError("Session-Key header is required", nil)
	}

	session, ok := h.auth.GetDatastore().GetSession(sessionID)
	if !ok {
		h.auth.GetLogger().Printf("[SECURITY] WebAuthn Register Finish: Invalid or expired session: %s", sessionID)
		return e.UnauthorizedError("Invalid or expired registration session", nil)
	}

	user, err := h.auth.GetDatastore().GetOrCreateUser(session.Email)
	if err != nil {
		h.auth.GetLogger().Printf("[ERROR] WebAuthn Register Finish: Failed to get user for email: %s, error: %v", session.Email, err)
		h.auth.GetDatastore().DeleteSession(sessionID)
		return e.InternalServerError("Failed to process user account", nil)
	}

	var ccr CredentialCreationResponse
	if err := e.BindBody(&ccr); err != nil {
		h.auth.GetLogger().Printf("[WARN] WebAuthn Register Finish: Invalid credential data for email: %s, error: %v", session.Email, err)
		h.auth.GetDatastore().DeleteSession(sessionID)
		return e.BadRequestError("Invalid credential data", nil)
	}

	credential, err := h.auth.GetWebAuthn().FinishRegistration(user, session.SessionData, e.Request)
	if err != nil {
		h.auth.GetLogger().Printf("[SECURITY] WebAuthn Register Finish: Failed to verify credential for email: %s, error: %v", session.Email, err)
		h.clearSessionCookie(e, sessionID)
		h.auth.GetDatastore().DeleteSession(sessionID)
		return e.BadRequestError("Failed to verify credential", nil)
	}

	if err := user.AddCredential(credential, session.Email); err != nil {
		h.auth.GetLogger().Printf("[ERROR] WebAuthn Register Finish: Failed to save credential for email: %s, error: %v", session.Email, err)
		h.auth.GetDatastore().DeleteSession(sessionID)
		return e.InternalServerError("Failed to save credential", nil)
	}

	h.auth.GetLogger().Printf("[INFO] WebAuthn Register: Successfully registered credential for email: %s", session.Email)

	h.auth.GetDatastore().DeleteSession(sessionID)
	h.clearSessionCookie(e, sessionID)

	return e.JSON(http.StatusOK, "Registration Success")
}

// HandleLoginStart begins WebAuthn authentication
func (h *WebAuthnHandlers) HandleLoginStart(e *core.RequestEvent) error {
	email, err := getEmail(e)
	if err != nil {
		h.auth.GetLogger().Printf("[WARN] WebAuthn Login: Invalid email in request: %v", err)
		return e.BadRequestError("Invalid email address", nil)
	}

	// Basic email validation
	if len(email) < 3 || !strings.Contains(email, "@") {
		h.auth.GetLogger().Printf("[WARN] WebAuthn Login: Invalid email format: %s", email)
		return e.BadRequestError("Valid email address is required", nil)
	}

	user, err := h.auth.GetDatastore().GetOrCreateUser(email)
	if err != nil {
		h.auth.GetLogger().Printf("[ERROR] WebAuthn Login: Failed to get user for email: %s, error: %v", email, err)
		return e.UnauthorizedError("Authentication failed", nil)
	}

	options, session, err := h.auth.GetWebAuthn().BeginLogin(user)
	if err != nil {
		h.auth.GetLogger().Printf("[ERROR] WebAuthn Login: Failed to begin login for email: %s, error: %v", email, err)
		return e.UnauthorizedError("Authentication failed", nil)
	}

	sessionID, err := h.auth.GetDatastore().GenSessionID()
	if err != nil {
		h.auth.GetLogger().Printf("[ERROR] WebAuthn Login: Failed to generate session ID for email: %s, error: %v", email, err)
		return e.InternalServerError("Failed to create login session", nil)
	}

	h.auth.GetLogger().Printf("[INFO] WebAuthn Login: Started authentication for email: %s", email)

	h.auth.GetDatastore().SaveSession(sessionID, LocalSession{
		SessionData: *session,
		Email:       email,
	})

	e.Response.Header().Set("Login-Key", sessionID)
	return e.JSON(http.StatusOK, options)
}

// HandleLoginFinish completes WebAuthn authentication
func (h *WebAuthnHandlers) HandleLoginFinish(e *core.RequestEvent) error {
	sessionID := e.Request.Header.Get("Login-Key")
	if sessionID == "" {
		h.auth.GetLogger().Printf("[WARN] WebAuthn Login Finish: Missing Login-Key header")
		return e.BadRequestError("Login-Key header is required", nil)
	}

	session, ok := h.auth.GetDatastore().GetSession(sessionID)
	if !ok {
		h.auth.GetLogger().Printf("[SECURITY] WebAuthn Login Finish: Invalid or expired session: %s", sessionID)
		return e.UnauthorizedError("Invalid or expired login session", nil)
	}

	user, err := h.auth.GetDatastore().GetOrCreateUser(session.Email)
	if err != nil {
		h.auth.GetLogger().Printf("[ERROR] WebAuthn Login Finish: Failed to get user for email: %s, error: %v", session.Email, err)
		h.auth.GetDatastore().DeleteSession(sessionID)
		return e.UnauthorizedError("Authentication failed", nil)
	}

	var ccr CredentialCreationResponse
	if err := e.BindBody(&ccr); err != nil {
		h.auth.GetLogger().Printf("[WARN] WebAuthn Login Finish: Invalid credential data for email: %s, error: %v", session.Email, err)
		h.auth.GetDatastore().DeleteSession(sessionID)
		return e.BadRequestError("Invalid credential data", nil)
	}

	credential, err := h.auth.GetWebAuthn().FinishLogin(user, session.SessionData, e.Request)
	if err != nil {
		h.auth.GetLogger().Printf("[SECURITY] WebAuthn Login Finish: Failed to verify credential for email: %s, error: %v", session.Email, err)
		h.auth.GetDatastore().DeleteSession(sessionID)
		return e.UnauthorizedError("Authentication failed", nil)
	}

	// Handle credential.Authenticator.CloneWarning
	if credential.Authenticator.CloneWarning {
		h.auth.GetLogger().Printf("[WARN] CloneWarning detected during login")
	}

	user.UpdateCredential(credential)

	userRecord, err := h.app.FindFirstRecordByData("users", "email", session.Email)
	if err != nil {
		h.auth.GetLogger().Printf("[ERROR] WebAuthn Login Finish: User record not found for email: %s, error: %v", session.Email, err)
		h.auth.GetDatastore().DeleteSession(sessionID)
		return e.UnauthorizedError("Authentication failed", nil)
	}

	h.auth.GetLogger().Printf("[INFO] WebAuthn Login: Successful authentication for email: %s", session.Email)
	h.auth.GetDatastore().DeleteSession(sessionID)

	return apis.RecordAuthResponse(e, userRecord, "passkeys", nil)
}

// clearSessionCookie clears the session cookie
func (h *WebAuthnHandlers) clearSessionCookie(e *core.RequestEvent, sessionID string) {
	e.SetCookie(&http.Cookie{
		Name:  "sid",
		Value: "",
	})
}