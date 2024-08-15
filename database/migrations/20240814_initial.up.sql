BEGIN;

CREATE TABLE webauthn_users (
    "_id" BIGSERIAL PRIMARY KEY,
    "ref_id" VARCHAR(255) NOT NULL UNIQUE,
    "name" VARCHAR(255) NOT NULL,
    "display_name" VARCHAR(255) NOT NULL
);

CREATE TABLE webauthn_credentials (
    "credential_id" bytea PRIMARY KEY,
    "user_id" BIGSERIAL REFERENCES webauthn_users("_id") ON DELETE CASCADE,
    "use_counter" INT NOT NULL DEFAULT 0,
    "public_key" bytea,
    "attestation_type" VARCHAR(25),
    "transport" JSON,
    "flags" JSON,
    "authenticator" JSON,
    "attestation" JSON
);

COMMIT;
