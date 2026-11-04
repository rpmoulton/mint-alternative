package budgets

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"

	"github.com/rpmoulton/mint-alternative/internal/api/common"
	"github.com/rpmoulton/mint-alternative/internal/db"
	"github.com/rpmoulton/mint-alternative/internal/db/sqlc"
	"github.com/rpmoulton/mint-alternative/internal/httpx"
)

type Service interface {
	List(ctx context.Context, userID string, input *ListBudgetsInput) ([]BudgetResponse, error)
	Get(ctx context.Context, id, userID string) (*BudgetResponse, error)
	Create(ctx context.Context, userID string, req *CreateBudgetRequest) (*BudgetResponse, error)
	Update(ctx context.Context, id, userID string, req *UpdateBudgetRequest) (*BudgetResponse, error)
	Delete(ctx context.Context, id, userID string) error
	UpsertCategory(ctx context.Context, id, userID string, req *UpsertBudgetCategoryRequest) (*BudgetResponse, error)
	DeleteCategory(ctx context.Context, id, userID, categoryID string) (*BudgetResponse, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) List(ctx context.Context, userID string, input *ListBudgetsInput) ([]BudgetResponse, error) {
	uid, err := common.StringToUUID(userID)
	if err != nil {
		return nil, httpx.NewBadRequestError("invalid user ID")
	}

	budgets, err := s.repo.ListBudgets(ctx, uid)
	if err != nil {
		slog.ErrorContext(ctx, "failed to list budgets", "error", err)
		return nil, httpx.NewInternalError("could not list budgets")
	}

	result := make([]BudgetResponse, 0, len(budgets))
	for _, b := range budgets {
		if input.Year != nil && b.Year != *input.Year {
			continue
		}
		if input.Month != nil && b.Month != *input.Month {
			continue
		}
		full, err := s.Get(ctx, common.UUIDToString(b.ID), userID)
		if err == nil && full != nil {
			result = append(result, *full)
		} else {
			result = append(result, budgetSummaryResponse(b))
		}
	}
	return result, nil
}

func (s *service) Get(ctx context.Context, id, userID string) (*BudgetResponse, error) {
	bID, err := common.StringToUUID(id)
	if err != nil {
		return nil, httpx.NewBadRequestError("invalid budget ID")
	}
	uid, err := common.StringToUUID(userID)
	if err != nil {
		return nil, httpx.NewBadRequestError("invalid user ID")
	}

	budget, err := s.repo.GetBudgetByID(ctx, bID, uid)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get budget", "id", id, "error", err)
		if db.IsNotFoundError(db.WrapDBError(ctx, err)) {
			return nil, httpx.NewNotFoundError("budget not found")
		}
		return nil, httpx.NewInternalError("could not retrieve budget")
	}

	catAllocations, err := s.repo.ListBudgetCategories(ctx, bID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to list budget categories", "budget_id", id, "error", err)
		return nil, httpx.NewInternalError("could not retrieve budget categories")
	}

	// Calculate date range for the budget month
	startDate := time.Date(int(budget.Year), time.Month(budget.Month), 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 1, -1)

	spendingRows, err := s.repo.GetMonthlyCategorySpending(ctx, sqlc.GetMonthlySpendingByCategoryParams{
		UserID:    uid,
		StartDate: pgtype.Date{Time: startDate, Valid: true},
		EndDate:   pgtype.Date{Time: endDate, Valid: true},
	})
	if err != nil {
		slog.WarnContext(ctx, "failed to get monthly category spending for budget", "error", err)
	}

	spendingByCat := make(map[string]decimal.Decimal)
	for _, sr := range spendingRows {
		catStr := common.UUIDToString(sr.CategoryID)
		dec, _ := decimal.NewFromString(common.NumericToString(sr.TotalSpent))
		spendingByCat[catStr] = dec
	}

	totalAllocated := decimal.Zero
	totalSpent := decimal.Zero

	catsResp := make([]BudgetCategoryResponse, len(catAllocations))
	for i, ca := range catAllocations {
		catIDStr := common.UUIDToString(ca.CategoryID)
		allocatedDec, _ := decimal.NewFromString(common.NumericToString(ca.AllocatedAmount))
		totalAllocated = totalAllocated.Add(allocatedDec)

		spentDec := spendingByCat[catIDStr]
		totalSpent = totalSpent.Add(spentDec)

		remainingDec := allocatedDec.Sub(spentDec)
		isOver := spentDec.GreaterThan(allocatedDec)

		var percentUsed float64
		if allocatedDec.IsPositive() {
			ratio, _ := spentDec.Div(allocatedDec).Float64()
			percentUsed = ratio * 100
		} else if spentDec.IsPositive() {
			percentUsed = 100
		}

		catsResp[i] = BudgetCategoryResponse{
			ID:              common.UUIDToString(ca.ID),
			BudgetID:        common.UUIDToString(ca.BudgetID),
			CategoryID:      catIDStr,
			CategoryName:    ca.CategoryName,
			CategoryType:    ca.CategoryType,
			CategoryIcon:    ca.CategoryIcon,
			CategoryColor:   ca.CategoryColor,
			AllocatedAmount: allocatedDec.StringFixed(2),
			SpentAmount:     spentDec.StringFixed(2),
			RemainingAmount: remainingDec.StringFixed(2),
			PercentUsed:     percentUsed,
			IsOverBudget:    isOver,
		}
	}

	return &BudgetResponse{
		ID:             common.UUIDToString(budget.ID),
		UserID:         common.UUIDToString(budget.UserID),
		Name:           budget.Name,
		Month:          budget.Month,
		Year:           budget.Year,
		TotalAllocated: totalAllocated.StringFixed(2),
		TotalSpent:     totalSpent.StringFixed(2),
		Categories:     catsResp,
		CreatedAt:      budget.CreatedAt.Time,
		UpdatedAt:      budget.UpdatedAt.Time,
	}, nil
}

