-- name: ListAccountsByUser :many
SELECT 
    a.id,
    a.user_id,
    a.institution_id,
    a.name,
    a.account_type,
    a.account_subtype,
    a.currency,
    a.current_balance,
    a.available_balance,
    a.credit_limit,
    a.interest_rate,
    a.source_type,
    a.archived_at,
    a.created_at,
    a.updated_at,
    i.name as institution_name,
    i.display_name as institution_display_name,
    i.logo_url as institution_logo_url
FROM accounts a
LEFT JOIN institutions i ON a.institution_id = i.id
WHERE a.user_id = @user_id 
    AND (sqlc.narg('include_archived')::boolean = true OR a.archived_at IS NULL)
ORDER BY a.created_at ASC;

-- name: GetAccountByID :one
SELECT 
    a.id,
    a.user_id,
    a.institution_id,
    a.name,
    a.account_type,
    a.account_subtype,
    a.currency,
    a.current_balance,
    a.available_balance,
    a.credit_limit,
    a.interest_rate,
    a.source_type,
    a.archived_at,
    a.created_at,
    a.updated_at,
    i.name as institution_name,
    i.display_name as institution_display_name,
    i.logo_url as institution_logo_url
FROM accounts a
LEFT JOIN institutions i ON a.institution_id = i.id
WHERE a.id = @id AND a.user_id = @user_id;

-- name: CreateAccount :one
INSERT INTO accounts (
    user_id,
    institution_id,
    name,
    account_type,
    account_subtype,
    currency,
    current_balance,
    available_balance,
    credit_limit,
    interest_rate,
    source_type
) VALUES (
    @user_id,
    @institution_id,
    @name,
    @account_type,
    @account_subtype,
    @currency,
    @current_balance,
    @available_balance,
    @credit_limit,
    @interest_rate,
    @source_type
) RETURNING *;

-- name: UpdateAccount :one
UPDATE accounts SET
    institution_id = COALESCE(sqlc.narg('institution_id'), institution_id),
    name = COALESCE(sqlc.narg('name'), name),
    account_type = COALESCE(sqlc.narg('account_type'), account_type),
    account_subtype = COALESCE(sqlc.narg('account_subtype'), account_subtype),
    currency = COALESCE(sqlc.narg('currency'), currency),
    current_balance = COALESCE(sqlc.narg('current_balance'), current_balance),
    available_balance = COALESCE(sqlc.narg('available_balance'), available_balance),
    credit_limit = COALESCE(sqlc.narg('credit_limit'), credit_limit),
    interest_rate = COALESCE(sqlc.narg('interest_rate'), interest_rate),
    archived_at = CASE 
        WHEN sqlc.narg('archive')::boolean IS TRUE THEN CURRENT_TIMESTAMP
        WHEN sqlc.narg('unarchive')::boolean IS TRUE THEN NULL
        ELSE archived_at
    END
WHERE id = @id AND user_id = @user_id
RETURNING *;

-- name: UpdateAccountBalance :one
UPDATE accounts SET
    current_balance = current_balance + @amount_delta
WHERE id = @id AND user_id = @user_id
RETURNING *;

-- name: DeleteAccount :exec
DELETE FROM accounts
WHERE id = @id AND user_id = @user_id;
