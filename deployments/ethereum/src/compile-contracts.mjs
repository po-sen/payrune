import path from "node:path";
import { writeArtifacts } from "./contracts.mjs";

const outputPath = process.env.ETHEREUM_CONTRACT_ARTIFACTS_OUTPUT
  ? path.resolve(process.cwd(), process.env.ETHEREUM_CONTRACT_ARTIFACTS_OUTPUT)
  : undefined;

const writtenPath = await writeArtifacts(outputPath);
console.log(`wrote contract artifacts to ${writtenPath}`);
