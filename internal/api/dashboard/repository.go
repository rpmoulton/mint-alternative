package dashboard

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/rpmoulton/mint-alternative/internal/db"
	"github.com/rpmoulton/mint-alternative/internal/db/sqlc"
)

type Repository interface {
	GetDashboardAccountsSummary(ctx context.Context, userID pgtype.UUID) ([]sqlc.GetDashboardAccountsSummaryRow, error)
	GetMonthlyIncomeExpenses(ctx context.Context, params sqlc.GetMonthlyIncomeExpensesParams) (sqlc.GetMonthlyIncomeExpensesRow, error)
	GetMonthlySpendingByCategory(ctx context.Context, params sqlc.GetMonthlySpendingByCategoryParams) ([]sqlc.GetMonthlySpendingByCategoryRow, error)
	GetRecentTransactions(ctx context.Context, params sqlc.GetRecentTransactionsParams) ([]sqlc.GetRecentTransactionsRow, error)
	GetUserByID(ctx context.Context, id pgtype.UUID) (*sqlc.User, error)
}

type repository struct {
	query *sqlc.Queries
}

func NewRepository(pg db.DB) Repository {
	return &repository{
		query: pg.RootStore().Queries(),
	}
}

func (r *repository) GetDashboardAccountsSummary(ctx context.Context, userID pgtype.UUID) ([]sqlc.GetDashboardAccountsSummaryRow, error) {
	rows, err := r.query.GetDashboardAccountsSummary(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get accounts summary: %w", err)
	}
	return rows, nil
}

func (r *repository) GetMonthlyIncomeExpenses(ctx context.Context, params sqlc.GetMonthlyIncomeExpensesParams) (sqlc.GetMonthlyIncomeExpensesRow, error) {
	row, err := r.query.GetMonthlyIncomeExpenses(ctx, params)
	if err != nil {
		return sqlc.GetMonthlyIncomeExpensesRow{}, fmt.Errorf("get income/expenses: %w", err)
	}
	return row, nil
}

func (r *repository) GetMonthlySpendingByCategory(ctx context.Context, params sqlc.GetMonthlySpendingByCategoryParams) ([]sqlc.GetMonthlySpendingByCategoryRow, error) {
	rows, err := r.query.GetMonthlySpendingByCategory(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("get spending by category: %w", err)
	}
	return rows, nil
}

func (r *repository) GetRecentTransactions(ctx context.Context, params sqlc.GetRecentTransactionsParams) ([]sqlc.GetRecentTransactionsRow, error) {
	rows, err := r.query.GetRecentTransactions(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("get recent transactions: %w", err)
	}
	return rows, nil
}

func (r *repository) GetUserByID(ctx context.Context, id pgtype.UUID) (*sqlc.User, error) {
	u, err := r.query.GetUserByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get user %v: %w", id, err)
	}
	return &u, nil
}
