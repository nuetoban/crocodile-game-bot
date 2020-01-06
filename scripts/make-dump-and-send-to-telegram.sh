#!/bin/bash

set -ex

DUMPERDIR=$(dirname "$0")
DBDUMPFILE="$DUMPERDIR/dumps/crocodile-dump-$(printf '%(%Y-%m-%d)T' -1)"
DBCONTAINERID="$(docker ps | grep crocodile-game-bot_postgres | awk '{ print $1 }')"

cd "$DUMPERDIR"

# Make database dump
docker exec -e \
	PGPASSWORD="$CROCODILE_GAME_DB_PASS" \
	"$DBCONTAINERID" \
	pg_dump -U "$CROCODILE_GAME_DB_USER" -W "$CROCODILE_GAME_DB_NAME" > "$DBDUMPFILE"

set +x
curl -F "chat_id=-371498376" \
    -F document=@"$DBDUMPFILE" \
    "https://api.telegram.org/bot$CROCODILE_GAME_BOT_TOKEN/sendDocument"
