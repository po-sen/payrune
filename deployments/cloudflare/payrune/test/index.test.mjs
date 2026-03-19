import test from 'node:test';
import assert from 'node:assert/strict';

import {
  buildPollerCycleErrorMessage,
  buildPollerCycleLogMessage,
  buildRequestEnvelope,
  buildScheduledEnvelope,
  buildScheduledEnv,
  buildWebhookDispatchCycleErrorMessage,
  buildWebhookDispatchCycleLogMessage,
  buildWorkerResponse,
  isPublicRoute,
  resolveScheduledJob,
  scheduledJobs,
  snapshotEnv,
} from '../src/index.mjs';
import {
  __test as webhookBridgeTest,
  registerWebhookNotifierBridge,
  unregisterWebhookNotifierBridge,
} from '../src/webhook-notifier-bridge.mjs';

test('isPublicRoute allows health and v1 routes', () => {
  assert.equal(isPublicRoute('/health'), true);
  assert.equal(isPublicRoute('/v1/chains/bitcoin/address-policies'), true);
  assert.equal(isPublicRoute('/private'), false);
});

test('buildRequestEnvelope keeps path, query, headers, and body', () => {
  const request = new Request('https://example.com/v1/test?foo=bar', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Idempotency-Key': 'abc',
    },
  });
  const env = {
    BITCOIN_MAINNET_TAPROOT_XPUB: 'xpub',
  };

  const envelope = buildRequestEnvelope(request, env, 'bridge-123', '{"ok":true}');
  assert.equal(envelope.bridgeId, 'bridge-123');
  assert.equal(envelope.request.path, '/v1/test');
  assert.equal(envelope.request.rawQuery, 'foo=bar');
  assert.equal(envelope.request.headers['content-type'], 'application/json');
  assert.equal(envelope.env.BITCOIN_MAINNET_TAPROOT_XPUB, 'xpub');
  assert.equal(envelope.request.body, '{"ok":true}');
});

test('buildWorkerResponse returns fetch Response', async () => {
  const response = buildWorkerResponse({
    status: 201,
    headers: {
      'Content-Type': 'application/json',
      'Idempotency-Replayed': 'true',
    },
    body: '{"ok":true}',
  });

  assert.equal(response.status, 201);
  assert.equal(response.headers.get('Idempotency-Replayed'), 'true');
  assert.equal(await response.text(), '{"ok":true}');
});

test('snapshotEnv keeps only string values', () => {
  const env = snapshotEnv({
    POLL_NETWORK: 'mainnet',
    NUMBER_VALUE: 123,
  });

  assert.deepEqual(env, {
    POLL_NETWORK: 'mainnet',
  });
});

test('buildScheduledEnv applies overrides on top of string snapshot', () => {
  const env = buildScheduledEnv(
    {
      POLL_NETWORK: 'mainnet',
      RECEIPT_WEBHOOK_MOCK: { fetch() {} },
    },
    {
      POLL_NETWORK: 'testnet4',
      POLL_CHAIN: 'bitcoin',
    },
  );

  assert.deepEqual(env, {
    POLL_NETWORK: 'testnet4',
    POLL_CHAIN: 'bitcoin',
  });
});

test('resolveScheduledJob returns poller and dispatcher mappings', () => {
  assert.deepEqual(resolveScheduledJob({ cron: '5,20,35,50 * * * *' }), scheduledJobs['5,20,35,50 * * * *']);
  assert.deepEqual(resolveScheduledJob({ cron: '*/15 * * * *' }), scheduledJobs['*/15 * * * *']);
  assert.deepEqual(resolveScheduledJob({ cron: '10,25,40,55 * * * *' }), scheduledJobs['10,25,40,55 * * * *']);
  assert.equal(resolveScheduledJob({ cron: '0 * * * *' }), null);
});

test('buildScheduledEnvelope keeps env, bridge IDs, and schedule metadata', () => {
  const envelope = buildScheduledEnvelope(
    {
      scheduledTime: Date.parse('2026-03-19T12:00:00Z'),
      cron: '5,20,35,50 * * * *',
    },
    { POLL_CHAIN: 'bitcoin' },
    { postgresBridgeId: 'pg-bridge', bitcoinBridgeId: 'btc-bridge' },
    { POLL_NETWORK: 'mainnet' },
  );

  assert.equal(envelope.postgresBridgeId, 'pg-bridge');
  assert.equal(envelope.bitcoinBridgeId, 'btc-bridge');
  assert.equal(envelope.scheduledTime, '2026-03-19T12:00:00.000Z');
  assert.equal(envelope.cron, '5,20,35,50 * * * *');
  assert.equal(envelope.env.POLL_CHAIN, 'bitcoin');
  assert.equal(envelope.env.POLL_NETWORK, 'mainnet');
});

test('buildPollerCycleLogMessage formats summary output', () => {
  assert.equal(
    buildPollerCycleLogMessage(
      { POLL_CHAIN: 'bitcoin', POLL_NETWORK: 'testnet4' },
      { claimedCount: 2, updatedCount: 1, terminalFailedCount: 0, processingErrorCount: 3 },
    ),
    'poll cycle complete chain=bitcoin network=testnet4 claimed=2 updated=1 terminal_failed=0 processing_errors=3',
  );
});

test('buildPollerCycleErrorMessage includes scope', () => {
  assert.equal(
    buildPollerCycleErrorMessage({ POLL_CHAIN: 'bitcoin', POLL_NETWORK: 'mainnet' }),
    'poll cycle failed chain=bitcoin network=mainnet',
  );
});

test('buildWebhookDispatchCycleLogMessage formats summary output', () => {
  assert.equal(
    buildWebhookDispatchCycleLogMessage({ claimedCount: 2, sentCount: 1, retriedCount: 3, failedCount: 4 }),
    'webhook dispatch cycle complete claimed=2 sent=1 retried=3 failed=4',
  );
});

test('buildWebhookDispatchCycleErrorMessage returns constant scope', () => {
  assert.equal(buildWebhookDispatchCycleErrorMessage(), 'webhook dispatch cycle failed');
});

test('webhook bridge uses service binding when binding target is provided', async () => {
  let called = false;
  const env = {
    RECEIPT_WEBHOOK_MOCK: {
      async fetch(request) {
        called = true;
        assert.equal(new URL(request.url).pathname, '/receipt-status');
        return new Response(null, { status: 204 });
      },
    },
  };

  registerWebhookNotifierBridge(env);
  try {
    await globalThis.__payruneWebhookPost(
      'RECEIPT_WEBHOOK_MOCK',
      '/receipt-status',
      1000,
      { 'Content-Type': 'application/json' },
      '{"ok":true}',
    );
  } finally {
    unregisterWebhookNotifierBridge();
  }

  assert.equal(called, true);
  assert.equal(webhookBridgeTest.defaultBindingPath, '/receipt-status');
});
