import { getGoWasmRuntime } from './go-wasm-runtime.mjs';
import { registerPostgresBridge, unregisterPostgresBridge } from './postgres-bridge.mjs';
import { registerBitcoinObserverBridge, unregisterBitcoinObserverBridge } from './bitcoin-observer-bridge.mjs';

function jsonError(status, message) {
  return new Response(JSON.stringify({ error: message }), {
    status,
    headers: { 'Content-Type': 'application/json' },
  });
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

export function buildScheduledEnvelope(controller, env, postgresBridgeId, bitcoinBridgeId) {
  return {
    env: snapshotEnv(env),
    postgresBridgeId,
    bitcoinBridgeId,
    scheduledTime: controller?.scheduledTime ? new Date(controller.scheduledTime).toISOString() : '',
    cron: typeof controller?.cron === 'string' ? controller.cron : '',
  };
}

export function buildCycleLogMessage(env, output = {}) {
  const chain = env.POLL_CHAIN ?? '';
  const network = env.POLL_NETWORK ?? '';
  return `poll cycle complete chain=${chain} network=${network} claimed=${output.claimedCount ?? 0} updated=${output.updatedCount ?? 0} terminal_failed=${output.terminalFailedCount ?? 0} processing_errors=${output.processingErrorCount ?? 0}`;
}

export function buildCycleErrorMessage(env) {
  const chain = env.POLL_CHAIN ?? '';
  const network = env.POLL_NETWORK ?? '';
  return `poll cycle failed chain=${chain} network=${network}`;
}

export default {
  async scheduled(controller, env, ctx) {
    const postgresBridgeId = registerPostgresBridge(env);
    const bitcoinBridgeId = registerBitcoinObserverBridge(env);

    try {
      const runtime = await getGoWasmRuntime();
      const envelope = buildScheduledEnvelope(controller, env, postgresBridgeId, bitcoinBridgeId);
      const encodedResponse = await runtime.handle(JSON.stringify(envelope));
      const payload = JSON.parse(encodedResponse);
      console.log(buildCycleLogMessage(env, payload.output));
    } catch (error) {
      console.error(buildCycleErrorMessage(env), error);
      throw error;
    } finally {
      ctx.waitUntil(Promise.allSettled([
        unregisterPostgresBridge(postgresBridgeId),
        unregisterBitcoinObserverBridge(bitcoinBridgeId),
      ]));
    }
  },

  async fetch() {
    return jsonError(404, 'not found');
  },
};
