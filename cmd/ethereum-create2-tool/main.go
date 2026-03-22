package main

import (
	"fmt"
	"os"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		return usageError("usage: go run ./cmd/ethereum-create2-tool <build|predict|verify-chain>")
	}

	switch args[0] {
	case "build":
		return runBuild(args[1:])
	case "predict":
		return runPredict(args[1:])
	case "verify-chain":
		return runVerifyChain(args[1:])
	default:
		return usageError("unknown subcommand %q", args[0])
	}
}
