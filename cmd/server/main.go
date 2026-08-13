package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/danieljmanningdev/go-starter-auth-app/internal/auth"
	"github.com/danieljmanningdev/go-starter-auth-app/routes"
	authmiddleware "github.com/danieljmanningdev/go-web-auth/middleware"
	"github.com/danieljmanningdev/go-web-auth/session"
	"github.com/danieljmanningdev/go-web-core/config"
	"github.com/danieljmanningdev/go-web-core/database"
	"github.com/danieljmanningdev/go-web-core/logging"
	"github.com/danieljmanningdev/go-web-core/rendering"
	"github.com/danieljmanningdev/go-web-security/csrf"
	"github.com/danieljmanningdev/go-web-security/headers"
	"github.com/danieljmanningdev/go-web-security/recovery"
)

func main() {
	cfg := config.Load()

	logger := logging.New(
		cfg.Environment,
		cfg.LogLevel,
	)

	db, err := database.Open(
		context.Background(),
		cfg.DatabasePath,
	)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := database.RunMigrations(
		db.SQL,
		"migrations",
	); err != nil {
		log.Fatal(err)
	}

	renderer, err := rendering.New(
		"web/templates/login.html",
		"web/templates/dashboard.html",
	)
	if err != nil {
		log.Fatal(err)
	}

	sessionConfig := session.DefaultConfig()

	if cfg.Environment != "production" {
		sessionConfig.Secure = false
	}

	authStore := auth.NewStore(db.SQL)

	router := routes.New(
		routes.Dependencies{
			Logger:        logger,
			Renderer:      renderer,
			AuthStore:     authStore,
			SessionConfig: sessionConfig,
		},
	)

	var handler http.Handler = router

	handler = authmiddleware.Authenticate(
		authmiddleware.Config{
			Session:  sessionConfig,
			LoginURL: "/login",
		},
		authStore.FindSessionUserID,
		handler,
	)

	csrfHandler, err := csrf.Protect(
		csrf.Config{},
		handler,
	)
	if err != nil {
		log.Fatal(err)
	}

	handler = csrfHandler
	handler = headers.Secure(handler)
	handler = recovery.Middleware(
		logger,
		handler,
	)

	address := fmt.Sprintf(
		":%d",
		cfg.Port,
	)

	logger.Info(
		"server starting",
		"address", address,
		"environment", cfg.Environment,
	)

	if err := http.ListenAndServe(
		address,
		handler,
	); err != nil {
		log.Fatal(err)
	}
}
