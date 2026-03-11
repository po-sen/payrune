import './generated/wasm_exec.js';

let runtimePromise;

async function initializeRuntime() {
  if (typeof globalThis.Go !== 'function') {
    throw new Error('Go runtime is not available');
  }

  const go = new globalThis.Go();
  const { default: wasmModule } = await import('./generated/payrune_api_worker.wasm');
  const instance = await WebAssembly.instantiate(wasmModule, go.importObject);

  const runPromise = go.run(instance);
  void runPromise.catch((error) => {
    console.error('payrune wasm runtime exited unexpectedly', error);
  });

  for (let attempts = 0; attempts < 50; attempts += 1) {
    if (typeof globalThis.payruneHandle === 'function') {
      return {
        async handle(encodedRequest) {
          return globalThis.payruneHandle(encodedRequest);
        },
      };
    }
    await Promise.resolve();
  }

  throw new Error('payrune wasm handler is not ready');
}

export function getGoWasmRuntime() {
  if (!runtimePromise) {
    runtimePromise = initializeRuntime();
  }
  return runtimePromise;
}
