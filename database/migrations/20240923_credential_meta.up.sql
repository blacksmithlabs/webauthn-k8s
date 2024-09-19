BEGIN;

ALTER TABLE webauthn_credentials
    ADD COLUMN "meta" JSON;

UPDATE webauthn_credentials
SET meta = '{"active":true}'
WHERE meta IS NULL;

COMMIT;
