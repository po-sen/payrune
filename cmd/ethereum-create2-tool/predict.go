package main

import (
	"context"
	"fmt"
	"os"
)

func runPredict(args []string) error {
	paths, err := resolveToolPaths()
	if err != nil {
		return err
	}

	flagSet := newFlagSet("predict")
	network := flagSet.String("network", defaultPredictNetwork, "Ethereum network label")
	factoryAddress := flagSet.String("factory", "", "CREATE2 factory address")
	collectorAddress := flagSet.String("collector", "", "fixed collector address")
	receiverArtifactPath := flagSet.String("receiver-artifact", paths.receiverArtifact, "path to receiver artifact JSON")
	salt := flagSet.String("salt", "", "32-byte CREATE2 salt hex; if omitted a random salt is generated")
	if err := flagSet.Parse(args); err != nil {
		return err
	}

	if *factoryAddress == "" {
		return fmt.Errorf("factory address is required")
	}
	if *collectorAddress == "" {
		return fmt.Errorf("collector address is required")
	}
	if *network == "" {
		return fmt.Errorf("network is required")
	}
	normalizedSalt, err := normalizeOrGenerateSalt(*salt)
	if err != nil {
		return err
	}

	result, err := predictFromArtifact(
		context.Background(),
		*network,
		*factoryAddress,
		*collectorAddress,
		*receiverArtifactPath,
		normalizedSalt,
	)
	if err != nil {
		return err
	}

	return writePrettyJSON(os.Stdout, result)
}
