FROM golang:1.14.1-alpine3.11 AS build

WORKDIR /build

RUN apk add --update make gcc libc-dev

COPY . .
RUN make


FROM alpine:3.11

RUN apk add --no-cache ca-certificates curl
COPY --from=build /build/crocodile-server /crocodile-server
COPY dictionaries /dictionaries/

ENTRYPOINT ["/crocodile-server"]
