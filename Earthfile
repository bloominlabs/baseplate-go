VERSION 0.6
FROM golang:1.19-alpine
WORKDIR /baseplate-go

deps:
    FROM +base
    COPY go.mod go.sum ./
    RUN go mod download

artifact:
    FROM +deps
    COPY . .
    SAVE ARTIFACT .
