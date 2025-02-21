package main

import (
	"bytes"
	"encoding/base32"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image/png"
	"log"
	"net/http"
	"os"
	"reflect"
	"strconv"

	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/shujink0/pocketbase-experiments/ui"

	"github.com/pquerna/otp/totp"

	"github.com/joho/godotenv"
)

var (
	webAuthn *webauthn.WebAuthn
	err      error

	datastore PasskeyStore
	//sessions  SessionStore
	l Logger
)

func main() {

	err := godotenv.Load()
	if err != nil {
		panic("Error loading .env file")
	}

	totpIssuer := os.Getenv("TOTP_ISSUER")

	if totpIssuer == "" {
		panic("env TOTP_ISSUER not found")
	}

	app := pocketbase.New()

	var indexFallback bool
	app.RootCmd.PersistentFlags().BoolVar(
		&indexFallback,
		"indexFallback",
		true,
		"fallback the request to index.html on missing static path, e.g. when pretty urls are used with SPA",
	)

	proto := os.Getenv("PROTO")
	host := os.Getenv("HOST")
	port := os.Getenv("PORT")
	origin := fmt.Sprintf("%s://%s%s", proto, host, port)

	wconfig := &webauthn.Config{
		RPDisplayName: "PB Expetiments WebAuthn", // Display Name for your site
		RPID:          host,                      // Generally the FQDN for your site
		RPOrigins:     []string{origin},          // The origin URLs allowed for WebAuthn
	}

	if webAuthn, err = webauthn.New(wconfig); err != nil {
		fmt.Printf("[FATA] %s", err.Error())
		os.Exit(1)
	}

	l = log.Default()

	app.OnServe().BindFunc(func(se *core.ServeEvent) error {

		datastore = NewInMem(l, app)

		if !se.Router.HasRoute(http.MethodGet, "/{path...}") {
			se.Router.GET("/{path...}", apis.Static(ui.DistDirFS, indexFallback)).
				Bind(apis.Gzip())
		}

		se.Router.GET("/api/pb-experiments/get-qr", func(e *core.RequestEvent) error {

			info, err := e.RequestInfo()
			userId := info.Query["userId"]

			str_regenerate := info.Query["regenerate"]

			regenerate, err := strconv.ParseBool(str_regenerate)
			if err != nil {
				return e.BadRequestError("regenerate not bool", nil)
			}

			if userId == "" {
				return e.BadRequestError("userId required", nil)
			}

			record, err := app.FindRecordById("users", userId)
			if err != nil {
				return err
			}

			canAccess, err := e.App.CanAccessRecord(record, info, record.Collection().ViewRule)
			if !canAccess {
				return e.ForbiddenError("", err)
			}

			opts := totp.GenerateOpts{
				Issuer:      totpIssuer,
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

				err = app.Save(record)
				if err != nil {
					return e.BadRequestError("Error saving otp secret", err)
				}
			}

			var buf bytes.Buffer
			img, err := key.Image(200, 200)
			if err != nil {
				return e.BadRequestError("Error generating otp image", err)
			}

			err = png.Encode(&buf, img)

			if err != nil {
				return e.BadRequestError("Error encoding image", err)
			}

			return e.Blob(http.StatusOK, "image/png", buf.Bytes())
		}).Bind(apis.RequireAuth())

		se.Router.POST("/api/pb-experiments/totp-login", func(e *core.RequestEvent) error {
			data := UserTotp{}
			if err := e.BindBody(&data); err != nil {
				return e.BadRequestError("Failed to read request data", err)
			}

			if data.MfaId == "" {
				return e.BadRequestError("mfaId required", nil)
			}

			if data.Passcode == "" {
				return e.BadRequestError("passcode required", nil)
			}

			record, err := app.FindRecordById("_mfas", data.MfaId)
			if err != nil {
				return e.BadRequestError("Mfa not found", err)
			}

			userId := record.GetString("recordRef")

			if userId == "" {
				return e.BadRequestError("Column not found", err)
			}

			userRecord, err := app.FindRecordById("users", userId)
			if err != nil {
				return e.BadRequestError("Mfa not found", err)
			}

			secret := userRecord.GetString("totpSecret")

			if secret == "" {
				return e.BadRequestError("Secret not found", err)
			}

			valid := totp.Validate(data.Passcode, secret)
			if !valid {
				return e.UnauthorizedError("Invalid passcode", err)
			}

			return apis.RecordAuthResponse(e, userRecord, "totp", nil)
		})

		se.Router.POST("/api/pb-experiments/passkey/registerStart", func(e *core.RequestEvent) error {

			app.Logger().Info("begin registration ----------------------\\")

			email, err := getEmail(e)
			if err != nil {
				return e.BadRequestError("[ERRO] can't get user name: %s", err.Error())
			}

			user, err := datastore.GetOrCreateUser(email)
			if err != nil {
				return e.BadRequestError("[ERRO] can't get user name: %s", err.Error())
			}

			options, session, err := webAuthn.BeginRegistration(user)
			if err != nil {
				return e.BadRequestError("can't begin registration: %s", err.Error())
			}

			t, err := datastore.GenSessionID()
			if err != nil {
				return e.BadRequestError("[ERRO] can't generate session id: %s", err.Error())
			}

			datastore.SaveSession(t, LocalSession{
				SessionData: *session,
				Email:       email,
			})

			e.Response.Header().Set("Session-Key", t)

			return e.JSON(http.StatusOK, options)
		})

		se.Router.POST("/api/pb-experiments/passkey/registerFinish", func(e *core.RequestEvent) error {

			sid := e.Request.Header.Get("Session-Key")
			if err != nil {
				return e.BadRequestError("[ERRO] can't get session id: %s", err.Error())
			}

			app.Logger().Info("sid %s ----------------------/", "sid", sid)

			session, ok := datastore.GetSession(sid)

			if !ok {
				return e.BadRequestError("[ERRO] can't get session id: %s from datastore", err.Error())
			}

			user, err := datastore.GetOrCreateUser(session.Email)
			if err != nil {
				return e.BadRequestError("[ERRO] can't get user id: %s", err.Error())
			}

			app.Logger().Info("user %s ----------------------/", "user", user.WebAuthnDisplayName())

			var ccr CredentialCreationResponse

			if err := e.BindBody(&ccr); err != nil {
				return e.BadRequestError("Failed to read request data", err)
			}

			app.Logger().Info("rawId %s ----------------------/", "rawId", ccr.PublicKeyCredential.RawID)

			credential, err := webAuthn.FinishRegistration(user, session.SessionData, e.Request)
			if err != nil {
				msg := fmt.Sprintf("can't finish registration: %s", err.Error())

				e.SetCookie(&http.Cookie{
					Name:  "sid",
					Value: "",
				})
				return e.BadRequestError(msg, err.Error())
			}

			err = user.AddCredential(credential, session.Email)
			if err != nil {
				return e.BadRequestError("Failed to add credential", err)
			}

			datastore.DeleteSession(sid)
			e.SetCookie(&http.Cookie{
				Name:  "sid",
				Value: "",
			})

			app.Logger().Info("finish registration ----------------------/")
			return e.JSON(http.StatusOK, "Registration Success")
		})

		se.Router.POST("/api/pb-experiments/passkey/loginStart", func(e *core.RequestEvent) error {
			l.Printf("[INFO] begin login ----------------------\\")

			email, err := getEmail(e)
			if err != nil {
				msg := fmt.Sprintf("[ERRO]can't get user name: %s", err.Error())
				return e.BadRequestError(msg, err.Error())
			}

			user, err := datastore.GetOrCreateUser(email) // Find the user

			if err != nil {
				msg := fmt.Sprintf("[ERRO]can't get user name: %s", err.Error())
				return e.BadRequestError(msg, err.Error())
			}

			user.WebAuthnCredentials()

			options, session, err := webAuthn.BeginLogin(user)
			if err != nil {
				msg := fmt.Sprintf("can't begin login: %s", err.Error())
				l.Printf("[ERRO] %s", msg)
				return e.BadRequestError(msg, err.Error())
			}

			// Make a session key and store the sessionData values
			t, err := datastore.GenSessionID()
			if err != nil {
				return e.BadRequestError("[ERRO] can't generate session id: %s", err.Error())
			}
			datastore.SaveSession(t, LocalSession{
				SessionData: *session,
				Email:       email,
			})

			e.Response.Header().Set("Login-Key", t)
			return e.JSON(http.StatusOK, options)
		})

		se.Router.POST("/api/pb-experiments/passkey/loginFinish", func(e *core.RequestEvent) error {
			// Get the session key from cookie
			sid := e.Request.Header.Get("Login-Key")
			if err != nil {
				return e.BadRequestError("[ERRO] can't get session id: %s", err.Error())
			}
			// Get the session data stored from the function above
			session, err := datastore.GetSession(sid) // FIXME: cover invalid session
			if !err {
				return e.BadRequestError("[ERRO] can't get session id: %s", sid)
			}

			// In out example username == userID, but in real world it should be different
			user, err2 := datastore.GetOrCreateUser(session.Email) // Get the user
			if err2 != nil {
				return e.BadRequestError("[ERRO] can't get user: %s", err2.Error())
			}

			var ccr CredentialCreationResponse

			if err := e.BindBody(&ccr); err != nil {
				return e.BadRequestError("Failed to read request data", err)
			}

			app.Logger().Info("rawId %s ----------------------/", "rawId", ccr.PublicKeyCredential.RawID)

			credential, errLogin := webAuthn.FinishLogin(user, session.SessionData, e.Request)
			if errLogin != nil {
				return e.BadRequestError("[ERRO] can't finish login: %s", errLogin.Error())
			}

			// Handle credential.Authenticator.CloneWarning
			if credential.Authenticator.CloneWarning {
				l.Printf("[WARN] can't finish login: %s", "CloneWarning")
			}

			user.UpdateCredential(credential)

			userRecord, recordErr := app.FindFirstRecordByData("users", "email", session.Email)
			if recordErr != nil {
				return e.BadRequestError("Mfa not found", err)
			}

			datastore.DeleteSession(sid)

			l.Printf("[INFO] finish login ----------------------/")

			return apis.RecordAuthResponse(e, userRecord, "passkeys", nil)
		})

		return se.Next()
	})

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}

