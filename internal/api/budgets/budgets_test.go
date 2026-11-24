package budgets

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"

	"github.com/rpmoulton/mint-alternative/internal/api/common"
	"github.com/rpmoulton/mint-alternative/internal/db/sqlc"
	"github.com/rpmoulton/mint-alternative/internal/httpx"
)

type mockBudgetRepo struct {
	budgets        map[string]*sqlc.Budget
	categories     map[string]*sqlc.Category
	allocations    map[string]map[string]*sqlc.BudgetCategory // budgetID -> catID -> alloc
	categorySpends map[string]decimal.Decimal
}

func newMockBudgetRepo() *mockBudgetRepo {
	return &mockBudgetRepo{
		budgets:        make(map[string]*sqlc.Budget),
		categories:     make(map[string]*sqlc.Category),
		allocations:    make(map[string]map[string]*sqlc.BudgetCategory),
		categorySpends: make(map[string]decimal.Decimal),
	}
}

func (m *mockBudgetRepo) addCategory(id string, userID *string, parentID *string, name, catType string) {
	cid, _ := common.StringToUUID(id)
	var uid pgtype.UUID
	if userID != nil {
		uid, _ = common.StringToUUID(*userID)
	}
	var pid pgtype.UUID
	if parentID != nil {
		pid, _ = common.StringToUUID(*parentID)
	}
	m.categories[id] = &sqlc.Category{
		ID:               cid,
		UserID:           uid,
		ParentCategoryID: pid,
		Name:             name,
		Type:             catType,
		CreatedAt:        pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true},
		UpdatedAt:        pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true},
	}
}

func (m *mockBudgetRepo) ListBudgets(ctx context.Context, userID pgtype.UUID) ([]sqlc.Budget, error) {
	var res []sqlc.Budget
	for _, b := range m.budgets {
		if b.UserID == userID {
			res = append(res, *b)
		}
	}
	return res, nil
}

func (m *mockBudgetRepo) GetBudgetByID(ctx context.Context, id, userID pgtype.UUID) (*sqlc.Budget, error) {
	idStr := common.UUIDToString(id)
	b, ok := m.budgets[idStr]
	if !ok || b.UserID != userID {
		return nil, pgx.ErrNoRows
	}
	return b, nil
}

func (m *mockBudgetRepo) GetBudgetByMonth(ctx context.Context, userID pgtype.UUID, year, month int16) (*sqlc.Budget, error) {
	for _, b := range m.budgets {
		if b.UserID == userID && b.Year == year && b.Month == month {
			return b, nil
		}
	}
	return nil, pgx.ErrNoRows
}

func (m *mockBudgetRepo) CreateBudget(ctx context.Context, params sqlc.CreateBudgetParams) (*sqlc.Budget, error) {
	for _, b := range m.budgets {
		if b.UserID == params.UserID && b.Year == params.Year && b.Month == params.Month {
			return nil, errors.New("unique constraint violation")
		}
	}
	newID := fmt.Sprintf("00000000-0000-0000-0000-0000000000%02d", len(m.budgets)+1)
	bid, _ := common.StringToUUID(newID)
	b := &sqlc.Budget{
		ID:        bid,
		UserID:    params.UserID,
		Name:      params.Name,
		Month:     params.Month,
		Year:      params.Year,
		CreatedAt: pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true},
		UpdatedAt: pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true},
	}
	m.budgets[newID] = b
	return b, nil
}

func (m *mockBudgetRepo) UpdateBudget(ctx context.Context, params sqlc.UpdateBudgetParams) (*sqlc.Budget, error) {
	idStr := common.UUIDToString(params.ID)
	b, ok := m.budgets[idStr]
	if !ok || b.UserID != params.UserID {
		return nil, pgx.ErrNoRows
	}
	if params.Name != nil {
		b.Name = *params.Name
	}
	return b, nil
}

func (m *mockBudgetRepo) DeleteBudget(ctx context.Context, id, userID pgtype.UUID) error {
	idStr := common.UUIDToString(id)
	b, ok := m.budgets[idStr]
	if !ok || b.UserID != userID {
		return pgx.ErrNoRows
	}
	delete(m.budgets, idStr)
	delete(m.allocations, idStr)
	return nil
}

