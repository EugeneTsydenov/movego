package grpc

import (
	"game/internal/application"
	gamev1 "protogen/game/v1"

	"github.com/google/uuid"
)

func toPlayerDTO(in *gamev1.Player) application.PlayerDTO {
	return application.PlayerDTO{
		ID:   uuid.Must(uuid.Parse(in.Id)),
		Name: in.Name,
	}
}

func toCreateGameInput(in *gamev1.CreateGameRequest) application.CreateGameInput {
	return application.CreateGameInput{
		WhitePlayer:   toPlayerDTO(in.WhitePlayer),
		BlackPlayer:   toPlayerDTO(in.BlackPlayer),
		TimeControlID: in.TimeControlId,
	}
}

func toCreateGameResponse(in application.CreateGameOutput) *gamev1.CreateGameResponse {
	return &gamev1.CreateGameResponse{
		GameId: in.GameID.String(),
	}
}
