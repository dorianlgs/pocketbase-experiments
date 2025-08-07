package main

import (
	"net/http"

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
		return e.BadRequestError("can't get user email", err)
	}

	user, err := h.auth.GetDatastore().GetOrCreateUser(email)
	if err != nil {
		return e.BadRequestError("can't get or create user", err)
	}

	options, session, err := h.auth.GetWebAuthn().BeginRegistration(user)
	if err != nil {
		return e.BadRequestError("can't begin registration", err)
	}

	sessionID, err := h.auth.GetDatastore().GenSessionID()
	if err != nil {
		return e.BadRequestError("can't generate session id", err)
	}

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
		return e.BadRequestError("Session-Key header required", nil)
	}

	session, ok := h.auth.GetDatastore().GetSession(sessionID)
	if !ok {
		return e.BadRequestError("invalid session", nil)
	}

	user, err := h.auth.GetDatastore().GetOrCreateUser(session.Email)
	if err != nil {
		return e.BadRequestError("can't get user", err)
	}

	var ccr CredentialCreationResponse
	if err := e.BindBody(&ccr); err != nil {
		return e.BadRequestError("Failed to read request data", err)
	}

	credential, err := h.auth.GetWebAuthn().FinishRegistration(user, session.SessionData, e.Request)
	if err != nil {
		h.clearSessionCookie(e, sessionID)
		return e.BadRequestError("can't finish registration", err)
	}

	if err := user.AddCredential(credential, session.Email); err != nil {
		return e.BadRequestError("Failed to add credential", err)
	}

	h.auth.GetDatastore().DeleteSession(sessionID)
	h.clearSessionCookie(e, sessionID)

	return e.JSON(http.StatusOK, "Registration Success")
}

// HandleLoginStart begins WebAuthn authentication
func (h *WebAuthnHandlers) HandleLoginStart(e *core.RequestEvent) error {
	email, err := getEmail(e)
	if err != nil {
		return e.BadRequestError("can't get user email", err)
	}

	user, err := h.auth.GetDatastore().GetOrCreateUser(email)
	if err != nil {
		return e.BadRequestError("can't get user", err)
	}

	options, session, err := h.auth.GetWebAuthn().BeginLogin(user)
	if err != nil {
		h.auth.GetLogger().Printf("[ERROR] can't begin login: %s", err.Error())
		return e.BadRequestError("can't begin login", err)
	}

	sessionID, err := h.auth.GetDatastore().GenSessionID()
	if err != nil {
		return e.BadRequestError("can't generate session id", err)
	}

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
		return e.BadRequestError("Login-Key header required", nil)
	}

	session, ok := h.auth.GetDatastore().GetSession(sessionID)
	if !ok {
		return e.BadRequestError("invalid session", nil)
	}

	user, err := h.auth.GetDatastore().GetOrCreateUser(session.Email)
	if err != nil {
		return e.BadRequestError("can't get user", err)
	}

	var ccr CredentialCreationResponse
	if err := e.BindBody(&ccr); err != nil {
		return e.BadRequestError("Failed to read request data", err)
	}

	credential, err := h.auth.GetWebAuthn().FinishLogin(user, session.SessionData, e.Request)
	if err != nil {
		return e.BadRequestError("can't finish login", err)
	}

	// Handle credential.Authenticator.CloneWarning
	if credential.Authenticator.CloneWarning {
		h.auth.GetLogger().Printf("[WARN] CloneWarning detected during login")
	}

	user.UpdateCredential(credential)

	userRecord, err := h.app.FindFirstRecordByData("users", "email", session.Email)
	if err != nil {
		return e.BadRequestError("User not found", err)
	}

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