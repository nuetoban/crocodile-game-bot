-- Possible TODO:
-- One DB; Tables: Users, Chats, Rating history, Settings.
-- Users table stores all the users data like name, id etc.
-- Chats table stores all the chats data like chat name, chat id etc.
-- History rating table stores data about seasons: current season number, 
-- historical rating in previous seasons etc.
-- Settings table consists of user and chat data, that may be affected by users.
-- For example, users may want to drop their current rating, but that rating
-- should remain in the overall top rating by chats.

BEGIN;

-- Create table with common users data (name, id), pulling data from user_in_chats table

CREATE TABLE IF NOT EXISTS users
AS SELECT
    DISTINCT (id), (array_agg("name"))[1] as name
FROM
    user_in_chats GROUP BY id;

ALTER TABLE users
ADD PRIMARY KEY (id);

ALTER TABLE user_in_chats
ADD CONSTRAINT
    user_in_chats_id_users_foreign
FOREIGN KEY (id) REFERENCES
    users (id);

ALTER TABLE user_in_chats
DROP COLUMN name;

-- Create table rating_history for rating history ( :D ), pulling data from user_in_chats table.
-- It stores information about game seasons

CREATE TABLE IF NOT EXISTS rating_history
AS SELECT
    id AS user_id,
    chat_id AS chat_id,
    0 AS season,
    was_host,
    success,
    guessed
FROM
    user_in_chats;

ALTER TABLE rating_history
ADD PRIMARY KEY (user_id, chat_id, season);

ALTER TABLE rating_history
ADD CONSTRAINT
    rating_history_chat_id_chats_id_foreign
FOREIGN KEY (chat_id) REFERENCES
    chats (id);

ALTER TABLE rating_history
ADD CONSTRAINT
    rating_history_user_id_users_id_foreign
FOREIGN KEY (user_id) REFERENCES
    users (id);

-- Create table settings (simple key-value)

CREATE TABLE IF NOT EXISTS settings(
    key TEXT,
    value JSONB
);

INSERT INTO settings VALUES ('season', '{ "value": 0 }');

COMMIT;
