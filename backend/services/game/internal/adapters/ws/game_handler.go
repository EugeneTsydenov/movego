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
	StartGame(ctx context.Context, game *domain.Game, now time.Time) error
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
		msg, code := toHTTPError(err)
		if code == http.StatusInternalServerError {
			h.logger.WarnContext(r.Context(), "failed to get game state", "game_id", gameID, "user_id", userID)
		}
		http.Error(w, msg, code)
		return
	}

	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		OriginPatterns: []string{"*"},
	})
	if err != nil {
		h.logger.WarnContext(r.Context(), "failed to accept websocket", "user_id", userID, "err", err)
		return
	}

	client := wsclient.New(conn)

	bgCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go client.WritePump(bgCtx)

	onDisconn := h.makeOnDisconnect(gameID.String())
	onTimeout := h.makeOnTimeout(gameID.String())

	defer h.manager.DisconnectClient(gameID.String(), userID.String(), onDisconn, onTimeout)

	if !h.manager.IsRoomCreated(gameID.String()) {
		h.manager.CreateRoom(gameID.String(), game.PlayerIDStrings(), onDisconn, onTimeout)
	}

	onConn := func(ctx context.Context, clientID string) error {
		msg, err := json.Marshal(toConnectEvent(clientID))
		if err != nil {
			return err
		}
		// 👈 ИСПРАВЛЕНИЕ: передаем независящий от HTTP-запроса контекст
		return h.manager.BroadcastToClients(context.Background(), gameID.String(), msg)
	}

	onStart := func(ctx context.Context) error {
		err := h.gameService.StartGame(ctx, game, time.Now())
		if err != nil {
			return err
		}
		msg, err := json.Marshal(toStartEvent())
		if err != nil {
			return err
		}
		// 👈 ИСПРАВЛЕНИЕ: передаем независящий от HTTP-запроса контекст
		return h.manager.BroadcastToClients(context.Background(), gameID.String(), msg)
	}

	err = h.manager.OnPlayerConnect(bgCtx, gameID.String(), userID.String(), client, onConn, onStart)
	if err != nil {
		h.logger.WarnContext(bgCtx, "failed to connect player", "user_id", userID, "game_id", gameID, "err", err)
		_ = conn.Close(websocket.StatusInternalError, "failed to join room")
		return
	}

	// Цикл чтения веб-сокета
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
				_ = conn.Close(websocket.StatusPolicyViolation, "forbidden")
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

func (h *GameHandler) makeOnDisconnect(gameID string) func(ctx context.Context, client *client) error {
	return func(ctx context.Context, client *client) error {
		msg, err := json.Marshal(toDisconnectEvent(client.ClientID(), client.DisconnExpiresAt()))
		if err != nil {
			return err
		}
		_ = h.manager.BroadcastToClients(context.Background(), gameID, msg)
		return nil
	}
}

func (h *GameHandler) disconnectClient(
	gameID,
	userID string,
	onDisconn func(ctx context.Context, client *client) error,
	onTimeout func(ctx context.Context) error,
) {
	h.manager.DisconnectClient(gameID, userID, onDisconn, onTimeout)
}

func (h *GameHandler) makeOnTimeout(gameID string) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		h.logger.Info("player timeout")
		return nil
	}
}

// func (h *GameHandler) clientSessions(roomID string, players []*domain.Player) []*clientSession {
// 	session := make([]*clientSession, len(players))
// 	for i, p := range players {
// 		if h.manager.IsConnected(roomID, p.ID().String()) {
// 			session[i] = newClientSession(p, true, false, nil)
// 			continue
// 		}
// 		expiresAt := time.Now().UTC().Add(1 * time.Minute)
// 		session[i] = newClientSession(p, false, true, &expiresAt)
// 	}
// 	return session
// }

func (h *GameHandler) handleClientMessage(ctx context.Context, gameID domain.GameID, userID domain.PlayerID, data []byte) error {
	var msg clientEvent
	if err := json.Unmarshal(data, &msg); err != nil {
		return err
	}

	switch msg.Type {
	// case "make_move":
	// 	return h.handleMakeMove(ctx, gameID, userID, msg.Payload)
	default:
		return errors.New("unknown message type")
	}
}

// func (h *GameHandler) handleMakeMove(ctx context.Context, gameID domain.GameID, playerID domain.PlayerID, data []byte) error {
// 	var payload MovePayload
// 	if err := json.Unmarshal(data, &payload); err != nil {
// 		return err
// 	}

// 	move, err := domain.NewMove(payload.From, payload.To, payload.Promotion)
// 	if err != nil {
// 		return err
// 	}

// 	out, err := h.gameService.MakeMove(ctx, gameID, playerID, move)
// 	if err != nil {
// 		return err
// 	}

// 	event := toMoveMadeEvent(out, gameID)
// 	eventJSON, err := json.Marshal(event)
// 	if err != nil {
// 		return err
// 	}

// 	if err := h.manager.BroadcastMessage(ctx, gameID.String(), eventJSON); err != nil {
// 		return err
// 	}
// 	return nil
// }
