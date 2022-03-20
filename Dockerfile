FROM golang:1.18rc1-alpine3.15

RUN apk add --no-cache git

WORKDIR /app/kraken

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN go test -cover -v ./...

RUN go build cmd/orderbook/main.go

CMD [ "./main" ]

