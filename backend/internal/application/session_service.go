package application

import (
	"context"
	"movego/internal/domain"

	"github.com/google/uuid"
)

type SessionService struct {
	sessionRepo domain.SessionRepo
}

func NewSessionService(sessionRepo domain.SessionRepo) *SessionService {
	return &SessionService{
		sessionRepo: sessionRepo,
	}
}

func (s *SessionService) GetActiveSessions(ctx context.Context, userID uuid.UUID, sessionID uuid.UUID) ([]SessionDTO, error) {
	sessions, err := s.sessionRepo.ListActiveByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	foundCurrent := false
	dtos := make([]SessionDTO, len(sessions))
	for i, session := range sessions {
		isCurrent := session.ID() == sessionID
		if isCurrent {
			foundCurrent = true
		}

		dtos[i] = SessionDTO{
			ID:           session.ID(),
			UserAgent:    session.UserAgent(),
			ClientIP:     session.ClientIP(),
			LastActiveAt: session.LastActiveAt(),
			CreatedAt:    session.CreatedAt(),
			ExpiresAt:    session.ExpiresAt(),
			IsCurrent:    isCurrent,
		}
	}

	if !foundCurrent {
		return nil, domain.ErrSessionNotFound
	}

	return dtos, nil
}
