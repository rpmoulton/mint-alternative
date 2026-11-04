package categories

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/rpmoulton/mint-alternative/internal/api/common"
	"github.com/rpmoulton/mint-alternative/internal/db"
	"github.com/rpmoulton/mint-alternative/internal/db/sqlc"
	"github.com/rpmoulton/mint-alternative/internal/httpx"
)

type Service interface {
	List(ctx context.Context, userID string) ([]CategoryResponse, error)
	Get(ctx context.Context, id, userID string) (*CategoryResponse, error)
	Create(ctx context.Context, userID string, req *CreateCategoryRequest) (*CategoryResponse, error)
	Update(ctx context.Context, id, userID string, req *UpdateCategoryRequest) (*CategoryResponse, error)
	Delete(ctx context.Context, id, userID string) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) List(ctx context.Context, userID string) ([]CategoryResponse, error) {
	uid, err := common.StringToUUID(userID)
	if err != nil {
		return nil, httpx.NewBadRequestError("invalid user ID")
	}

	cats, err := s.repo.ListCategories(ctx, uid)
	if err != nil {
		slog.ErrorContext(ctx, "failed to list categories", "error", err)
		return nil, httpx.NewInternalError("could not list categories")
	}

	result := make([]CategoryResponse, len(cats))
	for i, c := range cats {
		result[i] = categoryResponse(c)
	}
	return result, nil
}

func (s *service) Get(ctx context.Context, id, userID string) (*CategoryResponse, error) {
	catID, err := common.StringToUUID(id)
	if err != nil {
		return nil, httpx.NewBadRequestError("invalid category ID")
	}
	uid, err := common.StringToUUID(userID)
	if err != nil {
		return nil, httpx.NewBadRequestError("invalid user ID")
	}

	cat, err := s.repo.GetCategoryByID(ctx, catID, uid)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get category", "id", id, "error", err)
		if db.IsNotFoundError(db.WrapDBError(ctx, err)) {
			return nil, httpx.NewNotFoundError("category not found")
		}
		return nil, httpx.NewInternalError("could not retrieve category")
	}

	resp := categoryResponse(*cat)
	return &resp, nil
}

func (s *service) Create(ctx context.Context, userID string, req *CreateCategoryRequest) (*CategoryResponse, error) {
	uid, err := common.StringToUUID(userID)
	if err != nil {
		return nil, httpx.NewBadRequestError("invalid user ID")
	}

	var parentID pgtype.UUID
	if req.ParentCategoryID != nil && *req.ParentCategoryID != "" {
		parentID, err = common.StringToUUID(*req.ParentCategoryID)
		if err != nil {
			return nil, httpx.NewBadRequestError("invalid parent category ID")
		}
	}

	cat, err := s.repo.CreateCategory(ctx, sqlc.CreateCategoryParams{
		UserID:           uid,
		ParentCategoryID: parentID,
		Name:             req.Name,
		Type:             req.Type,
		Icon:             req.Icon,
		Color:            req.Color,
	})
	if err != nil {
		slog.ErrorContext(ctx, "failed to create category", "error", err)
		return nil, httpx.NewInternalError("could not create category")
	}

	resp := categoryResponse(*cat)
	return &resp, nil
}

func (s *service) Update(ctx context.Context, id, userID string, req *UpdateCategoryRequest) (*CategoryResponse, error) {
	catID, err := common.StringToUUID(id)
	if err != nil {
		return nil, httpx.NewBadRequestError("invalid category ID")
	}
	uid, err := common.StringToUUID(userID)
	if err != nil {
		return nil, httpx.NewBadRequestError("invalid user ID")
	}

	// First verify it's not a system category
	existing, err := s.repo.GetCategoryByID(ctx, catID, uid)
	if err != nil {
		if db.IsNotFoundError(db.WrapDBError(ctx, err)) {
			return nil, httpx.NewNotFoundError("category not found")
		}
		return nil, httpx.NewInternalError("could not check category")
	}
	if !existing.UserID.Valid {
		return nil, httpx.NewForbiddenError("system categories cannot be modified")
	}

	var parentID pgtype.UUID
	if req.ParentCategoryID != nil && *req.ParentCategoryID != "" {
		parentID, err = common.StringToUUID(*req.ParentCategoryID)
		if err != nil {
			return nil, httpx.NewBadRequestError("invalid parent category ID")
		}
	}

	updated, err := s.repo.UpdateCategory(ctx, sqlc.UpdateCategoryParams{
		ID:               catID,
		UserID:           uid,
		Name:             req.Name,
		Type:             req.Type,
		Icon:             req.Icon,
		Color:            req.Color,
		ParentCategoryID: parentID,
	})
	if err != nil {
		slog.ErrorContext(ctx, "failed to update category", "id", id, "error", err)
		return nil, httpx.NewInternalError("could not update category")
	}

	resp := categoryResponse(*updated)
	return &resp, nil
}

func (s *service) Delete(ctx context.Context, id, userID string) error {
	catID, err := common.StringToUUID(id)
	if err != nil {
		return httpx.NewBadRequestError("invalid category ID")
	}
	uid, err := common.StringToUUID(userID)
	if err != nil {
		return httpx.NewBadRequestError("invalid user ID")
	}

	// First verify it's not a system category
	existing, err := s.repo.GetCategoryByID(ctx, catID, uid)
	if err != nil {
		if db.IsNotFoundError(db.WrapDBError(ctx, err)) {
			return httpx.NewNotFoundError("category not found")
		}
		return httpx.NewInternalError("could not check category")
	}
	if !existing.UserID.Valid {
		return httpx.NewForbiddenError("system categories cannot be deleted")
	}

	if err := s.repo.DeleteCategory(ctx, catID, uid); err != nil {
		slog.ErrorContext(ctx, "failed to delete category", "id", id, "error", err)
		return httpx.NewInternalError("could not delete category")
	}
	return nil
}
