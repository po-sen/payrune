import { Pool } from 'pg';

const poolByConnectionString = new Map();
const bridgeContexts = new Map();

function resolveConnectionString(env) {
  if (env?.HYPERDRIVE?.connectionString) {
    return env.HYPERDRIVE.connectionString;
  }
  if (typeof env?.POSTGRES_CONNECTION_STRING === 'string' && env.POSTGRES_CONNECTION_STRING.trim() !== '') {
    return env.POSTGRES_CONNECTION_STRING.trim();
  }
  return '';
}

function getPool(connectionString) {
  let pool = poolByConnectionString.get(connectionString);
  if (!pool) {
    pool = new Pool({ connectionString });
    poolByConnectionString.set(connectionString, pool);
  }
  return pool;
}

function normalizeScalar(value) {
  if (value === null || value === undefined) {
    return null;
  }
  if (value instanceof Date) {
    return value.toISOString();
  }
  if (typeof value === 'bigint') {
    return value.toString();
  }
  return value;
}

function normalizeRow(row) {
  return row.map((value) => normalizeScalar(value));
}

function normalizePgError(error) {
  return {
    message: error?.message ?? 'postgres query failed',
    code: error?.code ?? '',
    constraint: error?.constraint ?? '',
  };
}

function requireBridgeContext(bridgeId) {
  if (!bridgeId) {
    throw new Error('postgres bridge is not configured');
  }
  const context = bridgeContexts.get(bridgeId);
  if (!context) {
    throw new Error('postgres bridge context is not found');
  }
  return context;
}

async function rollbackAndRelease(client) {
  try {
    await client.query('ROLLBACK');
  } catch (error) {
    console.error('cloudflare postgres rollback failed', error);
  } finally {
    client.release();
  }
}

async function queryWithRowMode(executor, text, values) {
  return executor.query({
    text,
    values,
    rowMode: 'array',
  });
}

globalThis.__payrunePgBeginTx = async (bridgeId) => {
  const context = requireBridgeContext(bridgeId);
  const client = await context.pool.connect();
  try {
    await client.query('BEGIN');
  } catch (error) {
    client.release();
    throw normalizePgError(error);
  }
  const txId = crypto.randomUUID();
  context.transactions.set(txId, client);
  return txId;
};

globalThis.__payrunePgCommitTx = async (bridgeId, txId) => {
  const context = requireBridgeContext(bridgeId);
  const client = context.transactions.get(txId);
  if (!client) {
    throw new Error('postgres transaction is not found');
  }
  try {
    await client.query('COMMIT');
  } catch (error) {
    throw normalizePgError(error);
  } finally {
    context.transactions.delete(txId);
    client.release();
  }
  return null;
};

globalThis.__payrunePgRollbackTx = async (bridgeId, txId) => {
  const context = requireBridgeContext(bridgeId);
  const client = context.transactions.get(txId);
  if (!client) {
    return null;
  }
  context.transactions.delete(txId);
  await rollbackAndRelease(client);
  return null;
};

globalThis.__payrunePgExec = async (bridgeId, txId, text, values) => {
  const context = requireBridgeContext(bridgeId);
  const executor = txId ? context.transactions.get(txId) : context.pool;
  if (!executor) {
    throw new Error('postgres executor is not found');
  }
  try {
    const result = await executor.query(text, values);
    return { rowCount: result.rowCount ?? 0 };
  } catch (error) {
    throw normalizePgError(error);
  }
};

globalThis.__payrunePgQuery = async (bridgeId, txId, text, values) => {
  const context = requireBridgeContext(bridgeId);
  const executor = txId ? context.transactions.get(txId) : context.pool;
  if (!executor) {
    throw new Error('postgres executor is not found');
  }
  try {
    const result = await queryWithRowMode(executor, text, values);
    return { rows: result.rows.map((row) => normalizeRow(row)) };
  } catch (error) {
    throw normalizePgError(error);
  }
};

globalThis.__payrunePgQueryRow = async (bridgeId, txId, text, values) => {
  const context = requireBridgeContext(bridgeId);
  const executor = txId ? context.transactions.get(txId) : context.pool;
  if (!executor) {
    throw new Error('postgres executor is not found');
  }
  try {
    const result = await queryWithRowMode(executor, text, values);
    if (result.rows.length === 0) {
      return { found: false, row: null };
    }
    return { found: true, row: normalizeRow(result.rows[0]) };
  } catch (error) {
    throw normalizePgError(error);
  }
};

export function registerPostgresBridge(env) {
  const connectionString = resolveConnectionString(env);
  if (!connectionString) {
    return '';
  }
  const bridgeId = crypto.randomUUID();
  bridgeContexts.set(bridgeId, {
    pool: getPool(connectionString),
    transactions: new Map(),
  });
  return bridgeId;
}

export async function unregisterPostgresBridge(bridgeId) {
  if (!bridgeId) {
    return;
  }
  const context = bridgeContexts.get(bridgeId);
  if (!context) {
    return;
  }

  bridgeContexts.delete(bridgeId);
  const rollbacks = [];
  for (const client of context.transactions.values()) {
    rollbacks.push(rollbackAndRelease(client));
  }
  await Promise.allSettled(rollbacks);
}
