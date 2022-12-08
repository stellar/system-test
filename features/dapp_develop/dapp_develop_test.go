package dapp_develop

import (
	"context"
	"fmt"
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
	ContractFunctionResponse string
	TestWorkingDir           string
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
		fmt.Printf("Failed to install CLI version %s, error: %v", e2eConfig.SorobanCLICrateVersion, err)
		os.Exit(1)
	}

	status := godog.TestSuite{
		Name:                 "soroban dapp e2e",
		Options:              opts,
		TestSuiteInitializer: initializeTestSuite,
		ScenarioInitializer:  initializeScenario,
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

	envCmd = cmd.NewCmd("/bin/sh", "-c", fmt.Sprintf("cd %s; git checkout %s", contractWorkingDirectory, testConfig.E2EConfig.SorobanExamplesGitHash))

	status, _, err = e2e.RunCommand(envCmd, testConfig.E2EConfig)

	if status != 0 || err != nil {
		return fmt.Errorf("git checkout %v of sample contracts repo %s had error %v, %v", testConfig.E2EConfig.SorobanExamplesGitHash, testConfig.E2EConfig.SorobanExamplesRepoURL, status, err)
	}

	envCmd = cmd.NewCmd("/bin/sh", "-c", fmt.Sprintf("cd %s/%s; cargo build --config net.git-fetch-with-cli=true --target wasm32-unknown-unknown --release", contractWorkingDirectory, contractExamplesSubPath))

	status, _, err = e2e.RunCommand(envCmd, testConfig.E2EConfig)

	if status != 0 || err != nil {
		return fmt.Errorf("cargo build of sample contract %v/%v had error %v, %v", testConfig.E2EConfig.SorobanExamplesRepoURL, contractExamplesSubPath, status, err)
	}

	return nil
}

func deployContract(ctx context.Context, compiledContractFileName string) error {

	testConfig := ctx.Value(e2e.TestConfigContextKey).(*testConfig)
	contractWorkingDirectory := fmt.Sprintf("%s/soroban_examples", testConfig.TestWorkingDir)

	envCmd := cmd.NewCmd("/bin/sh", "-c",
		fmt.Sprintf("soroban deploy --wasm ./%s/target/wasm32-unknown-unknown/release/%s --rpc-url %s --secret-key %s --network-passphrase %q",
			contractWorkingDirectory,
			compiledContractFileName,
			testConfig.E2EConfig.TargetNetworkRPCURL,
			testConfig.E2EConfig.TargetNetworkSecretKey,
			testConfig.E2EConfig.TargetNetworkPassPhrase))

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

	var param1Flag string
	var param1Value string

	if param1 != "" {
		param1Flag = "--arg"
		param1Value = fmt.Sprintf("%q", param1)
	}

	envCmd := cmd.NewCmd("/bin/sh", "-c",
		fmt.Sprintf("soroban invoke --fn %s %s %s --id %s --rpc-url %s --secret-key %s --network-passphrase %q",
			functionName,
			param1Flag,
			param1Value,
			testConfig.DeployedContractId,
			testConfig.E2EConfig.TargetNetworkRPCURL,
			testConfig.E2EConfig.TargetNetworkSecretKey,
			testConfig.E2EConfig.TargetNetworkPassPhrase))

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

func initializeScenario(ctx *godog.ScenarioContext) {
	ctx.Before(func(ctx context.Context, _ *godog.Scenario) (context.Context, error) {

		e2eConfig := ctx.Value(e2e.TestConfigContextKey).(*e2e.E2EConfig)

		testConfig := newTestConfig(e2eConfig)

		envCmd := cmd.NewCmd("/bin/sh", "-c", fmt.Sprintf("rm -rf %s; mkdir %s", e2e.TestTmpDirectory, e2e.TestTmpDirectory))

		status, _, err := e2e.RunCommand(envCmd, testConfig.E2EConfig)

		if status != 0 || err != nil {
			return nil, fmt.Errorf("could not initialize %s directory, had error %v, %v", e2e.TestTmpDirectory, status, err)
		}

		testConfig.TestWorkingDir = e2e.TestTmpDirectory
		ctx = context.WithValue(ctx, e2e.TestConfigContextKey, testConfig)
		return ctx, nil
	})

	// register resolvers to execute these steps found in .features file
	ctx.Step(`^I used cli to compile example contract (\S+)$`, compileContract)
	ctx.Step(`^I used cli to deploy contract (\S+) to network$`, deployContract)
	ctx.Step(`^I invoke function (\S+) on (\S+) with request parameter (\S*) from (\S+)$`, invokeContract)
	ctx.Step(`^the result should be (\S+)$`, theResultShouldBe)
}

func initializeTestSuite(ctx *godog.TestSuiteContext) {
	ctx.BeforeSuite(func() {})
}
