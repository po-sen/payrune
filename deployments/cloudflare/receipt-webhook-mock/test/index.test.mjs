import test from 'node:test';
import assert from 'node:assert/strict';

import {
  computeWebhookSignature,
  handleRequest,
  snapshotHeaders,
  verifyRequest,
} from '../src/index.mjs';

test('snapshotHeaders keeps request headers', () => {
  const headers = new Headers({
    'X-Test': '1',
    'Content-Type': 'application/json',
  });

  assert.deepEqual(snapshotHeaders(headers), {
    'content-type': 'application/json',
    'x-test': '1',
  });
});

test('verifyRequest validates signature when secret is present', async () => {
  const body = JSON.stringify({ notificationId: 9 });
  const signature = `sha256=${await computeWebhookSignature('secret-key', body)}`;
  const request = new Request('https://receipt-webhook-mock.example/receipt-status', {
    method: 'POST',
    headers: {
      'X-Payrune-Event': 'payment_receipt.status_changed',
      'X-Payrune-Event-Version': '1',
      'X-Payrune-Notification-ID': '9',
      'X-Payrune-Signature-256': signature,
    },
    body,
  });

  const verification = await verifyRequest('secret-key', request, body);
  assert.equal(verification.validJSON, true);
  assert.equal(verification.verificationEnabled, true);
  assert.equal(verification.signatureValid, true);
});

test('handleRequest returns 204 for a valid signed webhook', async () => {
  const body = JSON.stringify({ notificationId: 9 });
  const signature = `sha256=${await computeWebhookSignature('secret-key', body)}`;
  const request = new Request('https://receipt-webhook-mock.example/receipt-status', {
    method: 'POST',
    headers: {
      'X-Payrune-Event': 'payment_receipt.status_changed',
      'X-Payrune-Event-Version': '1',
      'X-Payrune-Notification-ID': '9',
      'X-Payrune-Signature-256': signature,
    },
    body,
  });

  const response = await handleRequest(request, { PAYMENT_RECEIPT_WEBHOOK_SECRET: 'secret-key' });
  assert.equal(response.status, 204);
});

test('handleRequest returns 401 for an invalid signature', async () => {
  const request = new Request('https://receipt-webhook-mock.example/receipt-status', {
    method: 'POST',
    headers: {
      'X-Payrune-Event': 'payment_receipt.status_changed',
      'X-Payrune-Event-Version': '1',
      'X-Payrune-Notification-ID': '9',
      'X-Payrune-Signature-256': 'sha256=invalid',
    },
    body: JSON.stringify({ notificationId: 9 }),
  });

  const response = await handleRequest(request, { PAYMENT_RECEIPT_WEBHOOK_SECRET: 'secret-key' });
  assert.equal(response.status, 401);
  assert.deepEqual(await response.json(), { error: 'invalid signature' });
});

test('handleRequest returns 400 for invalid json payload', async () => {
  const request = new Request('https://receipt-webhook-mock.example/receipt-status', {
    method: 'POST',
    headers: {
      'X-Payrune-Signature-256': 'sha256=invalid',
    },
    body: '{',
  });

  const response = await handleRequest(request, { PAYMENT_RECEIPT_WEBHOOK_SECRET: 'secret-key' });
  assert.equal(response.status, 400);
});

test('handleRequest returns 404 for unsupported routes', async () => {
  const request = new Request('https://receipt-webhook-mock.example/not-found', {
    method: 'GET',
  });

  const response = await handleRequest(request, {});
  assert.equal(response.status, 404);
});
