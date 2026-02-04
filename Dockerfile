FROM golang:1.24-alpine as buildbase

# Змінили шлях робочої директорії під новий проект
WORKDIR /go/src/github.com/alwayswannafeed/eth-ind

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /usr/local/bin/eth-ind .

FROM alpine:3.20

RUN apk add --no-cache ca-certificates

COPY --from=buildbase /usr/local/bin/eth-ind /usr/local/bin/eth-ind
COPY config.yaml /config.yaml

RUN chmod +x /usr/local/bin/eth-ind

ENTRYPOINT ["/usr/local/bin/eth-ind"]