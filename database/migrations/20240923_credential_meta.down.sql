BEGIN;

ALTER TABLE webauthn_credentials
    DROP COLUMN "meta";

COMMIT;
