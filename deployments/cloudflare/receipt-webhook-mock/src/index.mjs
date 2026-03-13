const signatureHeader = 'X-Payrune-Signature-256';

function jsonResponse(status, payload) {
  return new Response(JSON.stringify(payload), {
    status,
    headers: { 'Content-Type': 'application/json' },
  });
}

export function snapshotHeaders(headers) {
  const result = {};
  headers.forEach((value, key) => {
    result[key] = value;
  });
  return result;
}

export async function computeWebhookSignature(secret, body) {
  const key = await crypto.subtle.importKey(
    'raw',
    new TextEncoder().encode(secret),
    { name: 'HMAC', hash: 'SHA-256' },
    false,
    ['sign'],
  );
  const signature = await crypto.subtle.sign('HMAC', key, new TextEncoder().encode(body));
  return Array.from(new Uint8Array(signature))
    .map((byte) => byte.toString(16).padStart(2, '0'))
    .join('');
}

export async function verifyRequest(secret, request, body) {
  const providedSignature = request.headers.get(signatureHeader) ?? '';
  const event = request.headers.get('X-Payrune-Event') ?? '';
  const version = request.headers.get('X-Payrune-Event-Version') ?? '';
  const notificationID = request.headers.get('X-Payrune-Notification-ID') ?? '';

  let payload = null;
  try {
    payload = JSON.parse(body);
  } catch (error) {
    return {
      event,
      version,
      notificationID,
      providedSignature,
      validJSON: false,
      verificationEnabled: secret.trim() !== '',
      signatureValid: false,
      computedSignature: '',
      payload,
      error,
    };
  }

  if (secret.trim() === '') {
    return {
      event,
      version,
      notificationID,
      providedSignature,
      validJSON: true,
      verificationEnabled: false,
      signatureValid: true,
      computedSignature: '',
      payload,
      error: null,
    };
  }

  const computedSignature = `sha256=${await computeWebhookSignature(secret, body)}`;
  const signatureValid = computedSignature === providedSignature;
  return {
    event,
    version,
    notificationID,
    providedSignature,
    validJSON: true,
    verificationEnabled: true,
    signatureValid,
    computedSignature,
    payload,
    error: null,
  };
}

export async function handleRequest(request, env) {
  const url = new URL(request.url);
  if (request.method === 'GET' && url.pathname === '/health') {
    return new Response('ok', { status: 200 });
  }
  if (request.method !== 'POST' || url.pathname !== '/receipt-status') {
    return jsonResponse(404, { error: 'not found' });
  }

  const body = await request.text();
  const verification = await verifyRequest(env.PAYMENT_RECEIPT_WEBHOOK_SECRET ?? '', request, body);
  console.log('receipt webhook mock request', JSON.stringify({
    method: request.method,
    path: url.pathname,
    headers: snapshotHeaders(request.headers),
    verification: {
      enabled: verification.verificationEnabled,
      valid: verification.signatureValid,
      event: verification.event,
      version: verification.version,
      notificationID: verification.notificationID,
    },
    rawBody: body,
  }));

  if (!verification.validJSON) {
    return jsonResponse(400, { error: 'invalid json payload' });
  }

  if (verification.verificationEnabled && !verification.signatureValid) {
    console.warn('receipt webhook mock signature invalid', JSON.stringify({
      providedSignature: verification.providedSignature,
      computedSignature: verification.computedSignature,
      notificationID: verification.notificationID,
    }));
    return jsonResponse(401, { error: 'invalid signature' });
  }

  return new Response(null, { status: 204 });
}

export default {
  async fetch(request, env) {
    return handleRequest(request, env);
  },
};
