package categories

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/rpmoulton/mint-alternative/internal/db"
	"github.com/rpmoulton/mint-alternative/internal/db/sqlc"
)

type Repository interface {
	ListCategories(ctx context.Context, userID pgtype.UUID) ([]sqlc.Category, error)
	GetCategoryByID(ctx context.Context, id, userID pgtype.UUID) (*sqlc.Category, error)
	CreateCategory(ctx context.Context, params sqlc.CreateCategoryParams) (*sqlc.Category, error)
	UpdateCategory(ctx context.Context, params sqlc.UpdateCategoryParams) (*sqlc.Category, error)
	DeleteCategory(ctx context.Context, id, userID pgtype.UUID) error
}

type repository struct {
	query *sqlc.Queries
}

func NewRepository(pg db.DB) Repository {
	return &repository{
		query: pg.RootStore().Queries(),
	}
}

func (r *repository) ListCategories(ctx context.Context, userID pgtype.UUID) ([]sqlc.Category, error) {
	cats, err := r.query.ListCategories(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list categories: %w", err)
	}
	return cats, nil
}

func (r *repository) GetCategoryByID(ctx context.Context, id, userID pgtype.UUID) (*sqlc.Category, error) {
	cat, err := r.query.GetCategoryByID(ctx, sqlc.GetCategoryByIDParams{
		ID:     id,
		UserID: userID,
	})
	if err != nil {
		return nil, fmt.Errorf("get category %v: %w", id, err)
	}
	return &cat, nil
}

func (r *repository) CreateCategory(ctx context.Context, params sqlc.CreateCategoryParams) (*sqlc.Category, error) {
	cat, err := r.query.CreateCategory(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("create category: %w", err)
	}
	return &cat, nil
}

func (r *repository) UpdateCategory(ctx context.Context, params sqlc.UpdateCategoryParams) (*sqlc.Category, error) {
	cat, err := r.query.UpdateCategory(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("update category: %w", err)
	}
	return &cat, nil
}

func (r *repository) DeleteCategory(ctx context.Context, id, userID pgtype.UUID) error {
	err := r.query.DeleteCategory(ctx, sqlc.DeleteCategoryParams{
		ID:     id,
		UserID: userID,
	})
	if err != nil {
		return fmt.Errorf("delete category: %w", err)
	}
	return nil
}