func (m *mockBudgetRepo) ListBudgetCategories(ctx context.Context, budgetID pgtype.UUID) ([]sqlc.ListBudgetCategoriesRow, error) {
	idStr := common.UUIDToString(budgetID)
	var res []sqlc.ListBudgetCategoriesRow
	for _, alloc := range m.allocations[idStr] {
		catStr := common.UUIDToString(alloc.CategoryID)
		cat := m.categories[catStr]
		catName := ""
		catType := "expense"
		if cat != nil {
			catName = cat.Name
			catType = cat.Type
		}
		res = append(res, sqlc.ListBudgetCategoriesRow{
			ID:              alloc.ID,
			BudgetID:        alloc.BudgetID,
			CategoryID:      alloc.CategoryID,
			AllocatedAmount: alloc.AllocatedAmount,
			CategoryName:    catName,
			CategoryType:    catType,
		})
	}
	return res, nil
}

func (m *mockBudgetRepo) UpsertBudgetCategory(ctx context.Context, params sqlc.UpsertBudgetCategoryParams) (*sqlc.BudgetCategory, error) {
	bStr := common.UUIDToString(params.BudgetID)
	cStr := common.UUIDToString(params.CategoryID)
	if m.allocations[bStr] == nil {
		m.allocations[bStr] = make(map[string]*sqlc.BudgetCategory)
	}

	alloc := m.allocations[bStr][cStr]
	if alloc == nil {
		allocID, _ := common.StringToUUID(fmt.Sprintf("00000000-0000-0000-0000-000000000%03d", len(m.allocations[bStr])+1))
		alloc = &sqlc.BudgetCategory{
			ID:              allocID,
			BudgetID:        params.BudgetID,
			CategoryID:      params.CategoryID,
			AllocatedAmount: params.AllocatedAmount,
			CreatedAt:       pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true},
			UpdatedAt:       pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true},
		}
		m.allocations[bStr][cStr] = alloc
	} else {
		alloc.AllocatedAmount = params.AllocatedAmount
		alloc.UpdatedAt = pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true}
	}
	return alloc, nil
}

func (m *mockBudgetRepo) DeleteBudgetCategory(ctx context.Context, budgetID, categoryID pgtype.UUID) error {
	bStr := common.UUIDToString(budgetID)
	cStr := common.UUIDToString(categoryID)
	if m.allocations[bStr] != nil {
		delete(m.allocations[bStr], cStr)
	}
	return nil
}

func (m *mockBudgetRepo) GetMonthlyCategorySpending(ctx context.Context, params sqlc.GetMonthlySpendingByCategoryParams) ([]sqlc.GetMonthlySpendingByCategoryRow, error) {
	var res []sqlc.GetMonthlySpendingByCategoryRow
	for catID, spent := range m.categorySpends {
		cid, _ := common.StringToUUID(catID)
		spentNum, _ := common.StringToNumeric(spent.StringFixed(4))
		res = append(res, sqlc.GetMonthlySpendingByCategoryRow{
			CategoryID: cid,
			TotalSpent: spentNum,
		})
	}
	return res, nil
}

func (m *mockBudgetRepo) GetCategoryByID(ctx context.Context, id, userID pgtype.UUID) (*sqlc.Category, error) {
	idStr := common.UUIDToString(id)
	c, ok := m.categories[idStr]
	if !ok {
		return nil, pgx.ErrNoRows
	}
	if c.UserID.Valid && c.UserID != userID {
		return nil, pgx.ErrNoRows
	}
	return c, nil
}

func (m *mockBudgetRepo) ListCategories(ctx context.Context, userID pgtype.UUID) ([]sqlc.Category, error) {
	var res []sqlc.Category
	for _, c := range m.categories {
		if !c.UserID.Valid || c.UserID == userID {
			res = append(res, *c)
		}
	}
	return res, nil
}

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

