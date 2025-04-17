FROM golang:1.19 AS build-env

ENV GOOS=linux

WORKDIR /go/src/github.com/runetale/runetale-oidc-server

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -v .

FROM debian:bullseye

RUN apt-get update && apt-get install -y \
    ca-certificates \
    curl \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

COPY --from=build-env /go/src/github.com/runetale/runetale-oidc-server/runetale-oidc-server /runetale-oidc-server
RUN chmod u+x /runetale-oidc-server

COPY docker-entrypoint.sh /docker-entrypoint.sh
RUN chmod u+x ./docker-entrypoint.sh