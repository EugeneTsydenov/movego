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

func toSignInInput(req *movegov1.SignInRequest) application.SignInInput {
	return application.SignInInput{
		Email:     req.GetEmail(),
		Password:  req.GetPassword(),
		UserAgent: req.GetUserAgent(),
		ClientIP:  req.GetClientIp(),
	}
}

func toSignInResponse(out application.SignInOutput) *movegov1.SignInResponse {
	return &movegov1.SignInResponse{
		User:         toProtoUser(out.User),
		AccessToken:  out.AccessToken,
		RefreshToken: out.RefreshToken,
	}
}

func toRefreshInput(req *movegov1.RefreshRequest) application.RefreshInput {
	return application.RefreshInput{
		RefreshToken: req.GetRefreshToken(),
	}
}

func toRefreshResponse(out application.RefreshOutput) *movegov1.RefreshResponse {
	return &movegov1.RefreshResponse{
		AccessToken:  out.AccessToken,
		RefreshToken: out.RefreshToken,
	}
}

func toSignOutInput(req *movegov1.SignOutRequest) application.SignOutInput {
	return application.SignOutInput{
		RefreshToken: req.GetRefreshToken(),
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

func toProtoSession(session application.SessionDTO) *movegov1.SessionInfo {
	return &movegov1.SessionInfo{
		Id:           session.ID.String(),
		UserAgent:    session.UserAgent,
		ClientIp:     session.ClientIP,
		IsCurrent:    session.IsCurrent,
		LastActiveAt: timestamppb.New(session.LastActiveAt),
		CreatedAt:    timestamppb.New(session.CreatedAt),
		ExpiresAt:    timestamppb.New(session.ExpiresAt),
	}
}

func toProtoSessions(sessions []application.SessionDTO) []*movegov1.SessionInfo {
	protoSessions := make([]*movegov1.SessionInfo, len(sessions))
	for i := range sessions {
		protoSessions[i] = toProtoSession(sessions[i])
	}

	return protoSessions
}
