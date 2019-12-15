BEGIN;

CREATE TABLE IF NOT EXISTS user_in_chats(
    id INTEGER NOT NULL,
    chat_id BIGINT NOT NULL,
    was_host INTEGER,
    success INTEGER,
    guessed INTEGER,
    name TEXT,
    PRIMARY KEY(id, chat_id)
);

CREATE INDEX IF NOT EXISTS user_in_chats_id_idx ON user_in_chats(id);
CREATE INDEX IF NOT EXISTS user_in_chats_chat_id_idx ON user_in_chats(chat_id);

COMMIT;
