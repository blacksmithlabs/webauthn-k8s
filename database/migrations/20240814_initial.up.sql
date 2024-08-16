BEGIN;

CREATE TABLE webauthn_users (
    "_id" BIGSERIAL PRIMARY KEY,
    "ref_id" VARCHAR(100) NOT NULL UNIQUE,
    "raw_id" bytea NOT NULL,
    "name" VARCHAR(255) NOT NULL,
    "display_name" VARCHAR(255) NOT NULL
);

CREATE TABLE webauthn_credentials (
    "credential_id" bytea PRIMARY KEY,
    "user_id" BIGINT REFERENCES webauthn_users("_id") ON DELETE CASCADE,
    "use_counter" INT NOT NULL DEFAULT 0,
    "public_key" bytea,
    "attestation_type" VARCHAR(25),
    "transport" JSON,
    "flags" JSON,
    "authenticator" JSON,
    "attestation" JSON
);

COMMIT;
