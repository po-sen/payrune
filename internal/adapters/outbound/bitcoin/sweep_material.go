package bitcoin

import "encoding/json"

const bitcoinSweepMaterialVersion = 1

type sweepMaterial struct {
	MaterialType     string `json:"material_type"`
	MaterialVersion  int    `json:"material_version"`
	Chain            string `json:"chain"`
	Network          string `json:"network"`
	Address          string `json:"address"`
	HDDerivationPath string `json:"hd_derivation_path"`
	AccountXPub      string `json:"account_xpub"`
	ScriptType       string `json:"script_type"`
}

func buildSweepMaterialJSON(
	chain string,
	network string,
	address string,
	hdDerivationPath string,
	accountXPub string,
	scriptType string,
) (string, error) {
	raw, err := json.Marshal(sweepMaterial{
		MaterialType:     "bitcoin_hd",
		MaterialVersion:  bitcoinSweepMaterialVersion,
		Chain:            chain,
		Network:          network,
		Address:          address,
		HDDerivationPath: hdDerivationPath,
		AccountXPub:      accountXPub,
		ScriptType:       scriptType,
	})
	if err != nil {
		return "", err
	}
	return string(raw), nil
}
