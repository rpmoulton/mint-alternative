package transactions

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
)

func TestFinancialCalculationConventions(t *testing.T) {
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

	netWorth := newAssetBal.Sub(newDebt)
	if !netWorth.Equal(decimal.NewFromFloat(3024.50)) {
		t.Fatalf("expected net worth 3024.50, got %s", netWorth)
	}
}

func TestTransactionSignNormalization(t *testing.T) {
	amount1, _ := decimal.NewFromString("45.00")
	txType1 := "debit"
	if txType1 == "debit" && amount1.IsPositive() {
		amount1 = amount1.Neg()
	}
	if !amount1.Equal(decimal.NewFromFloat(-45.00)) {
		t.Fatalf("expected -45.00, got %s", amount1)
	}

	amount2, _ := decimal.NewFromString("-100.00")
	txType2 := "credit"
	if txType2 == "credit" && amount2.IsNegative() {
		amount2 = amount2.Abs()
	}
	if !amount2.Equal(decimal.NewFromFloat(100.00)) {
		t.Fatalf("expected +100.00, got %s", amount2)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Mock Repository for Financial Integrity & Regression Tests
// ─────────────────────────────────────────────────────────────────────────────

type mockTransactionRepo struct {
	accounts            map[string]*sqlc.GetAccountByIDRow
	transactions        map[string]*sqlc.GetTransactionByIDRow
	failUpdateTx        bool
	failDeleteTx        bool
	failUpdateBalance   bool
	nextTxID            int
}

func newMockTransactionRepo() *mockTransactionRepo {
	return &mockTransactionRepo{
		accounts:     make(map[string]*sqlc.GetAccountByIDRow),
		transactions: make(map[string]*sqlc.GetTransactionByIDRow),
	}
}

func (m *mockTransactionRepo) addAccount(id, userID, name, accType, currentBalance string, isArchived bool) {
	balNum, _ := common.StringToNumeric(currentBalance)
	uid, _ := common.StringToUUID(userID)
	aid, _ := common.StringToUUID(id)

	var archivedAt pgtype.Timestamptz
	if isArchived {
		archivedAt = pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true}
	}

	m.accounts[id] = &sqlc.GetAccountByIDRow{
		ID:             aid,
		UserID:         uid,
		Name:           name,
		AccountType:    accType,
		Currency:       "USD",
		CurrentBalance: balNum,
		ArchivedAt:     archivedAt,
	}
}

func (m *mockTransactionRepo) getAccountBalance(id string) decimal.Decimal {
	acc := m.accounts[id]
	if acc == nil {
		return decimal.Zero
	}
	d, _ := decimal.NewFromString(common.NumericToString(acc.CurrentBalance))
	return d
}

func (m *mockTransactionRepo) ListTransactions(ctx context.Context, params sqlc.ListTransactionsParams) ([]sqlc.ListTransactionsRow, error) {
	return nil, nil
}

func (m *mockTransactionRepo) CountTransactions(ctx context.Context, params sqlc.CountTransactionsParams) (int64, error) {
	return int64(len(m.transactions)), nil
}

func (m *mockTransactionRepo) GetTransactionByID(ctx context.Context, id, userID pgtype.UUID) (*sqlc.GetTransactionByIDRow, error) {
	idStr := common.UUIDToString(id)
	tx, ok := m.transactions[idStr]
	if !ok {
		return nil, pgx.ErrNoRows
	}
	if common.UUIDToString(tx.UserID) != common.UUIDToString(userID) {
		return nil, pgx.ErrNoRows
	}
	return tx, nil
}

func (m *mockTransactionRepo) CreateTransaction(ctx context.Context, params sqlc.CreateTransactionParams) (*sqlc.Transaction, error) {
	m.nextTxID++
	txIDStr := fmt.Sprintf("00000000-0000-0000-0000-%012d", m.nextTxID)
	txID, _ := common.StringToUUID(txIDStr)

	accIDStr := common.UUIDToString(params.AccountID)
	accName := "Unknown"
	if acc, ok := m.accounts[accIDStr]; ok {
		accName = acc.Name
	}

	txRow := &sqlc.GetTransactionByIDRow{
		ID:              txID,
		UserID:          params.UserID,
		AccountID:       params.AccountID,
		CategoryID:      params.CategoryID,
		Amount:          params.Amount,
		Currency:        params.Currency,
		TransactionDate: params.TransactionDate,
		PostedDate:      params.PostedDate,
		MerchantName:    params.MerchantName,
		Description:     params.Description,
		Notes:           params.Notes,
		TransactionType: params.TransactionType,
		Status:          params.Status,
		SourceType:      params.SourceType,
		AccountName:     accName,
	}
	m.transactions[txIDStr] = txRow

	return &sqlc.Transaction{
		ID:              txID,
		UserID:          params.UserID,
		AccountID:       params.AccountID,
		Amount:          params.Amount,
		Currency:        params.Currency,
		TransactionDate: params.TransactionDate,
		Description:     params.Description,
		TransactionType: params.TransactionType,
		Status:          params.Status,
	}, nil
}

func (m *mockTransactionRepo) UpdateTransaction(ctx context.Context, params sqlc.UpdateTransactionParams) (*sqlc.Transaction, error) {
	if m.failUpdateTx {
		return nil, errors.New("simulated update transaction database failure")
	}

	idStr := common.UUIDToString(params.ID)
	tx, ok := m.transactions[idStr]
	if !ok {
		return nil, pgx.ErrNoRows
	}

	if params.AccountID.Valid {
		tx.AccountID = params.AccountID
		accIDStr := common.UUIDToString(params.AccountID)
		if acc, ok := m.accounts[accIDStr]; ok {
			tx.AccountName = acc.Name
		}
	}
	if params.Amount.Valid {
		tx.Amount = params.Amount
	}
	if params.Description != nil {
		tx.Description = *params.Description
	}
	if params.MerchantName != nil {
		tx.MerchantName = params.MerchantName
	}
	if params.TransactionType != nil {
		tx.TransactionType = *params.TransactionType
	}

	return &sqlc.Transaction{
		ID:              tx.ID,
		UserID:          tx.UserID,
		AccountID:       tx.AccountID,
		Amount:          tx.Amount,
		Currency:        tx.Currency,
		TransactionDate: tx.TransactionDate,
		Description:     tx.Description,
		TransactionType: tx.TransactionType,
		Status:          tx.Status,
	}, nil
}

func (m *mockTransactionRepo) DeleteTransaction(ctx context.Context, id, userID pgtype.UUID) error {
	if m.failDeleteTx {
		return errors.New("simulated delete transaction database failure")
	}
	idStr := common.UUIDToString(id)
	delete(m.transactions, idStr)
	return nil
}

func (m *mockTransactionRepo) GetAccountByID(ctx context.Context, id, userID pgtype.UUID) (*sqlc.GetAccountByIDRow, error) {
	idStr := common.UUIDToString(id)
	acc, ok := m.accounts[idStr]
	if !ok {
		return nil, pgx.ErrNoRows
	}
	if common.UUIDToString(acc.UserID) != common.UUIDToString(userID) {
		return nil, pgx.ErrNoRows
	}
	return acc, nil
}

func (m *mockTransactionRepo) UpdateAccountBalance(ctx context.Context, id, userID pgtype.UUID, amountDelta pgtype.Numeric) (*sqlc.Account, error) {
	if m.failUpdateBalance {
		return nil, errors.New("simulated update account balance failure")
	}
	idStr := common.UUIDToString(id)
	acc, ok := m.accounts[idStr]
	if !ok {
		return nil, pgx.ErrNoRows
	}
	if common.UUIDToString(acc.UserID) != common.UUIDToString(userID) {
		return nil, pgx.ErrNoRows
	}

	curDec, _ := decimal.NewFromString(common.NumericToString(acc.CurrentBalance))
	deltaDec, _ := decimal.NewFromString(common.NumericToString(amountDelta))
	newDec := curDec.Add(deltaDec)

	newNum, _ := common.StringToNumeric(newDec.StringFixed(4))
	acc.CurrentBalance = newNum

	return &sqlc.Account{
		ID:             acc.ID,
		UserID:         acc.UserID,
		CurrentBalance: newNum,
		AccountType:    acc.AccountType,
	}, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Section 7: Regression Tests for Financial Integrity
// ─────────────────────────────────────────────────────────────────────────────

const (
	testUser1 = "00000000-0000-0000-0000-000000000001"
	testUser2 = "00000000-0000-0000-0000-000000000002"
	accAsset1 = "10000000-0000-0000-0000-000000000001"
	accAsset2 = "10000000-0000-0000-0000-000000000002"
	accLiab1  = "10000000-0000-0000-0000-000000000003"
	accOtherU = "20000000-0000-0000-0000-000000000001"
	accArch1  = "10000000-0000-0000-0000-000000000005"
)

func setupTestService() (*service, *mockTransactionRepo) {
	repo := newMockTransactionRepo()
	repo.addAccount(accAsset1, testUser1, "Checking 1", "depository", "1000.00", false)
	repo.addAccount(accAsset2, testUser1, "Savings 1", "depository", "500.00", false)
	repo.addAccount(accLiab1, testUser1, "Credit Card 1", "credit", "500.00", false)
	repo.addAccount(accOtherU, testUser2, "User 2 Checking", "depository", "2000.00", false)
	repo.addAccount(accArch1, testUser1, "Archived Savings", "depository", "300.00", true)

	svc := &service{repo: repo}
	return svc, repo
}

// 1. Editing a transaction amount
func TestFinancialIntegrity_EditTransactionAmount(t *testing.T) {
	ctx := context.Background()
	svc, repo := setupTestService()

	// Initial balance: 1000.00
	// Create debit tx of 50.00 -> balance becomes 950.00
	created, err := svc.Create(ctx, testUser1, &CreateTransactionRequest{
		AccountID:       accAsset1,
		Amount:          "50.00",
		TransactionType: "debit",
		TransactionDate: "2026-11-01",
		Description:     "Groceries",
	})
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}

	balAfterCreate := repo.getAccountBalance(accAsset1)
	if !balAfterCreate.Equal(decimal.NewFromFloat(950.00)) {
		t.Fatalf("expected balance 950.00, got %s", balAfterCreate)
	}

	// Update amount to 80.00 (debit: -80.00)
	newAmount := "80.00"
	_, err = svc.Update(ctx, created.ID, testUser1, &UpdateTransactionRequest{
		Amount: &newAmount,
	})
	if err != nil {
		t.Fatalf("update failed: %v", err)
	}

	// Reverts -50.00 (+50.00 -> 1000.00), applies -80.00 -> 920.00
	balAfterUpdate := repo.getAccountBalance(accAsset1)
	if !balAfterUpdate.Equal(decimal.NewFromFloat(920.00)) {
		t.Fatalf("expected balance 920.00 after amount edit, got %s", balAfterUpdate)
	}
}

// 2. Changing a debit to a credit
func TestFinancialIntegrity_ChangeDebitToCredit(t *testing.T) {
	ctx := context.Background()
	svc, repo := setupTestService()

	// Create debit tx of 50.00 on asset account (1000.00 -> 950.00)
	created, err := svc.Create(ctx, testUser1, &CreateTransactionRequest{
		AccountID:       accAsset1,
		Amount:          "50.00",
		TransactionType: "debit",
		TransactionDate: "2026-11-01",
		Description:     "Item purchase",
	})
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}

	// Change to credit of 50.00 (e.g. refund)
	creditType := "credit"
	creditAmount := "50.00"
	_, err = svc.Update(ctx, created.ID, testUser1, &UpdateTransactionRequest{
		TransactionType: &creditType,
		Amount:          &creditAmount,
	})
	if err != nil {
		t.Fatalf("update failed: %v", err)
	}

	// Balance: 950.00 + 50.00 (revert debit) + 50.00 (apply credit) = 1050.00
	bal := repo.getAccountBalance(accAsset1)
	if !bal.Equal(decimal.NewFromFloat(1050.00)) {
		t.Fatalf("expected balance 1050.00 after changing debit to credit, got %s", bal)
	}
}

