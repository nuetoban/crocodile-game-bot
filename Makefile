.PHONY: run docker get test graph

default: run

docker:
	docker build --rm -t github.com/nuetoban/crocodile-game-bot .

run:
	go run bot.go log.go exporter.go

get:
	go get -v ./...

test:
	go test ./...

graph:
	go get -u github.com/TrueFurby/go-callvis
	go-callvis -focus github.com/nuetoban/crocodile-game-bot/crocodile \
		-group pkg,type -nostd -format=png \
		-ignore github.com/sirupsen/logrus . | dot -Tpng -o crocodile.png
