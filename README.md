## Payrune

Payrune is a payment-address service for Bitcoin and Ethereum.

It does 3 things:

- allocate one payment address per payment request
- track current payment status by polling chain data
- emit status-change webhooks

It supports:

- `bitcoin` chain
- `ethereum` chain
- `mainnet` and `testnet4`
- `sepolia`
- address schemes: `legacy`, `segwit`, `nativeSegwit`, `taproot`
- `create2` on Ethereum
- assets: Bitcoin-native payments, Ethereum native ETH, and Ethereum ERC-20 USDT

## What You Can Call

Public API:

- `GET /health`
- `GET /v1/chains/bitcoin/address-policies`
- `GET /v1/chains/ethereum/address-policies`
- `GET /v1/chains/bitcoin/addresses?addressPolicyId=...&index=...`
- `POST /v1/chains/bitcoin/payment-addresses`
- `POST /v1/chains/ethereum/payment-addresses`
- `GET /v1/chains/bitcoin/payment-addresses/{paymentAddressId}`
- `GET /v1/chains/ethereum/payment-addresses/{paymentAddressId}`

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
  "decimals": 8,
  "address": "bc1qexamplepaymentaddress"
}
```

For Sepolia smoke testing, use Tether's published USD₮ test-token contract
`0xd077a400968890eacc75cdc901f0356c943e4fdb` and acquire test tokens from the Pimlico or Candide
faucets linked in Tether WDK docs before enabling `ETHEREUM_SEPOLIA_USDT_CREATE2_ENABLED=true`.

Create an Ethereum USDT payment address:

```bash
curl -X POST http://localhost:8080/v1/chains/ethereum/payment-addresses \
  -H 'Content-Type: application/json' \
  -H 'Idempotency-Key: order-20260405-usdt-001' \
  -d '{
    "addressPolicyId": "ethereum-sepolia-usdt-create2",
    "expectedAmountMinor": 2500000,
    "customerReference": "order-20260405-usdt-001"
  }'
