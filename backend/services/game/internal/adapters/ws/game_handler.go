package ws

import (
	"context"
	"encoding/json"
	"errors"
	"game/internal/application"
	"game/internal/domain"
	"log/slog"
	"net/http"
	"shared/coreerrors"
	"shared/wsclient"
	"time"

	"github.com/coder/websocket"
)

type GameService interface {
	GetState(ctx context.Context, gameID domain.GameID, playerID domain.PlayerID) (*domain.Game, error)
	MakeMove(ctx context.Context, gameID domain.GameID, playerID domain.PlayerID, move domain.Move) (application.MakeMoveOutput, error)
}

type GameHandler struct {
	logger      *slog.Logger
	manager     *GameManager
	gameService GameService
}

func NewGameHandler(manager *GameManager, gameService GameService, logger *slog.Logger) *GameHandler {
	return &GameHandler{
		logger:      logger,
		manager:     manager,
		gameService: gameService,
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

	game, err := h.gameService.GetState(r.Context(), gameID, userID)
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

	bgCtx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	h.manager.Register(bgCtx, gameID.String(), userID.String(), client)
	defer h.manager.Unregister(gameID.String(), userID.String())

	data, err := json.Marshal(toGameInitEvent(game))
	if err != nil {
		h.logger.InfoContext(bgCtx, "failed to marshal game init state", "user_id", userID, "game_id", gameID, "err", err)
		return
	}

	if err := client.Send(bgCtx, data); err != nil {
		h.logger.InfoContext(bgCtx, "failed to send game init state", "user_id", userID, "game_id", gameID, "err", err)
		return
	}

	for {
		msgType, data, err := conn.Read(bgCtx)
		if err != nil {
			h.logger.InfoContext(bgCtx, "failed to read from websocket", "user_id", userID, "game_id", gameID, "err", err)
			return
		}

		if msgType != websocket.MessageText {
			continue
		}

		if err := h.handleClientMessage(bgCtx, gameID, userID, data); err != nil {
			h.logger.WarnContext(bgCtx, "failed to handle client message", "user_id", userID, "game_id", gameID, "err", err)

			if errors.Is(err, coreerrors.ErrPermissionDenied) {
				err = conn.Close(websocket.StatusPolicyViolation, "forbidden")
				if err != nil {
					h.logger.WarnContext(bgCtx, "failed to close ws connection", "user_id", userID, "game_id", gameID, "err", err)
				}
				return
			}

			errorResp, _ := json.Marshal(map[string]any{
				"type": "error",
				"code": toWsErrorCode(err),
			})
			_ = client.Send(bgCtx, errorResp)
		}
	}
}

func (h *GameHandler) handleClientMessage(ctx context.Context, gameID domain.GameID, userID domain.PlayerID, data []byte) error {
	var msg ClientEvent
	if err := json.Unmarshal(data, &msg); err != nil {
		return err
	}

	switch msg.Type {
	case "make_move":
		return h.handleMakeMove(ctx, gameID, userID, msg.Payload)
	default:
		return errors.New("unknown message type")
	}
}

func (h *GameHandler) handleMakeMove(ctx context.Context, gameID domain.GameID, playerID domain.PlayerID, data []byte) error {
	var payload MovePayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return err
	}

	move, err := domain.NewMove(payload.From, payload.To, payload.Promotion)
	if err != nil {
		return err
	}

	out, err := h.gameService.MakeMove(ctx, gameID, playerID, move)
	if err != nil {
		return err
	}

	event := toMoveMadeEvent(out, gameID)
	eventJSON, err := json.Marshal(event)
	if err != nil {
		return err
	}

	if err := h.manager.BroadcastMessage(ctx, gameID.String(), eventJSON); err != nil {
		return err
	}

	return nil
}
