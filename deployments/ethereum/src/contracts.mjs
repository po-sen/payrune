import fs from "node:fs/promises";
import path from "node:path";
import { fileURLToPath } from "node:url";
import solc from "solc";

const projectDir = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "..");
const contractsDir = path.join(projectDir, "contracts");
const buildDir = path.join(projectDir, "build");
const contractFiles = ["DepositVaultFactory.sol", "TestUSDTMock.sol"];

function formatCompilerMessage(error) {
  const location = error.sourceLocation
    ? `${error.sourceLocation.file}:${error.sourceLocation.start}:${error.sourceLocation.end}`
    : "solc";
  return `${location} ${error.severity}: ${error.formattedMessage ?? error.message}`;
}

async function readSources() {
  const entries = await Promise.all(
    contractFiles.map(async (fileName) => {
      const fullPath = path.join(contractsDir, fileName);
      const content = await fs.readFile(fullPath, "utf8");
      return [fileName, { content }];
    }),
  );
  return Object.fromEntries(entries);
}

export async function compileContracts() {
  const input = {
    language: "Solidity",
    sources: await readSources(),
    settings: {
      evmVersion: "paris",
      optimizer: {
        enabled: true,
        runs: 200,
      },
      outputSelection: {
        "*": {
          "*": ["abi", "evm.bytecode.object", "evm.deployedBytecode.object"],
        },
      },
    },
  };

  const output = JSON.parse(solc.compile(JSON.stringify(input)));
  const messages = output.errors ?? [];
  const fatalMessages = messages.filter((entry) => entry.severity === "error");
  if (fatalMessages.length > 0) {
    throw new Error(fatalMessages.map(formatCompilerMessage).join("\n"));
  }

  return {
    compilerVersion: solc.version(),
    contracts: output.contracts,
    warnings: messages.filter((entry) => entry.severity !== "error").map(formatCompilerMessage),
  };
}

export async function getContractArtifact(sourceName, contractName) {
  const { compilerVersion, contracts, warnings } = await compileContracts();
  const sourceContracts = contracts[sourceName];
  if (!sourceContracts || !sourceContracts[contractName]) {
    throw new Error(`contract artifact not found: ${sourceName}:${contractName}`);
  }

  const artifact = sourceContracts[contractName];
  return {
    compilerVersion,
    warnings,
    contractName,
    sourceName,
    abi: artifact.abi,
    bytecode: `0x${artifact.evm.bytecode.object}`,
    deployedBytecode: `0x${artifact.evm.deployedBytecode.object}`,
  };
}

export async function writeArtifacts(outputPath = path.join(buildDir, "contracts.json")) {
  const compiled = await compileContracts();
  await fs.mkdir(path.dirname(outputPath), { recursive: true });
  await fs.writeFile(outputPath, JSON.stringify(compiled, null, 2));
  return outputPath;
}
