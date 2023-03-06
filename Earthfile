VERSION 0.6
FROM golang:1.20-alpine
WORKDIR /baseplate-go

deps:
    FROM +base
    COPY go.mod go.sum ./
    RUN go mod download

artifact-config:
    FROM +deps
    COPY ./config/ ./config/
    SAVE ARTIFACT ./config

artifact:
    FROM +deps
    COPY . .
    SAVE ARTIFACT .
