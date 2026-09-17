package ws

import (
	"game/internal/domain"
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

func toTimeoutEvent(playerID string) timeoutEvent {
	return timeoutEvent{
		Type: "player_connection_timeout",
		Payload: timeoutPaylod{
			PlayerID: playerID,
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

func toPlayerDTO(player *domain.Player) playerDTO {
	return playerDTO{
		ID:     player.ID().String(),
		Name:   player.Name(),
		Rating: player.Rating().Int(),
		Color:  player.Color(),
	}
}

func toPlayerDTOs(players []*domain.Player) []playerDTO {
	dtos := make([]playerDTO, len(players))
	for i, s := range players {
		dtos[i] = toPlayerDTO(s)
	}

	return dtos
}

func toRoomStateEvent(game *domain.Game) roomStateEvent {
	return roomStateEvent{
		Type: "room_state",
		Payload: roomStatePayload{
			RoomID:  game.ID().String(),
			Status:  game.Status().String(),
			Players: toPlayerDTOs(game.Players()),
			GameState: &gameState{
				FEN:      game.FEN(),
				Turn:     game.CurrentPosition().Turn().String(),
				SANMoves: game.MoveSANHistoryStrings(),
				LANMoves: game.MoveLANHistoryStrings(),
			},
		},
	}
}

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
