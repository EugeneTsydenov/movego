package ws

import (
	"errors"
	"game/internal/domain"
	"net/http"
	"shared/coreerrors"
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
func toHTTPError(err error) (msg string, code int) {
	switch {
	case errors.Is(err, coreerrors.ErrNotFound):
		msg = "not found"
		code = http.StatusNotFound
		return
	case errors.Is(err, coreerrors.ErrPermissionDenied):
		msg = "forbidden"
		code = http.StatusForbidden
		return
	default:
		msg = "internal server error"
		code = http.StatusInternalServerError
		return
	}
}
