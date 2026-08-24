package ws

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/coder/websocket"
)

type Handler struct {
	logger  *slog.Logger
	manager *Manager
}

func NewHandler(manager *Manager, logger *slog.Logger) *Handler {
	return &Handler{
		logger:  logger,
		manager: manager,
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		OriginPatterns: []string{"*"},
	})

	if err != nil {
		h.logger.InfoContext(r.Context(), "failed to accept websocket", "user_id", userID, "err", err)
		return
	}

	h.manager.Register(context.Background(), userID, NewClient(conn))
	defer h.manager.Unregister(userID)

	for {
		_, _, err := conn.Read(r.Context())
		if err != nil {
			h.logger.InfoContext(r.Context(), "failed to read from websocket", "user_id", userID, "err", err)
			return
		}

	}

}
