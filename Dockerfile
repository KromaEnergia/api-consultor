FROM golang:1.24-alpine

WORKDIR /app

# Dependências
COPY go.mod go.sum ./
RUN go mod download

# Código + chaves
COPY . .
# garante que a pasta keys vai pra imagem
# (precisa existir ./keys/private.pem no host)
COPY keys ./keys

# ENVs usadas pelo pacote internal/auth
ENV AUTH_RSA_PRIVATE_PATH=/app/keys/private.pem \
    AUTH_KID=kroma-dev-v1 \
    AUTH_ISSUER=http://localhost:8080 \
    AUTH_AUDIENCE=portal-consultor-local \
    COOKIE_SECURE=false

# build
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/main.go

EXPOSE 8080
CMD ["./main"]