func TestBudgetService_OwnershipAndCrossUser(t *testing.T) {
	repo := newMockBudgetRepo()
	svc := NewService(repo)
	ctx := context.Background()

	user1 := "11111111-1111-1111-1111-111111111111"
	user2 := "22222222-2222-2222-2222-222222222222"

	// Create budget for User 1
	b1, err := svc.Create(ctx, user1, &CreateBudgetRequest{
		Name:  "User 1 Budget",
		Month: 11,
		Year:  2026,
	})
	if err != nil {
		t.Fatalf("expected create budget success, got %v", err)
	}

	// User 2 cannot access User 1's budget
	_, err = svc.Get(ctx, b1.ID, user2)
	if err == nil {
		t.Fatalf("expected error when user 2 accesses user 1 budget, got nil")
	}

	// User 2 cannot update User 1's budget
	_, err = svc.Update(ctx, b1.ID, user2, &UpdateBudgetRequest{Name: "Hacked"})
	if err == nil {
		t.Fatalf("expected error when user 2 updates user 1 budget, got nil")
	}

	// User 2 cannot delete User 1's budget
	err = svc.Delete(ctx, b1.ID, user2)
	if err == nil {
		t.Fatalf("expected error when user 2 deletes user 1 budget, got nil")
	}
}

func TestBudgetService_DuplicateBudgetPrevention(t *testing.T) {
	repo := newMockBudgetRepo()
	svc := NewService(repo)
	ctx := context.Background()

	user := "11111111-1111-1111-1111-111111111111"

	_, err := svc.Create(ctx, user, &CreateBudgetRequest{
		Name:  "November 2026",
		Month: 11,
		Year:  2026,
	})
	if err != nil {
		t.Fatalf("unexpected error creating first budget: %v", err)
	}

	// Attempt duplicate creation for same year and month
	_, err = svc.Create(ctx, user, &CreateBudgetRequest{
		Name:  "November 2026 Duplicate",
		Month: 11,
		Year:  2026,
	})
	if err == nil {
		t.Fatalf("expected conflict error on duplicate month budget, got nil")
	}
	var prob *httpx.ProblemDetails
	if errors.As(err, &prob) && prob.Status != 409 {
		t.Fatalf("expected status 409 Conflict, got %d", prob.Status)
	}
}

func TestBudgetService_InvalidCategoryClassification(t *testing.T) {
	repo := newMockBudgetRepo()
	svc := NewService(repo)
	ctx := context.Background()

	user1 := "11111111-1111-1111-1111-111111111111"
	user2 := "22222222-2222-2222-2222-222222222222"

	b, _ := svc.Create(ctx, user1, &CreateBudgetRequest{
		Name:  "Budget 2026-11",
		Month: 11,
		Year:  2026,
	})

	// Add income category
	incomeCatID := "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
	repo.addCategory(incomeCatID, nil, nil, "Salary", "income")

	// Add transfer category
	transferCatID := "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"
	repo.addCategory(transferCatID, nil, nil, "Account Transfer", "transfer")

	// Add user2 private expense category
	user2CatID := "cccccccc-cccc-cccc-cccc-cccccccccccc"
	repo.addCategory(user2CatID, &user2, nil, "User 2 Category", "expense")

	// Add valid user1 expense category
	validCatID := "dddddddd-dddd-dddd-dddd-dddddddddddd"
	repo.addCategory(validCatID, &user1, nil, "Groceries", "expense")

	// 1. Try to allocate income category
	_, err := svc.UpsertCategory(ctx, b.ID, user1, &UpsertBudgetCategoryRequest{
		CategoryID:      incomeCatID,
		AllocatedAmount: "200.00",
	})
	if err == nil {
		t.Fatalf("expected error allocating income category, got nil")
	}

	// 2. Try to allocate transfer category
	_, err = svc.UpsertCategory(ctx, b.ID, user1, &UpsertBudgetCategoryRequest{
		CategoryID:      transferCatID,
		AllocatedAmount: "300.00",
	})
	if err == nil {
		t.Fatalf("expected error allocating transfer category, got nil")
	}

	// 3. Try to allocate category belonging to another user
	_, err = svc.UpsertCategory(ctx, b.ID, user1, &UpsertBudgetCategoryRequest{
		CategoryID:      user2CatID,
		AllocatedAmount: "100.00",
	})
	if err == nil {
		t.Fatalf("expected error allocating another user's category, got nil")
	}

	// 4. Allocate valid expense category succeeds
	updated, err := svc.UpsertCategory(ctx, b.ID, user1, &UpsertBudgetCategoryRequest{
		CategoryID:      validCatID,
		AllocatedAmount: "500.00",
	})
	if err != nil {
		t.Fatalf("expected success allocating valid expense category, got %v", err)
	}
	if len(updated.Categories) != 1 {
		t.Fatalf("expected 1 category allocation, got %d", len(updated.Categories))
	}
}

