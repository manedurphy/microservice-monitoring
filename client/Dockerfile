FROM golang:1.17.1-alpine3.14

WORKDIR /app

COPY ./ ./

RUN go build -o client .

ENTRYPOINT [ "/app/client" ]
