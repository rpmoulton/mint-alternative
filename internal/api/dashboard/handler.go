package dashboard

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

func (h *Handler) GetSummary(ctx context.Context, req *GetDashboardInput) (*httpx.Reply[DashboardSummaryResponse], error) {
	_, userID, err := jwt.ParseClaimsCtx(ctx)
	if err != nil {
		return nil, httpx.NewUnauthorizedError("authentication required")
	}

	summary, err := h.svc.GetSummary(ctx, userID, req)
	if err != nil {
		return nil, err
	}
	return httpx.OK("Dashboard summary retrieved", *summary), nil
}
