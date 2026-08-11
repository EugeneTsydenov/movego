package grpc

import (
	movegov1 "movego/gen/go/movego/v1"
	"movego/internal/application"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func toSignUpInput(req *movegov1.SignUpRequest) application.SignUpInput {
	return application.SignUpInput{
		Email:     req.GetEmail(),
		Password:  req.GetPassword(),
		UserAgent: req.GetUserAgent(),
		ClientIP:  req.GetClientIp(),
	}
}

func toSignUpResponse(out application.SignUpOutput) *movegov1.SignUpResponse {
	return &movegov1.SignUpResponse{
		User:         toProtoUser(out.User),
		AccessToken:  out.AccessToken,
		RefreshToken: out.RefreshToken,
	}
}

func toProtoUser(dto application.UserDTO) *movegov1.User {
	return &movegov1.User{
		Id:          dto.ID.String(),
		Email:       dto.Email,
		Tag:         dto.Tag,
		DisplayName: dto.DisplayName,
		Role:        dto.Role,
		CreatedAt:   timestamppb.New(dto.CreatedAt),
		UpdatedAt:   timestamppb.New(dto.UpdatedAt),
	}
}

func toUserDTO(user *movegov1.User) application.UserDTO {
	id, _ := uuid.Parse(user.GetId())

	return application.UserDTO{
		ID:          id,
		Tag:         user.GetTag(),
		Email:       user.GetEmail(),
		DisplayName: user.GetDisplayName(),
		Role:        user.GetRole(),
		CreatedAt:   user.GetCreatedAt().AsTime(),
		UpdatedAt:   user.GetUpdatedAt().AsTime(),
	}
}
