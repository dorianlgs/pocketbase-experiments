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

	app.OnServe().BindFunc(func(se *core.ServeEvent) error {
		// serves static files from the provided public dir (if exists)
		se.Router.GET("/{path...}", apis.Static(os.DirFS("./pb_public"), false))

		se.Router.GET("/api/pb-experiments/get-qr", func(e *core.RequestEvent) error {

			record, err := app.FindRecordById("users", "a7d90iil825ptia")
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

			if data.UserId == "" {
				return e.BadRequestError("userId required", nil)
			}

			if data.Passcode == "" {
				return e.BadRequestError("passcode required", nil)
			}

			record, err := app.FindRecordById("users", data.UserId)
			if err != nil {
				return e.BadRequestError("User not found", err)
			}

			secret := record.GetString("totpSecret")

			if secret == "" {
				return e.BadRequestError("Secret not found", err)
			}

			valid := totp.Validate(data.Passcode, secret)
			if !valid {
				return e.UnauthorizedError("Invalid passcode", err)
			}

			return apis.RecordAuthResponse(e, record, "totp", nil)
		})

		return se.Next()
	})

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}

type UserTotp struct {
	UserId   string `json:"userId" form:"userId"`
	Passcode string `json:"passcode" form:"passcode"`
}
