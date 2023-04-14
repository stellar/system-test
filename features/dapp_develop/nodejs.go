package dapp_develop

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/go-cmd/cmd"

	e2e "github.com/stellar/system-test"
)

// return the fn response as a serialized string
// uses secret-key and network-passphrase directly on command
func invokeContractFromNodeJSTool(deployedContractId, contractName, functionName, param1 string, e2eConfig *e2e.E2EConfig) (string, error) {
	args := []string{
		"--id", deployedContractId,
		"--rpc-url", e2eConfig.TargetNetworkRPCURL,
		"--source", e2eConfig.TargetNetworkSecretKey,
		"--network-passphrase", e2eConfig.TargetNetworkPassPhrase,
		"--", functionName,
	}
	if param1 != "" {
		args = append(args, "--param1", param1)
	}
	envCmd := cmd.NewCmd("./invoke.ts", args...)
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
func invokeContractFromNodeJSToolWithConfig(deployedContractId, contractName, functionName, parameters, identity, networkConfig string, e2eConfig *e2e.E2EConfig) (string, error) {
	return "", fmt.Errorf("invoke with named identity not supported for NODEJS tool")
}

func getEventsFromNodeJSTool(ledgerFrom uint32, deployedContractId string, size uint32, e2eConfig *e2e.E2EConfig) ([]map[string]interface{}, error) {
	args := []string{
		"--id", deployedContractId,
		"--rpc-url", e2eConfig.TargetNetworkRPCURL,
		"--size", fmt.Sprint(size),
		"--ledgerFrom", fmt.Sprint(ledgerFrom),
	}

	envCmd := cmd.NewCmd("./events.ts", args...)
	status, stdOutLines, err := e2e.RunCommand(envCmd, e2eConfig)

	var jsonEvents []map[string]interface{}

	if status != 0 || err != nil {
		return jsonEvents, fmt.Errorf("soroban js client get events had error %v, %v", status, err)
	}

	stdOutEvents := strings.TrimSpace(strings.Join(stdOutLines, "\n"))
	if stdOutEvents == "" {
		return jsonEvents, fmt.Errorf("soroban js client get events did not emit successful console response")
	}

	err = json.Unmarshal([]byte(stdOutEvents), &jsonEvents)
	if err != nil {
		return jsonEvents, fmt.Errorf("soroban js client get events response output %v was not parseable as event json, %v", stdOutEvents, err)
	}

	return jsonEvents, nil
}
