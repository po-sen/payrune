import './generated/wasm_exec.js';

let runtimePromise;

async function initializeRuntime() {
  if (typeof globalThis.Go !== 'function') {
    throw new Error('Go runtime is not available');
  }

  const go = new globalThis.Go();
  const { default: wasmModule } = await import('./generated/payrune_poller_worker.wasm');
  const instance = await WebAssembly.instantiate(wasmModule, go.importObject);

  const runPromise = go.run(instance);
  void runPromise.catch((error) => {
    console.error('payrune poller wasm runtime exited unexpectedly', error);
  });

  for (let attempts = 0; attempts < 50; attempts += 1) {
    if (typeof globalThis.payrunePollerHandle === 'function') {
      return {
        async handle(encodedRequest) {
          return globalThis.payrunePollerHandle(encodedRequest);
        },
      };
    }
    await Promise.resolve();
  }

  throw new Error('payrune poller wasm handler is not ready');
}

export function getGoWasmRuntime() {
  if (!runtimePromise) {
    runtimePromise = initializeRuntime();
  }
  return runtimePromise;
}
