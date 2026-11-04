package dashboard

import (
	"net/http"

	"github.com/rpmoulton/mint-alternative/internal/httpx"
	"github.com/rpmoulton/mint-alternative/internal/jwt"
)

func (h *Handler) Routes() *httpx.Group {
	g := httpx.NewGroup(httpx.Defaults{Tags: []string{"dashboard"}})

	httpx.Register(g, httpx.Operation[GetDashboardInput, DashboardSummaryResponse]{
		Method:             http.MethodGet,
		Path:               "/summary",
		Summary:            "Get financial dashboard summary",
		SuccessDescription: "Dashboard financial metrics, net worth, cash, liabilities, spending",
		Success:            httpx.Enveloped(http.StatusOK),
		Problems: []httpx.ProblemKind{
			httpx.BadRequest().Described("Invalid query parameters"),
		},
		Guards:  []httpx.Guard{jwt.AccessGuard()},
		Handler: h.GetSummary,
	})

	return g
}
