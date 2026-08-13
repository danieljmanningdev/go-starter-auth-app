package handlers

import (
	"errors"
	"log/slog"
	"net/http"
	"time"

	appauth "github.com/danieljmanningdev/go-starter-auth-app/internal/auth"
	"github.com/danieljmanningdev/go-web-auth/password"
	"github.com/danieljmanningdev/go-web-auth/session"
	"github.com/danieljmanningdev/go-web-core/rendering"
)

type LoginHandler struct {
	Logger        *slog.Logger
	Renderer      *rendering.Renderer
	AuthStore     *appauth.Store
	SessionConfig session.Config
}

type LoginPageData struct {
	Error string
}

func (h *LoginHandler) ServeHTTP(
	w http.ResponseWriter,
	r *http.Request,
) {
	switch r.Method {
	case http.MethodGet:
		h.showLogin(w)

	case http.MethodPost:
		h.handleLogin(w, r)

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
}

func (h *LoginHandler) showLogin(
	w http.ResponseWriter,
) {
	if err := h.Renderer.HTML(
		w,
		http.StatusOK,
		"login.html",
		nil,
	); err != nil {
		h.Logger.Error(
			"render login page",
			"error", err,
		)

		http.Error(
			w,
			http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError,
		)
	}
}

func (h *LoginHandler) handleLogin(
	w http.ResponseWriter,
	r *http.Request,
) {
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

	user, err := h.AuthStore.FindUserByEmail(
		r.Context(),
		email,
	)

	if err != nil {
		if !errors.Is(err, appauth.ErrUserNotFound) {
			h.Logger.Error(
				"find user during login",
				"error", err,
			)
		}

		h.renderError(
			w,
			"Invalid email or password.",
		)
		return
	}

	if !password.Compare(
		user.PasswordHash,
		plainPassword,
	) {
		h.renderError(
			w,
			"Invalid email or password.",
		)
		return
	}

	token, err := session.GenerateToken()
	if err != nil {
		h.Logger.Error(
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
		h.SessionConfig.MaxAge,
	)

	if err := h.AuthStore.CreateSession(
		r.Context(),
		user.ID,
		tokenHash,
		expiresAt,
	); err != nil {
		h.Logger.Error(
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
		h.SessionConfig,
		token,
	)

	http.Redirect(
		w,
		r,
		"/dashboard",
		http.StatusSeeOther,
	)
}

func (h *LoginHandler) renderError(
	w http.ResponseWriter,
	message string,
) {
	if err := h.Renderer.HTML(
		w,
		http.StatusUnauthorized,
		"login.html",
		LoginPageData{
			Error: message,
		},
	); err != nil {
		h.Logger.Error(
			"render login error",
			"error", err,
		)
	}
}
