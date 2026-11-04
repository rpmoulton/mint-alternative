package budgets

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/rpmoulton/mint-alternative/internal/db"
	"github.com/rpmoulton/mint-alternative/internal/db/sqlc"
)

type Repository interface {
	ListBudgets(ctx context.Context, userID pgtype.UUID) ([]sqlc.Budget, error)
	GetBudgetByID(ctx context.Context, id, userID pgtype.UUID) (*sqlc.Budget, error)
	GetBudgetByMonth(ctx context.Context, userID pgtype.UUID, year, month int16) (*sqlc.Budget, error)
	CreateBudget(ctx context.Context, params sqlc.CreateBudgetParams) (*sqlc.Budget, error)
	UpdateBudget(ctx context.Context, params sqlc.UpdateBudgetParams) (*sqlc.Budget, error)
	DeleteBudget(ctx context.Context, id, userID pgtype.UUID) error
	ListBudgetCategories(ctx context.Context, budgetID pgtype.UUID) ([]sqlc.ListBudgetCategoriesRow, error)
	UpsertBudgetCategory(ctx context.Context, params sqlc.UpsertBudgetCategoryParams) (*sqlc.BudgetCategory, error)
	DeleteBudgetCategory(ctx context.Context, budgetID, categoryID pgtype.UUID) error
	GetMonthlyCategorySpending(ctx context.Context, params sqlc.GetMonthlySpendingByCategoryParams) ([]sqlc.GetMonthlySpendingByCategoryRow, error)
}

type repository struct {
	query *sqlc.Queries
}

func NewRepository(pg db.DB) Repository {
	return &repository{
		query: pg.RootStore().Queries(),
	}
}

func (r *repository) ListBudgets(ctx context.Context, userID pgtype.UUID) ([]sqlc.Budget, error) {
	rows, err := r.query.ListBudgets(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list budgets: %w", err)
	}
	return rows, nil
}

func (r *repository) GetBudgetByID(ctx context.Context, id, userID pgtype.UUID) (*sqlc.Budget, error) {
	b, err := r.query.GetBudgetByID(ctx, sqlc.GetBudgetByIDParams{
		ID:     id,
		UserID: userID,
	})
	if err != nil {
		return nil, fmt.Errorf("get budget %v: %w", id, err)
	}
	return &b, nil
}

func (r *repository) GetBudgetByMonth(ctx context.Context, userID pgtype.UUID, year, month int16) (*sqlc.Budget, error) {
	b, err := r.query.GetBudgetByMonth(ctx, sqlc.GetBudgetByMonthParams{
		UserID: userID,
		Year:   year,
		Month:  month,
	})
	if err != nil {
		return nil, fmt.Errorf("get budget for %d-%d: %w", year, month, err)
	}
	return &b, nil
}

func (r *repository) CreateBudget(ctx context.Context, params sqlc.CreateBudgetParams) (*sqlc.Budget, error) {
	b, err := r.query.CreateBudget(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("create budget: %w", err)
	}
	return &b, nil
}

func (r *repository) UpdateBudget(ctx context.Context, params sqlc.UpdateBudgetParams) (*sqlc.Budget, error) {
	b, err := r.query.UpdateBudget(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("update budget: %w", err)
	}
	return &b, nil
}

func (r *repository) DeleteBudget(ctx context.Context, id, userID pgtype.UUID) error {
	err := r.query.DeleteBudget(ctx, sqlc.DeleteBudgetParams{
		ID:     id,
		UserID: userID,
	})
	if err != nil {
		return fmt.Errorf("delete budget %v: %w", id, err)
	}
	return nil
}

func (r *repository) ListBudgetCategories(ctx context.Context, budgetID pgtype.UUID) ([]sqlc.ListBudgetCategoriesRow, error) {
	rows, err := r.query.ListBudgetCategories(ctx, budgetID)
	if err != nil {
		return nil, fmt.Errorf("list budget categories: %w", err)
	}
	return rows, nil
}

func (r *repository) UpsertBudgetCategory(ctx context.Context, params sqlc.UpsertBudgetCategoryParams) (*sqlc.BudgetCategory, error) {
	bc, err := r.query.UpsertBudgetCategory(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("upsert budget category: %w", err)
	}
	return &bc, nil
}

func (r *repository) DeleteBudgetCategory(ctx context.Context, budgetID, categoryID pgtype.UUID) error {
	err := r.query.DeleteBudgetCategory(ctx, sqlc.DeleteBudgetCategoryParams{
		BudgetID:   budgetID,
		CategoryID: categoryID,
	})
	if err != nil {
		return fmt.Errorf("delete budget category: %w", err)
	}
	return nil
}

func (r *repository) GetMonthlyCategorySpending(ctx context.Context, params sqlc.GetMonthlySpendingByCategoryParams) ([]sqlc.GetMonthlySpendingByCategoryRow, error) {
	rows, err := r.query.GetMonthlySpendingByCategory(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("get monthly category spending: %w", err)
	}
	return rows, nil
}
