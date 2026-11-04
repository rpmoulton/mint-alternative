package budgets

import (
	"time"

	"github.com/rpmoulton/mint-alternative/internal/api/common"
	"github.com/rpmoulton/mint-alternative/internal/db/sqlc"
)

type BudgetCategoryResponse struct {
	ID              string  `json:"id"`
	BudgetID        string  `json:"budget_id"`
	CategoryID      string  `json:"category_id"`
	CategoryName    string  `json:"category_name"`
	CategoryType    string  `json:"category_type"`
	CategoryIcon    *string `json:"category_icon,omitempty"`
	CategoryColor   *string `json:"category_color,omitempty"`
	AllocatedAmount string  `json:"allocated_amount"`
	SpentAmount     string  `json:"spent_amount"`
	RemainingAmount string  `json:"remaining_amount"`
	PercentUsed     float64 `json:"percent_used"`
	IsOverBudget    bool    `json:"is_over_budget"`
}

type BudgetResponse struct {
	ID             string                   `json:"id"`
	UserID         string                   `json:"user_id"`
	Name           string                   `json:"name"`
	Month          int16                    `json:"month"`
	Year           int16                    `json:"year"`
	TotalAllocated string                   `json:"total_allocated"`
	TotalSpent     string                   `json:"total_spent"`
	Categories     []BudgetCategoryResponse `json:"categories,omitempty"`
	CreatedAt      time.Time                `json:"created_at"`
	UpdatedAt      time.Time                `json:"updated_at"`
}

type CreateBudgetRequest struct {
	Name  string `json:"name"  validate:"required,min=1,max=100"`
	Month int16  `json:"month" validate:"required,min=1,max=12"`
	Year  int16  `json:"year"  validate:"required,min=2000,max=2100"`
}

type UpdateBudgetRequest struct {
	Name string `json:"name" validate:"required,min=1,max=100"`
}

type UpsertBudgetCategoryRequest struct {
	CategoryID      string `json:"category_id"      validate:"required,uuid"`
	AllocatedAmount string `json:"allocated_amount" validate:"required"`
}

type CreateBudgetInput struct {
	Body CreateBudgetRequest `body:"required"`
}

type UpdateBudgetInput struct {
	ID   string              `path:"id"      validate:"required,uuid"`
	Body UpdateBudgetRequest `body:"required"`
}

type GetBudgetInput struct {
	ID string `path:"id" validate:"required,uuid"`
}

type DeleteBudgetInput struct {
	ID string `path:"id" validate:"required,uuid"`
}

type UpsertBudgetCategoryInput struct {
	ID   string                      `path:"id"      validate:"required,uuid"`
	Body UpsertBudgetCategoryRequest `body:"required"`
}

type DeleteBudgetCategoryInput struct {
	ID         string `path:"id"          validate:"required,uuid"`
	CategoryID string `path:"category_id" validate:"required,uuid"`
}

type ListBudgetsInput struct {
	Year  *int16 `query:"year"  validate:"omitempty,min=2000,max=2100"`
	Month *int16 `query:"month" validate:"omitempty,min=1,max=12"`
}

func budgetSummaryResponse(b sqlc.Budget) BudgetResponse {
	return BudgetResponse{
		ID:             common.UUIDToString(b.ID),
		UserID:         common.UUIDToString(b.UserID),
		Name:           b.Name,
		Month:          b.Month,
		Year:           b.Year,
		TotalAllocated: "0.00",
		TotalSpent:     "0.00",
		Categories:     []BudgetCategoryResponse{},
		CreatedAt:      b.CreatedAt.Time,
		UpdatedAt:      b.UpdatedAt.Time,
	}
}
