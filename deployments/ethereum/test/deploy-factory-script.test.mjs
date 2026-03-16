import test from "node:test";
import assert from "node:assert/strict";
import fs from "node:fs/promises";
import os from "node:os";
import path from "node:path";
import ganache from "ganache";
import { ethers } from "ethers";
import { deployFactory } from "../src/deploy-factory.mjs";

const testMnemonic = "test test test test test test test test test test test junk";

test("deployFactory deploys the factory contract and writes a manifest", async (t) => {
  const server = ganache.server({
    logging: { quiet: true },
    wallet: {
      mnemonic: testMnemonic,
      totalAccounts: 3,
      defaultBalance: 1_000,
    },
  });

  await new Promise((resolve, reject) => {
    server.listen(0, "127.0.0.1", (error) => {
      if (error) {
        reject(error);
        return;
      }
      resolve();
    });
  });

  t.after(async () => {
    await server.close();
  });

  const address = server.address();
  if (!address || typeof address === "string") {
    throw new Error("ganache server did not expose a numeric port");
  }

  const rpcURL = `http://127.0.0.1:${address.port}`;
  const provider = new ethers.JsonRpcProvider(rpcURL);
  const deployer = ethers.Wallet.fromPhrase(testMnemonic).connect(provider);
  const collector = await provider.getSigner(1);
  const collectorAddress = await collector.getAddress();
  const tempDir = await fs.mkdtemp(path.join(os.tmpdir(), "payrune-eth-deploy-"));
  const outputPath = path.join(tempDir, "factory-manifest.json");

  const manifest = await deployFactory({
    rpcURL,
    privateKey: deployer.privateKey,
    collector: collectorAddress,
    confirmations: 1,
    outputPath,
  });

  const persistedManifest = JSON.parse(await fs.readFile(outputPath, "utf8"));

  assert.equal(manifest.contractName, "DepositVaultFactory");
  assert.equal(manifest.collector, collectorAddress);
  assert.equal(manifest.deployer, await deployer.getAddress());
  assert.equal(manifest.contractAddress, persistedManifest.contractAddress);
  assert.equal(manifest.chainId, "1337");
  assert.notEqual(await provider.getCode(manifest.contractAddress), "0x");
});
