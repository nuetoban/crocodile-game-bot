BEGIN;

CREATE TABLE IF NOT EXISTS chats
AS SELECT
    DISTINCT chat_id AS id, '' AS title
FROM
    user_in_chats;

ALTER TABLE chats
ADD PRIMARY KEY (id);

ALTER TABLE user_in_chats
ADD CONSTRAINT
    user_in_chats_chat_id_chats_id_foreign
FOREIGN KEY (chat_id) REFERENCES
    chats (id);

COMMIT;
