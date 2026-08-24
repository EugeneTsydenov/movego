package game

import (
	gamev1 "gen/game/v1"
	"matchmaking/internal/domain"
)

func toProtoPlayer(player *domain.Player) *gamev1.Player {
	return &gamev1.Player{
		Id:     player.ID().String(),
		Name:   player.Name(),
		Rating: int32(player.Rating()),
	}
}

func toCreateGameRequest(whitePlayer, blackPlayer *domain.Player, timeControlID domain.TimeControlID) *gamev1.CreateGameRequest {
	return &gamev1.CreateGameRequest{
		WhitePlayer:   toProtoPlayer(whitePlayer),
		BlackPlayer:   toProtoPlayer(blackPlayer),
		TimeControlId: timeControlID.String(),
	}
}
