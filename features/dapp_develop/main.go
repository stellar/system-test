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

	envCmd = cmd.NewCmd("cargo", "build", "--config", "net.git-fetch-with-cli=true", "--target", "wasm32-unknown-unknown", "--release")
	envCmd.Dir = fmt.Sprintf("%s/%s", contractWorkingDirectory, contractExamplesSubPath)

	status, _, err = e2e.RunCommand(envCmd, e2eConfig)

	if status != 0 || err != nil {
		return fmt.Errorf("cargo build of sample contract %v/%v had error %v, %v", e2eConfig.SorobanExamplesRepoURL, contractExamplesSubPath, status, err)
	}

	return nil
}

// returns the deployed contract id
func deployContract(compiledContractFileName string, contractWorkingDirectory string, installedContractId string, e2eConfig *e2e.E2EConfig) (string, error) {
	var envCmd *cmd.Cmd

	if installedContractId != "" {
		envCmd = cmd.NewCmd("soroban",
			"contract",
			"deploy",
			"--wasm-hash", installedContractId,
			"--rpc-url", e2eConfig.TargetNetworkRPCURL,
			"--secret-key", e2eConfig.TargetNetworkSecretKey,
			"--network-passphrase", e2eConfig.TargetNetworkPassPhrase)
	} else {
		envCmd = cmd.NewCmd("soroban",
			"contract",
			"deploy",
			"--wasm", fmt.Sprintf("./%s/target/wasm32-unknown-unknown/release/%s", contractWorkingDirectory, compiledContractFileName),
			"--rpc-url", e2eConfig.TargetNetworkRPCURL,
			"--secret-key", e2eConfig.TargetNetworkSecretKey,
			"--network-passphrase", e2eConfig.TargetNetworkPassPhrase)
	}

	status, stdOut, err := e2e.RunCommand(envCmd, e2eConfig)

	if status != 0 || err != nil {
		return "", fmt.Errorf("soroban cli deployment of example contract %s had error %v, %v", compiledContractFileName, status, err)
	}

	if len(stdOut) < 1 {
		return "", fmt.Errorf("soroban cli deployment of example contract %s returned no contract id", compiledContractFileName)
	}

	return stdOut[0], nil
}

func deployContractUsingConfigParams(compiledContractFileName string, contractWorkingDirectory string, identityName string, networkConfigName string, e2eConfig *e2e.E2EConfig) (string, error) {
	envCmd := cmd.NewCmd("soroban",
		"contract",
		"deploy",
		"--wasm", fmt.Sprintf("./%s/target/wasm32-unknown-unknown/release/%s", contractWorkingDirectory, compiledContractFileName),
		"--network", networkConfigName,
		"--identity", identityName)

	status, stdOut, err := e2e.RunCommand(envCmd, e2eConfig)

	if status != 0 || err != nil {
		return "", fmt.Errorf("soroban cli deployment of example contract %s had error %v, %v", compiledContractFileName, status, err)
	}

	if len(stdOut) < 1 {
		return "", fmt.Errorf("soroban cli deployment of example contract %s returned no contract id", compiledContractFileName)
	}

	return stdOut[0], nil
}

// returns the installed contract id
func installContract(compiledContractFileName string, contractWorkingDirectory string, e2eConfig *e2e.E2EConfig) (string, error) {
	envCmd := cmd.NewCmd("soroban",
		"contract",
		"install",
		"--wasm", fmt.Sprintf("./%s/target/wasm32-unknown-unknown/release/%s", contractWorkingDirectory, compiledContractFileName),
		"--rpc-url", e2eConfig.TargetNetworkRPCURL,
		"--secret-key", e2eConfig.TargetNetworkSecretKey,
		"--network-passphrase", e2eConfig.TargetNetworkPassPhrase)

	status, stdOut, err := e2e.RunCommand(envCmd, e2eConfig)

	if status != 0 || err != nil {
		return "", fmt.Errorf("soroban cli install of example contract %s had error %v, %v", compiledContractFileName, status, err)
	}

	if len(stdOut) < 1 {
		return "", fmt.Errorf("soroban cli install of example contract %s returned no contract id", compiledContractFileName)
	}

	return stdOut[0], nil
}

func createNetworkConfig(configName string, rpcUrl string, networkPassphrase string, e2eConfig *e2e.E2EConfig) error {
	envCmd := cmd.NewCmd("soroban",
		"config",
		"network",
		"add",
		"--rpc-url", rpcUrl,
		"--network-passphrase", networkPassphrase,
		configName)

	status, _, err := e2e.RunCommand(envCmd, e2eConfig)

	if status != 0 || err != nil {
		return fmt.Errorf("soroban cli create network config %s had error %v, %v", configName, status, err)
	}

	return nil
}

// uses 'expect' cli tool program, to forward the secret to the tty that cli wait for input
func createIdentityConfig(identityName string, secretKey string, e2eConfig *e2e.E2EConfig) error {
	envCmd := cmd.NewCmd("expect",
		"soroban_config.exp",
		identityName,
		secretKey)

	status, _, err := e2e.RunCommand(envCmd, e2eConfig)

	if status != 0 || err != nil {
		return fmt.Errorf("soroban cli create identity config %s had error %v, %v", identityName, status, err)
	}

	return nil
}

// returns the contract fn invocation response payload as a serialized string
// uses secret-key and network-passphrase directly on command
func invokeContract(deployedContractId string, contractName string, functionName string, param1 string, tool string, e2eConfig *e2e.E2EConfig) (string, error) {
	var response string
	var err error

	switch tool {
	case "CLI":
		response, err = invokeContractFromCliTool(deployedContractId, contractName, functionName, param1, e2eConfig)
	case "NODEJS":
		response, err = invokeContractFromNodeJSTool(deployedContractId, contractName, functionName, param1, e2eConfig)
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
