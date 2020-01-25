BEGIN;

ALTER TABLE user_in_chats
DROP CONSTRAINT
    user_in_chats_chat_id_chats_id_foreign;

DROP TABLE chats;

COMMIT;