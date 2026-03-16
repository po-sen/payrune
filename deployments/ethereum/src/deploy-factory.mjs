import fs from "node:fs/promises";
import path from "node:path";
import { pathToFileURL } from "node:url";
import { ethers } from "ethers";
import { getContractArtifact } from "./contracts.mjs";

function requireEnv(env, key) {
  const value = env[key]?.trim();
  if (!value) {
    throw new Error(`${key} is required`);
  }
  return value;
}

function parseConfirmations(env) {
  const raw = env.ETHEREUM_DEPLOY_CONFIRMATIONS?.trim();
  if (!raw) {
    return 1;
  }
  const parsed = Number.parseInt(raw, 10);
  if (!Number.isInteger(parsed) || parsed < 0) {
    throw new Error("ETHEREUM_DEPLOY_CONFIRMATIONS must be a non-negative integer");
  }
  return parsed;
}

function defaultOutputPath(chainId) {
  return path.resolve(
    process.cwd(),
    "deployments/ethereum/build/deployments",
    `${chainId}-deposit-vault-factory.json`,
  );
}

export function loadDeployConfigFromEnv(env = process.env, cwd = process.cwd()) {
  const rpcURL = requireEnv(env, "ETHEREUM_DEPLOY_RPC_URL");
  const privateKey = requireEnv(env, "ETHEREUM_DEPLOY_PRIVATE_KEY");
  const collector = ethers.getAddress(requireEnv(env, "ETHEREUM_DEPLOY_COLLECTOR_ADDRESS"));
  const confirmations = parseConfirmations(env);
  const outputPath = env.ETHEREUM_DEPLOY_OUTPUT?.trim()
    ? path.resolve(cwd, env.ETHEREUM_DEPLOY_OUTPUT.trim())
    : null;

  return {
    rpcURL,
    privateKey,
    collector,
    confirmations,
    outputPath,
  };
}

export async function deployFactory(config) {
  const {
    rpcURL,
    privateKey,
    collector,
    confirmations = 1,
  } = config;

  const provider = new ethers.JsonRpcProvider(rpcURL);
  const signer = new ethers.Wallet(privateKey, provider);
  const network = await provider.getNetwork();

  const artifact = await getContractArtifact("DepositVaultFactory.sol", "DepositVaultFactory");
  const contractFactory = new ethers.ContractFactory(artifact.abi, artifact.bytecode, signer);
  const contract = await contractFactory.deploy(collector);
  const deployTx = contract.deploymentTransaction();
  if (!deployTx) {
    throw new Error("deployment transaction is not available");
  }
  await deployTx.wait(confirmations);

  const contractAddress = await contract.getAddress();
  const vaultCreationCodeHash = await contract.vaultCreationCodeHash();
  const outputPath = config.outputPath ?? defaultOutputPath(network.chainId.toString());

  const manifest = {
    contractName: artifact.contractName,
    sourceName: artifact.sourceName,
    compilerVersion: artifact.compilerVersion,
    chainId: network.chainId.toString(),
    contractAddress,
    collector,
    vaultCreationCodeHash,
    deployer: await signer.getAddress(),
    deploymentTransactionHash: deployTx.hash,
    confirmations,
    deployedAt: new Date().toISOString(),
    abi: artifact.abi,
  };

  await fs.mkdir(path.dirname(outputPath), { recursive: true });
  await fs.writeFile(outputPath, JSON.stringify(manifest, null, 2));

  return manifest;
}

async function main() {
  const manifest = await deployFactory(loadDeployConfigFromEnv());
  console.log(JSON.stringify(manifest, null, 2));
}

const entryFileURL = process.argv[1] ? pathToFileURL(process.argv[1]).href : null;
if (entryFileURL === import.meta.url) {
  await main();
}
