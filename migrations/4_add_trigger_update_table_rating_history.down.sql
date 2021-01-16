BEGIN;

DROP TRIGGER IF EXISTS user_in_chats_insert ON user_in_chats;

DROP FUNCTION IF EXISTS insert_rating_history;

DROP TRIGGER IF EXISTS user_in_chats_update ON user_in_chats;

DROP FUNCTION IF EXISTS update_rating_history;

COMMIT;
