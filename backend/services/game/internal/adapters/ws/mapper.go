package ws

import "game/internal/domain"

type PlayerDTO struct {
	PlayerID string `json:"player_id"`
	Name     string `json:"name"`
	Rating   int    `json:"rating"`
}

type GameInitState struct {
	Type        string    `json:"type"`
	GameID      string    `json:"game_id"`
	Status      string    `json:"status"`
	WhitePlayer PlayerDTO `json:"white_player"`
	BlackPlayer PlayerDTO `json:"black_player"`
	FEN         string    `json:"fen"`
	Turn        string    `json:"turn"`
	WhiteTimeMs int       `json:"white_time_ms"`
	BlackTimeMs int       `json:"black_time_ms"`
}

func toPlayerDTO(player *domain.Player) PlayerDTO {
	return PlayerDTO{
		PlayerID: player.ID().String(),
		Name:     player.Name(),
		Rating:   player.Rating().Int(),
	}
}

func toGameInitStateDTO(game *domain.Game) GameInitState {
	return GameInitState{
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
