package ws

import (
	"time"
)

func toDisconnectEvent(playerID string, timeoutExpiresAt time.Time) disconnectEvent {
	return disconnectEvent{
		Type: "player_disconnected",
		Payload: disconnectPayload{
			PlayerID:         playerID,
			TimeoutExpiresAt: timeoutExpiresAt.UnixMilli(),
		},
	}
}

func toConnectEvent(playerID string) connectEvent {
	return connectEvent{
		Type: "player_connected",
		Payload: connectPayload{
			PlayerID: playerID,
		},
	}
}

func toStartEvent() startEvent {
	return startEvent{
		Type: "game_started",
	}
}

// func toPlayerDTO(client *client) playerDTO {
// 	var expiresAt int64
// 	if !client.DisconnExpiresAt().IsZero() {
// 		expiresAt = client.DisconnExpiresAt().UnixMilli()
// 	}
// 	return playerDTO{
// 		ID:               session.Player.ID().String(),
// 		Name:             session.Player.Name(),
// 		Rating:           session.Player.Rating().Int(),
// 		Color:            session.Player.Color(),
// 		IsOnline:         session.IsOnline,
// 		TimeMs:           session.Player.TimeRemaining().Milliseconds(),
// 		TimeoutActive:    session.TimeoutActive,
// 		TimeoutExpiresAt: expiresAt,
// 	}
// }

// func toPlayerDTOs(sessions []*clientSession) []playerDTO {
// 	dtos := make([]playerDTO, len(sessions))
// 	for i, s := range sessions {
// 		dtos[i] = toPlayerDTO(s)
// 	}
// 	return dtos
// }

// func toRoomStateEvent(game *domain.Game, sessions []*clientSession) roomStateEvent {
// 	return roomStateEvent{
// 		Type: "room_state",
// 		Payload: roomStatePayload{
// 			RoomID:  game.ID().String(),
// 			Status:  game.Status().String(),
// 			Players: toPlayerDTOs(sessions),
// 			GameState: &gameState{
// 				FEN:      game.FEN(),
// 				Turn:     game.CurrentPosition().Turn().String(),
// 				SANMoves: game.MoveSANHistoryStrings(),
// 				LANMoves: game.MoveLANHistoryStrings(),
// 			},
// 		},
// 	}
// }

// func toMoveMadeEvent(out application.MakeMoveOutput, gameID domain.GameID) MoveMadeEvent {
// 	return MoveMadeEvent{
// 		Type:        "move_made",
// 		GameID:      gameID.String(),
// 		Move:        out.Move,
// 		Fen:         out.Fen,
// 		Turn:        out.Turn,
// 		WhiteTimeMs: out.WhiteTimeMs,
// 		BlackTimeMs: out.BlackTimeMs,
// 		Status:      out.Status.String(),
// 		Reason:      out.Reason,
// 	}
// }
