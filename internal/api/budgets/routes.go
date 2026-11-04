package budgets

import (
	"net/http"

	"github.com/rpmoulton/mint-alternative/internal/httpx"
	"github.com/rpmoulton/mint-alternative/internal/jwt"
)

func (h *Handler) Routes() *httpx.Group {
	g := httpx.NewGroup(httpx.Defaults{Tags: []string{"budgets"}})

	httpx.Register(g, httpx.Operation[ListBudgetsInput, []BudgetResponse]{
		Method:             http.MethodGet,
		Path:               "/",
		Summary:            "List user budgets",
		SuccessDescription: "List of monthly budgets",
		Success:            httpx.Enveloped(http.StatusOK),
		Guards:             []httpx.Guard{jwt.AccessGuard()},
		Handler:            h.List,
	})

	httpx.Register(g, httpx.Operation[GetBudgetInput, BudgetResponse]{
		Method:             http.MethodGet,
		Path:               "/{id}",
		Summary:            "Get budget by ID with spending details",
		SuccessDescription: "Budget details with category allocations and actual spending",
		Success:            httpx.Enveloped(http.StatusOK),
		Problems: []httpx.ProblemKind{
			httpx.NotFound().Described("Budget not found"),
			httpx.BadRequest().Described("Invalid budget ID"),
		},
		Guards:  []httpx.Guard{jwt.AccessGuard()},
		Handler: h.Get,
	})

	httpx.Register(g, httpx.Operation[CreateBudgetInput, BudgetResponse]{
		Method:             http.MethodPost,
		Path:               "/",
		Summary:            "Create a monthly budget",
		BodyDescription:    "Budget creation payload",
		SuccessDescription: "Budget created successfully",
		Success:            httpx.Enveloped(http.StatusCreated),
		Problems: []httpx.ProblemKind{
			httpx.Conflict().Described("A budget for this month already exists"),
			httpx.BadRequest().Described("Invalid budget parameters"),
		},
		Guards:  []httpx.Guard{jwt.AccessGuard()},
		Handler: h.Create,
	})

	httpx.Register(g, httpx.Operation[UpdateBudgetInput, BudgetResponse]{
		Method:             http.MethodPatch,
		Path:               "/{id}",
		Summary:            "Update budget name",
		BodyDescription:    "Budget update payload",
		SuccessDescription: "Budget updated successfully",
		Success:            httpx.Enveloped(http.StatusOK),
		Problems: []httpx.ProblemKind{
			httpx.NotFound().Described("Budget not found"),
			httpx.BadRequest().Described("Invalid budget parameters"),
		},
		Guards:  []httpx.Guard{jwt.AccessGuard()},
		Handler: h.Update,
	})

	httpx.Register(g, httpx.Operation[DeleteBudgetInput, httpx.NoBody]{
		Method:             http.MethodDelete,
		Path:               "/{id}",
		Summary:            "Delete a budget",
		SuccessDescription: "Budget deleted",
		Success:            httpx.Message(http.StatusOK),
		Problems: []httpx.ProblemKind{
			httpx.NotFound().Described("Budget not found"),
			httpx.BadRequest().Described("Invalid budget ID"),
		},
		Guards:  []httpx.Guard{jwt.AccessGuard()},
		Handler: h.Delete,
	})

	httpx.Register(g, httpx.Operation[UpsertBudgetCategoryInput, BudgetResponse]{
		Method:             http.MethodPost,
		Path:               "/{id}/categories",
		Summary:            "Upsert a budget category allocation",
		BodyDescription:    "Category allocation details",
		SuccessDescription: "Updated budget returned",
		Success:            httpx.Enveloped(http.StatusOK),
		Problems: []httpx.ProblemKind{
			httpx.NotFound().Described("Budget not found"),
			httpx.BadRequest().Described("Invalid allocation parameters"),
		},
		Guards:  []httpx.Guard{jwt.AccessGuard()},
		Handler: h.UpsertCategory,
	})

	httpx.Register(g, httpx.Operation[DeleteBudgetCategoryInput, BudgetResponse]{
		Method:             http.MethodDelete,
		Path:               "/{id}/categories/{category_id}",
		Summary:            "Remove a budget category allocation",
		SuccessDescription: "Updated budget returned",
		Success:            httpx.Enveloped(http.StatusOK),
		Problems: []httpx.ProblemKind{
			httpx.NotFound().Described("Budget not found"),
			httpx.BadRequest().Described("Invalid IDs"),
		},
		Guards:  []httpx.Guard{jwt.AccessGuard()},
		Handler: h.DeleteCategory,
	})

	return g
}
