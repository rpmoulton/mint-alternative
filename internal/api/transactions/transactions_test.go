package transactions

import (
	"testing"

	"github.com/shopspring/decimal"
)

func TestFinancialCalculationConventions(t *testing.T) {
	// Rule: Income & Credits are positive, Spending & Debits are negative.
	// For Depository/Asset accounts:
	//   Current balance = Balance + Amount
	//   So +$1000 income increases checking account by $1000.
	//   And -$50 groceries decreases checking account by $50.

	assetBalance := decimal.NewFromFloat(1500.00)
	incomeTx := decimal.NewFromFloat(2000.00)
	expenseTx := decimal.NewFromFloat(-75.50)

	newAssetBal := assetBalance.Add(incomeTx)
	if !newAssetBal.Equal(decimal.NewFromFloat(3500.00)) {
		t.Fatalf("expected asset balance 3500.00, got %s", newAssetBal)
	}

	newAssetBal = newAssetBal.Add(expenseTx)
	if !newAssetBal.Equal(decimal.NewFromFloat(3424.50)) {
		t.Fatalf("expected asset balance 3424.50, got %s", newAssetBal)
	}

	// For Liability accounts (Credit Card / Loan):
	//   Tracked as positive debt (e.g. $500 balance).
	//   A debit (expense) of -$100 increases the debt owed by $100:
	//   Debt = Debt - Amount = 500 - (-100) = 600.
	//   A credit (payment) of +$200 decreases the debt owed:
	//   Debt = Debt - Amount = 600 - (+200) = 400.

	creditCardDebt := decimal.NewFromFloat(500.00)
	purchaseTx := decimal.NewFromFloat(-100.00)
	paymentTx := decimal.NewFromFloat(200.00)

	newDebt := creditCardDebt.Sub(purchaseTx)
	if !newDebt.Equal(decimal.NewFromFloat(600.00)) {
		t.Fatalf("expected credit card debt 600.00, got %s", newDebt)
	}

	newDebt = newDebt.Sub(paymentTx)
	if !newDebt.Equal(decimal.NewFromFloat(400.00)) {
		t.Fatalf("expected credit card debt 400.00, got %s", newDebt)
	}

	// Net Worth calculation:
	// Net Worth = Total Assets - Total Liabilities
	// 3424.50 - 400.00 = 3024.50
	netWorth := newAssetBal.Sub(newDebt)
	if !netWorth.Equal(decimal.NewFromFloat(3024.50)) {
		t.Fatalf("expected net worth 3024.50, got %s", netWorth)
	}
}

func TestTransactionSignNormalization(t *testing.T) {
	// If a user inputs positive 45.00 for a debit, it should normalize to -45.00
	amount1, _ := decimal.NewFromString("45.00")
	txType1 := "debit"
	if txType1 == "debit" && amount1.IsPositive() {
		amount1 = amount1.Neg()
	}
	if !amount1.Equal(decimal.NewFromFloat(-45.00)) {
		t.Fatalf("expected -45.00, got %s", amount1)
	}

	// If a user inputs negative -100.00 for a credit, it should normalize to +100.00
	amount2, _ := decimal.NewFromString("-100.00")
	txType2 := "credit"
	if txType2 == "credit" && amount2.IsNegative() {
		amount2 = amount2.Abs()
	}
	if !amount2.Equal(decimal.NewFromFloat(100.00)) {
		t.Fatalf("expected +100.00, got %s", amount2)
	}
}
