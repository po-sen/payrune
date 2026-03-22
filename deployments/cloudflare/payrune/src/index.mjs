import { getGoWasmRuntime } from './go-wasm-runtime.mjs';
import { registerPostgresBridge, unregisterPostgresBridge } from './postgres-bridge.mjs';
import { registerBitcoinObserverBridge, unregisterBitcoinObserverBridge } from './bitcoin-observer-bridge.mjs';
import { registerWebhookNotifierBridge, unregisterWebhookNotifierBridge } from './webhook-notifier-bridge.mjs';

const publicPrefix = '/v1/';

export const scheduledJobs = Object.freeze({
  '2,17,32,47 * * * *': Object.freeze({
    type: 'poller',
    operation: 'poller',
    envOverrides: Object.freeze({
      POLL_CHAIN: 'ethereum',
      POLL_NETWORK: 'mainnet',
    }),
  }),
  '5,20,35,50 * * * *': Object.freeze({
    type: 'poller',
    operation: 'poller',
    envOverrides: Object.freeze({
      POLL_CHAIN: 'bitcoin',
      POLL_NETWORK: 'mainnet',
    }),
  }),
  '8,23,38,53 * * * *': Object.freeze({
    type: 'poller',
    operation: 'poller',
    envOverrides: Object.freeze({
      POLL_CHAIN: 'ethereum',
      POLL_NETWORK: 'sepolia',
    }),
  }),
  '*/15 * * * *': Object.freeze({
    type: 'poller',
    operation: 'poller',
    envOverrides: Object.freeze({
      POLL_CHAIN: 'bitcoin',
      POLL_NETWORK: 'testnet4',
    }),
  }),
  '10,25,40,55 * * * *': Object.freeze({
    type: 'webhook_dispatcher',
    operation: 'webhook_dispatcher',
    envOverrides: Object.freeze({}),
  }),
});

export function isPublicRoute(pathname) {
  return pathname === '/health' || pathname.startsWith(publicPrefix);
}

export function snapshotEnv(env) {
  const envSnapshot = {};
  for (const [key, value] of Object.entries(env)) {
    if (typeof value === 'string') {
      envSnapshot[key] = value;
    }
  }
  return envSnapshot;
}

export function buildScheduledEnv(env, envOverrides = {}) {
  return {
    ...snapshotEnv(env),
    ...envOverrides,
  };
}

export function buildRequestEnvelope(request, env, bridgeId, body) {
  const url = new URL(request.url);
  const headers = {};
  request.headers.forEach((value, key) => {
    headers[key] = value;
  });

  return {
    request: {
      method: request.method,
      path: url.pathname,
      rawQuery: url.search.length > 0 ? url.search.slice(1) : '',
      headers,
      body,
    },
    env: snapshotEnv(env),
    bridgeId,
  };
}

