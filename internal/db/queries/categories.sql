-- name: ListCategories :many
SELECT * FROM categories
WHERE user_id = @user_id OR user_id IS NULL
ORDER BY type ASC, name ASC;

-- name: GetCategoryByID :one
SELECT * FROM categories
WHERE id = @id AND (user_id = @user_id OR user_id IS NULL);

-- name: CreateCategory :one
INSERT INTO categories (
    user_id,
    parent_category_id,
    name,
    type,
    icon,
    color
) VALUES (
    @user_id,
    @parent_category_id,
    @name,
    @type,
    @icon,
    @color
) RETURNING *;

-- name: UpdateCategory :one
UPDATE categories SET
    name = COALESCE(sqlc.narg('name'), name),
    type = COALESCE(sqlc.narg('type'), type),
    icon = COALESCE(sqlc.narg('icon'), icon),
    color = COALESCE(sqlc.narg('color'), color),
    parent_category_id = COALESCE(sqlc.narg('parent_category_id'), parent_category_id)
WHERE id = @id AND user_id = @user_id
RETURNING *;

-- name: DeleteCategory :exec
DELETE FROM categories
WHERE id = @id AND user_id = @user_id;
