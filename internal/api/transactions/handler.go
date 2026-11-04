package transactions

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

func (h *Handler) List(ctx context.Context, req *ListTransactionsInput) (*httpx.Reply[[]TransactionResponse], error) {
	_, userID, err := jwt.ParseClaimsCtx(ctx)
	if err != nil {
		return nil, httpx.NewUnauthorizedError("authentication required")
	}

	txs, pagination, err := h.svc.List(ctx, userID, req)
	if err != nil {
		return nil, err
	}
	return httpx.Paged("Transactions retrieved", txs, pagination), nil
}

func (h *Handler) Get(ctx context.Context, req *GetTransactionInput) (*httpx.Reply[TransactionResponse], error) {
	_, userID, err := jwt.ParseClaimsCtx(ctx)
	if err != nil {
		return nil, httpx.NewUnauthorizedError("authentication required")
	}

	tx, err := h.svc.Get(ctx, req.ID, userID)
	if err != nil {
		return nil, err
	}
	return httpx.OK("Transaction retrieved", *tx), nil
}

func (h *Handler) Create(ctx context.Context, req *CreateTransactionInput) (*httpx.Reply[TransactionResponse], error) {
	_, userID, err := jwt.ParseClaimsCtx(ctx)
	if err != nil {
		return nil, httpx.NewUnauthorizedError("authentication required")
	}

	tx, err := h.svc.Create(ctx, userID, &req.Body)
	if err != nil {
		return nil, err
	}
	return httpx.OK("Transaction created successfully", *tx), nil
}

func (h *Handler) Update(ctx context.Context, req *UpdateTransactionInput) (*httpx.Reply[TransactionResponse], error) {
	_, userID, err := jwt.ParseClaimsCtx(ctx)
	if err != nil {
		return nil, httpx.NewUnauthorizedError("authentication required")
	}

	tx, err := h.svc.Update(ctx, req.ID, userID, &req.Body)
	if err != nil {
		return nil, err
	}
	return httpx.OK("Transaction updated successfully", *tx), nil
}

func (h *Handler) Delete(ctx context.Context, req *DeleteTransactionInput) (*httpx.Reply[httpx.NoBody], error) {
	_, userID, err := jwt.ParseClaimsCtx(ctx)
	if err != nil {
		return nil, httpx.NewUnauthorizedError("authentication required")
	}

	if err := h.svc.Delete(ctx, req.ID, userID); err != nil {
		return nil, err
	}
	return httpx.Done("Transaction deleted successfully"), nil
}
