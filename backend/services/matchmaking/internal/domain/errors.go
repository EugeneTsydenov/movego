package domain

import "shared/coreerrors"

var (
	// Validation errors
	ErrInvalidUserID        = coreerrors.New(coreerrors.ErrValidation, "invalid user id")
	ErrInvalidTimeControlID = coreerrors.New(coreerrors.ErrValidation, "invalid time control id")
	ErrInvalidRating        = coreerrors.New(coreerrors.ErrValidation, "invalid rating")
	ErrInvalidGameID        = coreerrors.New(coreerrors.ErrValidation, "invalid game id")

	// AlreadyExists errors
	ErrAlreadyInQueue = coreerrors.New(coreerrors.ErrAlreadyExists, "already in queue")
)
