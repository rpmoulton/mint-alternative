-- name: ListTransactions :many
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
    AND (sqlc.narg('account_id')::uuid IS NULL OR t.account_id = sqlc.narg('account_id'))
    AND (sqlc.narg('category_id')::uuid IS NULL OR t.category_id = sqlc.narg('category_id'))
    AND (sqlc.narg('transaction_type')::text IS NULL OR t.transaction_type = sqlc.narg('transaction_type'))
    AND (sqlc.narg('start_date')::date IS NULL OR t.transaction_date >= sqlc.narg('start_date'))
    AND (sqlc.narg('end_date')::date IS NULL OR t.transaction_date <= sqlc.narg('end_date'))
    AND (sqlc.narg('search')::text IS NULL OR (
        t.description ILIKE '%' || sqlc.narg('search')::text || '%'
        OR (t.merchant_name IS NOT NULL AND t.merchant_name ILIKE '%' || sqlc.narg('search')::text || '%')
    ))
ORDER BY t.transaction_date DESC, t.created_at DESC
LIMIT @limit_val OFFSET @offset_val;

-- name: CountTransactions :one
SELECT COUNT(*) FROM transactions t
WHERE t.user_id = @user_id
    AND (sqlc.narg('account_id')::uuid IS NULL OR t.account_id = sqlc.narg('account_id'))
    AND (sqlc.narg('category_id')::uuid IS NULL OR t.category_id = sqlc.narg('category_id'))
    AND (sqlc.narg('transaction_type')::text IS NULL OR t.transaction_type = sqlc.narg('transaction_type'))
    AND (sqlc.narg('start_date')::date IS NULL OR t.transaction_date >= sqlc.narg('start_date'))
    AND (sqlc.narg('end_date')::date IS NULL OR t.transaction_date <= sqlc.narg('end_date'))
    AND (sqlc.narg('search')::text IS NULL OR (
        t.description ILIKE '%' || sqlc.narg('search')::text || '%'
        OR (t.merchant_name IS NOT NULL AND t.merchant_name ILIKE '%' || sqlc.narg('search')::text || '%')
    ));

-- name: GetTransactionByID :one
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
WHERE t.id = @id AND t.user_id = @user_id;

-- name: CreateTransaction :one
INSERT INTO transactions (
    user_id,
    account_id,
    category_id,
    amount,
    currency,
    transaction_date,
    posted_date,
    merchant_name,
    description,
    notes,
    transaction_type,
    status,
    source_type,
    external_transaction_id
) VALUES (
    @user_id,
    @account_id,
    @category_id,
    @amount,
    @currency,
    @transaction_date,
    @posted_date,
    @merchant_name,
    @description,
    @notes,
    @transaction_type,
    @status,
    @source_type,
    @external_transaction_id
) RETURNING *;

-- name: UpdateTransaction :one
UPDATE transactions SET
    account_id = COALESCE(sqlc.narg('account_id'), account_id),
    category_id = COALESCE(sqlc.narg('category_id'), category_id),
    amount = COALESCE(sqlc.narg('amount'), amount),
    currency = COALESCE(sqlc.narg('currency'), currency),
    transaction_date = COALESCE(sqlc.narg('transaction_date'), transaction_date),
    posted_date = COALESCE(sqlc.narg('posted_date'), posted_date),
    merchant_name = COALESCE(sqlc.narg('merchant_name'), merchant_name),
    description = COALESCE(sqlc.narg('description'), description),
    notes = COALESCE(sqlc.narg('notes'), notes),
    transaction_type = COALESCE(sqlc.narg('transaction_type'), transaction_type),
    status = COALESCE(sqlc.narg('status'), status)
WHERE id = @id AND user_id = @user_id
RETURNING *;

-- name: DeleteTransaction :exec
DELETE FROM transactions
WHERE id = @id AND user_id = @user_id;
