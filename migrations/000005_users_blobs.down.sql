SET search_path TO voco;

ALTER TABLE rooms ALTER COLUMN owner_id TYPE TEXT USING owner_id::text;

ALTER TABLE blobs DROP CONSTRAINT IF EXISTS blobs_owner_user_id_fkey;

DROP INDEX IF EXISTS idx_users_email;
DROP INDEX IF EXISTS idx_users_nickname_lower;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS blobs;
