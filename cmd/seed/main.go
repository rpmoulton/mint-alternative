package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"

	"github.com/rpmoulton/mint-alternative/config"
	"github.com/rpmoulton/mint-alternative/internal/api/common"
	"github.com/rpmoulton/mint-alternative/internal/db"
	"github.com/rpmoulton/mint-alternative/internal/db/sqlc"
	"github.com/rpmoulton/mint-alternative/pkg/argon2id"
)

func main() {
	ctx := context.Background()
	conf, err := config.InitConfig()
	if err != nil {
		log.Fatalf("failed to load configuration: %v", err)
	}

	// Safety check: Dev seed must NEVER run in production!
	if conf.Server.AppEnv == "production" {
		log.Fatalf("ABORT: Seed script cannot run in production environment!")
	}

	database, err := db.New(ctx, &conf.DB, slog.Default())
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer database.Close()

	queries := database.RootStore().Queries()
	argon := argon2id.DefaultConfig()

	fmt.Println("🌱 Seeding realistic development data for Mint Alternative...")

	// 1. Create or get demo user
	demoEmail := "demo@mint.local"
	demoPass := "Password123!"
	hashedPassword, err := argon.HashPassword(demoPass)
	if err != nil {
		log.Fatalf("failed to hash password: %v", err)
	}

	user, err := queries.GetUserByEmail(ctx, demoEmail)
	var userID pgtype.UUID
	if err != nil {
		createdUser, err := queries.CreateUser(ctx, sqlc.CreateUserParams{
			Email:             demoEmail,
			PasswordHash:      hashedPassword,
			FirstName:         "Alex",
			LastName:          "Morgan",
			PreferredCurrency: "USD",
			Timezone:          "America/New_York",
		})
		if err != nil {
			log.Fatalf("failed to create demo user: %v", err)
		}
		userID = createdUser.ID
		fmt.Printf("✅ Created demo user: %s (Password: %s)\n", demoEmail, demoPass)
	} else {
		userID = user.ID
		fmt.Printf("ℹ️ Found existing demo user: %s\n", demoEmail)
	}

	// 2. Seed Institutions
	chaseInst, err := queries.CreateInstitution(ctx, sqlc.CreateInstitutionParams{
		Name:        "JPMorgan Chase",
		DisplayName: "Chase",
		Website:     ptr("https://www.chase.com"),
	})
	if err != nil {
		insts, _ := queries.ListInstitutions(ctx)
		if len(insts) > 0 {
			chaseInst = insts[0]
		}
	}

	allyInst, err := queries.CreateInstitution(ctx, sqlc.CreateInstitutionParams{
		Name:        "Ally Financial",
		DisplayName: "Ally Bank",
		Website:     ptr("https://www.ally.com"),
	})
	if err != nil {
		insts, _ := queries.ListInstitutions(ctx)
		if len(insts) > 1 {
			allyInst = insts[1]
		}
	}

	vanguardInst, err := queries.CreateInstitution(ctx, sqlc.CreateInstitutionParams{
		Name:        "The Vanguard Group",
		DisplayName: "Vanguard",
		Website:     ptr("https://www.vanguard.com"),
	})
	if err != nil {
		insts, _ := queries.ListInstitutions(ctx)
		if len(insts) > 2 {
			vanguardInst = insts[2]
		}
	}

	// 3. Check existing accounts
	existingAccounts, err := queries.ListAccountsByUser(ctx, sqlc.ListAccountsByUserParams{
		UserID: userID,
	})
	if err != nil {
		log.Fatalf("failed to list user accounts: %v", err)
	}

	var checkingAccID, creditAccID, savingsAccID pgtype.UUID
	if len(existingAccounts) == 0 {
		// Create accounts
		checking, err := queries.CreateAccount(ctx, sqlc.CreateAccountParams{
			UserID:         userID,
			InstitutionID:  chaseInst.ID,
			Name:           "Premier Checking",
			AccountType:    "depository",
			AccountSubtype: "checking",
			Currency:       "USD",
			CurrentBalance: num("4850.25"),
			SourceType:     "manual",
		})
		if err != nil {
			log.Fatalf("failed to create checking account: %v", err)
		}
		checkingAccID = checking.ID

		credit, err := queries.CreateAccount(ctx, sqlc.CreateAccountParams{
			UserID:         userID,
			InstitutionID:  chaseInst.ID,
			Name:           "Sapphire Preferred",
			AccountType:    "credit",
			AccountSubtype: "credit_card",
			Currency:       "USD",
			CurrentBalance: num("825.40"),
			CreditLimit:    num("15000.00"),
			InterestRate:   num("21.49"),
			SourceType:     "manual",
		})
		if err != nil {
			log.Fatalf("failed to create credit account: %v", err)
		}
		creditAccID = credit.ID

		savings, err := queries.CreateAccount(ctx, sqlc.CreateAccountParams{
			UserID:         userID,
			InstitutionID:  allyInst.ID,
			Name:           "High Yield Savings",
			AccountType:    "depository",
			AccountSubtype: "savings",
			Currency:       "USD",
			CurrentBalance: num("22400.00"),
			InterestRate:   num("4.25"),
			SourceType:     "manual",
		})
		if err != nil {
			log.Fatalf("failed to create savings account: %v", err)
		}
		savingsAccID = savings.ID

		_, _ = queries.CreateAccount(ctx, sqlc.CreateAccountParams{
			UserID:         userID,
			InstitutionID:  vanguardInst.ID,
			Name:           "Brokerage Total Stock",
			AccountType:    "investment",
			AccountSubtype: "brokerage",
			Currency:       "USD",
			CurrentBalance: num("68150.80"),
			SourceType:     "manual",
		})

		_, _ = queries.CreateAccount(ctx, sqlc.CreateAccountParams{
			UserID:         userID,
			InstitutionID:  chaseInst.ID,
			Name:           "Auto Loan",
			AccountType:    "loan",
			AccountSubtype: "auto",
			Currency:       "USD",
			CurrentBalance: num("14200.00"),
			InterestRate:   num("4.99"),
			SourceType:     "manual",
		})
		fmt.Println("✅ Created realistic checking, savings, credit, investment, and loan accounts")
	} else {
		for _, a := range existingAccounts {
			if a.AccountType == "depository" && checkingAccID.Valid == false {
				checkingAccID = a.ID
			} else if a.AccountType == "credit" {
				creditAccID = a.ID
			} else if a.AccountType == "depository" {
				savingsAccID = a.ID
			}
		}
	}

	// 4. Get categories
	cats, err := queries.ListCategories(ctx, userID)
	if err != nil {
		log.Fatalf("failed to list categories: %v", err)
	}

	catMap := make(map[string]pgtype.UUID)
	for _, c := range cats {
		catMap[c.Name] = c.ID
	}

	// 5. Seed Transactions if none exist
	existingTx, _ := queries.ListTransactions(ctx, sqlc.ListTransactionsParams{
		UserID:   userID,
		LimitVal: 5,
	})

	if len(existingTx) == 0 && checkingAccID.Valid && creditAccID.Valid {
		now := time.Now().UTC()
		sampleTransactions := []struct {
			accountID pgtype.UUID
			catName   string
			amount    string
			desc      string
			merchant  string
			daysAgo   int
			txType    string
		}{
			{checkingAccID, "Salary", "4250.00", "Acme Corp Direct Deposit", "Acme Corp", 3, "credit"},
			{creditAccID, "Groceries", "-142.30", "Weekly groceries haul", "Trader Joe's", 1, "debit"},
			{creditAccID, "Groceries", "-89.15", "Produce & organic goods", "Whole Foods", 4, "debit"},
			{creditAccID, "Restaurants", "-24.50", "Lunch burrito & drink", "Chipotle", 2, "debit"},
			{creditAccID, "Restaurants", "-85.00", "Dinner with friends", "Osteria Italian", 6, "debit"},
			{creditAccID, "Transportation", "-48.20", "Fuel fill-up", "Shell Oil", 5, "debit"},
			{creditAccID, "Transportation", "-22.40", "Ride home from airport", "Uber", 8, "debit"},
			{checkingAccID, "Housing", "-1650.00", "Monthly apartment rent", "Apex Properties", 3, "debit"},
			{checkingAccID, "Utilities", "-115.00", "Electric utility bill", "Pacific Power", 7, "debit"},
			{checkingAccID, "Utilities", "-70.00", "Fiber Gigabit Internet", "Metronet", 9, "debit"},
			{creditAccID, "Shopping", "-64.80", "Home supplies & clothing", "Target", 10, "debit"},
			{creditAccID, "Entertainment", "-38.00", "Movie tickets & popcorn", "AMC Theatres", 11, "debit"},
			{creditAccID, "Healthcare", "-18.50", "Prescription medicine", "CVS Pharmacy", 12, "debit"},
			{checkingAccID, "Salary", "4250.00", "Acme Corp Direct Deposit", "Acme Corp", 18, "credit"},
			{checkingAccID, "Transfers", "-500.00", "Automatic savings transfer", "Ally Savings", 19, "debit"},
			{savingsAccID, "Transfers", "500.00", "Transfer from checking", "Chase Checking", 19, "credit"},
		}

		for _, st := range sampleTransactions {
			catID := catMap[st.catName]
			txDate := now.AddDate(0, 0, -st.daysAgo)
			amountNum, _ := common.StringToNumeric(st.amount)

			_, err := queries.CreateTransaction(ctx, sqlc.CreateTransactionParams{
				UserID:          userID,
				AccountID:       st.accountID,
				CategoryID:      catID,
				Amount:          amountNum,
				Currency:        "USD",
				TransactionDate: pgtype.Date{Time: txDate, Valid: true},
				MerchantName:    ptr(st.merchant),
				Description:     st.desc,
				TransactionType: st.txType,
				Status:          "posted",
				SourceType:      "manual",
			})
			if err != nil {
				log.Printf("failed to seed transaction %s: %v", st.desc, err)
			}
		}
		fmt.Printf("✅ Seeded %d realistic transactions across accounts\n", len(sampleTransactions))
	}

	// 6. Seed Monthly Budget for Current Month
	now := time.Now().UTC()
	curMonth := int16(now.Month())
	curYear := int16(now.Year())

	budget, err := queries.GetBudgetByMonth(ctx, sqlc.GetBudgetByMonthParams{
		UserID: userID,
		Year:   curYear,
		Month:  curMonth,
	})
	if err != nil {
		createdBudget, err := queries.CreateBudget(ctx, sqlc.CreateBudgetParams{
			UserID: userID,
			Name:   fmt.Sprintf("%s %d Budget", now.Month().String(), curYear),
			Month:  curMonth,
			Year:   curYear,
		})
		if err == nil {
			budget = createdBudget
			fmt.Printf("✅ Created budget for %s %d\n", now.Month().String(), curYear)

			allocations := map[string]string{
				"Groceries":      "600.00",
				"Restaurants":    "300.00",
				"Housing":        "1650.00",
				"Utilities":      "250.00",
				"Transportation": "200.00",
				"Shopping":       "200.00",
				"Entertainment":  "150.00",
				"Healthcare":     "100.00",
			}

			for catName, allocStr := range allocations {
				if catID, ok := catMap[catName]; ok {
					allocNum, _ := common.StringToNumeric(allocStr)
					_, _ = queries.UpsertBudgetCategory(ctx, sqlc.UpsertBudgetCategoryParams{
						BudgetID:        budget.ID,
						CategoryID:      catID,
						AllocatedAmount: allocNum,
					})
				}
			}
			fmt.Printf("✅ Seeded budget category allocations\n")
		}
	}

	fmt.Println("\n✨ Database successfully seeded!")
	fmt.Println("--------------------------------------------------")
	fmt.Printf("Demo Login: %s\n", demoEmail)
	fmt.Printf("Password:   %s\n", demoPass)
	fmt.Println("--------------------------------------------------")
}

func num(s string) pgtype.Numeric {
	dec, _ := decimal.NewFromString(s)
	n, _ := common.StringToNumeric(dec.StringFixed(4))
	return n
}

func ptr(s string) *string {
	return &s
}
