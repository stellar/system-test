package dapp_develop

import (
	"fmt"

	"github.com/go-cmd/cmd"

	e2e "github.com/stellar/system-test"
)

func compileContract(contractExamplesSubPath string, contractWorkingDirectory string, e2eConfig *e2e.E2EConfig) error {
	envCmd := cmd.NewCmd("git", "clone", e2eConfig.SorobanExamplesRepoURL, contractWorkingDirectory)

	status, _, err := e2e.RunCommand(envCmd, e2eConfig)

	if status != 0 || err != nil {
		return fmt.Errorf("git clone of soroban example contracts from %s had error %v, %v", e2eConfig.SorobanExamplesRepoURL, status, err)
	}

	envCmd = cmd.NewCmd("git", "checkout", e2eConfig.SorobanExamplesGitHash)
	envCmd.Dir = contractWorkingDirectory

	status, _, err = e2e.RunCommand(envCmd, e2eConfig)

	if status != 0 || err != nil {
		return fmt.Errorf("git checkout %v of sample contracts repo %s had error %v, %v", e2eConfig.SorobanExamplesGitHash, e2eConfig.SorobanExamplesRepoURL, status, err)
	}

	envCmd = cmd.NewCmd("stellar", "contract", "build")
	envCmd.Dir = fmt.Sprintf("%s/%s", contractWorkingDirectory, contractExamplesSubPath)

	status, _, err = e2e.RunCommand(envCmd, e2eConfig)

	if status != 0 || err != nil {
		return fmt.Errorf("cargo build of sample contract %v/%v had error %v, %v", e2eConfig.SorobanExamplesRepoURL, contractExamplesSubPath, status, err)
	}

	return nil
}

// returns the deployed contract id
func deployContract(compiledContractFileName string, contractWorkingDirectory string, contractExamplesSubPath string, installedContractId string, e2eConfig *e2e.E2EConfig) (string, error) {
	var envCmd *cmd.Cmd

	if installedContractId != "" {
		envCmd = cmd.NewCmd("stellar",
			"contract",
			"deploy",
			"--quiet",
			"--wasm-hash", installedContractId,
			"--rpc-url", e2eConfig.TargetNetworkRPCURL,
			"--source", e2eConfig.TargetNetworkSecretKey,
			"--network-passphrase", e2eConfig.TargetNetworkPassPhrase)
	} else {
		envCmd = cmd.NewCmd("stellar",
			"contract",
			"deploy",
			"--quiet",
			"--wasm", fmt.Sprintf("./%s/%s/target/wasm32v1-none/release/%s", contractWorkingDirectory, contractExamplesSubPath, compiledContractFileName),
			"--rpc-url", e2eConfig.TargetNetworkRPCURL,
			"--source", e2eConfig.TargetNetworkSecretKey,
			"--network-passphrase", e2eConfig.TargetNetworkPassPhrase)
	}

	status, stdOut, err := e2e.RunCommand(envCmd, e2eConfig)

	if status != 0 || err != nil {
		return "", fmt.Errorf("stellar cli deployment of example contract %s had error %v, %v", compiledContractFileName, status, err)
	}

	if len(stdOut) < 1 {
		return "", fmt.Errorf("stellar cli deployment of example contract %s returned no contract id", compiledContractFileName)
	}

	return stdOut[0], nil
}

func deployContractUsingConfigParams(compiledContractFileName string, contractWorkingDirectory string, contractExamplesSubPath string, identityName string, networkConfigName string, e2eConfig *e2e.E2EConfig) (string, error) {
	envCmd := cmd.NewCmd("stellar",
		"contract",
		"deploy",
		"--quiet",
		"--wasm", fmt.Sprintf("./%s/%s/target/wasm32v1-none/release/%s", contractWorkingDirectory, contractExamplesSubPath, compiledContractFileName),
		"--network", networkConfigName,
		"--source", identityName)

	status, stdOut, err := e2e.RunCommand(envCmd, e2eConfig)

	if status != 0 || err != nil {
		return "", fmt.Errorf("stellar cli deployment of example contract %s had error %v, %v", compiledContractFileName, status, err)
	}

	if len(stdOut) < 1 {
		return "", fmt.Errorf("stellar cli deployment of example contract %s returned no contract id", compiledContractFileName)
	}

	return stdOut[0], nil
}

