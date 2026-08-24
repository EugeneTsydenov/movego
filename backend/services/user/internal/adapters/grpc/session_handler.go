package grpc

import (
	"context"
	userv1 "gen/user/v1"
	"user/internal/application"
	"user/internal/domain"
)

type sessionService interface {
	GetActiveSessions(ctx context.Context, userID domain.UserID, sessionID domain.SessionID) ([]application.SessionDTO, error)
	RevokeSession(ctx context.Context, userID domain.UserID, sessionID domain.SessionID) error
	RevokeOtherSessions(ctx context.Context, userID domain.UserID, currSessionID domain.SessionID) error
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
	userID, err := domain.NewUserID(userIDFromContext(ctx).String())
	if err != nil {
		return nil, err
	}
	sessionID, err := domain.NewSessionID(req.GetCurrentSessionId())
	if err != nil {
		return nil, err
	}
	sessions, err := h.sessionService.GetActiveSessions(ctx, userID, sessionID)
	if err != nil {
		return nil, err
	}
	return &userv1.GetActiveSessionsResponse{
		Sessions: toProtoSessions(sessions),
	}, nil
}

func (h *SessionHandler) RevokeSession(ctx context.Context, req *userv1.RevokeSessionRequest) (*userv1.RevokeSessionResponse, error) {
	userID, err := domain.NewUserID(userIDFromContext(ctx).String())
	if err != nil {
		return nil, err
	}
	sessionID, err := domain.NewSessionID(req.GetSessionId())
	if err != nil {
		return nil, err
	}
	err = h.sessionService.RevokeSession(ctx, userID, sessionID)
	if err != nil {
		return nil, err
	}
	return &userv1.RevokeSessionResponse{}, nil
}

func (h *SessionHandler) RevokeOtherSessions(ctx context.Context, req *userv1.RevokeOtherSessionsRequest) (*userv1.RevokeOtherSessionsResponse, error) {
	userID, err := domain.NewUserID(userIDFromContext(ctx).String())
	if err != nil {
		return nil, err
	}
	sessionID, err := domain.NewSessionID(req.GetCurrentSessionId())
	if err != nil {
		return nil, err
	}
	err = h.sessionService.RevokeOtherSessions(ctx, userID, sessionID)
	if err != nil {
		return nil, err
	}
	return &userv1.RevokeOtherSessionsResponse{}, nil
}
