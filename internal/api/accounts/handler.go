package accounts

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

func (h *Handler) List(ctx context.Context, req *ListAccountsInput) (*httpx.Reply[[]AccountResponse], error) {
	_, userID, err := jwt.ParseClaimsCtx(ctx)
	if err != nil {
		return nil, httpx.NewUnauthorizedError("authentication required")
	}

	accounts, err := h.svc.List(ctx, userID, req.IncludeArchived)
	if err != nil {
		return nil, err
	}
	return httpx.OK("Accounts retrieved", accounts), nil
}

func (h *Handler) Get(ctx context.Context, req *GetAccountInput) (*httpx.Reply[AccountResponse], error) {
	_, userID, err := jwt.ParseClaimsCtx(ctx)
	if err != nil {
		return nil, httpx.NewUnauthorizedError("authentication required")
	}

	account, err := h.svc.Get(ctx, req.ID, userID)
	if err != nil {
		return nil, err
	}
	return httpx.OK("Account retrieved", *account), nil
}

func (h *Handler) Create(ctx context.Context, req *CreateAccountInput) (*httpx.Reply[AccountResponse], error) {
	_, userID, err := jwt.ParseClaimsCtx(ctx)
	if err != nil {
		return nil, httpx.NewUnauthorizedError("authentication required")
	}

	account, err := h.svc.Create(ctx, userID, &req.Body)
	if err != nil {
		return nil, err
	}
	return httpx.OK("Account created successfully", *account), nil
}

func (h *Handler) Update(ctx context.Context, req *UpdateAccountInput) (*httpx.Reply[AccountResponse], error) {
	_, userID, err := jwt.ParseClaimsCtx(ctx)
	if err != nil {
		return nil, httpx.NewUnauthorizedError("authentication required")
	}

	account, err := h.svc.Update(ctx, req.ID, userID, &req.Body)
	if err != nil {
		return nil, err
	}
	return httpx.OK("Account updated successfully", *account), nil
}

func (h *Handler) Delete(ctx context.Context, req *DeleteAccountInput) (*httpx.Reply[httpx.NoBody], error) {
	_, userID, err := jwt.ParseClaimsCtx(ctx)
	if err != nil {
		return nil, httpx.NewUnauthorizedError("authentication required")
	}

	if err := h.svc.Delete(ctx, req.ID, userID); err != nil {
		return nil, err
	}
	return httpx.Done("Account deleted successfully"), nil
}
