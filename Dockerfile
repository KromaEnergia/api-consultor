FROM golang:1.24-alpine

WORKDIR /app

# Copia os arquivos de dependências e baixa módulos
COPY go.mod go.sum ./
RUN go mod download

# Copia o restante do código e compila
COPY . .
RUN go build -o main ./cmd/main.go

EXPOSE 8080

CMD ["./main"]
