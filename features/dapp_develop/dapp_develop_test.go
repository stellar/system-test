package dapp_develop

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/cucumber/godog/colors"
	"github.com/go-cmd/cmd"

	"github.com/cucumber/godog"
	"github.com/stellar/system-test"
	"github.com/stretchr/testify/assert"
)

/*

   Soroban Dapp Feature Test

*/

type testConfig struct {
	E2EConfig *e2e.E2EConfig

	// per scenario step results state
	DeployedContractId       string
	InstalledContractId      string
	ContractFunctionResponse string
	TestWorkingDir           string
}

type accountInfo struct {
	ID       string `json:"id"`
	Sequence int64  `json:"sequence,string"`
}

type accountResponse struct {
	Result accountInfo `json:"result"`
}

func TestDappDevelop(t *testing.T) {
	e2eConfig, err := e2e.InitEnvironment()

	if err != nil {
		t.Fatalf("Failed to setup environment for soroban dapp e2e tests, %v", err)
	}

	opts := &godog.Options{
		Format:         "pretty",
		Paths:          []string{"dapp_develop.feature"},
		Output:         colors.Colored(os.Stdout),
		StopOnFailure:  true,
		TestingT:       t,
		DefaultContext: context.WithValue(context.Background(), e2e.TestConfigContextKey, e2eConfig),
	}
	godog.BindCommandLineFlags("godog.", opts)

	err = e2e.InstallCli(e2eConfig)

	if err != nil {
		fmt.Printf("Failed to install CLI version %s, error: %v \n\n", e2eConfig.SorobanCLICrateVersion, err)
		os.Exit(1)
	}

	status := godog.TestSuite{
		Name:                "soroban dapp e2e",
		Options:             opts,
		ScenarioInitializer: initializeScenario,
	}.Run()

	if status != 0 {
		t.Fatal("Failed to pass all soroban dapp e2e tests")
	}
}

func newTestConfig(e2eConfig *e2e.E2EConfig) *testConfig {
	return &testConfig{
		E2EConfig: e2eConfig,
	}
}

func compileContract(ctx context.Context, contractExamplesSubPath string) error {

	testConfig := ctx.Value(e2e.TestConfigContextKey).(*testConfig)

	contractWorkingDirectory := fmt.Sprintf("%s/soroban_examples", testConfig.TestWorkingDir)
	envCmd := cmd.NewCmd("git", "clone", testConfig.E2EConfig.SorobanExamplesRepoURL, contractWorkingDirectory)

	status, _, err := e2e.RunCommand(envCmd, testConfig.E2EConfig)

	if status != 0 || err != nil {
		return fmt.Errorf("git clone of soroban example contracts from %s had error %v, %v", testConfig.E2EConfig.SorobanExamplesRepoURL, status, err)
	}

	envCmd = cmd.NewCmd("git", "checkout", testConfig.E2EConfig.SorobanExamplesGitHash)
	envCmd.Dir = contractWorkingDirectory

	status, _, err = e2e.RunCommand(envCmd, testConfig.E2EConfig)

	if status != 0 || err != nil {
		return fmt.Errorf("git checkout %v of sample contracts repo %s had error %v, %v", testConfig.E2EConfig.SorobanExamplesGitHash, testConfig.E2EConfig.SorobanExamplesRepoURL, status, err)
	}

	envCmd = cmd.NewCmd("cargo", "build", "--config", "net.git-fetch-with-cli=true", "--target", "wasm32-unknown-unknown", "--release")
	envCmd.Dir = fmt.Sprintf("%s/%s", contractWorkingDirectory, contractExamplesSubPath)

	status, _, err = e2e.RunCommand(envCmd, testConfig.E2EConfig)

	if status != 0 || err != nil {
		return fmt.Errorf("cargo build of sample contract %v/%v had error %v, %v", testConfig.E2EConfig.SorobanExamplesRepoURL, contractExamplesSubPath, status, err)
	}

	return nil
}

