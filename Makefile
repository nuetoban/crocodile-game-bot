.PHONY: build run docker-build docker-tag docker-push get test graph wc

default: build

docker-build:
	docker build -t crocodile .

docker-tag:
	docker tag crocodile docker.pkg.github.com/nuetoban/crocodile-game-bot/crocodile:latest

docker-push:
	docker push docker.pkg.github.com/nuetoban/crocodile-game-bot/crocodile:latest

docker-full: docker-build docker-tag docker-push

run:
	go run bot.go log.go exporter.go

build:
	go build -a \
		-ldflags '-linkmode external -extldflags "-static"' \
		-o crocodile-server \
		bot.go log.go exporter.go

get:
	go get -v ./...

test:
	go test ./...

graph:
	go get -u github.com/TrueFurby/go-callvis
	go-callvis -focus github.com/nuetoban/crocodile-game-bot/crocodile \
		-group pkg,type -nostd -format=png \
		-ignore github.com/sirupsen/logrus . | dot -Tpng -o crocodile.png

wc:
	find . -name '*.go' | xargs cat | wc -l
