package transactions

import (
	"net/http"

	"github.com/rpmoulton/mint-alternative/internal/httpx"
	"github.com/rpmoulton/mint-alternative/internal/jwt"
)

func (h *Handler) Routes() *httpx.Group {
	g := httpx.NewGroup(httpx.Defaults{Tags: []string{"transactions"}})

	httpx.Register(g, httpx.Operation[ListTransactionsInput, []TransactionResponse]{
		Method:             http.MethodGet,
		Path:               "/",
		Summary:            "List user transactions with filtering and pagination",
		SuccessDescription: "Page of transactions with pagination meta",
		Success:            httpx.PageOf(http.StatusOK),
		Problems: []httpx.ProblemKind{
			httpx.BadRequest().Described("Invalid filter parameters"),
		},
		Guards:  []httpx.Guard{jwt.AccessGuard()},
		Handler: h.List,
	})

	httpx.Register(g, httpx.Operation[GetTransactionInput, TransactionResponse]{
		Method:             http.MethodGet,
		Path:               "/{id}",
		Summary:            "Get transaction by ID",
		SuccessDescription: "Transaction details",
		Success:            httpx.Enveloped(http.StatusOK),
		Problems: []httpx.ProblemKind{
			httpx.NotFound().Described("Transaction not found"),
			httpx.BadRequest().Described("Invalid transaction ID"),
		},
		Guards:  []httpx.Guard{jwt.AccessGuard()},
		Handler: h.Get,
	})

	httpx.Register(g, httpx.Operation[CreateTransactionInput, TransactionResponse]{
		Method:             http.MethodPost,
		Path:               "/",
		Summary:            "Create a transaction",
		BodyDescription:    "Transaction creation payload",
		SuccessDescription: "Transaction created successfully",
		Success:            httpx.Enveloped(http.StatusCreated),
		Problems: []httpx.ProblemKind{
			httpx.BadRequest().Described("Invalid transaction parameters"),
		},
		Guards:  []httpx.Guard{jwt.AccessGuard()},
		Handler: h.Create,
	})

	httpx.Register(g, httpx.Operation[UpdateTransactionInput, TransactionResponse]{
		Method:             http.MethodPatch,
		Path:               "/{id}",
		Summary:            "Update a transaction",
		BodyDescription:    "Transaction update payload",
		SuccessDescription: "Transaction updated successfully",
		Success:            httpx.Enveloped(http.StatusOK),
		Problems: []httpx.ProblemKind{
			httpx.NotFound().Described("Transaction not found"),
			httpx.BadRequest().Described("Invalid transaction parameters"),
		},
		Guards:  []httpx.Guard{jwt.AccessGuard()},
		Handler: h.Update,
	})

	httpx.Register(g, httpx.Operation[DeleteTransactionInput, httpx.NoBody]{
		Method:             http.MethodDelete,
		Path:               "/{id}",
		Summary:            "Delete a transaction",
		SuccessDescription: "Transaction deleted",
		Success:            httpx.Message(http.StatusOK),
		Problems: []httpx.ProblemKind{
			httpx.NotFound().Described("Transaction not found"),
			httpx.BadRequest().Described("Invalid transaction ID"),
		},
		Guards:  []httpx.Guard{jwt.AccessGuard()},
		Handler: h.Delete,
	})

	return g
}
