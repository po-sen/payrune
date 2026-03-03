FROM golang:1.25-alpine AS builder

WORKDIR /src

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o /out/payrune ./cmd/app

FROM alpine:3.21

RUN addgroup -S app && adduser -S -G app app

WORKDIR /app

COPY --from=builder /out/payrune /app/payrune

EXPOSE 8080

USER app

ENTRYPOINT ["/app/payrune"]
