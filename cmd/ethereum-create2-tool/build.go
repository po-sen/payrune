package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
)

type contractArtifactDefinition struct {
	sourceName   string
	contractName string
	fileName     string
}

type solcStandardJSONInput struct {
	Language string                                `json:"language"`
	Sources  map[string]solcStandardJSONInputEntry `json:"sources"`
	Settings solcStandardJSONSettings              `json:"settings"`
}

type solcStandardJSONInputEntry struct {
	Content string `json:"content"`
}

type solcStandardJSONSettings struct {
	Optimizer       solcOptimizerSetting           `json:"optimizer"`
	OutputSelection map[string]map[string][]string `json:"outputSelection"`
}

type solcOptimizerSetting struct {
	Enabled bool `json:"enabled"`
	Runs    int  `json:"runs"`
}

type solcStandardJSONOutput struct {
	Contracts map[string]map[string]solcCompiledContract `json:"contracts"`
	Errors    []solcCompilerMessage                      `json:"errors"`
}

type solcCompiledContract struct {
	ABI json.RawMessage `json:"abi"`
	EVM struct {
		Bytecode struct {
			Object string `json:"object"`
		} `json:"bytecode"`
		DeployedBytecode struct {
			Object string `json:"object"`
		} `json:"deployedBytecode"`
	} `json:"evm"`
}

type solcCompilerMessage struct {
	Severity         string `json:"severity"`
	FormattedMessage string `json:"formattedMessage"`
}

var contractArtifactDefinitions = []contractArtifactDefinition{
	{
		sourceName:   "Create2ReceiverFactory.sol",
		contractName: "Create2ReceiverFactory",
		fileName:     "Create2ReceiverFactoryV1.json",
	},
	{
		sourceName:   "FixedCollectorReceiver.sol",
		contractName: "FixedCollectorReceiver",
		fileName:     "FixedCollectorReceiverV1.json",
	},
}

func runBuild(args []string) error {
	paths, err := resolveToolPaths()
	if err != nil {
		return err
	}

	flagSet := newFlagSet("build")
	solcImageRef := flagSet.String("solc-image", defaultSolcImageRef, "docker image ref for solc")
	solcPlatform := flagSet.String("solc-platform", defaultSolcPlatform, "docker platform for solc")
	if err := flagSet.Parse(args); err != nil {
		return err
	}

	input, err := loadSolcInput(paths.contractsDir)
	if err != nil {
		return err
	}

	inputJSON, err := json.Marshal(input)
	if err != nil {
		return err
	}

	outputJSON, stderr, err := runDockerSolc(*solcImageRef, *solcPlatform, "--standard-json", inputJSON)
	if err != nil {
		return fmt.Errorf("compile solidity contracts: %w\n%s", err, strings.TrimSpace(stderr))
	}

	var output solcStandardJSONOutput
	if err := json.Unmarshal(outputJSON, &output); err != nil {
		return fmt.Errorf("decode solc output: %w", err)
	}

	var fatalMessages []string
	for _, message := range output.Errors {
		line := strings.TrimSpace(message.FormattedMessage)
		if line == "" {
			continue
		}
		if strings.EqualFold(message.Severity, "error") {
			fatalMessages = append(fatalMessages, line)
			continue
		}
		fmt.Fprintln(os.Stderr, line)
	}
	if len(fatalMessages) > 0 {
		return fmt.Errorf("solidity compile errors:\n%s", strings.Join(fatalMessages, "\n"))
	}

	compilerVersion, err := readSolcCompilerVersion(*solcImageRef, *solcPlatform)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(paths.artifactsDir, 0o755); err != nil {
		return err
	}

	for _, definition := range contractArtifactDefinitions {
		compiled, ok := output.Contracts[definition.sourceName][definition.contractName]
		if !ok {
			return fmt.Errorf(
				"missing compiled contract %s from %s",
				definition.contractName,
				definition.sourceName,
			)
		}

		creationCode := strings.TrimSpace(compiled.EVM.Bytecode.Object)
		runtimeCode := strings.TrimSpace(compiled.EVM.DeployedBytecode.Object)
		if creationCode == "" || runtimeCode == "" {
			return fmt.Errorf("compiled contract %s is missing bytecode", definition.contractName)
		}

		payload := receiverArtifact{
			SourceName:      definition.sourceName,
			ContractName:    definition.contractName,
			CompilerVersion: compilerVersion,
			ABI:             compiled.ABI,
			CreationCodeHex: "0x" + creationCode,
			RuntimeCodeHex:  "0x" + runtimeCode,
		}

		raw, err := json.MarshalIndent(payload, "", "  ")
		if err != nil {
			return err
		}
		raw = append(raw, '\n')

		if err := os.WriteFile(filepath.Join(paths.artifactsDir, definition.fileName), raw, 0o644); err != nil {
			return err
		}
	}

	return nil
}

func loadSolcInput(contractsDir string) (solcStandardJSONInput, error) {
	entries, err := os.ReadDir(contractsDir)
	if err != nil {
		return solcStandardJSONInput{}, err
	}

	sources := make(map[string]solcStandardJSONInputEntry)
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".sol" {
			continue
		}

		raw, err := os.ReadFile(filepath.Join(contractsDir, entry.Name()))
		if err != nil {
			return solcStandardJSONInput{}, err
		}
		sources[entry.Name()] = solcStandardJSONInputEntry{Content: string(raw)}
	}

	if len(sources) == 0 {
		return solcStandardJSONInput{}, fmt.Errorf("no Solidity sources found in %s", contractsDir)
	}

	return solcStandardJSONInput{
		Language: "Solidity",
		Sources:  sources,
		Settings: solcStandardJSONSettings{
			Optimizer: solcOptimizerSetting{
				Enabled: true,
				Runs:    200,
			},
			OutputSelection: map[string]map[string][]string{
				"*": {
					"*": {"abi", "evm.bytecode.object", "evm.deployedBytecode.object"},
				},
			},
		},
	}, nil
}

func runDockerSolc(imageRef string, platform string, subcommand string, stdin []byte) ([]byte, string, error) {
	args := []string{"run", "--rm", "-i"}
	if strings.TrimSpace(platform) != "" {
		args = append(args, "--platform", platform)
	}
	args = append(args, imageRef, subcommand)

	cmd := exec.Command("docker", args...)
	cmd.Stdin = bytes.NewReader(stdin)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	return stdout.Bytes(), stderr.String(), err
}

func readSolcCompilerVersion(imageRef string, platform string) (string, error) {
	args := []string{"run", "--rm"}
	if strings.TrimSpace(platform) != "" {
		args = append(args, "--platform", platform)
	}
	args = append(args, imageRef, "--version")

	output, err := exec.Command("docker", args...).CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("read solc compiler version: %w\n%s", err, strings.TrimSpace(string(output)))
	}

	lines := strings.Split(string(output), "\n")
	slices.Reverse(lines)
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "Version:") {
			return strings.TrimSpace(strings.TrimPrefix(trimmed, "Version:")), nil
		}
	}

	return "", fmt.Errorf("read solc compiler version: missing version line")
}
