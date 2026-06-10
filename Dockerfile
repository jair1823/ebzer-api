FROM golang:1.25-bookworm AS build

WORKDIR /src

RUN apt-get update \
    && apt-get install -y --no-install-recommends build-essential ca-certificates \
    && rm -rf /var/lib/apt/lists/*

RUN apt-get update && apt-get install -y sqlite3

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 GOOS=linux go build -o /out/ebzer-api ./cmd/server

FROM debian:bookworm-slim

WORKDIR /app

RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates \
    && rm -rf /var/lib/apt/lists/* \
    && mkdir -p /app/data /app/internal/db

COPY --from=build /out/ebzer-api /app/ebzer-api
COPY internal/db/migrations /app/internal/db/migrations

ENV SQLITE_DB_PATH=/app/data/ebzer.db

EXPOSE 3000

CMD ["./ebzer-api"]