func deployContract(ctx context.Context, compiledContractFileName string) error {

	testConfig := ctx.Value(e2e.TestConfigContextKey).(*testConfig)
	contractWorkingDirectory := fmt.Sprintf("%s/soroban_examples", testConfig.TestWorkingDir)

	var envCmd *cmd.Cmd

	if testConfig.InstalledContractId != "" {
		envCmd = cmd.NewCmd("soroban",
			"contract",
			"deploy",
			"--wasm-hash", testConfig.InstalledContractId,
			"--rpc-url", testConfig.E2EConfig.TargetNetworkRPCURL,
			"--secret-key", testConfig.E2EConfig.TargetNetworkSecretKey,
			"--network-passphrase", testConfig.E2EConfig.TargetNetworkPassPhrase)
	} else {
		envCmd = cmd.NewCmd("soroban",
			"contract",
			"deploy",
			"--wasm", fmt.Sprintf("./%s/target/wasm32-unknown-unknown/release/%s", contractWorkingDirectory, compiledContractFileName),
			"--rpc-url", testConfig.E2EConfig.TargetNetworkRPCURL,
			"--secret-key", testConfig.E2EConfig.TargetNetworkSecretKey,
			"--network-passphrase", testConfig.E2EConfig.TargetNetworkPassPhrase)
	}

	status, stdOut, err := e2e.RunCommand(envCmd, testConfig.E2EConfig)

	if status != 0 || err != nil {
		return fmt.Errorf("soroban cli deployment of example contract %s had error %v, %v", compiledContractFileName, status, err)
	}

	if len(stdOut) < 1 {
		return fmt.Errorf("soroban cli deployment of example contract %s returned no contract id", compiledContractFileName)
	}

	testConfig.DeployedContractId = stdOut[0]
	return nil
}

func installContract(ctx context.Context, compiledContractFileName string) error {

	testConfig := ctx.Value(e2e.TestConfigContextKey).(*testConfig)
	contractWorkingDirectory := fmt.Sprintf("%s/soroban_examples", testConfig.TestWorkingDir)

	envCmd := cmd.NewCmd("soroban",
		"contract",
		"install",
		"--wasm", fmt.Sprintf("./%s/target/wasm32-unknown-unknown/release/%s", contractWorkingDirectory, compiledContractFileName),
		"--rpc-url", testConfig.E2EConfig.TargetNetworkRPCURL,
		"--secret-key", testConfig.E2EConfig.TargetNetworkSecretKey,
		"--network-passphrase", testConfig.E2EConfig.TargetNetworkPassPhrase)

	status, stdOut, err := e2e.RunCommand(envCmd, testConfig.E2EConfig)

	if status != 0 || err != nil {
		return fmt.Errorf("soroban cli install of example contract %s had error %v, %v", compiledContractFileName, status, err)
	}

	if len(stdOut) < 1 {
		return fmt.Errorf("soroban cli install of example contract %s returned no contract id", compiledContractFileName)
	}

	testConfig.InstalledContractId = stdOut[0]
	return nil
}

func invokeContract(ctx context.Context, functionName string, contractName string, param1 string, tool string) error {

	testConfig := ctx.Value(e2e.TestConfigContextKey).(*testConfig)

	if testConfig.DeployedContractId == "" {
		return fmt.Errorf("no deployed id found for contract %v", contractName)
	}

	var response string
	var err error

	if tool == "CLI" {
		response, err = invokeContractFromCliTool(testConfig, functionName, contractName, param1)
	} else {
		err = fmt.Errorf("%s tool not supported yet", tool)
	}

	if err != nil {
		return err
	}

	testConfig.ContractFunctionResponse = response
	return nil
}

func invokeContractFromCliTool(testConfig *testConfig, functionName string, contractName string, param1 string) (string, error) {

	args := []string{
		"contract",
		"invoke",
		"--id", testConfig.DeployedContractId,
		"--rpc-url", testConfig.E2EConfig.TargetNetworkRPCURL,
		"--secret-key", testConfig.E2EConfig.TargetNetworkSecretKey,
		"--network-passphrase", testConfig.E2EConfig.TargetNetworkPassPhrase,
		"--fn",
		functionName,
		"--"}

	if param1 != "" {
		args = append(args, param1)
	}

	envCmd := cmd.NewCmd("soroban", args...)

	status, stdOut, err := e2e.RunCommand(envCmd, testConfig.E2EConfig)

	if status != 0 || err != nil {
		return "", fmt.Errorf("soroban cli deployment of example contract %s had error %v, %v", contractName, status, err)
	}

	if len(stdOut) < 1 {
		return "", fmt.Errorf("soroban cli invoke of example contract %s did not emit successful response", contractName)
	}

	return stdOut[0], nil
}

func theResultShouldBe(ctx context.Context, expectedResult string) error {
	testConfig := ctx.Value(e2e.TestConfigContextKey).(*testConfig)

	var t e2e.Asserter
	assert.Equal(&t, expectedResult, testConfig.ContractFunctionResponse, "Expected %v but got %v", expectedResult, testConfig.ContractFunctionResponse)
	return t.Err
}

