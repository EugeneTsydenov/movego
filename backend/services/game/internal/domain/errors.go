package domain

import "shared/coreerrors"

var (
	// Validation
	ErrInvalidTimeControlID = coreerrors.New(coreerrors.ErrValidation, "invalid time control")
	ErrInvalidPlayerID      = coreerrors.New(coreerrors.ErrValidation, "invalid player id")
	ErrInvalidGameID        = coreerrors.New(coreerrors.ErrValidation, "invalid game id")
	ErrInvalidRating        = coreerrors.New(coreerrors.ErrValidation, "invalid rating")

	// NotFound
	ErrGameNotFound = coreerrors.New(coreerrors.ErrNotFound, "game not found")
)
