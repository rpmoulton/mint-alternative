package categories

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/rpmoulton/mint-alternative/internal/api/common"
	"github.com/rpmoulton/mint-alternative/internal/db/sqlc"
	"github.com/rpmoulton/mint-alternative/internal/httpx"
)

type mockCategoryRepo struct {
	categories        map[string]*sqlc.Category
	transactionsCount map[string]int64
	childrenCount     map[string]int64
}

func newMockCategoryRepo() *mockCategoryRepo {
	return &mockCategoryRepo{
		categories:        make(map[string]*sqlc.Category),
		transactionsCount: make(map[string]int64),
		childrenCount:     make(map[string]int64),
	}
}

func (m *mockCategoryRepo) addCategory(id string, userID *string, parentID *string, name, catType string) {
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

func (m *mockCategoryRepo) ListCategories(ctx context.Context, userID pgtype.UUID) ([]sqlc.Category, error) {
	var result []sqlc.Category
	for _, c := range m.categories {
		if !c.UserID.Valid || c.UserID == userID {
			result = append(result, *c)
		}
	}
	return result, nil
}

func (m *mockCategoryRepo) GetCategoryByID(ctx context.Context, id, userID pgtype.UUID) (*sqlc.Category, error) {
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

func (m *mockCategoryRepo) CreateCategory(ctx context.Context, params sqlc.CreateCategoryParams) (*sqlc.Category, error) {
	idStr := "10000000-0000-0000-0000-000000000099"
	cid, _ := common.StringToUUID(idStr)
	cat := &sqlc.Category{
		ID:               cid,
		UserID:           params.UserID,
		ParentCategoryID: params.ParentCategoryID,
		Name:             params.Name,
		Type:             params.Type,
		Icon:             params.Icon,
		Color:            params.Color,
		CreatedAt:        pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true},
		UpdatedAt:        pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true},
	}
	m.categories[idStr] = cat
	return cat, nil
}

func (m *mockCategoryRepo) UpdateCategory(ctx context.Context, params sqlc.UpdateCategoryParams) (*sqlc.Category, error) {
	idStr := common.UUIDToString(params.ID)
	c, ok := m.categories[idStr]
	if !ok {
		return nil, pgx.ErrNoRows
	}
	if params.Name != nil {
		c.Name = *params.Name
	}
	if params.Type != nil {
		c.Type = *params.Type
	}
	if params.Icon != nil {
		c.Icon = params.Icon
	}
	if params.Color != nil {
		c.Color = params.Color
	}
	if params.SetParentCategoryID {
		c.ParentCategoryID = params.ParentCategoryID
	}
	return c, nil
}

func (m *mockCategoryRepo) DeleteCategory(ctx context.Context, id, userID pgtype.UUID) error {
	idStr := common.UUIDToString(id)
	delete(m.categories, idStr)
	return nil
}

func (m *mockCategoryRepo) CountTransactionsByCategoryID(ctx context.Context, categoryID pgtype.UUID) (int64, error) {
	idStr := common.UUIDToString(categoryID)
	return m.transactionsCount[idStr], nil
}

