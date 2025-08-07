package main

import (
	"log"
	"net/http"

	"github.com/dorianlgs/pocketbase-experiments/ui"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

func main() {
	// Load configuration
	config, err := LoadConfig()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// Initialize PocketBase app
	app := pocketbase.New()

	// Add indexFallback flag for SPA routing
	var indexFallback bool
	app.RootCmd.PersistentFlags().BoolVar(
		&indexFallback,
		"indexFallback",
		true,
		"fallback the request to index.html on missing static path, e.g. when pretty urls are used with SPA",
	)

	// Initialize services
	logger := log.Default()
	authService, err := NewAuthService(config, logger)
	if err != nil {
		log.Fatal("Failed to initialize auth service:", err)
	}

	app.OnServe().BindFunc(func(se *core.ServeEvent) error {
		// Initialize datastore
		datastore := NewInMem(logger, app)
		authService.SetDatastore(datastore)

		// Setup static file serving
		if !se.Router.HasRoute(http.MethodGet, "/{path...}") {
			se.Router.GET("/{path...}", apis.Static(ui.DistDirFS, indexFallback)).
				Bind(apis.Gzip())
		}

		// Setup route handlers
		setupRoutes(se, app, authService)

		return se.Next()
	})

	// Start the application
	if err := app.Start(); err != nil {
		log.Fatal("Failed to start application:", err)
	}
}

// setupRoutes configures all API routes
func setupRoutes(se *core.ServeEvent, app *pocketbase.PocketBase, authService *AuthService) {
	// Initialize handlers
	totpHandlers := NewTOTPHandlers(app, authService)
	webauthnHandlers := NewWebAuthnHandlers(app, authService)

	// TOTP routes
	se.Router.GET("/api/pb-experiments/get-qr", totpHandlers.HandleGetQR).Bind(apis.RequireAuth())
	se.Router.POST("/api/pb-experiments/totp-login", totpHandlers.HandleTOTPLogin)

	// WebAuthn routes
	se.Router.POST("/api/pb-experiments/passkey/registerStart", webauthnHandlers.HandleRegisterStart)
	se.Router.POST("/api/pb-experiments/passkey/registerFinish", webauthnHandlers.HandleRegisterFinish)
	se.Router.POST("/api/pb-experiments/passkey/loginStart", webauthnHandlers.HandleLoginStart)
	se.Router.POST("/api/pb-experiments/passkey/loginFinish", webauthnHandlers.HandleLoginFinish)
}