type UserTotp struct {
	MfaId    string `json:"mfaId" form:"mfaId"`
	Passcode string `json:"passcode" form:"passcode"`
}

type LocalSession struct {
	SessionData webauthn.SessionData
	Email       string
}

type Logger interface {
	Printf(format string, v ...any)
}

type PasskeyUser interface {
	webauthn.User
	AddCredential(*webauthn.Credential, string) error
	UpdateCredential(*webauthn.Credential) error
}

type PasskeyStore interface {
	GetOrCreateUser(email string) (PasskeyUser, error)
	GenSessionID() (string, error)
	GetSession(token string) (LocalSession, bool)
	SaveSession(token string, data LocalSession)
	DeleteSession(token string)
}

// JSONResponse is a helper function to send json response
func JSONResponse(w http.ResponseWriter, data any, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

// getEmail is a helper function to extract the email from json request
func getEmail(e *core.RequestEvent) (string, error) {
	type User struct {
		Email string `json:"email"`
	}

	var u User
	if err := e.BindBody(&u); err != nil {
		return "", e.BadRequestError("Failed to read request data", err)
	}

	return u.Email, nil

}

type CredentialCreationResponse struct {
	PublicKeyCredential
}

type PublicKeyCredential struct {
	RawID URLEncodedBase64 `json:"rawId"`
}

type URLEncodedBase64 []byte

func (e URLEncodedBase64) String() string {
	return base64.RawURLEncoding.EncodeToString(e)
}

// UnmarshalJSON base64 decodes a URL-encoded value, storing the result in the
// provided byte slice.
func (e *URLEncodedBase64) UnmarshalJSON(data []byte) error {
	if bytes.Equal(data, []byte("null")) {
		return nil
	}

	// TODO: Investigate this line. It is commented as trimming the leading spaces but appears to trim the leading and trailing double quotes instead.
	// Trim the leading spaces.
	data = bytes.Trim(data, "\"")

	// Trim the trailing equal characters.
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

// MarshalJSON base64 encodes a non URL-encoded value, storing the result in the
// provided byte slice.
func (e URLEncodedBase64) MarshalJSON() ([]byte, error) {
	if e == nil {
		return []byte("null"), nil
	}

	return []byte(`"` + base64.RawURLEncoding.EncodeToString(e) + `"`), nil
}
