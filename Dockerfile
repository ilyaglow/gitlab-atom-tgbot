FROM golang:alpine AS build-env
LABEL maintainer "Ilya Glotov <ilya@ilyaglotov.com>" \
      repository "https://github.com/ilyaglow/gitlab-atom-tgbot"

ENV CGO_ENABLED 0

COPY main.go /go/src/gitlab-atom-tgbot/main.go

RUN apk --no-cache --update add git \
  && cd /go/src/gitlab-atom-tgbot \
  && go get -v -t . \
  && go build -ldflags="-s -w" -o gitlab-atom-tgbot

FROM alpine:edge

RUN apk -U --no-cache add ca-certificates \
  && adduser -D app

COPY --from=build-env /go/src/gitlab-atom-tgbot/gitlab-atom-tgbot /app/gitlab-atom-tgbot

RUN chmod +x /app/gitlab-atom-tgbot \
  && chown -R app /app

USER app

WORKDIR /app/data

ENTRYPOINT ["/app/gitlab-atom-tgbot"]
