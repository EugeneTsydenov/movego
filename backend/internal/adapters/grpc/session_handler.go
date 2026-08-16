package grpc

import (
	"context"
	movegov1 "movego/gen/go/movego/v1"
	"movego/internal/application"

	"github.com/google/uuid"
)

type sessionService interface {
	GetActiveSessions(ctx context.Context, userID uuid.UUID, sessionID uuid.UUID) ([]application.SessionDTO, error)
	RevokeSession(ctx context.Context, userID uuid.UUID, sessionID uuid.UUID) error
	RevokeOtherSessions(ctx context.Context, userID, currSessionID uuid.UUID) error
}

type SessionHandler struct {
	movegov1.UnimplementedSessionServiceServer
	sessionService sessionService
}

func NewSessionHandler(sessionService sessionService) *SessionHandler {
	return &SessionHandler{
		sessionService: sessionService,
	}
}

func (h *SessionHandler) GetActiveSessions(ctx context.Context, req *movegov1.GetActiveSessionsRequest) (*movegov1.GetActiveSessionsResponse, error) {
	userID := userIDFromContext(ctx)
	sessionID, _ := uuid.Parse(req.GetCurrentSessionId())
	sessions, err := h.sessionService.GetActiveSessions(ctx, userID, sessionID)
	if err != nil {
		return nil, err
	}
	return &movegov1.GetActiveSessionsResponse{
		Sessions: toProtoSessions(sessions),
	}, nil
}

func (h *SessionHandler) RevokeSession(ctx context.Context, req *movegov1.RevokeSessionRequest) (*movegov1.RevokeSessionResponse, error) {
	userID := userIDFromContext(ctx)
	sessionID, _ := uuid.Parse(req.GetSessionId())
	err := h.sessionService.RevokeSession(ctx, userID, sessionID)
	if err != nil {
		return nil, err
	}
	return &movegov1.RevokeSessionResponse{}, nil
}

func (h *SessionHandler) RevokeOtherSessions(ctx context.Context, req *movegov1.RevokeOtherSessionsRequest) (*movegov1.RevokeOtherSessionsResponse, error) {
	userID := userIDFromContext(ctx)
	sessionID, _ := uuid.Parse(req.GetCurrentSessionId())
	err := h.sessionService.RevokeOtherSessions(ctx, userID, sessionID)
	if err != nil {
		return nil, err
	}
	return &movegov1.RevokeOtherSessionsResponse{}, nil
}
