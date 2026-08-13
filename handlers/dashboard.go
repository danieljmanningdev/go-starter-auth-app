package handlers

import (
	"log/slog"
	"net/http"

	"github.com/danieljmanningdev/go-web-core/rendering"
)

type DashboardHandler struct {
	Logger   *slog.Logger
	Renderer *rendering.Renderer
}

func (h *DashboardHandler) ServeHTTP(
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

	if err := h.Renderer.HTML(
		w,
		http.StatusOK,
		"dashboard.html",
		nil,
	); err != nil {
		h.Logger.Error(
			"render dashboard",
			"error", err,
		)

		http.Error(
			w,
			http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError,
		)
	}
}
