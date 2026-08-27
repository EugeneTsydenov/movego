package grpc

import (
	matchmakingv1 "gen/matchmaking/v1"
	"matchmaking/internal/application"
	"matchmaking/internal/domain"

	"github.com/google/uuid"
)

func toJoinQueueInput(userID uuid.UUID, req *matchmakingv1.JoinQueueRequest) (application.JoinQueueInput, error) {
	playerID, err := domain.NewPlayerID(userID.String())
	if err != nil {
		return application.JoinQueueInput{}, err
	}

	timeControlID, err := domain.NewTimeControlID(req.TimeControlId)
	if err != nil {
		return application.JoinQueueInput{}, err
	}

	rating, err := domain.NewRating(int(req.PlayerRating))
	if err != nil {
		return application.JoinQueueInput{}, err
	}

	return application.JoinQueueInput{
		PlayerID:      playerID,
		PlayerName:    req.PlayerName,
		TimeControlID: timeControlID,
		Rating:        rating,
	}, nil
}
