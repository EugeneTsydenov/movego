package ws

import (
	"game/internal/application"
	"game/internal/domain"
)

func toPlayerDTO(player *domain.Player) PlayerDTO {
	return PlayerDTO{
		PlayerID: player.ID().String(),
		Name:     player.Name(),
		Rating:   player.Rating().Int(),
	}
}

func toGameInitEvent(game *domain.Game) GameInitEvent {
	return GameInitEvent{
		Type:        "game_init",
		GameID:      game.ID().String(),
		Status:      game.Status().String(),
		WhitePlayer: toPlayerDTO(game.WhitePlayer()),
		BlackPlayer: toPlayerDTO(game.BlackPlayer()),
		FEN:         game.FEN(),
		Turn:        game.Position().Turn().String(),
		WhiteTimeMs: int(game.WhiteTimeRemaining().Milliseconds()),
		BlackTimeMs: int(game.WhiteTimeRemaining().Milliseconds()),
	}
}

func toMoveMadeEvent(out application.MakeMoveOutput, gameID domain.GameID) MoveMadeEvent {
	return MoveMadeEvent{
		Type:        "move_made",
		GameID:      gameID.String(),
		Move:        out.Move,
		Fen:         out.Fen,
		Turn:        out.Turn,
		WhiteTimeMs: out.WhiteTimeMs,
		BlackTimeMs: out.BlackTimeMs,
		Status:      out.Status.String(),
		Reason:      out.Reason,
	}
}
