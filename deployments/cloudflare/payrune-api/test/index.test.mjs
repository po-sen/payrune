import test from 'node:test';
import assert from 'node:assert/strict';

import { buildRequestEnvelope, buildWorkerResponse, isPublicRoute } from '../src/index.mjs';

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
