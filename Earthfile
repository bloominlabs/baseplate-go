VERSION 0.6
FROM golang:1.20
WORKDIR /baseplate-go

deps:
    FROM +base
    COPY go.mod go.sum ./
    RUN go mod download

artifact-config:
    FROM +deps
    COPY ./config/ ./config/
    SAVE ARTIFACT ./config

artifact-http:
    FROM +deps
    COPY ./http/ ./http/
    SAVE ARTIFACT ./http

artifact:
    FROM +deps
    COPY . .
    SAVE ARTIFACT .
