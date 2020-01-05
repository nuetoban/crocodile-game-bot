#!/bin/bash

set -ex

cd $(dirname "$0")

CHATTONOTIFY="-1001493773956"
DBCONTAINERID="$(docker ps | grep crocodile-game-bot_postgres | awk '{ print $1 }')"

CHATSCOUNT=$(docker exec "$DBCONTAINERID" \
	psql -qtAX -U $CROCODILE_GAME_DB_USER -W $CROCODILE_GAME_DB_NAME -c 'SELECT COUNT(DISTINCT(chat_id)) FROM user_in_chats WHERE chat_id != id;')

PREVVALUE="$(cat /var/tmp/crocodile-chats-count.txt)"
ROUNDED=$(( $CHATSCOUNT / 100 ))

echo $(( $ROUNDED )) > /var/tmp/crocodile-chats-count.txt

if (( $PREVVALUE < $ROUNDED )); then
    set +x  # Let's hide the token
    curl -s "https://api.telegram.org/bot$CROCODILE_GAME_BOT_TOKEN/sendMessage" \
        --data-urlencode "chat_id=$CHATTONOTIFY" \
        --data-urlencode "text=Количество чатов достигло $(( $ROUNDED * 100 ))" \
        | jq '.ok'
fi
