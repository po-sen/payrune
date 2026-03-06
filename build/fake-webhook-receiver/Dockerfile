FROM golang:1.25-alpine AS builder

WORKDIR /src

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o /out/fake-webhook-receiver ./cmd/fake_webhook_receiver

FROM alpine:3.21

RUN addgroup -S app && adduser -S -G app app

WORKDIR /app

COPY --from=builder /out/fake-webhook-receiver /app/fake-webhook-receiver

USER app

ENTRYPOINT ["/app/fake-webhook-receiver"]
