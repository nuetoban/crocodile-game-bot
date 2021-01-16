BEGIN;

-- Trigger for UPDATE action

CREATE OR REPLACE FUNCTION update_rating_history()
    RETURNS trigger AS
$BODY$
DECLARE
    current_season INTEGER;
    rows_with_current_season INTEGER;

BEGIN
    SELECT INTO current_season value ->> 'value' AS value FROM settings WHERE key = 'season';

    SELECT INTO rows_with_current_season COUNT(*) FROM rating_history
    WHERE user_id = old.id AND chat_id = old.chat_id AND season = current_season;

    IF rows_with_current_season = 0 THEN
        INSERT INTO rating_history(user_id, chat_id, season, was_host, success, guessed)
        VALUES(new.id, new.chat_id, current_season, new.was_host - old.was_host, new.success - old.success, new.guessed - old.guessed);
    ELSE
        IF new.was_host > old.was_host THEN
            UPDATE rating_history
            SET was_host = was_host + 1
            WHERE rating_history.user_id = old.id AND rating_history.chat_id = old.chat_id AND season = current_season;
        END IF;
        IF new.success > old.success THEN
            UPDATE rating_history
            SET success = success + 1
            WHERE rating_history.user_id = old.id AND rating_history.chat_id = old.chat_id AND season = current_season;
        END IF;
        IF new.guessed > old.guessed THEN
            UPDATE rating_history
            SET guessed = guessed + 1
            WHERE rating_history.user_id = old.id AND rating_history.chat_id = old.chat_id AND season = current_season;
        END IF;
    END IF;

    RETURN NEW;
END;
$BODY$ LANGUAGE plpgsql;

CREATE TRIGGER user_in_chats_update
    AFTER UPDATE ON user_in_chats
    FOR EACH ROW
    EXECUTE PROCEDURE update_rating_history();

-- Trigger for INSERT action

CREATE OR REPLACE FUNCTION insert_rating_history()
    RETURNS trigger AS
$BODY$
DECLARE
    current_season INTEGER;

BEGIN
    SELECT INTO current_season value ->> 'value' AS value FROM settings WHERE key = 'season';

    INSERT INTO rating_history(user_id, chat_id, season, was_host, success, guessed)
    VALUES(new.id, new.chat_id, current_season, new.was_host, new.success, new.guessed);

    RETURN NEW;
END;
$BODY$ LANGUAGE plpgsql;

CREATE TRIGGER user_in_chats_insert
    AFTER INSERT ON user_in_chats
    FOR EACH ROW
    EXECUTE PROCEDURE insert_rating_history();

COMMIT;