// 3. Changing a credit to a debit
func TestFinancialIntegrity_ChangeCreditToDebit(t *testing.T) {
	ctx := context.Background()
	svc, repo := setupTestService()

	// Create credit tx of 100.00 on asset account (1000.00 -> 1100.00)
	created, err := svc.Create(ctx, testUser1, &CreateTransactionRequest{
		AccountID:       accAsset1,
		Amount:          "100.00",
		TransactionType: "credit",
		TransactionDate: "2026-11-01",
		Description:     "Bonus",
	})
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}

	balAfterCreate := repo.getAccountBalance(accAsset1)
	if !balAfterCreate.Equal(decimal.NewFromFloat(1100.00)) {
		t.Fatalf("expected balance 1100.00, got %s", balAfterCreate)
	}

	// Change to debit of 100.00
	debitType := "debit"
	debitAmount := "100.00"
	_, err = svc.Update(ctx, created.ID, testUser1, &UpdateTransactionRequest{
		TransactionType: &debitType,
		Amount:          &debitAmount,
	})
	if err != nil {
		t.Fatalf("update failed: %v", err)
	}

	// Balance: 1100.00 - 100.00 (revert credit) - 100.00 (apply debit) = 900.00
	bal := repo.getAccountBalance(accAsset1)
	if !bal.Equal(decimal.NewFromFloat(900.00)) {
		t.Fatalf("expected balance 900.00 after changing credit to debit, got %s", bal)
	}
}

