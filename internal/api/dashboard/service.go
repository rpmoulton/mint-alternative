package dashboard

import (
	"context"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"

	"github.com/rpmoulton/mint-alternative/internal/api/common"
	"github.com/rpmoulton/mint-alternative/internal/api/transactions"
	"github.com/rpmoulton/mint-alternative/internal/db/sqlc"
	"github.com/rpmoulton/mint-alternative/internal/httpx"
)

type Service interface {
	GetSummary(ctx context.Context, userID string, input *GetDashboardInput) (*DashboardSummaryResponse, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) GetSummary(ctx context.Context, userID string, input *GetDashboardInput) (*DashboardSummaryResponse, error) {
	uid, err := common.StringToUUID(userID)
	if err != nil {
		return nil, httpx.NewBadRequestError("invalid user ID")
	}

	user, err := s.repo.GetUserByID(ctx, uid)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get user", "error", err)
		return nil, httpx.NewInternalError("could not load user profile")
	}

	primaryCurrency := user.PreferredCurrency
	if primaryCurrency == "" {
		primaryCurrency = "USD"
	}

	// 1. Account summaries by account_type and currency
	accRows, err := s.repo.GetDashboardAccountsSummary(ctx, uid)
	if err != nil {
		slog.ErrorContext(ctx, "failed to load account summary", "error", err)
		return nil, httpx.NewInternalError("could not load account summary")
	}

	type currencyTotals struct {
		cash        decimal.Decimal
		assets      decimal.Decimal
		liabilities decimal.Decimal
	}
	currencyMap := make(map[string]*currencyTotals)

	for _, row := range accRows {
		curr := row.Currency
		if _, ok := currencyMap[curr]; !ok {
			currencyMap[curr] = &currencyTotals{}
		}
		balDec, _ := decimal.NewFromString(common.NumericToString(row.TotalBalance))

		switch row.AccountType {
		case "depository":
			currencyMap[curr].cash = currencyMap[curr].cash.Add(balDec)
			currencyMap[curr].assets = currencyMap[curr].assets.Add(balDec)
		case "investment", "other_asset":
			currencyMap[curr].assets = currencyMap[curr].assets.Add(balDec)
		case "credit", "loan", "other_liability":
			currencyMap[curr].liabilities = currencyMap[curr].liabilities.Add(balDec)
		}
	}

	var byCurrency []CurrencySummary
	primaryCash := decimal.Zero
	primaryAssets := decimal.Zero
	primaryLiabilities := decimal.Zero

	for curr, totals := range currencyMap {
		netWorth := totals.assets.Sub(totals.liabilities)
		byCurrency = append(byCurrency, CurrencySummary{
			Currency:         curr,
			TotalCash:        totals.cash.StringFixed(2),
			TotalAssets:      totals.assets.StringFixed(2),
			TotalLiabilities: totals.liabilities.StringFixed(2),
			NetWorth:         netWorth.StringFixed(2),
		})
		if curr == primaryCurrency {
			primaryCash = totals.cash
			primaryAssets = totals.assets
			primaryLiabilities = totals.liabilities
		}
	}

	// 2. Month Date Range
	now := time.Now().UTC()
	year := int(now.Year())
	month := int(now.Month())
	if input.Year != nil {
		year = int(*input.Year)
	}
	if input.Month != nil {
		month = int(*input.Month)
	}

	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 1, -1)

	// 3. Monthly Income & Expenses
	incExpRow, err := s.repo.GetMonthlyIncomeExpenses(ctx, sqlc.GetMonthlyIncomeExpensesParams{
		UserID:    uid,
		StartDate: pgtype.Date{Time: startDate, Valid: true},
		EndDate:   pgtype.Date{Time: endDate, Valid: true},
	})
	if err != nil {
		slog.ErrorContext(ctx, "failed to load income/expense summary", "error", err)
	}

	monthlyIncome, _ := decimal.NewFromString(common.NumericToString(incExpRow.TotalIncome))
	monthlyExpenses, _ := decimal.NewFromString(common.NumericToString(incExpRow.TotalExpenses))

	var savingsRate float64
	if monthlyIncome.IsPositive() {
		netSavings := monthlyIncome.Sub(monthlyExpenses)
		ratio, _ := netSavings.Div(monthlyIncome).Float64()
		savingsRate = ratio * 100
	}

	// 4. Monthly Category Spending
	catSpendingRows, err := s.repo.GetMonthlySpendingByCategory(ctx, sqlc.GetMonthlySpendingByCategoryParams{
		UserID:    uid,
		StartDate: pgtype.Date{Time: startDate, Valid: true},
		EndDate:   pgtype.Date{Time: endDate, Valid: true},
	})
	if err != nil {
		slog.ErrorContext(ctx, "failed to load category spending", "error", err)
	}

	var spendingByCategory []CategorySpendingResponse
	for _, csr := range catSpendingRows {
		spentDec, _ := decimal.NewFromString(common.NumericToString(csr.TotalSpent))
		var pct float64
		if monthlyExpenses.IsPositive() {
			ratio, _ := spentDec.Div(monthlyExpenses).Float64()
			pct = ratio * 100
		}
		spendingByCategory = append(spendingByCategory, CategorySpendingResponse{
			CategoryID:       common.UUIDToString(csr.CategoryID),
			CategoryName:     csr.CategoryName,
			CategoryColor:    csr.CategoryColor,
			CategoryIcon:     csr.CategoryIcon,
			TotalSpent:       spentDec.StringFixed(2),
			Percentage:       pct,
			TransactionCount: csr.TransactionCount,
		})
	}

	// 5. Recent Transactions
	recentRows, err := s.repo.GetRecentTransactions(ctx, sqlc.GetRecentTransactionsParams{
		UserID:   uid,
		LimitVal: 8,
	})
	if err != nil {
		slog.ErrorContext(ctx, "failed to load recent transactions", "error", err)
	}

	recentTransactions := make([]transactions.TransactionResponse, len(recentRows))
	for i, r := range recentRows {
		var catID *string
		if r.CategoryID.Valid {
			s := common.UUIDToString(r.CategoryID)
			catID = &s
		}
		var postedDate *string
		if r.PostedDate.Valid {
			s := common.DateToString(r.PostedDate)
			postedDate = &s
		}
		recentTransactions[i] = transactions.TransactionResponse{
			ID:                    common.UUIDToString(r.ID),
			UserID:                common.UUIDToString(r.UserID),
			AccountID:             common.UUIDToString(r.AccountID),
			AccountName:           r.AccountName,
			CategoryID:            catID,
			CategoryName:          r.CategoryName,
			CategoryColor:         r.CategoryColor,
			CategoryIcon:          r.CategoryIcon,
			CategoryType:          r.CategoryType,
			Amount:                common.NumericToString(r.Amount),
			Currency:              r.Currency,
			TransactionDate:       common.DateToString(r.TransactionDate),
			PostedDate:            postedDate,
			MerchantName:          r.MerchantName,
			Description:           r.Description,
			Notes:                 r.Notes,
			TransactionType:       r.TransactionType,
			Status:                r.Status,
			SourceType:            r.SourceType,
			ExternalTransactionID: r.ExternalTransactionID,
			CreatedAt:             r.CreatedAt.Time,
			UpdatedAt:             r.UpdatedAt.Time,
		}
	}

	primaryNetWorth := primaryAssets.Sub(primaryLiabilities)

	return &DashboardSummaryResponse{
		PrimaryCurrency:    primaryCurrency,
		TotalCash:          primaryCash.StringFixed(2),
		TotalAssets:        primaryAssets.StringFixed(2),
		TotalLiabilities:   primaryLiabilities.StringFixed(2),
		NetWorth:           primaryNetWorth.StringFixed(2),
		MonthlyIncome:      monthlyIncome.StringFixed(2),
		MonthlyExpenses:    monthlyExpenses.StringFixed(2),
		SavingsRate:        savingsRate,
		ByCurrency:         byCurrency,
		SpendingByCategory: spendingByCategory,
		RecentTransactions: recentTransactions,
	}, nil
}
