package budgets

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

func (h *Handler) List(ctx context.Context, req *ListBudgetsInput) (*httpx.Reply[[]BudgetResponse], error) {
	_, userID, err := jwt.ParseClaimsCtx(ctx)
	if err != nil {
		return nil, httpx.NewUnauthorizedError("authentication required")
	}

	budgets, err := h.svc.List(ctx, userID, req)
	if err != nil {
		return nil, err
	}
	return httpx.OK("Budgets retrieved", budgets), nil
}

func (h *Handler) Get(ctx context.Context, req *GetBudgetInput) (*httpx.Reply[BudgetResponse], error) {
	_, userID, err := jwt.ParseClaimsCtx(ctx)
	if err != nil {
		return nil, httpx.NewUnauthorizedError("authentication required")
	}

	budget, err := h.svc.Get(ctx, req.ID, userID)
	if err != nil {
		return nil, err
	}
	return httpx.OK("Budget retrieved", *budget), nil
}

func (h *Handler) Create(ctx context.Context, req *CreateBudgetInput) (*httpx.Reply[BudgetResponse], error) {
	_, userID, err := jwt.ParseClaimsCtx(ctx)
	if err != nil {
		return nil, httpx.NewUnauthorizedError("authentication required")
	}

	budget, err := h.svc.Create(ctx, userID, &req.Body)
	if err != nil {
		return nil, err
	}
	return httpx.OK("Budget created successfully", *budget), nil
}

func (h *Handler) Update(ctx context.Context, req *UpdateBudgetInput) (*httpx.Reply[BudgetResponse], error) {
	_, userID, err := jwt.ParseClaimsCtx(ctx)
	if err != nil {
		return nil, httpx.NewUnauthorizedError("authentication required")
	}

	budget, err := h.svc.Update(ctx, req.ID, userID, &req.Body)
	if err != nil {
		return nil, err
	}
	return httpx.OK("Budget updated successfully", *budget), nil
}

func (h *Handler) Delete(ctx context.Context, req *DeleteBudgetInput) (*httpx.Reply[httpx.NoBody], error) {
	_, userID, err := jwt.ParseClaimsCtx(ctx)
	if err != nil {
		return nil, httpx.NewUnauthorizedError("authentication required")
	}

	if err := h.svc.Delete(ctx, req.ID, userID); err != nil {
		return nil, err
	}
	return httpx.Done("Budget deleted successfully"), nil
}

func (h *Handler) UpsertCategory(ctx context.Context, req *UpsertBudgetCategoryInput) (*httpx.Reply[BudgetResponse], error) {
	_, userID, err := jwt.ParseClaimsCtx(ctx)
	if err != nil {
		return nil, httpx.NewUnauthorizedError("authentication required")
	}

	budget, err := h.svc.UpsertCategory(ctx, req.ID, userID, &req.Body)
	if err != nil {
		return nil, err
	}
	return httpx.OK("Budget category allocation saved", *budget), nil
}

func (h *Handler) DeleteCategory(ctx context.Context, req *DeleteBudgetCategoryInput) (*httpx.Reply[BudgetResponse], error) {
	_, userID, err := jwt.ParseClaimsCtx(ctx)
	if err != nil {
		return nil, httpx.NewUnauthorizedError("authentication required")
	}

	budget, err := h.svc.DeleteCategory(ctx, req.ID, userID, req.CategoryID)
	if err != nil {
		return nil, err
	}
	return httpx.OK("Budget category allocation removed", *budget), nil
}
