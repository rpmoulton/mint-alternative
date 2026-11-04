package categories

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

func (h *Handler) List(ctx context.Context, _ *ListCategoriesInput) (*httpx.Reply[[]CategoryResponse], error) {
	_, userID, err := jwt.ParseClaimsCtx(ctx)
	if err != nil {
		return nil, httpx.NewUnauthorizedError("authentication required")
	}

	cats, err := h.svc.List(ctx, userID)
	if err != nil {
		return nil, err
	}
	return httpx.OK("Categories retrieved", cats), nil
}

func (h *Handler) Get(ctx context.Context, req *GetCategoryInput) (*httpx.Reply[CategoryResponse], error) {
	_, userID, err := jwt.ParseClaimsCtx(ctx)
	if err != nil {
		return nil, httpx.NewUnauthorizedError("authentication required")
	}

	cat, err := h.svc.Get(ctx, req.ID, userID)
	if err != nil {
		return nil, err
	}
	return httpx.OK("Category retrieved", *cat), nil
}

func (h *Handler) Create(ctx context.Context, req *CreateCategoryInput) (*httpx.Reply[CategoryResponse], error) {
	_, userID, err := jwt.ParseClaimsCtx(ctx)
	if err != nil {
		return nil, httpx.NewUnauthorizedError("authentication required")
	}

	cat, err := h.svc.Create(ctx, userID, &req.Body)
	if err != nil {
		return nil, err
	}
	return httpx.OK("Category created successfully", *cat), nil
}

func (h *Handler) Update(ctx context.Context, req *UpdateCategoryInput) (*httpx.Reply[CategoryResponse], error) {
	_, userID, err := jwt.ParseClaimsCtx(ctx)
	if err != nil {
		return nil, httpx.NewUnauthorizedError("authentication required")
	}

	cat, err := h.svc.Update(ctx, req.ID, userID, &req.Body)
	if err != nil {
		return nil, err
	}
	return httpx.OK("Category updated successfully", *cat), nil
}

func (h *Handler) Delete(ctx context.Context, req *DeleteCategoryInput) (*httpx.Reply[httpx.NoBody], error) {
	_, userID, err := jwt.ParseClaimsCtx(ctx)
	if err != nil {
		return nil, httpx.NewUnauthorizedError("authentication required")
	}

	if err := h.svc.Delete(ctx, req.ID, userID); err != nil {
		return nil, err
	}
	return httpx.Done("Category deleted successfully"), nil
}