// 4. Moving a transaction between accounts
func TestFinancialIntegrity_MoveTransactionBetweenAccounts(t *testing.T) {
	ctx := context.Background()
	svc, repo := setupTestService()

	// Account 1: 1000.00, Account 2: 500.00
	// Create debit of 50.00 on Account 1 -> Account 1 becomes 950.00
	created, err := svc.Create(ctx, testUser1, &CreateTransactionRequest{
		AccountID:       accAsset1,
		Amount:          "50.00",
		TransactionType: "debit",
		TransactionDate: "2026-11-01",
		Description:     "Store purchase",
	})
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}

	// Move transaction to Account 2
	targetAcc := accAsset2
	_, err = svc.Update(ctx, created.ID, testUser1, &UpdateTransactionRequest{
		AccountID: &targetAcc,
	})
	if err != nil {
		t.Fatalf("update failed: %v", err)
	}

	// Account 1 balance should be restored to 1000.00
	bal1 := repo.getAccountBalance(accAsset1)
	if !bal1.Equal(decimal.NewFromFloat(1000.00)) {
		t.Fatalf("expected Account 1 balance restored to 1000.00, got %s", bal1)
	}

	// Account 2 balance should be 500.00 - 50.00 = 450.00
	bal2 := repo.getAccountBalance(accAsset2)
	if !bal2.Equal(decimal.NewFromFloat(450.00)) {
		t.Fatalf("expected Account 2 balance 450.00, got %s", bal2)
	}
}

