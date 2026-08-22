package grpc

import (
	"context"
	userv1 "protogen/user/v1"
	"user/internal/application"
)

type authService interface {
	SignUp(ctx context.Context, in application.SignUpInput) (application.SignUpOutput, error)
	SignIn(ctx context.Context, in application.SignInInput) (application.SignInOutput, error)
	Refresh(ctx context.Context, in application.RefreshInput) (application.RefreshOutput, error)
	SignOut(ctx context.Context, in application.SignOutInput) error
}

type AuthHandler struct {
	userv1.UnimplementedAuthServiceServer
	authService authService
}

func NewAuthHandler(authService authService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

func (h *AuthHandler) SignUp(ctx context.Context, req *userv1.SignUpRequest) (*userv1.SignUpResponse, error) {
	out, err := h.authService.SignUp(ctx, toSignUpInput(req))
	if err != nil {
		return nil, err
	}
	return toSignUpResponse(out), nil
}

func (h *AuthHandler) SignIn(ctx context.Context, req *userv1.SignInRequest) (*userv1.SignInResponse, error) {
	out, err := h.authService.SignIn(ctx, toSignInInput(req))
	if err != nil {
		return nil, err
	}
	return toSignInResponse(out), err
}

func (h *AuthHandler) Refresh(ctx context.Context, req *userv1.RefreshRequest) (*userv1.RefreshResponse, error) {
	out, err := h.authService.Refresh(ctx, toRefreshInput(req))
	if err != nil {
		return nil, err
	}
	return toRefreshResponse(out), err
}

func (h *AuthHandler) SignOut(ctx context.Context, req *userv1.SignOutRequest) (*userv1.SignOutResponse, error) {
	err := h.authService.SignOut(ctx, toSignOutInput(req))
	if err != nil {
		return nil, err
	}
	return &userv1.SignOutResponse{}, nil
}
