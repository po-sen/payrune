import { getGoWasmRuntime } from './go-wasm-runtime.mjs';
import { registerPostgresBridge, unregisterPostgresBridge } from './postgres-bridge.mjs';
import { registerWebhookNotifierBridge, unregisterWebhookNotifierBridge } from './webhook-notifier-bridge.mjs';

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

export function buildScheduledEnvelope(controller, env, postgresBridgeId) {
  return {
    env: snapshotEnv(env),
    postgresBridgeId,
    scheduledTime: controller?.scheduledTime ? new Date(controller.scheduledTime).toISOString() : '',
    cron: typeof controller?.cron === 'string' ? controller.cron : '',
  };
}

export function buildCycleLogMessage(output = {}) {
  return `webhook dispatch cycle complete claimed=${output.claimedCount ?? 0} sent=${output.sentCount ?? 0} retried=${output.retriedCount ?? 0} failed=${output.failedCount ?? 0}`;
}

export function buildCycleErrorMessage() {
  return 'webhook dispatch cycle failed';
}

export default {
  async scheduled(controller, env, ctx) {
    const postgresBridgeId = registerPostgresBridge(env);
    registerWebhookNotifierBridge(env);

    try {
      const runtime = await getGoWasmRuntime();
      const envelope = buildScheduledEnvelope(controller, env, postgresBridgeId);
      const encodedResponse = await runtime.handle(JSON.stringify(envelope));
      const payload = JSON.parse(encodedResponse);
      console.log(buildCycleLogMessage(payload.output));
    } catch (error) {
      console.error(buildCycleErrorMessage(), error);
      throw error;
    } finally {
      unregisterWebhookNotifierBridge();
      ctx.waitUntil(unregisterPostgresBridge(postgresBridgeId));
    }
  },

  async fetch() {
    return jsonError(404, 'not found');
  },
};