// 5. Moving transactions between asset and liability account types
func TestFinancialIntegrity_MoveBetweenAssetAndLiability(t *testing.T) {
	ctx := context.Background()
	svc, repo := setupTestService()

	// Account 1 (Asset): 1000.00
	// Account 3 (Credit Liability): debt 500.00
	// Create debit of 100.00 on Asset -> Asset balance becomes 900.00
	created, err := svc.Create(ctx, testUser1, &CreateTransactionRequest{
		AccountID:       accAsset1,
		Amount:          "100.00",
		TransactionType: "debit",
		TransactionDate: "2026-11-01",
		Description:     "Electronics",
	})
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}

	if bal := repo.getAccountBalance(accAsset1); !bal.Equal(decimal.NewFromFloat(900.00)) {
		t.Fatalf("expected asset balance 900.00, got %s", bal)
	}

	// Move transaction to Credit Card (Liability)
	creditCardAcc := accLiab1
	_, err = svc.Update(ctx, created.ID, testUser1, &UpdateTransactionRequest{
		AccountID: &creditCardAcc,
	})
	if err != nil {
		t.Fatalf("update failed: %v", err)
	}

	// Asset balance restored to 1000.00
	if bal := repo.getAccountBalance(accAsset1); !bal.Equal(decimal.NewFromFloat(1000.00)) {
		t.Fatalf("expected asset balance restored to 1000.00, got %s", bal)
	}

	// Liability debt was 500.00, now increased by 100.00 debit purchase -> 600.00
	if bal := repo.getAccountBalance(accLiab1); !bal.Equal(decimal.NewFromFloat(600.00)) {
		t.Fatalf("expected liability debt 600.00, got %s", bal)
	}
}

// 6. Updating transaction metadata without changing balances
func TestFinancialIntegrity_UpdateMetadataPreservesBalances(t *testing.T) {
	ctx := context.Background()
	svc, repo := setupTestService()

	created, err := svc.Create(ctx, testUser1, &CreateTransactionRequest{
		AccountID:       accAsset1,
		Amount:          "75.00",
		TransactionType: "debit",
		TransactionDate: "2026-11-01",
		Description:     "Old Description",
	})
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}

	expectedBal := decimal.NewFromFloat(925.00)
	if bal := repo.getAccountBalance(accAsset1); !bal.Equal(expectedBal) {
		t.Fatalf("expected balance %s, got %s", expectedBal, bal)
	}

	// Update only metadata: description and notes
	newDesc := "New Description"
	newNotes := "Some notes"
	newMerchant := "New Merchant"
	_, err = svc.Update(ctx, created.ID, testUser1, &UpdateTransactionRequest{
		Description:  &newDesc,
		Notes:        &newNotes,
		MerchantName: &newMerchant,
	})
	if err != nil {
		t.Fatalf("update failed: %v", err)
	}

	// Balance MUST remain exactly 925.00
	if bal := repo.getAccountBalance(accAsset1); !bal.Equal(expectedBal) {
		t.Fatalf("expected balance %s untouched, got %s", expectedBal, bal)
	}
}

