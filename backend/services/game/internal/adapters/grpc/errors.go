package grpc

import (
	"errors"
	"game/internal/domain"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// type errorSpec struct {
// 	Code   codes.Code
// 	Field  string
// 	Reason string
// }

// var errorRegistry = map[error]errorSpec{
// 	domain.ErrEmailTaken:            {Code: codes.AlreadyExists, Field: "email", Reason: "EMAIL_TAKEN"},
// 	domain.ErrTagTaken:              {Code: codes.AlreadyExists, Field: "tag", Reason: "TAG_TAKEN"},
// 	domain.ErrProviderAlreadyLinked: {Code: codes.AlreadyExists, Reason: "PROVIDER_ALREADY_LINKED"},
// 	domain.ErrProviderKeyTaken:      {Code: codes.AlreadyExists, Reason: "PROVIDER_KEY_TAKEN"},
// }

func mapDomainErrorToGRPC(err error) error {
	// for targetErr, spec := range errorRegistry {
	// 	if errors.Is(err, targetErr) {
	// 		st := status.New(spec.Code, err.Error())
	// 		br := &errdetails.BadRequest{
	// 			FieldViolations: []*errdetails.BadRequest_FieldViolation{
	// 				{Field: spec.Field, Description: err.Error()},
	// 			},
	// 		}
	// 		errInfo := &errdetails.ErrorInfo{
	// 			Reason: spec.Reason,
	// 			Domain: "movego.v1",
	// 		}

	// 		stWithDetails, _ := st.WithDetails(br, errInfo)
	// 		return stWithDetails.Err()
	// 	}
	// }

	switch {
	case errors.Is(err, domain.ErrGameNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, domain.ErrValidation):
		return status.Error(codes.InvalidArgument, err.Error())
	}

	return status.Error(codes.Internal, "internal server error")
}
