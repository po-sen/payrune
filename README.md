## Payrune

Payrune is a Bitcoin payment-address service.

It does 3 things:

- allocate one payment address per payment request
- track current payment status by polling chain data
- emit status-change webhooks

It supports:

- `bitcoin` chain
- `mainnet` and `testnet4`
- address schemes: `legacy`, `segwit`, `nativeSegwit`, `taproot`

## What You Can Call

Public API:

- `GET /health`
- `GET /v1/chains/bitcoin/address-policies`
- `GET /v1/chains/bitcoin/addresses?addressPolicyId=...&index=...`
- `POST /v1/chains/bitcoin/payment-addresses`
- `GET /v1/chains/bitcoin/payment-addresses/{paymentAddressId}`

Webhook event:

- `payment_receipt.status_changed`

Full HTTP schema:

- `deployments/swagger/openapi.yaml`

## Fast Integration

Create a payment address:

```bash
curl -X POST http://localhost:8080/v1/chains/bitcoin/payment-addresses \
  -H 'Content-Type: application/json' \
  -H 'Idempotency-Key: order-20260313-001' \
  -d '{
    "addressPolicyId": "bitcoin-mainnet-native-segwit",
    "expectedAmountMinor": 120000,
    "customerReference": "order-20260313-001"
  }'
```

Example success response:

```json
{
  "paymentAddressId": "101",
  "addressPolicyId": "bitcoin-mainnet-native-segwit",
  "expectedAmountMinor": 120000,
  "customerReference": "order-20260313-001",
  "chain": "bitcoin",
  "network": "mainnet",
  "scheme": "nativeSegwit",
  "minorUnit": "satoshi",
  "decimals": 8,
  "address": "bc1qexamplepaymentaddress"
}
```

Read current payment status:

```bash
curl http://localhost:8080/v1/chains/bitcoin/payment-addresses/101
```

Possible `paymentStatus` values:

- `watching`
- `partially_paid`
- `paid_unconfirmed`
- `paid_unconfirmed_reverted`
- `paid_confirmed`
- `failed_expired`

## Webhook Contract

Headers:

- `Content-Type: application/json`
- `X-Payrune-Event: payment_receipt.status_changed`
- `X-Payrune-Event-Version: 1`
- `X-Payrune-Notification-ID: <int64>`
- `X-Payrune-Signature-256: sha256=<hex>`

Payload:

```json
{
  "eventType": "payment_receipt.status_changed",
  "eventVersion": 1,
  "notificationId": 1,
  "paymentAddressId": 101,
  "customerReference": "order-20260313-001",
  "previousStatus": "watching",
  "currentStatus": "partially_paid",
  "observedTotalMinor": 80000,
  "confirmedTotalMinor": 40000,
  "unconfirmedTotalMinor": 40000,
  "statusChangedAt": "2026-03-13T12:00:00Z"
}
```

Signature verification:

- algorithm: `HMAC-SHA256`
- signed content: raw request body
- secret: `PAYMENT_RECEIPT_WEBHOOK_SECRET`

Minimal verifier example:

```text
expected = "sha256=" + hex(hmac_sha256(secret, raw_body))
secure_compare(expected, request.headers["X-Payrune-Signature-256"])
```

## Main Parameters

API / core:

- `DATABASE_URL`: PostgreSQL for local/process runtime
- `BITCOIN_MAINNET_*_XPUB`: enable mainnet address policies
- `BITCOIN_TESTNET4_*_XPUB`: enable testnet4 address policies
- `BITCOIN_MAINNET_REQUIRED_CONFIRMATIONS`: default `2`
- `BITCOIN_TESTNET4_REQUIRED_CONFIRMATIONS`: default `2`
- `BITCOIN_MAINNET_RECEIPT_EXPIRES_AFTER`: default `24h`
- `BITCOIN_TESTNET4_RECEIPT_EXPIRES_AFTER`: default `24h`

Poller:

- `POLL_RESCHEDULE_INTERVAL`: how often one payment address is re-polled
- `POLL_BATCH_SIZE`: how many due rows to claim per run
- `POLL_CLAIM_TTL`: lease time for claimed rows
- `BITCOIN_*_ESPLORA_URL`: Esplora endpoint
- `BITCOIN_*_ESPLORA_USER` / `BITCOIN_*_ESPLORA_PASSWORD`: optional provider auth

Webhook:

- `PAYMENT_RECEIPT_WEBHOOK_SECRET`: shared signing secret
- `RECEIPT_WEBHOOK_DISPATCH_BATCH_SIZE`
- `RECEIPT_WEBHOOK_DISPATCH_CLAIM_TTL`
- `RECEIPT_WEBHOOK_DISPATCH_MAX_ATTEMPTS`
- `RECEIPT_WEBHOOK_DISPATCH_RETRY_DELAY`
- `PAYMENT_RECEIPT_WEBHOOK_TIMEOUT`

Cloudflare helper env file:

- copy [`.env.cloudflare.example`](/Users/posen/Desktop/payrune/.env.cloudflare.example) to `.env.cloudflare`

## Deploy

Local Docker Compose:

```bash
make up
make down
```

Cloudflare Workers:

```bash
wrangler login
cp .env.cloudflare.example .env.cloudflare
make cf-up
make cf-down
```

`make cf-up` does:

1. migrate PostgreSQL
2. deploy `payrune-api`
3. deploy `payrune-poller-mainnet`
4. deploy `payrune-poller-testnet4`
5. deploy `receipt-webhook-mock`
6. deploy `payrune-webhook-dispatcher`