```

Example success response:

```json
{
  "paymentAddressId": "145",
  "addressPolicyId": "ethereum-sepolia-usdt-create2",
  "expectedAmountMinor": 2500000,
  "customerReference": "order-20260405-usdt-001",
  "chain": "ethereum",
  "network": "sepolia",
  "scheme": "create2",
  "assetReference": "0xd077a400968890eacc75cdc901f0356c943e4fdb",
  "decimals": 6,
  "address": "0x1234567890abcdef1234567890abcdef12345678"
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
  "assetReference": "0xd077a400968890eacc75cdc901f0356c943e4fdb",
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
that each selected receiver still has a non-zero asset balance, and prints the recovery command in
dry-run mode. For native ETH rows it prints one factory batch sweep. For ERC-20 rows it validates
that every selected receiver shares the same asset reference, deploys any missing receiver
contracts through the factory, and sweeps all compatible token receivers through one factory batch
transaction. It compares the selected row factory against checked-in metadata for that network and
uses the row-owned factory recorded in recovery material. Use
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
correct. The script deploys the checked-in `Create2ReceiverFactory` artifact, resolves the target
network from chain ID, and updates checked-in metadata with the deployed factory address.

After CREATE2 factory ABI changes, redeploy the factory on that network before using the current
one-signature ERC-20 batch recovery path. The current checked-in metadata also points new issuance
at the unified `FixedCollectorReceiver` artifact, so redeploy the factory before issuing fresh
Ethereum CREATE2 rows that you expect to recover through the current path.

1. Dry-run a sweep with explicit IDs or addresses:

```bash
DATABASE_URL=postgres://...
ETHEREUM_SWEEP_PAYMENT_ADDRESS_IDS=145,146
ETHEREUM_SWEEP_RPC_URL=https://...
ETHEREUM_SWEEP_FROM_ADDRESS=0xYourLedgerSender
ETHEREUM_SWEEP_DERIVATION_PATH="m/44'/60'/0'/0/0" \
  bash scripts/ethereum_create2_sweep.sh
```

The sweep helper stays Ledger-only, rejects mixed or malformed selections, rejects zero-balance
receivers, and prints one batch recovery command in dry-run mode. The factory derives each
receiver from CREATE2 recovery payload, deploys the receiver if code is still missing, and then
sweeps it. For compatible ERC-20 rows, the helper now prints one batch factory command and
estimates one Ledger signature, even when the selected rows belong to an older factory namespace.
Use
`bash scripts/ethereum_create2_sweep.sh --broadcast` only after the dry-run output is correct. For
one-off recovery, pass exactly one ID or address and the same helper will send a batch of size 1.

## Ledger USDT Payment Helper

Use [`scripts/ethereum_usdt_pay_with_ledger.sh`](/Users/posen/Desktop/payrune/scripts/ethereum_usdt_pay_with_ledger.sh)
to send one USDT payment with a Ledger signer.

Dry-run:

```bash
ETHEREUM_PAYMENT_RPC_URL=https://ethereum-sepolia-rpc.publicnode.com
ETHEREUM_PAYMENT_FROM_ADDRESS=0xYourLedgerSender
ETHEREUM_PAYMENT_TO_ADDRESS=0xRecipientAddress
ETHEREUM_PAYMENT_AMOUNT_MINOR=2500000
ETHEREUM_PAYMENT_DERIVATION_PATH="m/44'/60'/0'/0/0" \
  bash scripts/ethereum_usdt_pay_with_ledger.sh
```

Broadcast:

```bash
ETHEREUM_PAYMENT_RPC_URL=https://ethereum-sepolia-rpc.publicnode.com
ETHEREUM_PAYMENT_FROM_ADDRESS=0xYourLedgerSender
ETHEREUM_PAYMENT_TO_ADDRESS=0xRecipientAddress
ETHEREUM_PAYMENT_AMOUNT_MINOR=2500000
ETHEREUM_PAYMENT_DERIVATION_PATH="m/44'/60'/0'/0/0" \
  bash scripts/ethereum_usdt_pay_with_ledger.sh --broadcast
```

The helper resolves `mainnet` or `sepolia` from `ETHEREUM_PAYMENT_RPC_URL`, defaults the USDT
asset reference to the known network contract address when no override is set, and only validates
the connected Ledger sender in `--broadcast` mode.

## Sepolia Test USDt

For Sepolia test flows, use Tether's published USD₮ test-token contract:

- `0xd077a400968890eacc75cdc901f0356c943e4fdb`

Get free test USD₮ from:

- Pimlico faucet: `https://dashboard.pimlico.io/test-erc20-faucet`
- Candide faucet: `https://dashboard.candide.dev/faucet`

The checked-in [`deployments/compose/compose.dev.env`](/Users/posen/Desktop/payrune/deployments/compose/compose.dev.env)
already points `ETHEREUM_SEPOLIA_USDT_ASSET_REFERENCE` at that address and enables the Sepolia
CREATE2 policies explicitly.

## Main Parameters

API / core:

- `DATABASE_URL`: PostgreSQL for local/process runtime
- `*_ENABLED`: explicit operator intent for each address policy; disabled policies are skipped by config validation and Ethereum startup readiness
- `BITCOIN_MAINNET_*_XPUB`: required when the matching mainnet Bitcoin policy is enabled
- `BITCOIN_TESTNET4_*_XPUB`: required when the matching testnet4 Bitcoin policy is enabled
- `BITCOIN_MAINNET_REQUIRED_CONFIRMATIONS`: default `2`
- `BITCOIN_TESTNET4_REQUIRED_CONFIRMATIONS`: default `2`
- `BITCOIN_MAINNET_RECEIPT_EXPIRES_AFTER`: default `24h`
- `BITCOIN_TESTNET4_RECEIPT_EXPIRES_AFTER`: default `24h`
- `ETHEREUM_MAINNET_CREATE2_COLLECTOR_ADDRESS`: required when `ETHEREUM_MAINNET_CREATE2_ENABLED=true`
- `ETHEREUM_MAINNET_CREATE2_DERIVATION_KEY`: required when `ETHEREUM_MAINNET_CREATE2_ENABLED=true` or `ETHEREUM_MAINNET_USDT_CREATE2_ENABLED=true`
- `ETHEREUM_MAINNET_USDT_ASSET_REFERENCE`: required when `ETHEREUM_MAINNET_USDT_CREATE2_ENABLED=true`
- `ETHEREUM_SEPOLIA_CREATE2_COLLECTOR_ADDRESS`: required when `ETHEREUM_SEPOLIA_CREATE2_ENABLED=true`
- `ETHEREUM_SEPOLIA_CREATE2_DERIVATION_KEY`: required when `ETHEREUM_SEPOLIA_CREATE2_ENABLED=true` or `ETHEREUM_SEPOLIA_USDT_CREATE2_ENABLED=true`
- `ETHEREUM_SEPOLIA_USDT_ASSET_REFERENCE`: required when `ETHEREUM_SEPOLIA_USDT_CREATE2_ENABLED=true`
- `ETHEREUM_PAYMENT_RPC_URL`
- `ETHEREUM_PAYMENT_FROM_ADDRESS`
- `ETHEREUM_PAYMENT_TO_ADDRESS`
- `ETHEREUM_PAYMENT_AMOUNT_MINOR`
- `ETHEREUM_PAYMENT_DERIVATION_PATH`
- `ETHEREUM_PAYMENT_ASSET_REFERENCE`

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

- copy [`cloudflare.env.example`](/Users/posen/Desktop/payrune/deployments/cloudflare/cloudflare.env.example) to [`cloudflare.env`](/Users/posen/Desktop/payrune/deployments/cloudflare/cloudflare.env)

## Deploy

Local Docker Compose:

Default behavior:

- `make up`, `make down`, and `make config` use [`compose.dev.env`](/Users/posen/Desktop/payrune/deployments/compose/compose.dev.env) and add the `development` profile on top of the base stack
- `make up-mainnet`, `make down-mainnet`, and `make config-mainnet` use [`compose.env`](/Users/posen/Desktop/payrune/deployments/compose/compose.env) with the base stack only and no extra profile
- `make help` prints the supported local and Cloudflare entrypoints

Base-stack-only local compose:

```bash
cp deployments/compose/compose.env.example deployments/compose/compose.env
make up-mainnet
make down-mainnet
make config-mainnet
```

Unified example env:

- [`deployments/compose/compose.env.example`](/Users/posen/Desktop/payrune/deployments/compose/compose.env.example) includes both base mainnet blocks and local development-chain blocks (`bitcoin testnet4`, `ethereum sepolia`)
- the base-stack-only path keeps the local development policy flags disabled by default in that example
- checked-in [`deployments/compose/compose.dev.env`](/Users/posen/Desktop/payrune/deployments/compose/compose.dev.env) remains the ready-to-run local development env file and now keeps only the required development overrides

Local development path:

```bash
rm -f deployments/compose/compose.env
make up
make down
make config
```

Notes:

- the base stack still includes the mainnet pollers and other unprofiled services
- the `development` profile adds the extra local development services on top of that base stack, including the testnet4/sepolia pollers, webhook mock, and Swagger UI

Cloudflare Workers:

```bash
wrangler login
cp deployments/cloudflare/cloudflare.env.example deployments/cloudflare/cloudflare.env
make cf-up
make cf-down
```

Cloudflare credentials:

- you can keep `CLOUDFLARE_ACCOUNT_ID` and `CLOUDFLARE_API_TOKEN` in `deployments/cloudflare/cloudflare.env`
- local interactive deploy can still use `wrangler login`
- CI or non-interactive deploy can also source the same values from CI secrets

`make cf-up` does:

1. migrate PostgreSQL
2. deploy `receipt-webhook-mock`
3. deploy the unified `payrune` worker for API, poller, and dispatcher
