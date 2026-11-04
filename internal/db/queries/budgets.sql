-- name: ListBudgets :many
SELECT * FROM budgets
WHERE user_id = @user_id
ORDER BY year DESC, month DESC;

-- name: GetBudgetByID :one
SELECT * FROM budgets
WHERE id = @id AND user_id = @user_id;

-- name: GetBudgetByMonth :one
SELECT * FROM budgets
WHERE user_id = @user_id AND year = @year AND month = @month;

-- name: CreateBudget :one
INSERT INTO budgets (
    user_id,
    name,
    month,
    year
) VALUES (
    @user_id,
    @name,
    @month,
    @year
) RETURNING *;

-- name: UpdateBudget :one
UPDATE budgets SET
    name = COALESCE(sqlc.narg('name'), name)
WHERE id = @id AND user_id = @user_id
RETURNING *;

-- name: DeleteBudget :exec
DELETE FROM budgets
WHERE id = @id AND user_id = @user_id;

-- name: ListBudgetCategories :many
SELECT 
    bc.id,
    bc.budget_id,
    bc.category_id,
    bc.allocated_amount,
    bc.created_at,
    bc.updated_at,
    c.name as category_name,
    c.type as category_type,
    c.icon as category_icon,
    c.color as category_color
FROM budget_categories bc
JOIN categories c ON bc.category_id = c.id
WHERE bc.budget_id = @budget_id
ORDER BY c.name ASC;

-- name: UpsertBudgetCategory :one
INSERT INTO budget_categories (
    budget_id,
    category_id,
    allocated_amount
) VALUES (
    @budget_id,
    @category_id,
    @allocated_amount
) ON CONFLICT (budget_id, category_id) DO UPDATE SET
    allocated_amount = EXCLUDED.allocated_amount
RETURNING *;

-- name: DeleteBudgetCategory :exec
DELETE FROM budget_categories
WHERE budget_id = @budget_id AND category_id = @category_id;