func (s *service) Create(ctx context.Context, userID string, req *CreateBudgetRequest) (*BudgetResponse, error) {
	uid, err := common.StringToUUID(userID)
	if err != nil {
		return nil, httpx.NewBadRequestError("invalid user ID")
	}

	existing, _ := s.repo.GetBudgetByMonth(ctx, uid, req.Year, req.Month)
	if existing != nil {
		return nil, httpx.NewConflictError(fmt.Sprintf("A budget already exists for %d-%02d", req.Year, req.Month))
	}

	created, err := s.repo.CreateBudget(ctx, sqlc.CreateBudgetParams{
		UserID: uid,
		Name:   req.Name,
		Month:  req.Month,
		Year:   req.Year,
	})
	if err != nil {
		slog.ErrorContext(ctx, "failed to create budget", "error", err)
		return nil, httpx.NewInternalError("could not create budget")
	}

	return s.Get(ctx, common.UUIDToString(created.ID), userID)
}

func (s *service) Update(ctx context.Context, id, userID string, req *UpdateBudgetRequest) (*BudgetResponse, error) {
	bID, err := common.StringToUUID(id)
	if err != nil {
		return nil, httpx.NewBadRequestError("invalid budget ID")
	}
	uid, err := common.StringToUUID(userID)
	if err != nil {
		return nil, httpx.NewBadRequestError("invalid user ID")
	}

	_, err = s.repo.UpdateBudget(ctx, sqlc.UpdateBudgetParams{
		ID:     bID,
		UserID: uid,
		Name:   &req.Name,
	})
	if err != nil {
		slog.ErrorContext(ctx, "failed to update budget", "id", id, "error", err)
		if db.IsNotFoundError(db.WrapDBError(ctx, err)) {
			return nil, httpx.NewNotFoundError("budget not found")
		}
		return nil, httpx.NewInternalError("could not update budget")
	}

	return s.Get(ctx, id, userID)
}

func (s *service) Delete(ctx context.Context, id, userID string) error {
	bID, err := common.StringToUUID(id)
	if err != nil {
		return httpx.NewBadRequestError("invalid budget ID")
	}
	uid, err := common.StringToUUID(userID)
	if err != nil {
		return httpx.NewBadRequestError("invalid user ID")
	}

	if err := s.repo.DeleteBudget(ctx, bID, uid); err != nil {
		slog.ErrorContext(ctx, "failed to delete budget", "id", id, "error", err)
		return httpx.NewInternalError("could not delete budget")
	}
	return nil
}

func (s *service) UpsertCategory(ctx context.Context, id, userID string, req *UpsertBudgetCategoryRequest) (*BudgetResponse, error) {
	bID, err := common.StringToUUID(id)
	if err != nil {
		return nil, httpx.NewBadRequestError("invalid budget ID")
	}
	uid, err := common.StringToUUID(userID)
	if err != nil {
		return nil, httpx.NewBadRequestError("invalid user ID")
	}

	// Verify budget ownership
	_, err = s.repo.GetBudgetByID(ctx, bID, uid)
	if err != nil {
		return nil, httpx.NewNotFoundError("budget not found")
	}

	catID, err := common.StringToUUID(req.CategoryID)
	if err != nil {
		return nil, httpx.NewBadRequestError("invalid category ID")
	}

	allocDec, err := decimal.NewFromString(req.AllocatedAmount)
	if err != nil || allocDec.IsNegative() {
		return nil, httpx.NewBadRequestError("allocated amount must be a non-negative number")
	}

	allocNumeric, _ := common.StringToNumeric(allocDec.StringFixed(4))

	_, err = s.repo.UpsertBudgetCategory(ctx, sqlc.UpsertBudgetCategoryParams{
		BudgetID:        bID,
		CategoryID:      catID,
		AllocatedAmount: allocNumeric,
	})
	if err != nil {
		slog.ErrorContext(ctx, "failed to upsert budget category", "error", err)
		return nil, httpx.NewInternalError("could not save budget allocation")
	}

	return s.Get(ctx, id, userID)
}

func (s *service) DeleteCategory(ctx context.Context, id, userID, categoryID string) (*BudgetResponse, error) {
	bID, err := common.StringToUUID(id)
	if err != nil {
		return nil, httpx.NewBadRequestError("invalid budget ID")
	}
	uid, err := common.StringToUUID(userID)
	if err != nil {
		return nil, httpx.NewBadRequestError("invalid user ID")
	}

	// Verify budget ownership
	_, err = s.repo.GetBudgetByID(ctx, bID, uid)
	if err != nil {
		return nil, httpx.NewNotFoundError("budget not found")
	}

	catID, err := common.StringToUUID(categoryID)
	if err != nil {
		return nil, httpx.NewBadRequestError("invalid category ID")
	}

	if err := s.repo.DeleteBudgetCategory(ctx, bID, catID); err != nil {
		slog.ErrorContext(ctx, "failed to delete budget category", "error", err)
		return nil, httpx.NewInternalError("could not delete budget allocation")
	}

	return s.Get(ctx, id, userID)
}
