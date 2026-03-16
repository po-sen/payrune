import test from "node:test";
import assert from "node:assert/strict";
import ganache from "ganache";
import { ethers } from "ethers";
import { getContractArtifact } from "../src/contracts.mjs";

async function deployContract(signer, sourceName, contractName, args = []) {
  const artifact = await getContractArtifact(sourceName, contractName);
  const factory = new ethers.ContractFactory(artifact.abi, artifact.bytecode, signer);
  const contract = await factory.deploy(...args);
  await contract.waitForDeployment();
  return contract;
}

async function createTestContext() {
  const ganacheProvider = ganache.provider({
    logging: { quiet: true },
    wallet: { totalAccounts: 3, defaultBalance: 1_000 },
  });
  const provider = new ethers.BrowserProvider(ganacheProvider);
  const deployer = await provider.getSigner(0);
  const collector = await provider.getSigner(1);
  const collectorAddress = await collector.getAddress();
  const factory = await deployContract(
    deployer,
    "DepositVaultFactory.sol",
    "DepositVaultFactory",
    [collectorAddress],
  );

  return {
    provider,
    transport: ganacheProvider,
    deployer,
    collector,
    collectorAddress,
    factory,
  };
}

async function getNativeBalance(transport, address) {
  return BigInt(await transport.request({
    method: "eth_getBalance",
    params: [address, "latest"],
  }));
}

function getEventArgs(receipt, contract, eventName) {
  for (const log of receipt.logs) {
    try {
      const parsed = contract.interface.parseLog(log);
      if (parsed?.name === eventName) {
        return parsed.args;
      }
    } catch {
      // Ignore logs emitted by other contracts in the same transaction.
    }
  }
  throw new Error(`event not found: ${eventName}`);
}

test("DepositVaultFactory deploys a deterministic vault and sweeps native ETH", async () => {
  const { transport, deployer, collectorAddress, factory } = await createTestContext();
  const salt = ethers.keccak256(ethers.toUtf8Bytes("native-sweep"));
  const predictedAddress = await factory.predictVaultAddress(salt);
  const deployReceipt = await (await factory.deployVault(salt)).wait();
  const deployedEvent = getEventArgs(deployReceipt, factory, "VaultDeployed");

  assert.equal(deployedEvent.salt, salt);
  assert.equal(deployedEvent.vault, predictedAddress);

  const amount = ethers.parseEther("1.25");
  await (await deployer.sendTransaction({ to: predictedAddress, value: amount })).wait();

  const collectorBalanceBefore = await getNativeBalance(transport, collectorAddress);
  const receipt = await (await factory.batchDeployAndSweepNative([salt])).wait();
  const sweptEvent = getEventArgs(receipt, factory, "NativeSwept");
  const collectorBalanceAfter = await getNativeBalance(transport, collectorAddress);

  assert.equal(sweptEvent.salt, salt);
  assert.equal(sweptEvent.vault, predictedAddress);
  assert.equal(sweptEvent.amount, amount);
  assert.equal(collectorBalanceAfter - collectorBalanceBefore, amount);
  assert.equal(await getNativeBalance(transport, predictedAddress), 0n);
});

test("DepositVaultFactory deploys and sweeps a USDT-like token without boolean transfer return", async () => {
  const { deployer, collectorAddress, factory } = await createTestContext();
  const token = await deployContract(deployer, "TestUSDTMock.sol", "TestUSDTMock");
  const salt = ethers.keccak256(ethers.toUtf8Bytes("token-sweep"));
  const predictedAddress = await factory.predictVaultAddress(salt);
  const tokenAddress = await token.getAddress();
  const amount = ethers.parseUnits("42.5", 6);

  await (await token.mint(await deployer.getAddress(), amount)).wait();
  await (await token.transfer(predictedAddress, amount)).wait();
  await (await factory.batchDeployAndSweepToken([salt], tokenAddress)).wait();

  assert.equal(await token.balanceOf(collectorAddress), amount);
  assert.equal(await token.balanceOf(predictedAddress), 0n);
});
