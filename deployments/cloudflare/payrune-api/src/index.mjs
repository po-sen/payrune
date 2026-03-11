import { getGoWasmRuntime } from './go-wasm-runtime.mjs';
import { registerPostgresBridge, unregisterPostgresBridge } from './postgres-bridge.mjs';

const publicPrefix = '/v1/';
export function isPublicRoute(pathname) {
  return pathname === '/health' || pathname.startsWith(publicPrefix);
}

export function buildRequestEnvelope(request, env, bridgeId, body) {
  const url = new URL(request.url);
  const headers = {};
  request.headers.forEach((value, key) => {
    headers[key] = value;
  });

  const envSnapshot = {};
  for (const [key, value] of Object.entries(env)) {
    if (typeof value === 'string') {
      envSnapshot[key] = value;
    }
  }

  return {
    request: {
      method: request.method,
      path: url.pathname,
      rawQuery: url.search.length > 0 ? url.search.slice(1) : '',
      headers,
      body,
    },
    env: envSnapshot,
    bridgeId,
  };
}

export function buildWorkerResponse(payload) {
  const headers = new Headers();
  for (const [key, value] of Object.entries(payload.headers ?? {})) {
    headers.set(key, value);
  }
  return new Response(payload.body ?? '', {
    status: payload.status ?? 500,
    headers,
  });
}

function jsonError(status, message) {
  return new Response(JSON.stringify({ error: message }), {
    status,
    headers: { 'Content-Type': 'application/json' },
  });
}

export default {
  async fetch(request, env, ctx) {
    const url = new URL(request.url);
    if (!isPublicRoute(url.pathname)) {
      return jsonError(404, 'not found');
    }

    const body = request.method === 'GET' || request.method === 'HEAD' ? '' : await request.text();
    const bridgeId = registerPostgresBridge(env);

    try {
      const runtime = await getGoWasmRuntime();
      const envelope = buildRequestEnvelope(request, env, bridgeId, body);
      const encodedResponse = await runtime.handle(JSON.stringify(envelope));
      const payload = JSON.parse(encodedResponse);
      return buildWorkerResponse(payload.response ?? {});
    } catch (error) {
      console.error('payrune worker request failed', error);
      return jsonError(500, 'internal server error');
    } finally {
      ctx.waitUntil(unregisterPostgresBridge(bridgeId));
    }
  },
};
