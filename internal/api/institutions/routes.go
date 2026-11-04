package institutions

import (
	"net/http"

	"github.com/rpmoulton/mint-alternative/internal/httpx"
	"github.com/rpmoulton/mint-alternative/internal/jwt"
)

func (h *Handler) Routes() *httpx.Group {
	g := httpx.NewGroup(httpx.Defaults{Tags: []string{"institutions"}})

	httpx.Register(g, httpx.Operation[ListInstitutionsInput, []InstitutionResponse]{
		Method:             http.MethodGet,
		Path:               "/",
		Summary:            "List financial institutions",
		SuccessDescription: "List of financial institutions",
		Success:            httpx.Enveloped(http.StatusOK),
		Guards:             []httpx.Guard{jwt.AccessGuard()},
		Handler:            h.List,
	})

	httpx.Register(g, httpx.Operation[GetInstitutionInput, InstitutionResponse]{
		Method:             http.MethodGet,
		Path:               "/{id}",
		Summary:            "Get financial institution by ID",
		SuccessDescription: "Financial institution details",
		Success:            httpx.Enveloped(http.StatusOK),
		Problems: []httpx.ProblemKind{
			httpx.NotFound().Described("Institution not found"),
			httpx.BadRequest().Described("Invalid institution ID"),
		},
		Guards:  []httpx.Guard{jwt.AccessGuard()},
		Handler: h.Get,
	})

	httpx.Register(g, httpx.Operation[CreateInstitutionInput, InstitutionResponse]{
		Method:             http.MethodPost,
		Path:               "/",
		Summary:            "Create financial institution",
		BodyDescription:    "Financial institution details",
		SuccessDescription: "Financial institution created",
		Success:            httpx.Enveloped(http.StatusCreated),
		Guards:             []httpx.Guard{jwt.AccessGuard()},
		Handler:            h.Create,
	})

	return g
}
