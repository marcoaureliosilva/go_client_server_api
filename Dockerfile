# Build para o servidor
FROM golang:1.19 AS builder-server

WORKDIR /app/server

COPY ./server/go.mod ./server/go.sum ./
RUN go mod download

COPY ./server/*.go ./
RUN go build -o /server

# Build para o cliente
FROM golang:1.19 AS builder-client

WORKDIR /app/client

COPY ./client/go.mod ./
RUN go mod download || true

COPY ./client/*.go ./
RUN go build -o /client

# Imagem final
FROM golang:1.19

WORKDIR /app

COPY --from=builder-server /server .
COPY --from=builder-client /client .

CMD ["sh", "-c", "./server & ./client"]
