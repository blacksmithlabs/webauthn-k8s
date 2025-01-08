BEGIN;

ALTER TABLE webauthn_credentials
    ADD COLUMN "meta" JSON;

UPDATE webauthn_credentials
SET meta = '{"status":"active"}'
WHERE meta IS NULL;

COMMIT;
