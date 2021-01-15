# Crocodile Game Bot
This is a Crocodile Game bot for Telegram.
I'm not sure that such game is called "Crocodile" in English,
but in Russian it's called that way. `¯\_(ツ)_/¯`

The main bot instance is here: https://t.me/Crocodile_Game_Bot

# ⚠️ This repo is no longer maintained
We've migrated to a private repository.

## Installation
```
go get -u github.com/nuetoban/crocodile-game-bot
```

## Running
1. Copy .env.example to .env and fill variables
```
cp .env{.example,}                               # Copy
vim .env                                         # Edit
source <(cat .env | awk '{print "export", $1}')  # Set variables
```

2. Run redis and postgresql
```
docker-compose up -d redis
docker-compose up -d postgresql
```

3. Apply migrations — see "Database Migrations" section
```
make migrate-up
```

4. Run application
```
make run
```

## Testing
Execute this command:
```
make test
```

## Database Migrations
We use https://github.com/golang-migrate/migrate to perform migrations.
Them are stored in ./migrations/ folder.

To apply migrations to your database do the following:

1. Install `migrate` tool: https://github.com/golang-migrate/migrate/tree/master/cmd/migrate
2. Apply migrations
```
migrate -source file://migrations -database 'postgres://user:pass@localhost:5432/postgres?sslmode=disable' up
```

If you need to downgrade the database schema, run this command
```
migrate -source file://migrations -database 'postgres://user:pass@localhost:5432/postgres?sslmode=disable' down
```
Just `down` instead of `up` in the end of the command.
