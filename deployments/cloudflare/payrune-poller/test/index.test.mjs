import test from 'node:test';
import assert from 'node:assert/strict';

import {
  buildCycleErrorMessage,
  buildCycleLogMessage,
  buildScheduledEnvelope,
  snapshotEnv,
} from '../src/index.mjs';

test('snapshotEnv keeps only string values', () => {
  const env = snapshotEnv({
    POLL_NETWORK: 'mainnet',
    NUMBER_VALUE: 123,
  });

  assert.deepEqual(env, {
    POLL_NETWORK: 'mainnet',
  });
});

test('buildScheduledEnvelope keeps env and schedule metadata', () => {
  const envelope = buildScheduledEnvelope(
    {
      scheduledTime: Date.parse('2026-03-11T12:00:00Z'),
      cron: '* * * * *',
    },
    { POLL_CHAIN: 'bitcoin', POLL_NETWORK: 'mainnet' },
    'pg-bridge',
    'btc-bridge',
  );

  assert.equal(envelope.postgresBridgeId, 'pg-bridge');
  assert.equal(envelope.bitcoinBridgeId, 'btc-bridge');
  assert.equal(envelope.scheduledTime, '2026-03-11T12:00:00.000Z');
  assert.equal(envelope.cron, '* * * * *');
  assert.equal(envelope.env.POLL_NETWORK, 'mainnet');
});

test('buildCycleLogMessage formats summary output', () => {
  assert.equal(
    buildCycleLogMessage(
      { POLL_CHAIN: 'bitcoin', POLL_NETWORK: 'testnet4' },
      { claimedCount: 2, updatedCount: 1, terminalFailedCount: 0, processingErrorCount: 3 },
    ),
    'poll cycle complete chain=bitcoin network=testnet4 claimed=2 updated=1 terminal_failed=0 processing_errors=3',
  );
});

test('buildCycleErrorMessage includes scope', () => {
  assert.equal(
    buildCycleErrorMessage({ POLL_CHAIN: 'bitcoin', POLL_NETWORK: 'mainnet' }),
    'poll cycle failed chain=bitcoin network=mainnet',
  );
});
