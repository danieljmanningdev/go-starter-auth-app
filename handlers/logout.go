package handlers

import (
	"log/slog"
	"net/http"

	appauth "github.com/danieljmanningdev/go-starter-auth-app/internal/auth"
	"github.com/danieljmanningdev/go-web-auth/session"
)

type LogoutHandler struct {
	Logger        *slog.Logger
	AuthStore     *appauth.Store
	SessionConfig session.Config
}

func (h *LogoutHandler) ServeHTTP(
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
		h.SessionConfig,
	)

	if err == nil {
		tokenHash := session.HashToken(token)

		if err := h.AuthStore.DeleteSession(
			r.Context(),
			tokenHash,
		); err != nil {
			h.Logger.Error(
				"delete session",
				"error", err,
			)
		}
	}

	session.ClearCookie(
		w,
		h.SessionConfig,
	)

	http.Redirect(
		w,
		r,
		"/login",
		http.StatusSeeOther,
	)
}
