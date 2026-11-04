-- name: ListInstitutions :many
SELECT * FROM institutions
ORDER BY name ASC;

-- name: GetInstitutionByID :one
SELECT * FROM institutions
WHERE id = @id;

-- name: CreateInstitution :one
INSERT INTO institutions (
    name,
    display_name,
    logo_url,
    website
) VALUES (
    @name,
    @display_name,
    @logo_url,
    @website
) RETURNING *;
