package institutions

import (
	"context"

	"github.com/rpmoulton/mint-alternative/internal/httpx"
)

type Handler struct {
	svc Service
}

func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) List(ctx context.Context, _ *ListInstitutionsInput) (*httpx.Reply[[]InstitutionResponse], error) {
	list, err := h.svc.List(ctx)
	if err != nil {
		return nil, err
	}
	return httpx.OK("Institutions retrieved", list), nil
}

func (h *Handler) Get(ctx context.Context, req *GetInstitutionInput) (*httpx.Reply[InstitutionResponse], error) {
	inst, err := h.svc.Get(ctx, req.ID)
	if err != nil {
		return nil, err
	}
	return httpx.OK("Institution retrieved", *inst), nil
}

func (h *Handler) Create(ctx context.Context, req *CreateInstitutionInput) (*httpx.Reply[InstitutionResponse], error) {
	inst, err := h.svc.Create(ctx, &req.Body)
	if err != nil {
		return nil, err
	}
	return httpx.OK("Institution created successfully", *inst), nil
}
