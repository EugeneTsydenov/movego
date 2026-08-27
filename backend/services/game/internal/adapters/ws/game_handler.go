package ws

import (
	"context"
	"encoding/json"
	"errors"
	"game/internal/domain"
	"log/slog"
	"net/http"
	"shared/coreerrors"
	"shared/wsclient"

	"github.com/coder/websocket"
)

type GameService interface {
	GetState(ctx context.Context, gameID domain.GameID, playerID domain.PlayerID) (*domain.Game, error)
}

type GameHandler struct {
	logger  *slog.Logger
	manager *GameManager
	service GameService
}

func NewGameHandler(manager *GameManager, service GameService, logger *slog.Logger) *GameHandler {
	return &GameHandler{
		logger:  logger,
		manager: manager,
		service: service,
	}
}

func (h *GameHandler) Handle(mux *http.ServeMux) {
	mux.Handle("/ws", h)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
}

func (h *GameHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	userID, err := domain.NewPlayerID(r.Header.Get("X-User-ID"))
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	gameID, err := domain.NewGameID(r.URL.Query().Get("game_id"))
	if err != nil {
		http.Error(w, "invalid game_id format", http.StatusBadRequest)
		return
	}

	game, err := h.service.GetState(r.Context(), gameID, userID)
	if err != nil {
		switch {
		case errors.Is(err, coreerrors.ErrNotFound):
			http.Error(w, "game not found", http.StatusNotFound)
		case errors.Is(err, coreerrors.ErrPermissionDenied):
			http.Error(w, "forbidden", http.StatusForbidden)
		default:
			h.logger.WarnContext(r.Context(), "failed to get game state", "err", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}

	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		OriginPatterns: []string{"*"},
	})
	if err != nil {
		h.logger.InfoContext(r.Context(), "failed to accept websocket", "user_id", userID, "err", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	client := wsclient.New(conn)

	bgCtx := context.Background()

	h.manager.Register(bgCtx, gameID.String(), userID.String(), client)
	defer h.manager.Unregister(gameID.String(), userID.String())

	data, err := json.Marshal(toGameInitStateDTO(game))
	if err != nil {
		h.logger.InfoContext(bgCtx, "failed to marshal game init state", "user_id", userID, "game_id", gameID, "err", err)
		return
	}

	if err := client.Send(bgCtx, data); err != nil {
		h.logger.InfoContext(bgCtx, "failed to send game init state", "user_id", userID, "game_id", gameID, "err", err)
		return
	}

	for {
		_, _, err := conn.Read(bgCtx)
		if err != nil {
			h.logger.InfoContext(bgCtx, "failed to read from websocket", "user_id", userID, "game_id", gameID, "err", err)
			return
		}
	}
}
