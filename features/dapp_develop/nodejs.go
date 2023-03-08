package dapp_develop

import (
	"fmt"
	"strings"

	"github.com/go-cmd/cmd"

	e2e "github.com/stellar/system-test"
)

// return the fn response as a serialized string
// uses secret-key and network-passphrase directly on command
func invokeContractFromNodeJSTool(deployedContractId, contractName, functionName, param1 string, e2eConfig *e2e.E2EConfig) (string, error) {
	args := []string{
		"run",
		"-s",
		"invoke",
		"--",
		"--id", deployedContractId,
		"--rpc-url", e2eConfig.TargetNetworkRPCURL,
		"--account", e2eConfig.TargetNetworkPublicKey,
		"--secret-key", e2eConfig.TargetNetworkSecretKey,
		"--network-passphrase", e2eConfig.TargetNetworkPassPhrase,
		"--fn", functionName,
	}
	if param1 != "" {
		args = append(args, "--param1", param1)
	}
	envCmd := cmd.NewCmd("yarn", args...)
	status, stdOutLines, err := e2e.RunCommand(envCmd, e2eConfig)
	stdOut := strings.TrimSpace(strings.Join(stdOutLines, "\n"))

	if status != 0 || err != nil {
		return "", fmt.Errorf("nodejs invoke of example contract %s had error %v, %v, stdout: %v", contractName, status, err, stdOut)
	}

	if stdOut == "" {
		return "", fmt.Errorf("nodejs invoke of example contract %s did not print any response", contractName)
	}

	return stdOut, nil
}

// invokes the contract using identities and network from prior setup of config state in cli
func invokeContractFromNodeJSToolWithConfig(deployedContractId string, contractName string, functionName string, parameters string, identity string, networkConfig string, e2eConfig *e2e.E2EConfig) (string, error) {
	return "", fmt.Errorf("invoke with named identity not supported for NODEJS tool")
}
