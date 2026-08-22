package grpc

import (
	"context"
	userv1 "protogen/user/v1"
	"user/internal/application"

	"github.com/google/uuid"
)

type sessionService interface {
	GetActiveSessions(ctx context.Context, userID uuid.UUID, sessionID uuid.UUID) ([]application.SessionDTO, error)
	RevokeSession(ctx context.Context, userID uuid.UUID, sessionID uuid.UUID) error
	RevokeOtherSessions(ctx context.Context, userID, currSessionID uuid.UUID) error
}

type SessionHandler struct {
	userv1.UnimplementedSessionServiceServer
	sessionService sessionService
}

func NewSessionHandler(sessionService sessionService) *SessionHandler {
	return &SessionHandler{
		sessionService: sessionService,
	}
}

func (h *SessionHandler) GetActiveSessions(ctx context.Context, req *userv1.GetActiveSessionsRequest) (*userv1.GetActiveSessionsResponse, error) {
	userID := userIDFromContext(ctx)
	sessionID, _ := uuid.Parse(req.GetCurrentSessionId())
	sessions, err := h.sessionService.GetActiveSessions(ctx, userID, sessionID)
	if err != nil {
		return nil, err
	}
	return &userv1.GetActiveSessionsResponse{
		Sessions: toProtoSessions(sessions),
	}, nil
}

func (h *SessionHandler) RevokeSession(ctx context.Context, req *userv1.RevokeSessionRequest) (*userv1.RevokeSessionResponse, error) {
	userID := userIDFromContext(ctx)
	sessionID, _ := uuid.Parse(req.GetSessionId())
	err := h.sessionService.RevokeSession(ctx, userID, sessionID)
	if err != nil {
		return nil, err
	}
	return &userv1.RevokeSessionResponse{}, nil
}

func (h *SessionHandler) RevokeOtherSessions(ctx context.Context, req *userv1.RevokeOtherSessionsRequest) (*userv1.RevokeOtherSessionsResponse, error) {
	userID := userIDFromContext(ctx)
	sessionID, _ := uuid.Parse(req.GetCurrentSessionId())
	err := h.sessionService.RevokeOtherSessions(ctx, userID, sessionID)
	if err != nil {
		return nil, err
	}
	return &userv1.RevokeOtherSessionsResponse{}, nil
}
