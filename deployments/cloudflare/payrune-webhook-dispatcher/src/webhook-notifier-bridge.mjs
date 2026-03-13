function toErrorMessage(error) {
  if (error instanceof Error && typeof error.message === 'string' && error.message !== '') {
    return error.message;
  }
  return String(error ?? 'webhook request failed');
}

const defaultBindingPath = '/receipt-status';

function buildRequestInit(timeoutMs, headers, body) {
  const controller = new AbortController();
  let timeoutHandle = null;

  if (typeof timeoutMs === 'number' && timeoutMs > 0) {
    timeoutHandle = setTimeout(() => controller.abort(new Error('webhook request timed out')), timeoutMs);
  }

  return {
    requestInit: {
      method: 'POST',
      headers,
      body,
      signal: controller.signal,
    },
    cleanup() {
      if (timeoutHandle !== null) {
        clearTimeout(timeoutHandle);
      }
    },
  };
}

function resolveRequestTarget(env, binding) {
  if (binding && env?.[binding]?.fetch) {
    return env[binding].fetch.bind(env[binding]);
  }
  throw new Error(`cloudflare webhook binding ${binding || '(missing)'} is not configured`);
}

export function registerWebhookNotifierBridge(env) {
  globalThis.__payruneWebhookPost = async (binding, bindingPath, timeoutMs, headers, body) => {
    const { requestInit, cleanup } = buildRequestInit(timeoutMs, headers, body);
    const targetURL = `https://internal${bindingPath || defaultBindingPath}`;
    const request = new Request(targetURL, requestInit);
    const send = resolveRequestTarget(env, binding);

    try {
      const response = await send(request);

      if (response.status < 200 || response.status >= 300) {
        throw new Error(`webhook returned status ${response.status}`);
      }
      return null;
    } catch (error) {
      if (error?.name === 'AbortError') {
        throw new Error('webhook request timed out');
      }
      throw new Error(toErrorMessage(error));
    } finally {
      cleanup();
    }
  };
}

export function unregisterWebhookNotifierBridge() {
  delete globalThis.__payruneWebhookPost;
}

export const __test = {
  defaultBindingPath,
  resolveRequestTarget,
};

registerWebhookNotifierBridge(undefined);
