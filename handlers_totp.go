package main

import (
	"bytes"
	"encoding/base32"
	"image/png"
	"net/http"
	"strconv"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pquerna/otp/totp"
)

// TOTPHandlers contains TOTP-related HTTP handlers
type TOTPHandlers struct {
	app     *pocketbase.PocketBase
	auth    *AuthService
}

// NewTOTPHandlers creates new TOTP handlers
func NewTOTPHandlers(app *pocketbase.PocketBase, auth *AuthService) *TOTPHandlers {
	return &TOTPHandlers{
		app:  app,
		auth: auth,
	}
}

// HandleGetQR generates TOTP QR code
func (h *TOTPHandlers) HandleGetQR(e *core.RequestEvent) error {
	info, err := e.RequestInfo()
	if err != nil {
		h.auth.GetLogger().Printf("[ERROR] TOTP QR: Failed to get request info: %v", err)
		return e.BadRequestError("Invalid request", nil)
	}

	userId := info.Query["userId"]
	if userId == "" {
		h.auth.GetLogger().Printf("[WARN] TOTP QR: Missing userId parameter")
		return e.BadRequestError("userId parameter is required", nil)
	}

	// Validate userId format (basic check)
	if len(userId) < 15 { // PocketBase IDs are typically 15 chars
		h.auth.GetLogger().Printf("[WARN] TOTP QR: Invalid userId format: %s", userId)
		return e.BadRequestError("Invalid userId format", nil)
	}

	strRegenerate := info.Query["regenerate"]
	if strRegenerate == "" {
		strRegenerate = "false" // Default to false
	}
	regenerate, err := strconv.ParseBool(strRegenerate)
	if err != nil {
		h.auth.GetLogger().Printf("[WARN] TOTP QR: Invalid regenerate parameter: %s", strRegenerate)
		return e.BadRequestError("regenerate parameter must be true or false", nil)
	}

	record, err := h.app.FindRecordById("users", userId)
	if err != nil {
		h.auth.GetLogger().Printf("[WARN] TOTP QR: User not found for ID: %s", userId)
		return e.NotFoundError("User not found", nil)
	}

	canAccess, err := e.App.CanAccessRecord(record, info, record.Collection().ViewRule)
	if !canAccess {
		h.auth.GetLogger().Printf("[SECURITY] TOTP QR: Access denied for user: %s", userId)
		return e.ForbiddenError("Insufficient permissions to access this resource", nil)
	}

	opts := totp.GenerateOpts{
		Issuer:      h.auth.GetTOTPIssuer(),
		AccountName: record.Email(),
	}

	if !regenerate {
		totpSecret := record.GetString("totpSecret")
		if totpSecret == "" {
			h.auth.GetLogger().Printf("[WARN] TOTP QR: No existing TOTP secret for user: %s", userId)
			return e.BadRequestError("No TOTP configuration found. Please regenerate.", nil)
		}
		secretBytes, err := base32.StdEncoding.DecodeString(totpSecret)
		if err != nil {
			h.auth.GetLogger().Printf("[ERROR] TOTP QR: Invalid TOTP secret format for user: %s", userId)
			return e.InternalServerError("Invalid TOTP configuration", nil)
		}
		opts.Secret = secretBytes
	}

	key, err := totp.Generate(opts)
	if err != nil {
		h.auth.GetLogger().Printf("[ERROR] TOTP QR: Failed to generate TOTP key for user: %s, error: %v", userId, err)
		return e.InternalServerError("Failed to generate TOTP configuration", nil)
	}

	if regenerate {
		record.Set("totpSecret", key.Secret())
		record.Set("multiFactorAuth", true)

		if err := h.app.Save(record); err != nil {
			h.auth.GetLogger().Printf("[ERROR] TOTP QR: Failed to save TOTP secret for user: %s, error: %v", userId, err)
			return e.InternalServerError("Failed to save TOTP configuration", nil)
		}
		h.auth.GetLogger().Printf("[INFO] TOTP QR: Successfully regenerated TOTP secret for user: %s", userId)
	}

	var buf bytes.Buffer
	img, err := key.Image(200, 200)
	if err != nil {
		h.auth.GetLogger().Printf("[ERROR] TOTP QR: Failed to generate QR image for user: %s, error: %v", userId, err)
		return e.InternalServerError("Failed to generate QR code image", nil)
	}

	if err := png.Encode(&buf, img); err != nil {
		h.auth.GetLogger().Printf("[ERROR] TOTP QR: Failed to encode PNG for user: %s, error: %v", userId, err)
		return e.InternalServerError("Failed to encode QR code image", nil)
	}

	return e.Blob(http.StatusOK, "image/png", buf.Bytes())
}

// HandleTOTPLogin validates TOTP passcode and logs in user
func (h *TOTPHandlers) HandleTOTPLogin(e *core.RequestEvent) error {
	var data UserTotp
	if err := e.BindBody(&data); err != nil {
		h.auth.GetLogger().Printf("[WARN] TOTP Login: Invalid request body: %v", err)
		return e.BadRequestError("Invalid request format", nil)
	}

	if data.MfaId == "" {
		h.auth.GetLogger().Printf("[WARN] TOTP Login: Missing mfaId")
		return e.BadRequestError("mfaId is required", nil)
	}

	if data.Passcode == "" {
		h.auth.GetLogger().Printf("[WARN] TOTP Login: Missing passcode for mfaId: %s", data.MfaId)
		return e.BadRequestError("passcode is required", nil)
	}

	// Validate passcode format (6 digits)
	if len(data.Passcode) != 6 {
		h.auth.GetLogger().Printf("[WARN] TOTP Login: Invalid passcode length for mfaId: %s", data.MfaId)
		return e.BadRequestError("passcode must be 6 digits", nil)
	}

	record, err := h.app.FindRecordById("_mfas", data.MfaId)
	if err != nil {
		h.auth.GetLogger().Printf("[SECURITY] TOTP Login: Invalid MFA record: %s", data.MfaId)
		return e.UnauthorizedError("Invalid authentication request", nil)
	}

	userId := record.GetString("recordRef")
	if userId == "" {
		h.auth.GetLogger().Printf("[ERROR] TOTP Login: Missing recordRef in MFA record: %s", data.MfaId)
		return e.InternalServerError("Invalid MFA configuration", nil)
	}

	userRecord, err := h.app.FindRecordById("users", userId)
	if err != nil {
		h.auth.GetLogger().Printf("[ERROR] TOTP Login: User not found for ID: %s", userId)
		return e.UnauthorizedError("Invalid authentication request", nil)
	}

	secret := userRecord.GetString("totpSecret")
	if secret == "" {
		h.auth.GetLogger().Printf("[ERROR] TOTP Login: No TOTP secret configured for user: %s", userId)
		return e.UnauthorizedError("TOTP not configured for this account", nil)
	}

	if !totp.Validate(data.Passcode, secret) {
		h.auth.GetLogger().Printf("[SECURITY] TOTP Login: Invalid passcode attempt for user: %s", userId)
		return e.UnauthorizedError("Invalid TOTP passcode", nil)
	}

	h.auth.GetLogger().Printf("[INFO] TOTP Login: Successful authentication for user: %s", userId)

	return apis.RecordAuthResponse(e, userRecord, "totp", nil)
}