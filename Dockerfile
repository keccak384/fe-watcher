FROM golang:1.18.9 as builder

WORKDIR /src

COPY go.mod go.mod
COPY go.sum go.sum

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o fe-watcher .

FROM alpine:3.14

RUN apk --no-cache add ca-certificates git

COPY --from=builder /src/fe-watcher /usr/local/bin/fe-watcher

ENTRYPOINT [ "/usr/local/bin/fe-watcher" ]

