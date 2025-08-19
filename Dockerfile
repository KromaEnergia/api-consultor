# syntax=docker/dockerfile:1

##########
# BUILD  #
##########
FROM golang:1.24-alpine AS builder
WORKDIR /src

RUN apk add --no-cache git ca-certificates

# deps
COPY go.mod go.sum ./
RUN go mod download

# código
COPY . .

# ✔️ Falha o build se a chave não estiver presente no contexto
RUN test -f /src/keys/private.pem || (echo "ERROR: faltando keys/private.pem no build"; exit 1)

# ✔️ Ajusta permissão segura ainda no builder
RUN chmod 0400 /src/keys/private.pem

# build estático
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -buildvcs=false -ldflags="-s -w" -o app ./cmd/main.go


############
# RUNTIME  #
############
FROM gcr.io/distroless/base-debian12
WORKDIR /app

# binário
COPY --from=builder /src/app /app/app

# ✔️ Copia a pasta keys e já muda owner para o usuário nonroot do distroless
COPY --from=builder --chown=nonroot:nonroot /src/keys /app/keys

ENV PORT=8080
EXPOSE 8080

# roda como não-root
USER nonroot:nonroot
ENTRYPOINT ["/app/app"]
