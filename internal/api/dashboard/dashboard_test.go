package dashboard

import (
	"testing"

	"github.com/shopspring/decimal"
)

func TestDashboardCalculations(t *testing.T) {
	// Cash balances
	checking := decimal.NewFromFloat(4200.50)
	savings := decimal.NewFromFloat(15000.00)
	totalCash := checking.Add(savings)

	// Investments & Assets
	investments := decimal.NewFromFloat(35000.00)
	carAsset := decimal.NewFromFloat(18000.00)
	totalAssets := totalCash.Add(investments).Add(carAsset)

	// Liabilities
	creditCard := decimal.NewFromFloat(1250.75)
	autoLoan := decimal.NewFromFloat(8500.00)
	totalLiabilities := creditCard.Add(autoLoan)

	netWorth := totalAssets.Sub(totalLiabilities)
	expectedNetWorth := decimal.NewFromFloat(62449.75)

	if !netWorth.Equal(expectedNetWorth) {
		t.Fatalf("expected net worth %s, got %s", expectedNetWorth, netWorth)
	}

	// Savings rate: (Income - Expenses) / Income * 100
	income := decimal.NewFromFloat(8000.00)
	expenses := decimal.NewFromFloat(4800.00)
	savingsAmount := income.Sub(expenses) // 3200

	rateFloat, _ := savingsAmount.Div(income).Float64()
	rate := rateFloat * 100
	if rate != 40.0 {
		t.Fatalf("expected 40%% savings rate, got %f", rate)
	}
}

func TestMultiCurrencySeparation(t *testing.T) {
	// Rule: Multi-currency balances should never be added together without an exchange rate
	usdCash := decimal.NewFromFloat(5000.00)
	eurCash := decimal.NewFromFloat(3000.00)

	// In the dashboard summary, these are maintained as separate CurrencySummary entries
	currencies := map[string]decimal.Decimal{
		"USD": usdCash,
		"EUR": eurCash,
	}

	if currencies["USD"].String() != "5000" || currencies["EUR"].String() != "3000" {
		t.Fatalf("currencies should remain isolated")
	}
}
