package application

import (
	"context"
	"user/internal/domain"
)

type SessionService struct {
	sessionRepo domain.SessionRepo
}

func NewSessionService(sessionRepo domain.SessionRepo) *SessionService {
	return &SessionService{
		sessionRepo: sessionRepo,
	}
}

func (s *SessionService) GetActiveSessions(ctx context.Context, userID domain.UserID, sessionID domain.SessionID) ([]SessionDTO, error) {
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

func (s *SessionService) RevokeSession(ctx context.Context, userID domain.UserID, sessionID domain.SessionID) error {
	session, err := s.sessionRepo.FindByID(ctx, sessionID)
	if err != nil {
		return err
	}

	if !session.CanBeRevoked(userID) {
		return domain.ErrSessionNotFound
	}

	err = s.sessionRepo.Delete(ctx, sessionID)
	if err != nil {
		return err
	}

	return nil
}

func (s *SessionService) RevokeOtherSessions(ctx context.Context, userID domain.UserID, currSessionID domain.SessionID) error {
	return s.sessionRepo.DeleteAllExcept(ctx, userID, currSessionID)
}
