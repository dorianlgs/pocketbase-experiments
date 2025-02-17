package main

import (
	"bytes"
	"image/png"
	"log"
	"net/http"
	"os"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/shujink0/pocketbase-experiments/ui"

	"github.com/pquerna/otp/totp"

	"github.com/joho/godotenv"
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

	app.OnServe().BindFunc(func(se *core.ServeEvent) error {

		if !se.Router.HasRoute(http.MethodGet, "/{path...}") {
			se.Router.GET("/{path...}", apis.Static(ui.DistDirFS, indexFallback)).
				Bind(apis.Gzip())
		}

		se.Router.GET("/api/pb-experiments/get-qr", func(e *core.RequestEvent) error {

			info, err := e.RequestInfo()
			userId := info.Query["userId"]

			if userId == "" {
				return e.BadRequestError("userId required", nil)
			}

			record, err := app.FindRecordById("users", userId)
			if err != nil {
				return err
			}

			key, err := totp.Generate(totp.GenerateOpts{
				Issuer:      totpIssuer,
				AccountName: record.Email(),
			})
			if err != nil {
				return e.BadRequestError("Error generating otp", err)
			}

			record.Set("totpSecret", key.Secret())

			err = app.Save(record)
			if err != nil {
				return e.BadRequestError("Error saving otp secret", err)
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
		})

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
