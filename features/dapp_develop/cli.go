package dapp_develop

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/go-cmd/cmd"

	e2e "github.com/stellar/system-test"
)

// return the fn response as a serialized string
// uses secret-key and network-passphrase directly on command
func invokeContractFromCliTool(deployedContractId, contractName, functionName, functionParams string, e2eConfig *e2e.E2EConfig) (string, error) {
	args := []string{
		"contract",
		"invoke",
		"--id", deployedContractId,
		"--rpc-url", e2eConfig.TargetNetworkRPCURL,
		"--source", e2eConfig.TargetNetworkSecretKey,
		"--network-passphrase", e2eConfig.TargetNetworkPassPhrase,
		"--send", "yes",
		"--",
		functionName,
	}

	if functionParams != "" {
		args = append(args, functionParams)
	}

	envCmd := cmd.NewCmd("soroban", args...)

	status, stdOutLines, err := e2e.RunCommand(envCmd, e2eConfig)
	stdOut := strings.TrimSpace(strings.Join(stdOutLines, "\n"))

	if status != 0 || err != nil {
		return "", fmt.Errorf("soroban cli invoke of example contract %s had error %v, %v, stdout: %v", contractName, status, err, stdOut)
	}

	if stdOut == "" {
		return "", fmt.Errorf("soroban cli invoke of example contract %s did not emit successful response", contractName)
	}

	return stdOut, nil
}

// invokes the contract using identities and network from prior setup of config state in cli
func invokeContractFromCliToolWithConfig(deployedContractId, contractName, functionName, parameters, identity, networkConfig string, e2eConfig *e2e.E2EConfig) (string, error) {
	args := []string{
		"contract",
		"invoke",
		"--id", deployedContractId,
		"--source", identity,
		"--network", networkConfig,
		"--",
		functionName,
	}

	if parameters != "" {
		args = append(args, strings.Split(parameters, " ")...)
	}

	envCmd := cmd.NewCmd("soroban", args...)

	status, stdOutLines, err := e2e.RunCommand(envCmd, e2eConfig)
	stdOut := strings.TrimSpace(strings.Join(stdOutLines, "\n"))

	if status != 0 || err != nil {
		return "", fmt.Errorf("soroban cli invoke of example contract with config states, %s had error %v, %v, stdout: %v", contractName, status, err, stdOut)
	}

	if stdOut == "" {
		return "", fmt.Errorf("soroban cli invoke of example contract with config states, %s did not emit successful response", contractName)
	}

	return stdOut, nil
}

func getEventsFromCliTool(ledgerFrom uint32, deployedContractId string, size uint32, e2eConfig *e2e.E2EConfig) ([]map[string]interface{}, error) {

	args := []string{
		"events",
		"--start-ledger", fmt.Sprint(ledgerFrom),
		"--count", fmt.Sprint(size),
		"--id", deployedContractId,
		"--rpc-url", e2eConfig.TargetNetworkRPCURL,
		"--network-passphrase", e2eConfig.TargetNetworkPassPhrase,
		"--output", "json",
	}

	envCmd := cmd.NewCmd("soroban", args...)

	status, stdOutLines, err := e2e.RunCommand(envCmd, e2eConfig)
	var jsonEvents []map[string]interface{}

	if status != 0 || err != nil {
		return jsonEvents, fmt.Errorf("soroban cli get events had error %v, %v", status, err)
	}

	// put commas between any json event objects if more than one found
	regex := regexp.MustCompile(`\n}\n{\n`) // Create a regex pattern
	stdOutEventsValidJson := regex.ReplaceAllString(strings.Join(stdOutLines, "\n"), "\n}\n,\n{\n")
	// wrap the json objects in json array brackets
	stdOutEventsValidJson = "[" + stdOutEventsValidJson + "]"

	err = json.Unmarshal([]byte(stdOutEventsValidJson), &jsonEvents)
	if err != nil {
		return jsonEvents, fmt.Errorf("soroban cli get events console output %v was not parseable as event json, %e", strings.Join(stdOutLines, "\n"), err)
	}

	return jsonEvents, nil
}
