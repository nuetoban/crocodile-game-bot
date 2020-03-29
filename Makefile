IMAGE := "nuetoban/crocodile"
TAG := $(shell git describe --tags)

.PHONY: build run docker-build docker-tag docker-push migrate-up migrate-down get test graph wc

default: build

docker-build:
	docker build -t crocodile .

docker-tag:
	docker tag crocodile nuetoban/crocodile:$(TAG)

docker-push:
	docker push nuetoban/crocodile:$(TAG)

docker-full: docker-build docker-tag docker-push

deploy: docker-full
	helm \
		--namespace crocodile-prod \
		upgrade crocodile ./helm \
		-f helm/values-prod.yaml \
		--set tag=$(TAG)

migrate-up:
	migrate -source file://migrations \
		-database 'postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable' up

migrate-down:
	migrate -source file://migrations -database \
		'postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable' down 1

run:
	go run .

build:
	go build -a \
		-ldflags '-linkmode external -extldflags "-static"' \
		-o crocodile-server .

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