func queryAccount(ctx context.Context) error {
	testConfig := ctx.Value(e2e.TestConfigContextKey).(*testConfig)

	getAccountRequest := []byte(`{
           "jsonrpc": "2.0",
           "id": 10235,
           "method": "getAccount",
           "params": { 
               "address": "` + testConfig.E2EConfig.TargetNetworkPublicKey + `"
            }
        }`)

	resp, err := http.Post(testConfig.E2EConfig.TargetNetworkRPCURL, "application/json", bytes.NewBuffer(getAccountRequest))
	if err != nil {
		return fmt.Errorf("soroban rpc get account had error %e", err)
	}

	var rpcResponse accountResponse
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&rpcResponse)
	if err != nil {
		return fmt.Errorf("soroban rpc get account, not able to parse response, %v, %e", resp.Body, err)
	}

	var t e2e.Asserter
	assert.Equal(&t, testConfig.E2EConfig.TargetNetworkPublicKey, rpcResponse.Result.ID, "RPC get account, Expected %v but got %v", testConfig.E2EConfig.TargetNetworkPublicKey, rpcResponse.Result.ID)
	return t.Err
}

func initializeScenario(scenarioCtx *godog.ScenarioContext) {
	scenarioCtx.Before(func(ctx context.Context, scenario *godog.Scenario) (context.Context, error) {

		e2eConfig := ctx.Value(e2e.TestConfigContextKey).(*e2e.E2EConfig)

		testConfig := newTestConfig(e2eConfig)

		envCmd := cmd.NewCmd("rm", "-rf", e2e.TestTmpDirectory)
		status, _, err := e2e.RunCommand(envCmd, testConfig.E2EConfig)

		if status != 0 || err != nil {
			return nil, fmt.Errorf("could not remove %s directory, had error %v, %v", e2e.TestTmpDirectory, status, err)
		}

		envCmd = cmd.NewCmd("mkdir", e2e.TestTmpDirectory)
		status, _, err = e2e.RunCommand(envCmd, testConfig.E2EConfig)

		if status != 0 || err != nil {
			return nil, fmt.Errorf("could not initialize %s directory, had error %v, %v", e2e.TestTmpDirectory, status, err)
		}

		testConfig.TestWorkingDir = e2e.TestTmpDirectory
		ctx = context.WithValue(ctx, e2e.TestConfigContextKey, testConfig)

		switch scenario.Name {
		case "DApp developer compiles, installs, deploys and invokes a contract":
			scenarioCtx.Step(`^I used cli to compile example contract ([\S|\s]+)$`, compileContract)
			scenarioCtx.Step(`^I used rpc to verify my account is on the network`, queryAccount)
			scenarioCtx.Step(`^I used cli to install contract ([\S|\s]+) on ledger using my account to network$`, installContract)
			scenarioCtx.Step(`^I used cli to deploy contract ([\S|\s]+) by installed hash using my account to network$`, deployContract)
			scenarioCtx.Step(`^I invoke function ([\S|\s]+) on ([\S|\s]+) with request parameter ([\S|\s]*) from ([\S|\s]+)$`, invokeContract)
			scenarioCtx.Step(`^the result should be (\S+)$`, theResultShouldBe)
		case "DApp developer compiles, deploys and invokes a contract":
			scenarioCtx.Step(`^I used cli to compile example contract ([\S|\s]+)$`, compileContract)
			scenarioCtx.Step(`^I used rpc to verify my account is on the network`, queryAccount)
			scenarioCtx.Step(`^I used cli to deploy contract ([\S|\s]+) using my account to network$`, deployContract)
			scenarioCtx.Step(`^I invoke function ([\S|\s]+) on ([\S|\s]+) with request parameter ([\S|\s]*) from ([\S|\s]+)$`, invokeContract)
			scenarioCtx.Step(`^the result should be (\S+)$`, theResultShouldBe)
		}

		return ctx, nil
	})
	scenarioCtx.After(func(ctx context.Context, _ *godog.Scenario, _ error) (context.Context, error) {
		e2eConfig := ctx.Value(e2e.TestConfigContextKey).(*testConfig)
		envCmd := cmd.NewCmd("rm", "-rf", e2e.TestTmpDirectory)
		status, _, err := e2e.RunCommand(envCmd, e2eConfig.E2EConfig)

		if status != 0 || err != nil {
			return nil, fmt.Errorf("could not remove %s directory, had error %v, %v", e2e.TestTmpDirectory, status, err)
		}
		return ctx, nil
	})
}
