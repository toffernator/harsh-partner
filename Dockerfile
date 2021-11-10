# Starts the server in a docker container.
FROM golang:1.17-alpine AS base

COPY . /src
WORKDIR /src

RUN go mod download
RUN go build -o "bin/out" server/main.go
EXPOSE 4042
ENTRYPOINT ["bin/out"]