// returns the installed contract id
func installContract(compiledContractFileName string, contractWorkingDirectory string, contractExamplesSubPath string, e2eConfig *e2e.E2EConfig) (string, error) {
	envCmd := cmd.NewCmd("stellar",
		"contract",
		"install",
		"--quiet",
		"--wasm", fmt.Sprintf("./%s/%s/target/wasm32v1-none/release/%s", contractWorkingDirectory, contractExamplesSubPath, compiledContractFileName),
		"--rpc-url", e2eConfig.TargetNetworkRPCURL,
		"--source", e2eConfig.TargetNetworkSecretKey,
		"--network-passphrase", e2eConfig.TargetNetworkPassPhrase)

	status, stdOut, err := e2e.RunCommand(envCmd, e2eConfig)

	if status != 0 || err != nil {
		return "", fmt.Errorf("stellar cli install of example contract %s had error %v, %v", compiledContractFileName, status, err)
	}

	if len(stdOut) < 1 {
		return "", fmt.Errorf("stellar cli install of example contract %s returned no contract id", compiledContractFileName)
	}

	return stdOut[0], nil
}

func createNetworkConfig(configName string, rpcUrl string, networkPassphrase string, e2eConfig *e2e.E2EConfig) error {
	envCmd := cmd.NewCmd("stellar",
		"network",
		"add",
		"--rpc-url", rpcUrl,
		"--network-passphrase", networkPassphrase,
		configName)

	status, _, err := e2e.RunCommand(envCmd, e2eConfig)

	if status != 0 || err != nil {
		return fmt.Errorf("stellar cli create network config %s had error %v, %v", configName, status, err)
	}

	return nil
}

// uses 'expect' cli tool program, to forward the secret to the tty that cli wait for input
func createIdentityConfig(identityName string, secretKey string, e2eConfig *e2e.E2EConfig) error {
	envCmd := cmd.NewCmd("expect",
		e2eConfig.FeaturePath+"/soroban_config.exp",
		identityName,
		secretKey)

	status, _, err := e2e.RunCommand(envCmd, e2eConfig)

	if status != 0 || err != nil {
		return fmt.Errorf("stellar cli create identity config %s had error %v, %v", identityName, status, err)
	}

	return nil
}

// returns the contract fn invocation response payload as a serialized string
// uses secret-key and network-passphrase directly on command
func invokeContract(deployedContractId string, contractName string, functionName string, functionParams string, tool string, e2eConfig *e2e.E2EConfig) (string, error) {
	var response string
	var err error

	switch tool {
	case "CLI":
		response, err = invokeContractFromCliTool(deployedContractId, contractName, functionName, functionParams, e2eConfig)
	case "NODEJS":
		response, err = invokeContractFromNodeJSTool(deployedContractId, contractName, functionName, functionParams, e2eConfig)
	default:
		err = fmt.Errorf("%s tool not supported for invoke yet", tool)
	}

	if err != nil {
		return "", err
	}

	return response, nil
}

// invokes the contract using identities and network from prior setup of config state in cli
func invokeContractWithConfig(deployedContractId string, contractName string, functionName string, parameters string, tool string, identity string, networkConfig string, e2eConfig *e2e.E2EConfig) (string, error) {
	var response string
	var err error

	switch tool {
	case "CLI":
		response, err = invokeContractFromCliToolWithConfig(deployedContractId, contractName, functionName, parameters, identity, networkConfig, e2eConfig)
	case "NODEJS":
		response, err = invokeContractFromNodeJSToolWithConfig(deployedContractId, contractName, functionName, parameters, identity, networkConfig, e2eConfig)
	default:
		err = fmt.Errorf("%s tool not supported yet for invoker auth contract", tool)
	}

	if err != nil {
		return "", err
	}

	return response, nil
}

// returns all events as json array
// ledgerFrom - required, starting point
// deployedContractId - optional, the id of contract to filter events for or nil
// tool - required, which tool to use to get events
func getEvents(ledgerFrom uint32, deployedContractId string, tool string, size uint32, e2eConfig *e2e.E2EConfig) ([]map[string]interface{}, error) {
	var response []map[string]interface{}
	var err error

	switch tool {
	case "CLI":
		response, err = getEventsFromCliTool(ledgerFrom, deployedContractId, size, e2eConfig)
	case "NODEJS":
		response, err = getEventsFromNodeJSTool(ledgerFrom, deployedContractId, size, e2eConfig)
	default:
		err = fmt.Errorf("%s tool not supported for events retrieval yet", tool)
	}

	return response, err
}
