package grpc

import (
	userv1 "gen/user/v1"
	"user/internal/application"
	"user/internal/domain"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func toSignUpInput(req *userv1.SignUpRequest) (application.SignUpInput, error) {
	email, err := domain.NewEmail(req.Email)
	if err != nil {
		return application.SignUpInput{}, err
	}
	password, err := domain.NewPlainPassword(req.Password)
	if err != nil {
		return application.SignUpInput{}, err
	}

	return application.SignUpInput{
		Email:     email,
		Password:  password,
		UserAgent: req.GetUserAgent(),
		ClientIP:  req.GetClientIp(),
	}, nil
}

func toSignUpResponse(out application.SignUpOutput) *userv1.SignUpResponse {
	return &userv1.SignUpResponse{
		User:         toProtoUser(out.User),
		AccessToken:  out.AccessToken,
		RefreshToken: out.RefreshToken,
	}
}

func toSignInInput(req *userv1.SignInRequest) (application.SignInInput, error) {
	email, err := domain.NewEmail(req.Email)
	if err != nil {
		return application.SignInInput{}, err
	}
	password, err := domain.NewPlainPassword(req.Password)
	if err != nil {
		return application.SignInInput{}, err
	}

	return application.SignInInput{
		Email:     email,
		Password:  password,
		UserAgent: req.GetUserAgent(),
		ClientIP:  req.GetClientIp(),
	}, nil
}

func toSignInResponse(out application.SignInOutput) *userv1.SignInResponse {
	return &userv1.SignInResponse{
		User:         toProtoUser(out.User),
		AccessToken:  out.AccessToken,
		RefreshToken: out.RefreshToken,
	}
}

func toRefreshInput(req *userv1.RefreshRequest) application.RefreshInput {
	return application.RefreshInput{
		RefreshToken: req.GetRefreshToken(),
	}
}

func toRefreshResponse(out application.RefreshOutput) *userv1.RefreshResponse {
	return &userv1.RefreshResponse{
		AccessToken:  out.AccessToken,
		RefreshToken: out.RefreshToken,
	}
}

func toSignOutInput(req *userv1.SignOutRequest) application.SignOutInput {
	return application.SignOutInput{
		RefreshToken: req.GetRefreshToken(),
	}
}

func toProtoUser(dto application.UserDTO) *userv1.User {
	return &userv1.User{
		Id:          dto.ID.String(),
		Email:       dto.Email.String(),
		Tag:         dto.Tag.String(),
		DisplayName: dto.DisplayName.String(),
		Role:        dto.Role.String(),
		CreatedAt:   timestamppb.New(dto.CreatedAt),
		UpdatedAt:   timestamppb.New(dto.UpdatedAt),
	}
}

func toUserDTO(user *userv1.User) (application.UserDTO, error) {
	id, err := domain.NewUserID(user.Id)
	if err != nil {
		return application.UserDTO{}, err
	}

	tag, err := domain.NewTag(user.GetTag())
	if err != nil {
		return application.UserDTO{}, err
	}

	email, err := domain.NewEmail(user.GetEmail())
	if err != nil {
		return application.UserDTO{}, err
	}

	displayName, err := domain.NewDisplayName(user.GetDisplayName())
	if err != nil {
		return application.UserDTO{}, err
	}

	role, err := domain.NewRole(user.GetRole())
	if err != nil {
		return application.UserDTO{}, err
	}

	return application.UserDTO{
		ID:          id,
		Tag:         tag,
		Email:       email,
		DisplayName: displayName,
		Role:        role,
		CreatedAt:   user.GetCreatedAt().AsTime(),
		UpdatedAt:   user.GetUpdatedAt().AsTime(),
	}, nil
}

func toProtoSession(session application.SessionDTO) *userv1.SessionInfo {
	return &userv1.SessionInfo{
		Id:           session.ID.String(),
		UserAgent:    session.UserAgent,
		ClientIp:     session.ClientIP,
		IsCurrent:    session.IsCurrent,
		LastActiveAt: timestamppb.New(session.LastActiveAt),
		CreatedAt:    timestamppb.New(session.CreatedAt),
		ExpiresAt:    timestamppb.New(session.ExpiresAt),
	}
}

func toProtoSessions(sessions []application.SessionDTO) []*userv1.SessionInfo {
	protoSessions := make([]*userv1.SessionInfo, len(sessions))
	for i := range sessions {
		protoSessions[i] = toProtoSession(sessions[i])
	}

	return protoSessions
}
