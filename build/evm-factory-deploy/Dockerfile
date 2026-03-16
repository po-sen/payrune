FROM golang:1.25-alpine AS builder

WORKDIR /src

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o /out/evm-factory-deploy ./cmd/evm-factory-deploy

FROM alpine:3.21

RUN addgroup -S api && adduser -S -G api api

WORKDIR /api

COPY --from=builder /out/evm-factory-deploy /api/evm-factory-deploy
COPY deployments/ethereum/build/contracts.json /api/deployments/ethereum/build/contracts.json

USER api

ENTRYPOINT ["/api/evm-factory-deploy"]