// 7. Deleting an edited transaction
func TestFinancialIntegrity_DeleteEditedTransaction(t *testing.T) {
	ctx := context.Background()
	svc, repo := setupTestService()

	// Initial balance: 1000.00
	created, err := svc.Create(ctx, testUser1, &CreateTransactionRequest{
		AccountID:       accAsset1,
		Amount:          "50.00",
		TransactionType: "debit",
		TransactionDate: "2026-11-01",
		Description:     "Initial tx",
	})
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}

	// Edit to 120.00 -> balance becomes 880.00
	editedAmount := "120.00"
	_, err = svc.Update(ctx, created.ID, testUser1, &UpdateTransactionRequest{
		Amount: &editedAmount,
	})
	if err != nil {
		t.Fatalf("update failed: %v", err)
	}

	if bal := repo.getAccountBalance(accAsset1); !bal.Equal(decimal.NewFromFloat(880.00)) {
		t.Fatalf("expected balance 880.00, got %s", bal)
	}

	// Delete the edited transaction
	err = svc.Delete(ctx, created.ID, testUser1)
	if err != nil {
		t.Fatalf("delete failed: %v", err)
	}

	// Account balance MUST be restored to original 1000.00!
	if bal := repo.getAccountBalance(accAsset1); !bal.Equal(decimal.NewFromFloat(1000.00)) {
		t.Fatalf("expected balance fully restored to 1000.00 after deleting edited tx, got %s", bal)
	}
}

// 8. Rejecting updates that reference accounts owned by another user
func TestFinancialIntegrity_RejectCrossUserAccountUpdate(t *testing.T) {
	ctx := context.Background()
	svc, repo := setupTestService()

	// Create transaction for User 1 on Account 1 (balance 1000 -> 950)
	created, err := svc.Create(ctx, testUser1, &CreateTransactionRequest{
		AccountID:       accAsset1,
		Amount:          "50.00",
		TransactionType: "debit",
		TransactionDate: "2026-11-01",
		Description:     "User 1 tx",
	})
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}

	// User 1 attempts to update transaction to reference User 2's account
	user2Acc := accOtherU
	_, err = svc.Update(ctx, created.ID, testUser1, &UpdateTransactionRequest{
		AccountID: &user2Acc,
	})
	if err == nil {
		t.Fatalf("expected error when moving transaction to another user's account")
	}

	// Account 1 balance MUST remain 950.00 (not reverted prematurely)
	if bal := repo.getAccountBalance(accAsset1); !bal.Equal(decimal.NewFromFloat(950.00)) {
		t.Fatalf("expected User 1 account balance untouched at 950.00, got %s", bal)
	}

	// User 2 account balance MUST remain 2000.00
	if bal := repo.getAccountBalance(accOtherU); !bal.Equal(decimal.NewFromFloat(2000.00)) {
		t.Fatalf("expected User 2 account balance untouched at 2000.00, got %s", bal)
	}
}

// 9. Ensuring failed updates do not leave partial balance changes
func TestFinancialIntegrity_FailedUpdatesRollbackBalances(t *testing.T) {
	ctx := context.Background()
	svc, repo := setupTestService()

	// Account 1: 1000.00, Account 2: 500.00
	created, err := svc.Create(ctx, testUser1, &CreateTransactionRequest{
		AccountID:       accAsset1,
		Amount:          "50.00",
		TransactionType: "debit",
		TransactionDate: "2026-11-01",
		Description:     "Tx before fail",
	})
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}

	// Account 1 is now 950.00, Account 2 is 500.00
	if bal := repo.getAccountBalance(accAsset1); !bal.Equal(decimal.NewFromFloat(950.00)) {
		t.Fatalf("expected balance 950.00, got %s", bal)
	}

	// Simulate database failure during transaction record update
	repo.failUpdateTx = true

	targetAcc := accAsset2
	newAmount := "100.00"
	_, err = svc.Update(ctx, created.ID, testUser1, &UpdateTransactionRequest{
		AccountID: &targetAcc,
		Amount:    &newAmount,
	})
	if err == nil {
		t.Fatalf("expected update to fail due to simulated db error")
	}

	// Neither account should have been modified!
	// Account 1 should still be 950.00
	if bal := repo.getAccountBalance(accAsset1); !bal.Equal(decimal.NewFromFloat(950.00)) {
		t.Fatalf("expected Account 1 balance 950.00 after rollback, got %s", bal)
	}

	// Account 2 should still be 500.00
	if bal := repo.getAccountBalance(accAsset2); !bal.Equal(decimal.NewFromFloat(500.00)) {
		t.Fatalf("expected Account 2 balance 500.00 after rollback, got %s", bal)
	}
}