export function buildScheduledEnvelope(controller, env, extra = {}, envOverrides = {}) {
  return {
    env: buildScheduledEnv(env, envOverrides),
    scheduledTime: controller?.scheduledTime ? new Date(controller.scheduledTime).toISOString() : '',
    cron: typeof controller?.cron === 'string' ? controller.cron : '',
    ...extra,
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

export function resolveScheduledJob(controller) {
  const cron = typeof controller?.cron === 'string' ? controller.cron.trim() : '';
  if (!cron) {
    return null;
  }
  return scheduledJobs[cron] ?? null;
}

export function buildPollerCycleLogMessage(env, output = {}) {
  const chain = env.POLL_CHAIN ?? '';
  const network = env.POLL_NETWORK ?? '';
  return `poll cycle complete chain=${chain} network=${network} claimed=${output.claimedCount ?? 0} updated=${output.updatedCount ?? 0} terminal_failed=${output.terminalFailedCount ?? 0} processing_errors=${output.processingErrorCount ?? 0}`;
}

export function buildPollerCycleErrorMessage(env) {
  const chain = env.POLL_CHAIN ?? '';
  const network = env.POLL_NETWORK ?? '';
  return `poll cycle failed chain=${chain} network=${network}`;
}

export function buildWebhookDispatchCycleLogMessage(output = {}) {
  return `webhook dispatch cycle complete claimed=${output.claimedCount ?? 0} sent=${output.sentCount ?? 0} retried=${output.retriedCount ?? 0} failed=${output.failedCount ?? 0}`;
}

export function buildWebhookDispatchCycleErrorMessage() {
  return 'webhook dispatch cycle failed';
}

async function handleAPIRequest(request, env, ctx) {
  const url = new URL(request.url);
  if (!isPublicRoute(url.pathname)) {
    return jsonError(404, 'not found');
  }

  const body = request.method === 'GET' || request.method === 'HEAD' ? '' : await request.text();
  const bridgeId = registerPostgresBridge(env);

  try {
    const runtime = await getGoWasmRuntime();
    const envelope = buildRequestEnvelope(request, env, bridgeId, body);
    const encodedResponse = await runtime.handle('api', JSON.stringify(envelope));
    const payload = JSON.parse(encodedResponse);
    return buildWorkerResponse(payload.response ?? {});
  } catch (error) {
    console.error('payrune worker request failed', error);
    return jsonError(500, 'internal server error');
  } finally {
    ctx.waitUntil(unregisterPostgresBridge(bridgeId));
  }
}

async function runPollerJob(controller, env, ctx, job) {
  const scopedEnv = buildScheduledEnv(env, job.envOverrides);
  const postgresBridgeId = registerPostgresBridge(env);
  const bitcoinBridgeId = scopedEnv.POLL_CHAIN === 'bitcoin'
    ? registerBitcoinObserverBridge(scopedEnv)
    : '';

  try {
    const runtime = await getGoWasmRuntime();
    const envelope = buildScheduledEnvelope(
      controller,
      env,
      { postgresBridgeId, bitcoinBridgeId },
      job.envOverrides,
    );
    const encodedResponse = await runtime.handle(job.operation, JSON.stringify(envelope));
    const payload = JSON.parse(encodedResponse);
    console.log(buildPollerCycleLogMessage(scopedEnv, payload.output));
  } catch (error) {
    console.error(buildPollerCycleErrorMessage(scopedEnv), error);
    throw error;
  } finally {
    ctx.waitUntil(Promise.allSettled([
      unregisterPostgresBridge(postgresBridgeId),
      unregisterBitcoinObserverBridge(bitcoinBridgeId),
    ]));
  }
}

async function runWebhookDispatcherJob(controller, env, ctx, job) {
  const postgresBridgeId = registerPostgresBridge(env);
  registerWebhookNotifierBridge(env);

  try {
    const runtime = await getGoWasmRuntime();
    const envelope = buildScheduledEnvelope(
      controller,
      env,
      { postgresBridgeId },
      job.envOverrides,
    );
    const encodedResponse = await runtime.handle(job.operation, JSON.stringify(envelope));
    const payload = JSON.parse(encodedResponse);
    console.log(buildWebhookDispatchCycleLogMessage(payload.output));
  } catch (error) {
    console.error(buildWebhookDispatchCycleErrorMessage(), error);
    throw error;
  } finally {
    unregisterWebhookNotifierBridge();
    ctx.waitUntil(unregisterPostgresBridge(postgresBridgeId));
  }
}

export default {
  async fetch(request, env, ctx) {
    return handleAPIRequest(request, env, ctx);
  },

  async scheduled(controller, env, ctx) {
    const job = resolveScheduledJob(controller);
    if (!job) {
      const cron = typeof controller?.cron === 'string' ? controller.cron : '';
      const error = new Error(`payrune scheduled cron is not configured: ${cron || '(missing)'}`);
      console.error(error.message);
      throw error;
    }

    if (job.type === 'poller') {
      return runPollerJob(controller, env, ctx, job);
    }
    if (job.type === 'webhook_dispatcher') {
      return runWebhookDispatcherJob(controller, env, ctx, job);
    }

    throw new Error(`payrune scheduled job type is not supported: ${job.type}`);
  },
};
