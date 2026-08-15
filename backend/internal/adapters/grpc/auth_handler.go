package grpc

import (
	"context"
	movegov1 "movego/gen/go/movego/v1"
	"movego/internal/application"
)

type AuthService interface {
	SignUp(ctx context.Context, in application.SignUpInput) (application.SignUpOutput, error)
	SignIn(ctx context.Context, in application.SignInInput) (application.SignInOutput, error)
	Refresh(ctx context.Context, in application.RefreshInput) (application.RefreshOutput, error)
}

type AuthHandler struct {
	movegov1.UnimplementedAuthServiceServer
	authService AuthService
}

func NewAuthHandler(authService AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

func (h *AuthHandler) SignUp(ctx context.Context, req *movegov1.SignUpRequest) (*movegov1.SignUpResponse, error) {
	out, err := h.authService.SignUp(ctx, toSignUpInput(req))
	if err != nil {
		return nil, err
	}
	return toSignUpResponse(out), nil
}

func (h *AuthHandler) SignIn(ctx context.Context, req *movegov1.SignInRequest) (*movegov1.SignInResponse, error) {
	out, err := h.authService.SignIn(ctx, toSignInInput(req))
	if err != nil {
		return nil, err
	}
	return toSignInResponse(out), err
}

func (h *AuthHandler) Refresh(ctx context.Context, req *movegov1.RefreshRequest) (*movegov1.RefreshResponse, error) {
	out, err := h.authService.Refresh(ctx, toRefreshInput(req))
	if err != nil {
		return nil, err
	}
	return toRefreshResponse(out), err
}
