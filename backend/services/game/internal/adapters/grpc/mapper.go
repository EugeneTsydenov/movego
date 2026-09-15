package grpc

import (
	"game/internal/application"
	"game/internal/domain"
	gamev1 "gen/game/v1"
	"shared/timecontrol"
)

func toPlayerDTO(in *gamev1.Player) (application.PlayerDTO, error) {
	id, err := domain.NewPlayerID(in.Id)
	if err != nil {
		return application.PlayerDTO{}, err
	}
	rating, err := domain.NewRating(int(in.Rating))
	if err != nil {
		return application.PlayerDTO{}, err
	}
	return application.PlayerDTO{
		ID:     id,
		Name:   in.Name,
		Rating: rating,
		Color:  in.Color,
	}, nil
}

func toTimeControlID(timeControlID string) (domain.TimeControlID, error) {
	if !timecontrol.IsValidTimeControlID(timeControlID) {
		return domain.TimeControlID(""), domain.ErrInvalidTimeControlID
	}
	return domain.TimeControlID(timeControlID), nil
}

func toCreateGameInput(req *gamev1.CreateGameRequest) (application.CreateGameInput, error) {
	timeControlID, err := toTimeControlID(req.TimeControlId)
	if err != nil {
		return application.CreateGameInput{}, err
	}

	players := make([]application.PlayerDTO, len(req.Players))
	for i := 0; i < len(req.Players); i++ {
		player, err := toPlayerDTO(req.Players[i])
		if err != nil {
			return application.CreateGameInput{}, err
		}
		players[i] = player
	}

	return application.CreateGameInput{
		Players:       players,
		TimeControlID: timeControlID,
	}, nil
}
