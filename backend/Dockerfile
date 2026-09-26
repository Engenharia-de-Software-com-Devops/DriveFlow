# ---- Etapa de construcao -------------------------------------------------
FROM golang:1.24-alpine AS construcao

WORKDIR /src

# As dependencias entram antes do codigo para aproveitar o cache de camadas:
# alterar o codigo nao refaz o download dos modulos.
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Version reported by GET /health; the pipeline passes the release being built.
# Left empty (local builds), the default in server.go stays.
ARG VERSION=

# Binario estatico: a imagem final nao precisa de toolchain Go nem de libc.
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath \
    -ldflags="-s -w ${VERSION:+-X driveflow/backend/internal/delivery/http.Version=$VERSION}" \
    -o /bin/api ./cmd/app

# ---- Imagem final --------------------------------------------------------
FROM alpine:3.20

RUN apk add --no-cache wget ca-certificates \
    && adduser -D -u 10001 driveflow

COPY --from=construcao /bin/api /usr/local/bin/api

# A api roda sem privilegios de root.
USER driveflow

ENV API_PORT=8080
EXPOSE 8080

HEALTHCHECK --interval=10s --timeout=3s --start-period=15s --retries=5 \
    CMD wget --quiet --tries=1 --spider http://localhost:8080/health || exit 1

ENTRYPOINT ["api"]
