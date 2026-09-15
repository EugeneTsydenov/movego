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
		Color:  player.Color(),
	}
}

func toCreateGameRequest(players []*domain.Player, timeControlID domain.TimeControlID) *gamev1.CreateGameRequest {
	if len(players) < 2 {
		return nil
	}

	return &gamev1.CreateGameRequest{
		Players: []*gamev1.Player{
			toProtoPlayer(players[0]),
			toProtoPlayer(players[1]),
		},
		TimeControlId: timeControlID.String(),
	}
}