func (m *mockCategoryRepo) CountChildCategories(ctx context.Context, categoryID pgtype.UUID) (int64, error) {
	idStr := common.UUIDToString(categoryID)
	return m.childrenCount[idStr], nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Tests
// ─────────────────────────────────────────────────────────────────────────────

func TestSystemCategoryProtection(t *testing.T) {
	repo := newMockCategoryRepo()
	svc := NewService(repo)
	ctx := context.Background()

	// Seed system category
	sysCatID := "00000000-0000-0000-0000-000000000001"
	repo.addCategory(sysCatID, nil, nil, "Salary", "income")

	userID := "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"

	// 1. User cannot modify system category
	newName := "My Salary"
	_, err := svc.Update(ctx, sysCatID, userID, &UpdateCategoryRequest{
		Name: &newName,
	})
	if err == nil {
		t.Fatal("expected error modifying system category, got nil")
	}
	var prob *httpx.ProblemDetails
	if !errors.As(err, &prob) || prob.Status != 403 {
		t.Fatalf("expected 403 Forbidden, got %v", err)
	}

	// 2. User cannot delete system category
	err = svc.Delete(ctx, sysCatID, userID)
	if err == nil {
		t.Fatal("expected error deleting system category, got nil")
	}
	if !errors.As(err, &prob) || prob.Status != 403 {
		t.Fatalf("expected 403 Forbidden, got %v", err)
	}
}

func TestCategoryOwnershipAndCrossUserIsolation(t *testing.T) {
	repo := newMockCategoryRepo()
	svc := NewService(repo)
	ctx := context.Background()

	userA := "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
	userB := "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"

	catA := "11111111-1111-1111-1111-111111111111"
	repo.addCategory(catA, &userA, nil, "User A Groceries", "expense")

	// User B tries to update User A's category -> 404
	newName := "Hacked Name"
	_, err := svc.Update(ctx, catA, userB, &UpdateCategoryRequest{
		Name: &newName,
	})
	if err == nil {
		t.Fatal("expected error updating other user's category, got nil")
	}
	var prob *httpx.ProblemDetails
	if !errors.As(err, &prob) || prob.Status != 404 {
		t.Fatalf("expected 404 Not Found, got %v", err)
	}

	// User B tries to delete User A's category -> 404
	err = svc.Delete(ctx, catA, userB)
	if err == nil {
		t.Fatal("expected error deleting other user's category, got nil")
	}
	if !errors.As(err, &prob) || prob.Status != 404 {
		t.Fatalf("expected 404 Not Found, got %v", err)
	}

	// User B tries to create child category using User A's category as parent -> 400
	_, err = svc.Create(ctx, userB, &CreateCategoryRequest{
		Name:             "User B Subcat",
		Type:             "expense",
		ParentCategoryID: &catA,
	})
	if err == nil {
		t.Fatal("expected error creating category with cross-user parent, got nil")
	}
	if !errors.As(err, &prob) || prob.Status != 400 {
		t.Fatalf("expected 400 Bad Request, got %v", err)
	}
}

func TestParentChildTypeCompatibility(t *testing.T) {
	repo := newMockCategoryRepo()
	svc := NewService(repo)
	ctx := context.Background()

	userA := "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
	parentExp := "11111111-1111-1111-1111-111111111111"
	repo.addCategory(parentExp, &userA, nil, "Living Expenses", "expense")

	// Attempt to create income category under expense parent -> 400
	_, err := svc.Create(ctx, userA, &CreateCategoryRequest{
		Name:             "Child Income",
		Type:             "income",
		ParentCategoryID: &parentExp,
	})
	if err == nil {
		t.Fatal("expected error creating incompatible type child, got nil")
	}
	var prob *httpx.ProblemDetails
	if !errors.As(err, &prob) || prob.Status != 400 {
		t.Fatalf("expected 400 Bad Request, got %v", err)
	}

	// Valid creation with matching type
	created, err := svc.Create(ctx, userA, &CreateCategoryRequest{
		Name:             "Child Expense",
		Type:             "expense",
		ParentCategoryID: &parentExp,
	})
	if err != nil {
		t.Fatalf("unexpected error creating compatible child: %v", err)
	}
	if created.Name != "Child Expense" {
		t.Fatalf("expected Child Expense, got %s", created.Name)
	}
}

func TestPreventSelfParentAndCycles(t *testing.T) {
	repo := newMockCategoryRepo()
	svc := NewService(repo)
	ctx := context.Background()

	userA := "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
	cat1 := "11111111-1111-1111-1111-111111111111"
	cat2 := "22222222-2222-2222-2222-222222222222"
	repo.addCategory(cat1, &userA, nil, "Category 1", "expense")
	repo.addCategory(cat2, &userA, &cat1, "Category 2", "expense")

	// 1. Self parent
	_, err := svc.Update(ctx, cat1, userA, &UpdateCategoryRequest{
		ParentCategoryID: &cat1,
	})
	if err == nil {
		t.Fatal("expected error setting category as self parent, got nil")
	}
	var prob *httpx.ProblemDetails
	if !errors.As(err, &prob) || prob.Status != 400 {
		t.Fatalf("expected 400 Bad Request, got %v", err)
	}

	// 2. Cycle: make cat1 child of cat2 when cat2 is already child of cat1
	_, err = svc.Update(ctx, cat1, userA, &UpdateCategoryRequest{
		ParentCategoryID: &cat2,
	})
	if err == nil {
		t.Fatal("expected error detecting hierarchy cycle, got nil")
	}
	if !errors.As(err, &prob) || prob.Status != 400 {
		t.Fatalf("expected 400 Bad Request, got %v", err)
	}
}

func TestSafeDeletionProtections(t *testing.T) {
	repo := newMockCategoryRepo()
	svc := NewService(repo)
	ctx := context.Background()

	userA := "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
	catWithTx := "11111111-1111-1111-1111-111111111111"
	catWithChildren := "22222222-2222-2222-2222-222222222222"
	unusedCat := "33333333-3333-3333-3333-333333333333"

	repo.addCategory(catWithTx, &userA, nil, "Has Transactions", "expense")
	repo.transactionsCount[catWithTx] = 5

	repo.addCategory(catWithChildren, &userA, nil, "Has Children", "expense")
	repo.childrenCount[catWithChildren] = 2

	repo.addCategory(unusedCat, &userA, nil, "Unused Category", "expense")

	// 1. Deleting category with transactions is blocked
	err := svc.Delete(ctx, catWithTx, userA)
	if err == nil {
		t.Fatal("expected error deleting category with transactions, got nil")
	}
	var prob *httpx.ProblemDetails
	if !errors.As(err, &prob) || prob.Status != 400 {
		t.Fatalf("expected 400 Bad Request, got %v", err)
	}

	// 2. Deleting category with child categories is blocked
	err = svc.Delete(ctx, catWithChildren, userA)
	if err == nil {
		t.Fatal("expected error deleting category with children, got nil")
	}
	if !errors.As(err, &prob) || prob.Status != 400 {
		t.Fatalf("expected 400 Bad Request, got %v", err)
	}

	// 3. Deleting unused category succeeds
	err = svc.Delete(ctx, unusedCat, userA)
	if err != nil {
		t.Fatalf("unexpected error deleting unused category: %v", err)
	}
}

func TestInvalidParentRelationship(t *testing.T) {
	repo := newMockCategoryRepo()
	svc := NewService(repo)
	ctx := context.Background()

	userA := "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
	nonExistentParent := "99999999-9999-9999-9999-999999999999"

	// 1. Create with non-existent parent returns 400 Bad Request
	_, err := svc.Create(ctx, userA, &CreateCategoryRequest{
		Name:             "Child Category",
		Type:             "expense",
		ParentCategoryID: &nonExistentParent,
	})
	if err == nil {
		t.Fatal("expected error creating category with non-existent parent, got nil")
	}
	var prob *httpx.ProblemDetails
	if !errors.As(err, &prob) || prob.Status != 400 {
		t.Fatalf("expected 400 Bad Request, got %v", err)
	}

	// 2. Update with non-existent parent returns 400 Bad Request
	catID := "11111111-1111-1111-1111-111111111111"
	repo.addCategory(catID, &userA, nil, "User Category", "expense")

	_, err = svc.Update(ctx, catID, userA, &UpdateCategoryRequest{
		ParentCategoryID: &nonExistentParent,
	})
	if err == nil {
		t.Fatal("expected error updating category with non-existent parent, got nil")
	}
	if !errors.As(err, &prob) || prob.Status != 400 {
		t.Fatalf("expected 400 Bad Request, got %v", err)
	}
}
