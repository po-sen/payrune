FROM golang:1.25-alpine AS builder

WORKDIR /src

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o /out/migrate ./cmd/migrate

FROM alpine:3.21

RUN addgroup -S app && adduser -S -G app app

WORKDIR /app

COPY --from=builder /out/migrate /app/migrate
COPY deployments/postgresql/migrations /app/migrations

USER app

ENTRYPOINT ["/app/migrate"]
CMD ["up"]
