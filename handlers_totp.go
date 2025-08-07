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
		return e.BadRequestError("Failed to get request info", err)
	}

	userId := info.Query["userId"]
	if userId == "" {
		return e.BadRequestError("userId required", nil)
	}

	strRegenerate := info.Query["regenerate"]
	regenerate, err := strconv.ParseBool(strRegenerate)
	if err != nil {
		return e.BadRequestError("regenerate not bool", err)
	}

	record, err := h.app.FindRecordById("users", userId)
	if err != nil {
		return e.BadRequestError("User not found", err)
	}

	canAccess, err := e.App.CanAccessRecord(record, info, record.Collection().ViewRule)
	if !canAccess {
		return e.ForbiddenError("Access denied", err)
	}

	opts := totp.GenerateOpts{
		Issuer:      h.auth.GetTOTPIssuer(),
		AccountName: record.Email(),
	}

	if !regenerate {
		secretBytes, err := base32.StdEncoding.DecodeString(record.GetString("totpSecret"))
		if err != nil {
			return e.BadRequestError("Error decoding totp secret", err)
		}
		opts.Secret = secretBytes
	}

	key, err := totp.Generate(opts)
	if err != nil {
		return e.BadRequestError("Error generating otp", err)
	}

	if regenerate {
		record.Set("totpSecret", key.Secret())
		record.Set("multiFactorAuth", true)

		if err := h.app.Save(record); err != nil {
			return e.BadRequestError("Error saving otp secret", err)
		}
	}

	var buf bytes.Buffer
	img, err := key.Image(200, 200)
	if err != nil {
		return e.BadRequestError("Error generating otp image", err)
	}

	if err := png.Encode(&buf, img); err != nil {
		return e.BadRequestError("Error encoding image", err)
	}

	return e.Blob(http.StatusOK, "image/png", buf.Bytes())
}

// HandleTOTPLogin validates TOTP passcode and logs in user
func (h *TOTPHandlers) HandleTOTPLogin(e *core.RequestEvent) error {
	var data UserTotp
	if err := e.BindBody(&data); err != nil {
		return e.BadRequestError("Failed to read request data", err)
	}

	if data.MfaId == "" {
		return e.BadRequestError("mfaId required", nil)
	}

	if data.Passcode == "" {
		return e.BadRequestError("passcode required", nil)
	}

	record, err := h.app.FindRecordById("_mfas", data.MfaId)
	if err != nil {
		return e.BadRequestError("Mfa not found", err)
	}

	userId := record.GetString("recordRef")
	if userId == "" {
		return e.BadRequestError("Column not found", nil)
	}

	userRecord, err := h.app.FindRecordById("users", userId)
	if err != nil {
		return e.BadRequestError("User not found", err)
	}

	secret := userRecord.GetString("totpSecret")
	if secret == "" {
		return e.BadRequestError("Secret not found", nil)
	}

	if !totp.Validate(data.Passcode, secret) {
		return e.UnauthorizedError("Invalid passcode", nil)
	}

	return apis.RecordAuthResponse(e, userRecord, "totp", nil)
}