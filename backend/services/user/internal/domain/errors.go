package domain

import "shared/coreerrors"

var (
	// Not Found
	ErrUserNotFound       = coreerrors.New(coreerrors.ErrNotFound, "user not found")
	ErrSessionNotFound    = coreerrors.New(coreerrors.ErrNotFound, "session not found")
	ErrCredentialNotFound = coreerrors.New(coreerrors.ErrNotFound, "credential not found")

	// Already Exists
	ErrEmailTaken            = coreerrors.New(coreerrors.ErrAlreadyExists, "email is already taken")
	ErrTagTaken              = coreerrors.New(coreerrors.ErrAlreadyExists, "tag is already taken")
	ErrProviderAlreadyLinked = coreerrors.New(coreerrors.ErrAlreadyExists, "provider is already linked")
	ErrProviderKeyTaken      = coreerrors.New(coreerrors.ErrAlreadyExists, "social account is already linked")

	// Validation
	ErrInvalidUserID       = coreerrors.New(coreerrors.ErrValidation, "invalid user id")
	ErrInvalidSessionID    = coreerrors.New(coreerrors.ErrValidation, "invalid session id")
	ErrInvalidCredentialID = coreerrors.New(coreerrors.ErrValidation, "invalid credential id")
	ErrInvalidDisplayName  = coreerrors.New(coreerrors.ErrValidation, "invalid display name")
	ErrInvalidEmail        = coreerrors.New(coreerrors.ErrValidation, "invalid email")
	ErrInvalidTag          = coreerrors.New(coreerrors.ErrValidation, "invalid tag")
	ErrWeakPassword        = coreerrors.New(coreerrors.ErrValidation, "password is too weak")
	ErrInvalidRole         = coreerrors.New(coreerrors.ErrValidation, "invalid role")
	ErrInvalidProvider     = coreerrors.New(coreerrors.ErrValidation, "invalid provider")

	// Authentication
	ErrInvalidCredentials = coreerrors.New(coreerrors.ErrAuthentication, "invalid credentials")
)
