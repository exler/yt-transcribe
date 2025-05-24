ARG GO_VERSION=1.24

FROM golang:${GO_VERSION}-alpine AS build_go

RUN apk add git

WORKDIR /app
COPY . /app

ENV GO111MODULE=on
ENV CGO_ENABLED=0

RUN go build -tags urfave_cli_no_docs -ldflags "-X github.com/exler/yt-transcribe/cmd.Version=$(git describe --tags)" -o /yt-transcribe

FROM alpine:edge

WORKDIR /app
COPY --from=build_go /yt-transcribe /app/yt-transcribe

RUN apk update && apk add yt-dlp

ENTRYPOINT ["/app/yt-transcribe", "runserver"]

EXPOSE 8000
