FROM golang:1.25-alpine AS builder

WORKDIR /src

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o /out/payrune ./cmd/api

FROM alpine:3.21

RUN addgroup -S api && adduser -S -G api api

WORKDIR /api

COPY --from=builder /out/payrune /api/payrune

EXPOSE 8080

USER api

ENTRYPOINT ["/api/payrune"]
