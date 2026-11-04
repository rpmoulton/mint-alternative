package categories

import (
	"net/http"

	"github.com/rpmoulton/mint-alternative/internal/httpx"
	"github.com/rpmoulton/mint-alternative/internal/jwt"
)

func (h *Handler) Routes() *httpx.Group {
	g := httpx.NewGroup(httpx.Defaults{Tags: []string{"categories"}})

	httpx.Register(g, httpx.Operation[ListCategoriesInput, []CategoryResponse]{
		Method:             http.MethodGet,
		Path:               "/",
		Summary:            "List transaction categories",
		SuccessDescription: "List of system and user categories",
		Success:            httpx.Enveloped(http.StatusOK),
		Guards:             []httpx.Guard{jwt.AccessGuard()},
		Handler:            h.List,
	})

	httpx.Register(g, httpx.Operation[GetCategoryInput, CategoryResponse]{
		Method:             http.MethodGet,
		Path:               "/{id}",
		Summary:            "Get category by ID",
		SuccessDescription: "Category details",
		Success:            httpx.Enveloped(http.StatusOK),
		Problems: []httpx.ProblemKind{
			httpx.NotFound().Described("Category not found"),
			httpx.BadRequest().Described("Invalid category ID"),
		},
		Guards:  []httpx.Guard{jwt.AccessGuard()},
		Handler: h.Get,
	})

	httpx.Register(g, httpx.Operation[CreateCategoryInput, CategoryResponse]{
		Method:             http.MethodPost,
		Path:               "/",
		Summary:            "Create a custom category",
		BodyDescription:    "Category creation payload",
		SuccessDescription: "Category created",
		Success:            httpx.Enveloped(http.StatusCreated),
		Problems: []httpx.ProblemKind{
			httpx.BadRequest().Described("Invalid category parameters"),
		},
		Guards:  []httpx.Guard{jwt.AccessGuard()},
		Handler: h.Create,
	})

	httpx.Register(g, httpx.Operation[UpdateCategoryInput, CategoryResponse]{
		Method:             http.MethodPatch,
		Path:               "/{id}",
		Summary:            "Update a custom category",
		BodyDescription:    "Category update payload",
		SuccessDescription: "Category updated",
		Success:            httpx.Enveloped(http.StatusOK),
		Problems: []httpx.ProblemKind{
			httpx.NotFound().Described("Category not found"),
			httpx.Forbidden().Described("System categories cannot be modified"),
			httpx.BadRequest().Described("Invalid category parameters"),
		},
		Guards:  []httpx.Guard{jwt.AccessGuard()},
		Handler: h.Update,
	})

	httpx.Register(g, httpx.Operation[DeleteCategoryInput, httpx.NoBody]{
		Method:             http.MethodDelete,
		Path:               "/{id}",
		Summary:            "Delete a custom category",
		SuccessDescription: "Category deleted",
		Success:            httpx.Message(http.StatusOK),
		Problems: []httpx.ProblemKind{
			httpx.NotFound().Described("Category not found"),
			httpx.Forbidden().Described("System categories cannot be deleted"),
			httpx.BadRequest().Described("Invalid category ID"),
		},
		Guards:  []httpx.Guard{jwt.AccessGuard()},
		Handler: h.Delete,
	})

	return g
}
