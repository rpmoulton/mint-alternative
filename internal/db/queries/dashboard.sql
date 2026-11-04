-- name: GetDashboardAccountsSummary :many
SELECT 
    account_type,
    currency,
    COALESCE(SUM(current_balance), 0)::numeric(18, 4) as total_balance,
    COUNT(*) as account_count
FROM accounts
WHERE user_id = @user_id AND archived_at IS NULL
GROUP BY account_type, currency;

-- name: GetMonthlyIncomeExpenses :one
SELECT
    COALESCE(SUM(CASE WHEN t.transaction_type = 'credit' AND (c.type = 'income' OR c.type IS NULL) THEN t.amount ELSE 0 END), 0)::numeric(18, 4) as total_income,
    COALESCE(SUM(CASE WHEN t.transaction_type = 'debit' AND (c.type = 'expense' OR c.type IS NULL) THEN ABS(t.amount) ELSE 0 END), 0)::numeric(18, 4) as total_expenses
FROM transactions t
LEFT JOIN categories c ON t.category_id = c.id
WHERE t.user_id = @user_id
    AND t.transaction_date >= @start_date
    AND t.transaction_date <= @end_date;

-- name: GetMonthlySpendingByCategory :many
SELECT 
    c.id as category_id,
    c.name as category_name,
    c.color as category_color,
    c.icon as category_icon,
    COALESCE(SUM(ABS(t.amount)), 0)::numeric(18, 4) as total_spent,
    COUNT(t.id) as transaction_count
FROM transactions t
JOIN categories c ON t.category_id = c.id
WHERE t.user_id = @user_id
    AND t.transaction_type = 'debit'
    AND c.type = 'expense'
    AND t.transaction_date >= @start_date
    AND t.transaction_date <= @end_date
GROUP BY c.id, c.name, c.color, c.icon
ORDER BY total_spent DESC;

-- name: GetRecentTransactions :many
SELECT 
    t.id,
    t.user_id,
    t.account_id,
    t.category_id,
    t.amount,
    t.currency,
    t.transaction_date,
    t.posted_date,
    t.merchant_name,
    t.description,
    t.notes,
    t.transaction_type,
    t.status,
    t.source_type,
    t.external_transaction_id,
    t.created_at,
    t.updated_at,
    a.name as account_name,
    c.name as category_name,
    c.color as category_color,
    c.icon as category_icon,
    c.type as category_type
FROM transactions t
JOIN accounts a ON t.account_id = a.id
LEFT JOIN categories c ON t.category_id = c.id
WHERE t.user_id = @user_id
ORDER BY t.transaction_date DESC, t.created_at DESC
LIMIT @limit_val;
