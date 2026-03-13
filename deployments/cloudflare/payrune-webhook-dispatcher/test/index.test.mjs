import test from 'node:test';
import assert from 'node:assert/strict';

import {
  buildCycleErrorMessage,
  buildCycleLogMessage,
  buildScheduledEnvelope,
  snapshotEnv,
} from '../src/index.mjs';
import { __test as webhookBridgeTest, registerWebhookNotifierBridge, unregisterWebhookNotifierBridge } from '../src/webhook-notifier-bridge.mjs';

test('snapshotEnv keeps only string values', () => {
  const env = snapshotEnv({
    PAYMENT_RECEIPT_WEBHOOK_SECRET: 'secret',
    NUMBER_VALUE: 123,
  });

  assert.deepEqual(env, {
    PAYMENT_RECEIPT_WEBHOOK_SECRET: 'secret',
  });
});

test('buildScheduledEnvelope keeps env and schedule metadata', () => {
  const envelope = buildScheduledEnvelope(
    {
      scheduledTime: Date.parse('2026-03-13T12:00:00Z'),
      cron: '* * * * *',
    },
    { RECEIPT_WEBHOOK_DISPATCH_BATCH_SIZE: '50' },
    'pg-bridge',
  );

  assert.equal(envelope.postgresBridgeId, 'pg-bridge');
  assert.equal(envelope.scheduledTime, '2026-03-13T12:00:00.000Z');
  assert.equal(envelope.cron, '* * * * *');
  assert.equal(envelope.env.RECEIPT_WEBHOOK_DISPATCH_BATCH_SIZE, '50');
});

test('buildCycleLogMessage formats summary output', () => {
  assert.equal(
    buildCycleLogMessage({ claimedCount: 2, sentCount: 1, retriedCount: 3, failedCount: 4 }),
    'webhook dispatch cycle complete claimed=2 sent=1 retried=3 failed=4',
  );
});

test('buildCycleErrorMessage returns constant scope', () => {
  assert.equal(buildCycleErrorMessage(), 'webhook dispatch cycle failed');
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
