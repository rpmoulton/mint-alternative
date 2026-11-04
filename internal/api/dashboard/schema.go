package dashboard

import (
	"github.com/rpmoulton/mint-alternative/internal/api/transactions"
)

type CategorySpendingResponse struct {
	CategoryID       string  `json:"category_id"`
	CategoryName     string  `json:"category_name"`
	CategoryColor    *string `json:"category_color,omitempty"`
	CategoryIcon     *string `json:"category_icon,omitempty"`
	TotalSpent       string  `json:"total_spent"`
	Percentage       float64 `json:"percentage"`
	TransactionCount int64   `json:"transaction_count"`
}

type CurrencySummary struct {
	Currency         string `json:"currency"`
	TotalCash        string `json:"total_cash"`
	TotalAssets      string `json:"total_assets"`
	TotalLiabilities string `json:"total_liabilities"`
	NetWorth         string `json:"net_worth"`
}

type DashboardSummaryResponse struct {
	PrimaryCurrency    string                              `json:"primary_currency"`
	TotalCash          string                              `json:"total_cash"`
	TotalAssets        string                              `json:"total_assets"`
	TotalLiabilities   string                              `json:"total_liabilities"`
	NetWorth           string                              `json:"net_worth"`
	MonthlyIncome      string                              `json:"monthly_income"`
	MonthlyExpenses    string                              `json:"monthly_expenses"`
	SavingsRate        float64                             `json:"savings_rate"`
	ByCurrency         []CurrencySummary                   `json:"by_currency"`
	SpendingByCategory []CategorySpendingResponse          `json:"spending_by_category"`
	RecentTransactions []transactions.TransactionResponse `json:"recent_transactions"`
}

type GetDashboardInput struct {
	Month *int16 `query:"month" validate:"omitempty,min=1,max=12"`
	Year  *int16 `query:"year"  validate:"omitempty,min=2000,max=2100"`
}
