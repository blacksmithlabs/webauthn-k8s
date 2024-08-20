-- name: InsertUser :one
INSERT INTO webauthn_users (
    "ref_id", "raw_id", "name", "display_name"
) VALUES (
    $1, $2, $3, $4
) RETURNING *;

-- name: UpsertUser :one
INSERT INTO webauthn_users (
    "ref_id", "raw_id", "name", "display_name"
) VALUES (
    $1, $2, $3, $4
)
ON CONFLICT (ref_id)
DO UPDATE set ref_id = EXCLUDED.ref_id
RETURNING *;

-- name: ListUsers :many
SELECT *
FROM webauthn_users;

-- name: GetUserByID :one
SELECT *
FROM webauthn_users
WHERE _id = $1;

-- name: GetUserByRef :one
SELECT *
FROM webauthn_users
WHERE ref_id = $1;

-- name: GetUserWithCredentialsByID :many
SELECT sqlc.embed(webauthn_users), sqlc.embed(webauthn_credentials)
FROM webauthn_users
LEFT JOIN webauthn_credentials ON webauthn_users._id = webauthn_credentials.user_id
WHERE webauthn_users._id = $1
ORDER BY webauthn_credentials.credential_id;

-- name: GetUserWithCredentialsByRef :many
SELECT sqlc.embed(webauthn_users), sqlc.embed(webauthn_credentials)
FROM webauthn_users
LEFT JOIN webauthn_credentials ON webauthn_users._id = webauthn_credentials.user_id
WHERE webauthn_users.ref_id = $1
ORDER BY webauthn_credentials.credential_id;

-- name: UpdateUser :one
UPDATE webauthn_users
SET "name" = $2, display_name = $3
WHERE ref_id = $1
RETURNING *;


-- name: InsertCredential :one
INSERT INTO webauthn_credentials (
    "credential_id", "user_id", "public_key", "attestation_type", "transport", "flags", "authenticator", "attestation"
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
) RETURNING *;

-- name: ListCredentialsByUser :many
SELECT sqlc.embed(webauthn_credentials), sqlc.embed(webauthn_users)
FROM webauthn_credentials
INNER JOIN webauthn_users ON webauthn_credentials.user_id = webauthn_users._id
WHERE user_id = $1;

-- name: GetCredential :one
SELECT sqlc.embed(webauthn_credentials), sqlc.embed(webauthn_users)
FROM webauthn_credentials
INNER JOIN webauthn_users ON webauthn_credentials.user_id = webauthn_users._id
WHERE credential_id = $1;

-- name: IncrementCredentialUseCounter :one
UPDATE webauthn_credentials
SET use_counter = use_counter + 1
WHERE credential_id = $1
RETURNING use_counter;
