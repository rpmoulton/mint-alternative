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

		// Verify parent exists and belongs to this user or is system
		parent, err := s.repo.GetCategoryByID(ctx, parentID, uid)
		if err != nil {
			if db.IsNotFoundError(db.WrapDBError(ctx, err)) {
				return nil, httpx.NewBadRequestError("parent category not found")
			}
			return nil, httpx.NewInternalError("could not check parent category")
		}

		// Verify parent category type matches child category type
		if parent.Type != req.Type {
			return nil, httpx.NewBadRequestError("parent category type must match child category type")
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

	effectiveType := existing.Type
	if req.Type != nil {
		effectiveType = *req.Type
	}

	var parentID pgtype.UUID
	setParent := false
	if req.ParentCategoryID != nil {
		setParent = true
		if *req.ParentCategoryID != "" {
			parentID, err = common.StringToUUID(*req.ParentCategoryID)
			if err != nil {
				return nil, httpx.NewBadRequestError("invalid parent category ID")
			}

			// Cannot be its own parent
			if parentID == catID {
				return nil, httpx.NewBadRequestError("category cannot be its own parent")
			}

			// Verify parent category exists and belongs to this user or is system
			parent, err := s.repo.GetCategoryByID(ctx, parentID, uid)
			if err != nil {
				if db.IsNotFoundError(db.WrapDBError(ctx, err)) {
					return nil, httpx.NewBadRequestError("parent category not found")
				}
				return nil, httpx.NewInternalError("could not check parent category")
			}

			// Verify type compatibility
			if parent.Type != effectiveType {
				return nil, httpx.NewBadRequestError("parent category type must match child category type")
			}

			// Prevent parent-child cycles: walk up the parent tree
			currentParentID := parent.ParentCategoryID
			visited := map[pgtype.UUID]bool{catID: true, parentID: true}
			for currentParentID.Valid {
				if currentParentID == catID {
					return nil, httpx.NewBadRequestError("category hierarchy cycle detected")
				}
				if visited[currentParentID] {
					break
				}
				visited[currentParentID] = true

				ancestor, err := s.repo.GetCategoryByID(ctx, currentParentID, uid)
				if err != nil {
					break
				}
				currentParentID = ancestor.ParentCategoryID
			}
		}
	} else if req.Type != nil && existing.ParentCategoryID.Valid {
		// Parent didn't change, but type changed. Verify it matches existing parent
		parent, err := s.repo.GetCategoryByID(ctx, existing.ParentCategoryID, uid)
		if err == nil && parent.Type != effectiveType {
			return nil, httpx.NewBadRequestError("parent category type must match child category type")
		}
	}

	// If type is changing, check if this category has child categories whose type would become incompatible
	if req.Type != nil && *req.Type != existing.Type {
		childCount, err := s.repo.CountChildCategories(ctx, catID)
		if err == nil && childCount > 0 {
			return nil, httpx.NewBadRequestError("cannot change category type while subcategories exist")
		}
	}

	updated, err := s.repo.UpdateCategory(ctx, sqlc.UpdateCategoryParams{
		ID:                  catID,
		UserID:              uid,
		Name:                req.Name,
		Type:                req.Type,
		Icon:                req.Icon,
		Color:               req.Color,
		SetParentCategoryID: setParent,
		ParentCategoryID:    parentID,
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

	// Check if category is referenced by any transactions
	txCount, err := s.repo.CountTransactionsByCategoryID(ctx, catID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to check category transactions", "id", id, "error", err)
		return httpx.NewInternalError("could not check category usage")
	}
	if txCount > 0 {
		return httpx.NewBadRequestError("cannot delete category referenced by existing transactions")
	}

	// Check if category has subcategories
	childCount, err := s.repo.CountChildCategories(ctx, catID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to check child categories", "id", id, "error", err)
		return httpx.NewInternalError("could not check category hierarchy")
	}
	if childCount > 0 {
		return httpx.NewBadRequestError("cannot delete category that has subcategories")
	}

	if err := s.repo.DeleteCategory(ctx, catID, uid); err != nil {
		slog.ErrorContext(ctx, "failed to delete category", "id", id, "error", err)
		return httpx.NewInternalError("could not delete category")
	}
	return nil
}
