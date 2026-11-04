package budgets

import (
	"testing"

	"github.com/shopspring/decimal"
)

func TestBudgetAllocationAndSpendingMath(t *testing.T) {
	allocated, _ := decimal.NewFromString("500.00")
	spent, _ := decimal.NewFromString("350.00")

	remaining := allocated.Sub(spent)
	if !remaining.Equal(decimal.NewFromFloat(150.00)) {
		t.Fatalf("expected remaining 150.00, got %s", remaining)
	}

	pctFloat, _ := spent.Div(allocated).Float64()
	pct := pctFloat * 100
	if pct != 70.0 {
		t.Fatalf("expected 70%% used, got %f", pct)
	}

	isOver := spent.GreaterThan(allocated)
	if isOver {
		t.Fatalf("expected isOver to be false")
	}

	// Now spend exceeds allocation: $600 spent on $500 budget
	overSpent, _ := decimal.NewFromString("600.00")
	isOverNow := overSpent.GreaterThan(allocated)
	if !isOverNow {
		t.Fatalf("expected isOverNow to be true")
	}

	remainingOver := allocated.Sub(overSpent)
	if !remainingOver.Equal(decimal.NewFromFloat(-100.00)) {
		t.Fatalf("expected -100.00, got %s", remainingOver)
	}
}
