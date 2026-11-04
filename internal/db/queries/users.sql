-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = LOWER(@email);

-- name: GetUserByID :one
SELECT * FROM users
WHERE id = @id;

-- name: CreateUser :one
INSERT INTO users (
    email,
    password_hash,
    first_name,
    last_name,
    preferred_currency,
    timezone
) VALUES (
    LOWER(@email),
    @password_hash,
    @first_name,
    @last_name,
    @preferred_currency,
    @timezone
) RETURNING *;

-- name: UpdateUser :one
UPDATE users SET
    first_name = @first_name,
    last_name = @last_name,
    preferred_currency = @preferred_currency,
    timezone = @timezone
WHERE id = @id
RETURNING *;
