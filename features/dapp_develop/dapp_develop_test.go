package dapp_develop

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"
	"text/template"

	"github.com/cucumber/godog/colors"
	"github.com/go-cmd/cmd"

	"github.com/cucumber/godog"
	"github.com/pkg/errors"
	"github.com/stellar/go/strkey"
	"github.com/stellar/go/xdr"
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

type ledgerEntryResult struct {
	XDR string `json:"xdr"`
}

type getLedgerEntryResponse struct {
	Result ledgerEntryResult `json:"result"`
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

func invokeContract(ctx context.Context, functionName, contractName, param1, tool string) error {

	testConfig := ctx.Value(e2e.TestConfigContextKey).(*testConfig)

	if testConfig.DeployedContractId == "" {
		return fmt.Errorf("no deployed id found for contract %v", contractName)
	}

	var response string
	var err error

	switch tool {
	case "CLI":
		response, err = invokeContractFromCliTool(testConfig, functionName, contractName, param1)
	case "NODEJS":
		response, err = invokeContractFromNodeJSTool(testConfig, functionName, contractName, param1)
	default:
		err = fmt.Errorf("%s tool not supported yet", tool)
	}

	if err != nil {
		return err
	}

	testConfig.ContractFunctionResponse = response
	return nil
}

func invokeContractFromCliTool(testConfig *testConfig, functionName, contractName, param1 string) (string, error) {
	args := []string{
		"contract",
		"invoke",
		"--id", testConfig.DeployedContractId,
		"--rpc-url", testConfig.E2EConfig.TargetNetworkRPCURL,
		"--secret-key", testConfig.E2EConfig.TargetNetworkSecretKey,
		"--network-passphrase", testConfig.E2EConfig.TargetNetworkPassPhrase,
		"--",
		functionName,
	}

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

func invokeContractFromNodeJSTool(testConfig *testConfig, functionName, contractName, param1 string) (string, error) {
	script := `
		const SorobanClient = require('soroban-client');
		const xdr = SorobanClient.xdr;

		(async () => {
			const contract = new SorobanClient.Contract("{{js .contractId}}");
			const server = new SorobanClient.Server("{{js .rpcUrl}}", { allowHttp: true });
			const source = await server.getAccount("{{js .account}}")
			let txn = new SorobanClient.TransactionBuilder(source, {
					fee: 100,
					networkPassphrase: "{{js .networkPassphrase}}",
				})
				.addOperation(contract.call("{{js .functionName}}", "{{js .param1}}"))
				.setTimeout(30)
				.build();
			txn = await server.prepareTransaction(txn, "{{js .networkPassphrase}}");
			txn.sign(SorobanClient.Keypair.fromSecret("{{js .secretKey}}"));
			let response = server.sendTransaction(txn);
			let i = 0;
			while (response.status === "pending") {
				i += 1;
				if (i > 10) {
					throw new Error("Transaction timed out");
				}
				await new Promise(resolve => setTimeout(resolve, 1000));
				response = await server.getTransaction(response.id);
				if (response.status != "pending") {
					console.log(response.resultXdr);
					return;
				}
			}
		})();
	`

	stdin := &bytes.Buffer{}
	tmpl, err := template.New("script").Parse(script)
	if err != nil {
		return "", errors.Wrap(err, "failed to parse javascript template")
	}
	err = tmpl.Execute(stdin, map[string]string{
		"contractId":        testConfig.DeployedContractId,
		"rpcUrl":            testConfig.E2EConfig.TargetNetworkRPCURL,
		"account":           testConfig.E2EConfig.TargetNetworkPublicKey,
		"secretKey":         testConfig.E2EConfig.TargetNetworkSecretKey,
		"networkPassphrase": testConfig.E2EConfig.TargetNetworkPassPhrase,
		"functionName":      functionName,
		"param1":            param1,
	})
	if err != nil {
		return "", errors.Wrap(err, "failed to execute javascript template")
	}

	envCmd := cmd.NewCmd("node")
	status, stdOut, err := e2e.RunCommandWithStdin(envCmd, testConfig.E2EConfig, stdin)

	if status != 0 || err != nil {
		return "", fmt.Errorf("nodejs incoke of example contract %s had error %v, %v", contractName, status, err)
	}

	if len(stdOut) < 1 {
		return "", fmt.Errorf("nodejs invoke of example contract %s did not print any response", contractName)
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
	decoded, err := strkey.Decode(strkey.VersionByteAccountID, testConfig.E2EConfig.TargetNetworkPublicKey)
	if err != nil {
		return fmt.Errorf("invalid account address: %v", err)
	}
	var key xdr.Uint256
	copy(key[:], decoded)
	keyXdr, err := xdr.LedgerKey{
		Type: xdr.LedgerEntryTypeAccount,
		Account: &xdr.LedgerKeyAccount{
			AccountId: xdr.AccountId(xdr.PublicKey{
				Type:    xdr.PublicKeyTypePublicKeyTypeEd25519,
				Ed25519: &key,
			}),
		},
	}.MarshalBinaryBase64()
	if err != nil {
		return fmt.Errorf("error encoding account ledger key xdr: %v", err)
	}

	getAccountRequest := []byte(`{
           "jsonrpc": "2.0",
           "id": 10235,
           "method": "getLedgerEntry",
           "params": { 
               "key": "` + keyXdr + `"
            }
        }`)

	resp, err := http.Post(testConfig.E2EConfig.TargetNetworkRPCURL, "application/json", bytes.NewBuffer(getAccountRequest))
	if err != nil {
		return fmt.Errorf("soroban rpc get account had error %e", err)
	}

	var rpcResponse getLedgerEntryResponse
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&rpcResponse)
	if err != nil {
		return fmt.Errorf("soroban rpc get account, not able to parse response, %v, %e", resp.Body, err)
	}

	var t e2e.Asserter
	assert.NotEmpty(&t, rpcResponse.Result.XDR, "RPC get account, account not found")
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
