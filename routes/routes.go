package routes

import (
	"log/slog"
	"net/http"

	"github.com/danieljmanningdev/go-starter-auth-app/handlers"
	appauth "github.com/danieljmanningdev/go-starter-auth-app/internal/auth"
	authmiddleware "github.com/danieljmanningdev/go-web-auth/middleware"
	"github.com/danieljmanningdev/go-web-auth/session"
	"github.com/danieljmanningdev/go-web-core/rendering"
)

type Dependencies struct {
	Logger        *slog.Logger
	Renderer      *rendering.Renderer
	AuthStore     *appauth.Store
	SessionConfig session.Config
}

func New(deps Dependencies) http.Handler {
	mux := http.NewServeMux()

	loginHandler := &handlers.LoginHandler{
		Logger:        deps.Logger,
		Renderer:      deps.Renderer,
		AuthStore:     deps.AuthStore,
		SessionConfig: deps.SessionConfig,
	}

	dashboardHandler := &handlers.DashboardHandler{
		Logger:   deps.Logger,
		Renderer: deps.Renderer,
	}

	logoutHandler := &handlers.LogoutHandler{
		Logger:        deps.Logger,
		AuthStore:     deps.AuthStore,
		SessionConfig: deps.SessionConfig,
	}

	authConfig := authmiddleware.Config{
		Session:  deps.SessionConfig,
		LoginURL: "/login",
	}

	mux.Handle(
		"/login",
		loginHandler,
	)

	mux.Handle(
		"/dashboard",
		authmiddleware.RequireAuthentication(
			authConfig,
			dashboardHandler,
		),
	)

	mux.Handle(
		"/logout",
		authmiddleware.RequireAuthentication(
			authConfig,
			logoutHandler,
		),
	)

	return mux
}
