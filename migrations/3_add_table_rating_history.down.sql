BEGIN;

DROP TABLE IF EXISTS rating_history;

ALTER TABLE user_in_chats
DROP CONSTRAINT user_in_chats_id_users_foreign;

DROP TABLE IF EXISTS users;

DROP TABLE IF EXISTS settings;

COMMIT;
