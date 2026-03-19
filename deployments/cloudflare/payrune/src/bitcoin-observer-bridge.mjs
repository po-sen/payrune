const defaultEsploraChainPageSize = 25;
const bridgeContexts = new Map();

function normalizeNetwork(network) {
  return typeof network === 'string' ? network.trim().toLowerCase() : '';
}

function resolveConfig(env, network) {
  const normalizedNetwork = normalizeNetwork(network);
  if (normalizedNetwork === 'mainnet') {
    return buildConfig(env, {
      url: 'BITCOIN_MAINNET_ESPLORA_URL',
      user: 'BITCOIN_MAINNET_ESPLORA_USER',
      password: 'BITCOIN_MAINNET_ESPLORA_PASSWORD',
      timeout: 'BITCOIN_MAINNET_ESPLORA_TIMEOUT',
      timeoutSeconds: 'BITCOIN_MAINNET_ESPLORA_TIMEOUT_SECONDS',
    });
  }
  if (normalizedNetwork === 'testnet4') {
    return buildConfig(env, {
      url: 'BITCOIN_TESTNET4_ESPLORA_URL',
      user: 'BITCOIN_TESTNET4_ESPLORA_USER',
      password: 'BITCOIN_TESTNET4_ESPLORA_PASSWORD',
      timeout: 'BITCOIN_TESTNET4_ESPLORA_TIMEOUT',
      timeoutSeconds: 'BITCOIN_TESTNET4_ESPLORA_TIMEOUT_SECONDS',
    });
  }
  throw new Error(`bitcoin network is not supported: ${network}`);
}

function buildConfig(env, keys) {
  const endpoint = typeof env?.[keys.url] === 'string' ? env[keys.url].trim() : '';
  if (!endpoint) {
    throw new Error(`bitcoin endpoint is not configured: ${keys.url}`);
  }

  let timeoutMs = 10_000;
  const rawTimeout = typeof env?.[keys.timeout] === 'string' ? env[keys.timeout].trim() : '';
  if (rawTimeout) {
    timeoutMs = parseDurationToMs(rawTimeout, keys.timeout);
  }
  const rawTimeoutSeconds = typeof env?.[keys.timeoutSeconds] === 'string' ? env[keys.timeoutSeconds].trim() : '';
  if (rawTimeoutSeconds) {
    const parsed = Number.parseInt(rawTimeoutSeconds, 10);
    if (!Number.isFinite(parsed) || parsed <= 0) {
      throw new Error(`${keys.timeoutSeconds} must be a positive integer`);
    }
    timeoutMs = parsed * 1000;
  }

  return {
    endpoint,
    username: typeof env?.[keys.user] === 'string' ? env[keys.user].trim() : '',
    password: typeof env?.[keys.password] === 'string' ? env[keys.password] : '',
    timeoutMs,
  };
}

function parseDurationToMs(rawValue, key) {
  const value = rawValue.trim();
  const match = value.match(/^(\d+)(ms|s|m|h)$/);
  if (!match) {
    throw new Error(`${key} must be a duration like 500ms, 10s, 5m, or 1h`);
  }

  const amount = Number.parseInt(match[1], 10);
  const unit = match[2];
  if (!Number.isFinite(amount) || amount <= 0) {
    throw new Error(`${key} must be greater than zero`);
  }

  switch (unit) {
    case 'ms':
      return amount;
    case 's':
      return amount * 1000;
    case 'm':
      return amount * 60 * 1000;
    case 'h':
      return amount * 60 * 60 * 1000;
    default:
      throw new Error(`${key} unit is not supported`);
  }
}

function requireBridgeContext(bridgeId) {
  if (!bridgeId) {
    throw new Error('bitcoin observer bridge is not configured');
  }
  const context = bridgeContexts.get(bridgeId);
  if (!context) {
    throw new Error('bitcoin observer bridge context is not found');
  }
  return context;
}

function buildHeaders(config) {
  const headers = new Headers();
  if (config.username) {
    const credentials = btoa(`${config.username}:${config.password}`);
    headers.set('Authorization', `Basic ${credentials}`);
  }
  return headers;
}

async function fetchText(config, path) {
  const controller = new AbortController();
  const timeoutId = setTimeout(() => controller.abort(new Error('request timed out')), config.timeoutMs);
  try {
    const response = await fetch(`${config.endpoint.replace(/\/+$/, '')}${path}`, {
      method: 'GET',
      headers: buildHeaders(config),
      signal: controller.signal,
    });
    const body = await response.text();
    if (!response.ok) {
      throw new Error(`bitcoin endpoint http status ${response.status}`);
    }
    return body;
  } finally {
    clearTimeout(timeoutId);
  }
}

async function fetchJSON(config, path) {
  const text = await fetchText(config, path);
  try {
    return JSON.parse(text);
  } catch (error) {
    throw new Error(`decode endpoint response: ${error.message}`);
  }
}

async function fetchAddressChainTransactions(config, address) {
  const encodedAddress = encodeURIComponent(address.trim());
  const basePath = `/address/${encodedAddress}/txs/chain`;

  const transactions = [];
  let lastSeenTxId = '';

  for (let page = 0; page < 10000; page += 1) {
    const path = lastSeenTxId ? `${basePath}/${encodeURIComponent(lastSeenTxId)}` : basePath;
    const pageTransactions = await fetchJSON(config, path);
    if (!Array.isArray(pageTransactions) || pageTransactions.length === 0) {
      break;
    }

    transactions.push(...pageTransactions);
    if (pageTransactions.length < defaultEsploraChainPageSize) {
      break;
    }

    const nextTxId = typeof pageTransactions[pageTransactions.length - 1]?.txid === 'string'
      ? pageTransactions[pageTransactions.length - 1].txid.trim()
      : '';
    if (!nextTxId || nextTxId === lastSeenTxId) {
      break;
    }
    lastSeenTxId = nextTxId;
  }

  return transactions;
}

async function fetchAddressMempoolTransactions(config, address) {
  const encodedAddress = encodeURIComponent(address.trim());
  return fetchJSON(config, `/address/${encodedAddress}/txs/mempool`);
}

globalThis.__payruneBitcoinFetchLatestBlockHeight = async (bridgeId, network) => {
  const context = requireBridgeContext(bridgeId);
  const config = resolveConfig(context.env, network);
  const text = await fetchText(config, '/blocks/tip/height');
  const height = Number.parseInt(text.trim(), 10);
  if (!Number.isFinite(height) || height < 0) {
    throw new Error('latest block height must be non-negative');
  }
  return height;
};

globalThis.__payruneBitcoinFetchAddressChainTransactions = async (bridgeId, network, address) => {
  const context = requireBridgeContext(bridgeId);
  const config = resolveConfig(context.env, network);
  const transactions = await fetchAddressChainTransactions(config, address);
  return JSON.stringify(transactions);
};

globalThis.__payruneBitcoinFetchAddressMempoolTransactions = async (bridgeId, network, address) => {
  const context = requireBridgeContext(bridgeId);
  const config = resolveConfig(context.env, network);
  const transactions = await fetchAddressMempoolTransactions(config, address);
  return JSON.stringify(transactions);
};

export function registerBitcoinObserverBridge(env) {
  const bridgeId = crypto.randomUUID();
  bridgeContexts.set(bridgeId, { env });
  return bridgeId;
}

export async function unregisterBitcoinObserverBridge(bridgeId) {
  if (!bridgeId) {
    return;
  }
  bridgeContexts.delete(bridgeId);
}
