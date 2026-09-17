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
	"github.com/google/uuid"
)

type GameService interface {
	GetState(ctx context.Context, gameID domain.GameID, playerID domain.PlayerID) (*domain.Game, error)
	StartGame(ctx context.Context, gameID domain.GameID, now time.Time) (*domain.Game, error)
	MakeMove(
		ctx context.Context,
		gameID domain.GameID,
		playerID domain.PlayerID,
		move domain.Move,
	) (application.MakeMoveOutput, error)
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
	userID, gameID, ok := h.parseRequest(w, r)
	if !ok {
		return
	}
	game, ok := h.fetchGameState(w, r, gameID, userID)
	if !ok {
		return
	}
	conn, ok := h.acceptWebSocket(w, r, userID)
	if !ok {
		return
	}
	h.runSession(conn, gameID, userID, game)
}

func (h *GameHandler) parseRequest(w http.ResponseWriter, r *http.Request) (domain.PlayerID, domain.GameID, bool) {
	userID, err := domain.NewPlayerID(r.Header.Get("X-User-ID"))
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)

		return domain.PlayerID{}, domain.GameID{}, false
	}

	gameID, err := domain.NewGameID(r.URL.Query().Get("game_id"))
	if err != nil {
		http.Error(w, "invalid game_id format", http.StatusBadRequest)

		return domain.PlayerID{}, domain.GameID{}, false
	}

	return userID, gameID, true
}

func (h *GameHandler) fetchGameState(
	w http.ResponseWriter,
	r *http.Request,
	gameID domain.GameID,
	userID domain.PlayerID,
) (*domain.Game, bool) {
	game, err := h.gameService.GetState(r.Context(), gameID, userID)
	if err != nil {
		msg, code := toHTTPError(err)
		if code == http.StatusInternalServerError {
			h.logger.WarnContext(
				r.Context(),
				"failed to get game state",
				"game_id", gameID,
				"user_id", userID,
			)
		}
		http.Error(w, msg, code)

		return nil, false
	}

	return game, true
}

func (h *GameHandler) acceptWebSocket(
	w http.ResponseWriter,
	r *http.Request,
	userID domain.PlayerID,
) (*websocket.Conn, bool) {
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		OriginPatterns: []string{"*"},
	})
	if err != nil {
		h.logger.WarnContext(r.Context(), "failed to accept websocket", "user_id", userID, "err", err)

		return nil, false
	}

	return conn, true
}

func (h *GameHandler) runSession(
	conn *websocket.Conn,
	gameID domain.GameID,
	userID domain.PlayerID,
	game *domain.Game,
) {
	sessionID := uuid.Must(uuid.NewV7())
	client := wsclient.New(conn)

	bgCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go client.WritePump(bgCtx)

	onConn := h.makeOnConnection(gameID.String())
	onStart := h.makeOnStart(gameID)
	onDisconn := h.makeOnDisconnect(gameID.String())
	onTimeout := h.makeOnTimeout(gameID.String())

	defer h.manager.DisconnectClient(managerDisconnectClientArgs{
		roomID:    gameID.String(),
		clientID:  userID.String(),
		sessionID: sessionID.String(),
		onDisconn: onDisconn,
		onTimeout: onTimeout,
	})

	h.manager.CreateRoom(gameID.String(), game.PlayerIDStrings(), onTimeout)

	err := h.manager.OnClientConnect(bgCtx, onClientConnectArgs{
		roomID:    gameID.String(),
		clientID:  userID.String(),
		sessionID: sessionID.String(),
		wsClient:  client,
		onConn:    onConn,
		onStart:   onStart,
	})
	if err != nil {
		h.logger.WarnContext(
			bgCtx,
			"failed to connect player",
			"user_id", userID,
			"game_id", gameID,
			"err", err,
		)
		_ = conn.Close(websocket.StatusInternalError, "failed to join room")

		return
	}

	h.serveClientMessages(bgCtx, conn, client, gameID, userID)
}

func (h *GameHandler) serveClientMessages(
	ctx context.Context,
	conn *websocket.Conn,
	client *wsclient.Client,
	gameID domain.GameID,
	userID domain.PlayerID,
) {
	for {
		msgType, data, err := conn.Read(ctx)
		if err != nil {
			h.logger.InfoContext(
				ctx,
				"failed to read from websocket",
				"user_id",
				userID,
				"game_id",
				gameID,
				"err",
				err,
			)

			return
		}

		if msgType != websocket.MessageText {
			continue
		}

		if err := h.handleClientMessage(ctx, gameID, userID, data); err != nil {
			h.logger.WarnContext(
				ctx,
				"failed to handle client message",
				"user_id",
				userID,
				"game_id",
				gameID,
				"err",
				err,
			)

			if errors.Is(err, coreerrors.ErrPermissionDenied) {
				_ = conn.Close(websocket.StatusPolicyViolation, "forbidden")

				return
			}

			errorResp, _ := json.Marshal(map[string]any{
				"type": "error",
				"code": toWsErrorCode(err),
			})
			_ = client.Send(ctx, errorResp)
		}
	}
}

func (h *GameHandler) makeOnConnection(gameID string) onConn {
	return func(ctx context.Context, clientID string) error {
		msg, err := json.Marshal(toConnectEvent(clientID))
		if err != nil {
			return err
		}

		return h.manager.BroadcastExcept(context.Background(), gameID, clientID, msg)
	}
}

func (h *GameHandler) makeOnStart(gameID domain.GameID) onStart {
	return func(ctx context.Context) error {
		game, err := h.gameService.StartGame(ctx, gameID, time.Now())
		if err != nil {
			return err
		}

		startEventMsg, err := json.Marshal(toStartEvent())
		if err != nil {
			return err
		}

		roomStateMsg, err := json.Marshal(toRoomStateEvent(game))
		if err != nil {
			return err
		}

		if err := h.manager.Broadcast(context.Background(), gameID.String(), startEventMsg); err != nil {
			return err
		}

		return h.manager.Broadcast(context.Background(), gameID.String(), roomStateMsg)
	}
}

func (h *GameHandler) makeOnDisconnect(gameID string) onDisconn {
	return func(ctx context.Context, client *client) error {
		msg, err := json.Marshal(toDisconnectEvent(client.ID(), client.DisconnExpiresAt()))
		if err != nil {
			return err
		}
		_ = h.manager.BroadcastExcept(context.Background(), gameID, client.ID(), msg)

		return nil
	}
}

func (h *GameHandler) makeOnTimeout(gameID string) onTimeout {
	return func(ctx context.Context, clientID string) error {
		msg, err := json.Marshal(toTimeoutEvent(clientID))
		if err != nil {
			return nil
		}

		_ = h.manager.BroadcastExcept(context.Background(), gameID, clientID, msg)

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

func (h *GameHandler) handleClientMessage(
	ctx context.Context,
	gameID domain.GameID,
	userID domain.PlayerID,
	data []byte,
) error {
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