func TestBudgetService_ParentChildConflictAndRollup(t *testing.T) {
	repo := newMockBudgetRepo()
	svc := NewService(repo)
	ctx := context.Background()

	user := "11111111-1111-1111-1111-111111111111"

	b, _ := svc.Create(ctx, user, &CreateBudgetRequest{
		Name:  "Budget 2026-11",
		Month: 11,
		Year:  2026,
	})

	parentID := "10000000-0000-0000-0000-000000000001"
	child1ID := "10000000-0000-0000-0000-000000000002"
	child2ID := "10000000-0000-0000-0000-000000000003"

	repo.addCategory(parentID, &user, nil, "Food & Dining", "expense")
	repo.addCategory(child1ID, &user, &parentID, "Groceries", "expense")
	repo.addCategory(child2ID, &user, &parentID, "Restaurants", "expense")

	// Set spending rows: $150 Groceries, $80 Restaurants, $20 direct Food
	repo.categorySpends[child1ID] = decimal.NewFromFloat(150.00)
	repo.categorySpends[child2ID] = decimal.NewFromFloat(80.00)
	repo.categorySpends[parentID] = decimal.NewFromFloat(20.00)

	// Scenario A: Allocate Parent Category ("Food & Dining")
	updated, err := svc.UpsertCategory(ctx, b.ID, user, &UpsertBudgetCategoryRequest{
		CategoryID:      parentID,
		AllocatedAmount: "500.00",
	})
	if err != nil {
		t.Fatalf("expected parent category allocation to succeed: %v", err)
	}

	// Verify child spending rolled up into parent: 20 + 150 + 80 = 250
	if len(updated.Categories) != 1 {
		t.Fatalf("expected 1 category, got %d", len(updated.Categories))
	}
	if updated.Categories[0].SpentAmount != "250.00" {
		t.Fatalf("expected rolled-up spent amount 250.00, got %s", updated.Categories[0].SpentAmount)
	}
	if updated.TotalSpent != "250.00" {
		t.Fatalf("expected total spent 250.00, got %s", updated.TotalSpent)
	}

	// Attempting to allocate child category while parent is allocated must be rejected
	_, err = svc.UpsertCategory(ctx, b.ID, user, &UpsertBudgetCategoryRequest{
		CategoryID:      child1ID,
		AllocatedAmount: "200.00",
	})
	if err == nil {
		t.Fatalf("expected conflict allocating child when parent is allocated, got nil")
	}

	// Remove parent allocation
	_, err = svc.DeleteCategory(ctx, b.ID, user, parentID)
	if err != nil {
		t.Fatalf("failed to delete parent category allocation: %v", err)
	}

	// Scenario B: Allocate Child Categories instead
	_, err = svc.UpsertCategory(ctx, b.ID, user, &UpsertBudgetCategoryRequest{
		CategoryID:      child1ID,
		AllocatedAmount: "300.00",
	})
	if err != nil {
		t.Fatalf("expected child1 allocation to succeed: %v", err)
	}

	_, err = svc.UpsertCategory(ctx, b.ID, user, &UpsertBudgetCategoryRequest{
		CategoryID:      child2ID,
		AllocatedAmount: "200.00",
	})
	if err != nil {
		t.Fatalf("expected child2 allocation to succeed: %v", err)
	}

	// Now attempting to allocate parent must be rejected
	_, err = svc.UpsertCategory(ctx, b.ID, user, &UpsertBudgetCategoryRequest{
		CategoryID:      parentID,
		AllocatedAmount: "600.00",
	})
	if err == nil {
		t.Fatalf("expected conflict allocating parent when child is allocated, got nil")
	}

	// Verify spending in child allocations has no double-counting:
	// Groceries = 150.00, Restaurants = 80.00, Total = 230.00
	budget, err := svc.Get(ctx, b.ID, user)
	if err != nil {
		t.Fatalf("failed to get budget: %v", err)
	}
	if budget.TotalSpent != "230.00" {
		t.Fatalf("expected total spent 230.00, got %s", budget.TotalSpent)
	}
	if budget.TotalAllocated != "500.00" {
		t.Fatalf("expected total allocated 500.00, got %s", budget.TotalAllocated)
	}
}
