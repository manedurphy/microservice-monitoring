FROM golang:1.17.1-alpine3.14 as builder

WORKDIR /app

COPY ./ ./

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o server .

FROM alpine:3.14.1

WORKDIR /app

COPY --from=builder /app/server /app/server

USER 1000

ENTRYPOINT [ "/app/server" ]
