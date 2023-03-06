package dapp_develop

import (
	"fmt"
	"strings"

	"github.com/go-cmd/cmd"

	e2e "github.com/stellar/system-test"
)

// return the fn response as a serialized string
// uses secret-key and network-passphrase directly on command
func invokeContractFromCliTool(deployedContractId string, contractName string, functionName string, param1 string, e2eConfig *e2e.E2EConfig) (string, error) {
	args := []string{
		"contract",
		"invoke",
		"--id", deployedContractId,
		"--rpc-url", e2eConfig.TargetNetworkRPCURL,
		"--secret-key", e2eConfig.TargetNetworkSecretKey,
		"--network-passphrase", e2eConfig.TargetNetworkPassPhrase,
		"--",
		functionName,
	}

	if param1 != "" {
		args = append(args, param1)
	}

	envCmd := cmd.NewCmd("soroban", args...)

	status, stdOut, err := e2e.RunCommand(envCmd, e2eConfig)

	if status != 0 || err != nil {
		return "", fmt.Errorf("soroban cli invoke of example contract %s had error %v, %v", contractName, status, err)
	}

	if len(stdOut) < 1 {
		return "", fmt.Errorf("soroban cli invoke of example contract %s did not emit successful response", contractName)
	}

	return stdOut[0], nil
}

// invokes the contract using identities and network from prior setup of config state in cli
func invokeContractFromCliToolWithConfig(deployedContractId string, contractName string, functionName string, parameters string, identity string, networkConfig string, e2eConfig *e2e.E2EConfig) (string, error) {
	args := []string{
		"contract",
		"invoke",
		"--id", deployedContractId,
		"--identity", identity,
		"--network", networkConfig,
		"--",
		functionName,
	}

	if parameters != "" {
		args = append(args, strings.Split(parameters, " ")...)
	}

	envCmd := cmd.NewCmd("soroban", args...)

	status, stdOut, err := e2e.RunCommand(envCmd, e2eConfig)

	if status != 0 || err != nil {
		return "", fmt.Errorf("soroban cli invoke of example contract with config states, %s had error %v, %v", contractName, status, err)
	}

	if len(stdOut) < 1 {
		return "", fmt.Errorf("soroban cli invoke of example contract with config states, %s did not emit successful response", contractName)
	}

	return stdOut[0], nil
}
