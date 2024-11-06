ARG GO_VERSION=1
FROM golang:${GO_VERSION}-bookworm as builder

WORKDIR /usr/src/app
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY . .
# RUN go build -v -o /run-app .
RUN go build -v -o /reply-bot ./bot/cmd


FROM debian:bookworm

COPY --from=builder /reply-bot /usr/local/bin/
COPY --from=builder /bsky-feeds.json /bsky-feeds.json
CMD ["reply-bot"]
