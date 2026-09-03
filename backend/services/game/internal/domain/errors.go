package domain

import "shared/coreerrors"

var (
	// Validation
	ErrInvalidTimeControlID = coreerrors.New(coreerrors.ErrValidation, "invalid time control")
	ErrInvalidPlayerID      = coreerrors.New(coreerrors.ErrValidation, "invalid player id")
	ErrInvalidGameID        = coreerrors.New(coreerrors.ErrValidation, "invalid game id")
	ErrInvalidRating        = coreerrors.New(coreerrors.ErrValidation, "invalid rating")
	ErrInvalidSquare        = coreerrors.New(coreerrors.ErrValidation, "invalid chess square")
	ErrInvalidPromotion     = coreerrors.New(coreerrors.ErrValidation, "invalid promotion piece")
	ErrIllegalMove          = coreerrors.New(coreerrors.ErrValidation, "illegal move")

	// NotFound
	ErrGameNotFound = coreerrors.New(coreerrors.ErrNotFound, "game not found")

	// PermissionDenied
	ErrPlayerNotInGame = coreerrors.New(coreerrors.ErrPermissionDenied, "player not in game")

	// Conflict
	ErrGameAlreadyStarted = coreerrors.New(coreerrors.ErrConflict, "game already started")
)
