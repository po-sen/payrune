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

## Operator Recovery

Use `address_policy_allocations.sweep_material_json` as the only operator-facing recover material.
Phase-1 recovery should not depend on `issuance_ref`, `issuance_ref_kind`, or `address_space_ref`.

Inspect one issued row:

```bash
psql "$DATABASE_URL" -X -c "
  SELECT id, chain, network, address, sweep_material_json
    FROM address_policy_allocations
   WHERE id = 101
     AND allocation_status = 'issued';
"
```

ETH CREATE2 sweep helper:

```bash
DATABASE_URL=postgres://...
ETHEREUM_SWEEP_PAYMENT_ADDRESS_IDS=145
ETHEREUM_SWEEP_RPC_URL=https://...
ETHEREUM_SWEEP_FROM_ADDRESS=0xYourLedgerSender
ETHEREUM_SWEEP_DERIVATION_PATH="m/44'/60'/0'/0/0" \
  bash scripts/ethereum_create2_sweep.sh
```

The helper reads `sweep_material_json` from the DB, validates the connected Ledger sender, checks
that each selected receiver still has a non-zero on-chain balance, and prints the batch recovery
command in dry-run mode. It validates the recorded payload against the active metadata factory for
that network, and the factory derives each receiver from CREATE2 recovery payload, deploys the
receiver if code is still missing, and then sweeps it. Use
`bash scripts/ethereum_create2_sweep.sh --broadcast` only after the dry-run output is correct.

ETH CREATE2 batch sweep:

1. Build the checked-in CREATE2 artifacts:

```bash
bash scripts/ethereum_create2_build_artifacts.sh
```

1. Deploy the CREATE2 factory once per network with Ledger:

```bash
ETHEREUM_SWEEP_RPC_URL=https://...
ETHEREUM_SWEEP_FROM_ADDRESS=0xYourLedgerSender
ETHEREUM_SWEEP_DERIVATION_PATH="m/44'/60'/0'/0/0" \
  bash scripts/ethereum_create2_factory_deploy.sh
```

Use `bash scripts/ethereum_create2_factory_deploy.sh --broadcast` only after the dry-run output is
correct. The script deploys the checked-in `Create2ReceiverFactoryV1` artifact, resolves the target
network from chain ID, and updates checked-in metadata with the deployed factory address.

1. Dry-run a sweep with explicit IDs or addresses:

```bash
DATABASE_URL=postgres://...
ETHEREUM_SWEEP_PAYMENT_ADDRESS_IDS=145,146
ETHEREUM_SWEEP_RPC_URL=https://...
ETHEREUM_SWEEP_FROM_ADDRESS=0xYourLedgerSender
ETHEREUM_SWEEP_DERIVATION_PATH="m/44'/60'/0'/0/0" \
  bash scripts/ethereum_create2_sweep.sh
```

The sweep helper stays Ledger-only, rejects mixed or malformed selections, rejects stale rows whose
recorded `factory_address` no longer matches the active metadata factory, rejects zero-balance
receivers, and prints one batch recovery command in dry-run mode. The factory derives each receiver
from CREATE2 recovery payload, deploys the receiver if code is still missing, and then sweeps it. Use
`bash scripts/ethereum_create2_sweep.sh --broadcast` only after the dry-run output is correct. For
one-off recovery, pass exactly one ID or address and the same helper will send a batch of size 1.

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

Cloudflare credentials:

- you can keep `CLOUDFLARE_ACCOUNT_ID` and `CLOUDFLARE_API_TOKEN` in `.env.cloudflare`
- local interactive deploy can still use `wrangler login`
- CI or non-interactive deploy can also source the same values from CI secrets

`make cf-up` does:

1. migrate PostgreSQL
2. deploy `receipt-webhook-mock`
3. deploy the unified `payrune` worker for API, poller, and dispatcher
