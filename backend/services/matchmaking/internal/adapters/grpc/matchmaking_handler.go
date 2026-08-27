package grpc

import (
	"context"
	matchmakingv1 "gen/matchmaking/v1"
	"matchmaking/internal/application"
	"shared/auth"
)

type matchmakingService interface {
	JoinQueue(ctx context.Context, in application.JoinQueueInput) error
}

type MatchmakingHandler struct {
	matchmakingv1.UnimplementedMatchmakingServiceServer
	matchmakingService matchmakingService
}

func NewMatchmakingHandler(matchmakingService matchmakingService) *MatchmakingHandler {
	return &MatchmakingHandler{
		matchmakingService: matchmakingService,
	}
}

func (h *MatchmakingHandler) JoinQueue(ctx context.Context, req *matchmakingv1.JoinQueueRequest) (*matchmakingv1.JoinQueueResponse, error) {
	userID := auth.UserIDFromContext(ctx)

	in, err := toJoinQueueInput(userID, req)
	if err != nil {
		return nil, err
	}
	err = h.matchmakingService.JoinQueue(ctx, in)
	if err != nil {
		return nil, err
	}
	return &matchmakingv1.JoinQueueResponse{}, nil
}
