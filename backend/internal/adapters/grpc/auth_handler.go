package grpc

import (
	"context"
	movegov1 "movego/gen/go/movego/v1"
	"movego/internal/application"
)

type AuthService interface {
	SignUp(ctx context.Context, dto application.SignUpInput) (application.SignUpOutput, error)
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
	res, err := h.authService.SignUp(ctx, toSignUpInput(req))
	if err != nil {
		return nil, err
	}

	return toSignUpResponse(res), nil
}
