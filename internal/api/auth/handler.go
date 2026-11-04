package auth

import (
	"context"

	"github.com/rpmoulton/mint-alternative/internal/httpx"
	"github.com/rpmoulton/mint-alternative/internal/jwt"
)

type Handler struct {
	svc Service
}

func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Register(ctx context.Context, req *RegisterInput) (*httpx.Reply[AuthResponse], error) {
	resp, err := h.svc.Register(ctx, &req.Body)
	if err != nil {
		return nil, err
	}
	return httpx.OK("User registered successfully", *resp), nil
}

func (h *Handler) Login(ctx context.Context, req *LoginInput) (*httpx.Reply[AuthResponse], error) {
	resp, err := h.svc.Login(ctx, &req.Body)
	if err != nil {
		return nil, err
	}
	return httpx.OK("Login successful", *resp), nil
}

func (h *Handler) Me(ctx context.Context, _ *EmptyInput) (*httpx.Reply[UserResponse], error) {
	_, userID, err := jwt.ParseClaimsCtx(ctx)
	if err != nil {
		return nil, httpx.NewUnauthorizedError("authentication required")
	}
	user, err := h.svc.Me(ctx, userID)
	if err != nil {
		return nil, err
	}
	return httpx.OK("Current user retrieved", *user), nil
}

func (h *Handler) Refresh(ctx context.Context, _ *EmptyInput) (*httpx.Reply[jwt.Tokens], error) {
	claims, userID, err := jwt.ParseClaimsCtx(ctx)
	if err != nil {
		return nil, httpx.NewUnauthorizedError("authentication required")
	}
	tokens, err := h.svc.Refresh(ctx, claims.Subject, userID)
	if err != nil {
		return nil, err
	}
	return httpx.OK("Token refreshed successfully", *tokens), nil
}

func (h *Handler) Logout(ctx context.Context, _ *EmptyInput) (*httpx.Reply[httpx.NoBody], error) {
	return httpx.Done("Logged out successfully"), nil
}
