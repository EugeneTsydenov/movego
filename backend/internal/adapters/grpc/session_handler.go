package grpc

import (
	"context"
	movegov1 "movego/gen/go/movego/v1"
	"movego/internal/application"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type sessionService interface {
	GetActiveSessions(ctx context.Context, userID uuid.UUID, sessionID uuid.UUID) ([]application.SessionDTO, error)
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
	userID, ok := userIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing user identifier")
	}

	sessionID, err := uuid.Parse(req.GetCurrentSessionId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid session_id argument")
	}

	sessions, err := h.sessionService.GetActiveSessions(ctx, userID, sessionID)
	if err != nil {
		return nil, err
	}

	return &movegov1.GetActiveSessionsResponse{
		Sessions: toProtoSessions(sessions),
	}, nil
}
