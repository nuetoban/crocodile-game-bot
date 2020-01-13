FROM golang:1.13.6-alpine3.11 AS build

WORKDIR /build
COPY . .

RUN apk add --update make gcc libc-dev
RUN make


FROM scratch AS final

COPY --from=build /build/crocodile-server /crocodile-server
COPY dictionaries /dictionaries/

ENTRYPOINT ["/crocodile-server"]
