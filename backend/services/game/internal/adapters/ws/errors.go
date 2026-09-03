package ws

import (
	"errors"
	"game/internal/domain"
)

const (
	illegalMove      = "ILLEGAL_MOVE"
	invalidSquare    = "INVALID_SQUARE"
	invalidPromotion = "INVALID_PROMOTION"
)

func toWsErrorCode(err error) string {
	switch {
	case errors.Is(err, domain.ErrIllegalMove):
		return illegalMove
	case errors.Is(err, domain.ErrInvalidSquare):
		return invalidSquare
	case errors.Is(err, domain.ErrInvalidPromotion):
		return invalidPromotion
	default:
		return "UNKNOWN_ERROR"
	}
}
