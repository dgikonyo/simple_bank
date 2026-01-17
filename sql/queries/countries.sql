-- COUNTRIES
-- name: GetCountry :one
SELECT * FROM countries
WHERE code = $1 LIMIT 1;

-- name: ListCountries :many
SELECT * FROM countries
ORDER BY name;

-- name: CreateCountry :one
INSERT INTO countries (
  code, name, continent_name
) VALUES (
  $1, $2, $3
)
RETURNING *;

-- name: UpdateCountry :exec
UPDATE countries
SET name = $2,
    continent_name = $3
WHERE code = $1;

-- name: DeleteCountry :exec
DELETE FROM countries
WHERE code = $1;