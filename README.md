# Crocodile Game Bot
This is a Crocodile Game bot for Telegram.
I'm not sure that such game is called "Crocodile" in English,
but in Russian it's called that way. `¯\_(ツ)_/¯`

The main bot instance is here: https://t.me/Crocodile_Game_Bot

## Installation
```
go get -u github.com/nuetoban/crocodile-game-bot
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

1. Install `migrate` tool
```
go get -u https://github.com/golang-migrate/migrate
```
2. Apply migrations
```
migrate -source file://migrations -database 'postgres://user:pass@localhost:5432/postgres?sslmode=disable' up
```

If you need to downgrade the database schema, run this command
```
migrate -source file://migrations -database 'postgres://user:pass@localhost:5432/postgres?sslmode=disable' down
```
Just `down` instead of `up` in the end of the command.

## Contributing

1. Fork it (<https://github.com/nuetoban/crocodile-game-bot>)
2. Create your feature branch (`git checkout -b feature/fooBar`)
3. Commit your changes (`git commit -am 'Add some fooBar'`)
4. Push to the branch (`git push origin feature/fooBar`)
5. Create a new Pull Request

## Contact
- Telegram: https://t.me/blackstalin
- Mail: nu3toban@gmail.com

- Project link: https://github.com/nuetoban/crocodile-game-bot

## Architecture
![Architecture](crocodile.png)
