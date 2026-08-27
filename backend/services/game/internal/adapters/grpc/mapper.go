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

	whitePlayer, err := toPlayerDTO(req.WhitePlayer)
	if err != nil {
		return application.CreateGameInput{}, err
	}
	blackPlayer, err := toPlayerDTO(req.BlackPlayer)
	if err != nil {
		return application.CreateGameInput{}, err
	}

	return application.CreateGameInput{
		WhitePlayer:   whitePlayer,
		BlackPlayer:   blackPlayer,
		TimeControlID: timeControlID,
	}, nil
}
