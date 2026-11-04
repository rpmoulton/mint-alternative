package accounts

import (
	"net/http"

	"github.com/rpmoulton/mint-alternative/internal/httpx"
	"github.com/rpmoulton/mint-alternative/internal/jwt"
)

func (h *Handler) Routes() *httpx.Group {
	g := httpx.NewGroup(httpx.Defaults{Tags: []string{"accounts"}})

	httpx.Register(g, httpx.Operation[ListAccountsInput, []AccountResponse]{
		Method:             http.MethodGet,
		Path:               "/",
		Summary:            "List user financial accounts",
		SuccessDescription: "List of financial accounts",
		Success:            httpx.Enveloped(http.StatusOK),
		Guards:             []httpx.Guard{jwt.AccessGuard()},
		Handler:            h.List,
	})

	httpx.Register(g, httpx.Operation[GetAccountInput, AccountResponse]{
		Method:             http.MethodGet,
		Path:               "/{id}",
		Summary:            "Get account by ID",
		SuccessDescription: "Account details",
		Success:            httpx.Enveloped(http.StatusOK),
		Problems: []httpx.ProblemKind{
			httpx.NotFound().Described("Account not found"),
			httpx.BadRequest().Described("Invalid account ID"),
		},
		Guards:  []httpx.Guard{jwt.AccessGuard()},
		Handler: h.Get,
	})

	httpx.Register(g, httpx.Operation[CreateAccountInput, AccountResponse]{
		Method:             http.MethodPost,
		Path:               "/",
		Summary:            "Create a new financial account",
		BodyDescription:    "Account creation payload",
		SuccessDescription: "Account created successfully",
		Success:            httpx.Enveloped(http.StatusCreated),
		Problems: []httpx.ProblemKind{
			httpx.BadRequest().Described("Invalid account parameters"),
		},
		Guards:  []httpx.Guard{jwt.AccessGuard()},
		Handler: h.Create,
	})

	httpx.Register(g, httpx.Operation[UpdateAccountInput, AccountResponse]{
		Method:             http.MethodPatch,
		Path:               "/{id}",
		Summary:            "Update an account",
		BodyDescription:    "Account update payload",
		SuccessDescription: "Account updated successfully",
		Success:            httpx.Enveloped(http.StatusOK),
		Problems: []httpx.ProblemKind{
			httpx.NotFound().Described("Account not found"),
			httpx.BadRequest().Described("Invalid account parameters"),
		},
		Guards:  []httpx.Guard{jwt.AccessGuard()},
		Handler: h.Update,
	})

	httpx.Register(g, httpx.Operation[DeleteAccountInput, httpx.NoBody]{
		Method:             http.MethodDelete,
		Path:               "/{id}",
		Summary:            "Delete an account",
		SuccessDescription: "Account deleted",
		Success:            httpx.Message(http.StatusOK),
		Problems: []httpx.ProblemKind{
			httpx.NotFound().Described("Account not found"),
			httpx.BadRequest().Described("Invalid account ID"),
		},
		Guards:  []httpx.Guard{jwt.AccessGuard()},
		Handler: h.Delete,
	})

	return g
}
