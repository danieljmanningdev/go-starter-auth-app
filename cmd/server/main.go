package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	appauth "github.com/danieljmanningdev/go-starter-auth-app/internal/auth"
	authmiddleware "github.com/danieljmanningdev/go-web-auth/middleware"
	"github.com/danieljmanningdev/go-web-auth/password"
	"github.com/danieljmanningdev/go-web-auth/session"
	"github.com/danieljmanningdev/go-web-core/config"
	"github.com/danieljmanningdev/go-web-core/database"
	"github.com/danieljmanningdev/go-web-core/logging"
	"github.com/danieljmanningdev/go-web-core/rendering"
	"github.com/danieljmanningdev/go-web-security/csrf"
	"github.com/danieljmanningdev/go-web-security/headers"
	"github.com/danieljmanningdev/go-web-security/recovery"
)

type LoginPageData struct {
	Error string
}

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

	authStore := appauth.NewStore(db.SQL)

	mux := http.NewServeMux()

	mux.HandleFunc("/login", func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		switch r.Method {
		case http.MethodGet:
			if err := renderer.HTML(
				w,
				http.StatusOK,
				"login.html",
				nil,
			); err != nil {
				logger.Error(
					"render login page",
					"error", err,
				)

				http.Error(
					w,
					http.StatusText(http.StatusInternalServerError),
					http.StatusInternalServerError,
				)
			}

		case http.MethodPost:
			if err := r.ParseForm(); err != nil {
				http.Error(
					w,
					http.StatusText(http.StatusBadRequest),
					http.StatusBadRequest,
				)
				return
			}

			email := r.FormValue("email")
			plainPassword := r.FormValue("password")

			user, err := authStore.FindUserByEmail(
				r.Context(),
				email,
			)

			if err != nil {
				if !errors.Is(err, appauth.ErrUserNotFound) {
					logger.Error(
						"find user during login",
						"error", err,
					)
				}

				renderLoginError(
					renderer,
					w,
					"Invalid email or password.",
				)
				return
			}

			if !password.Compare(
				user.PasswordHash,
				plainPassword,
			) {
				renderLoginError(
					renderer,
					w,
					"Invalid email or password.",
				)
				return
			}

			token, err := session.GenerateToken()
			if err != nil {
				logger.Error(
					"generate session token",
					"error", err,
				)

				http.Error(
					w,
					http.StatusText(http.StatusInternalServerError),
					http.StatusInternalServerError,
				)
				return
			}

			tokenHash := session.HashToken(token)

			expiresAt := time.Now().Add(
				sessionConfig.MaxAge,
			)

			if err := authStore.CreateSession(
				r.Context(),
				user.ID,
				tokenHash,
				expiresAt,
			); err != nil {
				logger.Error(
					"create session",
					"error", err,
				)

				http.Error(
					w,
					http.StatusText(http.StatusInternalServerError),
					http.StatusInternalServerError,
				)
				return
			}

			session.SetCookie(
				w,
				sessionConfig,
				token,
			)

			http.Redirect(
				w,
				r,
				"/dashboard",
				http.StatusSeeOther,
			)

		default:
			w.Header().Set(
				"Allow",
				http.MethodGet+", "+http.MethodPost,
			)

			http.Error(
				w,
				http.StatusText(http.StatusMethodNotAllowed),
				http.StatusMethodNotAllowed,
			)
		}
	})

	dashboard := http.HandlerFunc(func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		if r.Method != http.MethodGet {
			w.Header().Set(
				"Allow",
				http.MethodGet,
			)

			http.Error(
				w,
				http.StatusText(http.StatusMethodNotAllowed),
				http.StatusMethodNotAllowed,
			)
			return
		}

		if err := renderer.HTML(
			w,
			http.StatusOK,
			"dashboard.html",
			nil,
		); err != nil {
			logger.Error(
				"render dashboard",
				"error", err,
			)

			http.Error(
				w,
				http.StatusText(http.StatusInternalServerError),
				http.StatusInternalServerError,
			)
		}
	})

	protectedDashboard := authmiddleware.RequireAuthentication(
		authmiddleware.Config{
			Session:  sessionConfig,
			LoginURL: "/login",
		},
		dashboard,
	)

	mux.Handle(
		"/dashboard",
		protectedDashboard,
	)

	mux.HandleFunc("/logout", func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		if r.Method != http.MethodPost {
			w.Header().Set(
				"Allow",
				http.MethodPost,
			)

			http.Error(
				w,
				http.StatusText(http.StatusMethodNotAllowed),
				http.StatusMethodNotAllowed,
			)
			return
		}

		token, err := session.TokenFromRequest(
			r,
			sessionConfig,
		)

		if err == nil {
			tokenHash := session.HashToken(token)

			if err := authStore.DeleteSession(
				r.Context(),
				tokenHash,
			); err != nil {
				logger.Error(
					"delete session",
					"error", err,
				)
			}
		}

		session.ClearCookie(
			w,
			sessionConfig,
		)

		http.Redirect(
			w,
			r,
			"/login",
			http.StatusSeeOther,
		)
	})

	lookupSession := authStore.FindSessionUserID

	var handler http.Handler = mux

	handler = authmiddleware.Authenticate(
		authmiddleware.Config{
			Session:  sessionConfig,
			LoginURL: "/login",
		},
		lookupSession,
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

func renderLoginError(
	renderer *rendering.Renderer,
	w http.ResponseWriter,
	message string,
) {
	_ = renderer.HTML(
		w,
		http.StatusUnauthorized,
		"login.html",
		LoginPageData{
			Error: message,
		},
	)
}
