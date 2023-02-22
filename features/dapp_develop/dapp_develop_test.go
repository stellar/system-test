package dapp_develop

import (
	"context"
	"strings"

	"fmt"

	"os"
	"testing"

	"github.com/cucumber/godog/colors"
	"github.com/go-cmd/cmd"

	"github.com/cucumber/godog"

	"github.com/stellar/go/keypair"
	"github.com/stellar/go/txnbuild"
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
	TesterAccountPublicKey   string
	TesterAccountPrivateKey  string
	Identities               map[string]string
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

func compileContractStep(ctx context.Context, contractExamplesSubPath string) error {
	testConfig := ctx.Value(e2e.TestConfigContextKey).(*testConfig)
	contractWorkingDirectory := fmt.Sprintf("%s/soroban_examples", testConfig.TestWorkingDir)
	return compileContract(contractExamplesSubPath, contractWorkingDirectory, testConfig.E2EConfig)
}

func deployContractStep(ctx context.Context, compiledContractFileName string) error {
	testConfig := ctx.Value(e2e.TestConfigContextKey).(*testConfig)
	contractWorkingDirectory := fmt.Sprintf("%s/soroban_examples", testConfig.TestWorkingDir)

	var err error
	if testConfig.DeployedContractId, err = deployContract(compiledContractFileName, contractWorkingDirectory, testConfig.InstalledContractId, testConfig.E2EConfig); err != nil {
		return err
	}

	return nil
}

func deployContractUsingConfigParamsStep(ctx context.Context, compiledContractFileName string, identityName string, networkConfigName string) error {
	testConfig := ctx.Value(e2e.TestConfigContextKey).(*testConfig)
	contractWorkingDirectory := fmt.Sprintf("%s/soroban_examples", testConfig.TestWorkingDir)

	var err error
	if testConfig.DeployedContractId, err = deployContractUsingConfigParams(compiledContractFileName, contractWorkingDirectory, identityName, networkConfigName, testConfig.E2EConfig); err != nil {
		return err
	}

	return nil
}

func installContractStep(ctx context.Context, compiledContractFileName string) error {

	testConfig := ctx.Value(e2e.TestConfigContextKey).(*testConfig)
	contractWorkingDirectory := fmt.Sprintf("%s/soroban_examples", testConfig.TestWorkingDir)

	var err error
	if testConfig.InstalledContractId, err = installContract(compiledContractFileName, contractWorkingDirectory, testConfig.E2EConfig); err != nil {
		return err
	}

	return nil
}

func invokeNonAuthContractStep(ctx context.Context, functionName string, contractName string, param1 string, tool string) error {

	testConfig := ctx.Value(e2e.TestConfigContextKey).(*testConfig)

	var err error
	if testConfig.ContractFunctionResponse, err = invokeContract(testConfig.DeployedContractId, contractName, functionName, param1, tool, testConfig.E2EConfig); err != nil {
		return err
	}

	return nil
}

func invokeInvokerAuthContractStep(ctx context.Context, functionName string, contractName string, parameters string, tool string, identity string, networkConfig string) error {
	testConfig := ctx.Value(e2e.TestConfigContextKey).(*testConfig)

	invokerPubKey, has := testConfig.Identities[identity]
	if !has {
		return fmt.Errorf("invocation of invoker auth contract could not proceed, no public key for identity %v", identity)
	}

	parameters = strings.Replace(parameters, "<tester_identity_pub_key>", invokerPubKey, 1)
	var err error
	if testConfig.ContractFunctionResponse, err = invokeContractWithConfig(testConfig.DeployedContractId, contractName, functionName, parameters, tool, identity, networkConfig, testConfig.E2EConfig); err != nil {
		return err
	}

	return nil
}

func createNetworkConfigStep(ctx context.Context, configName string) error {

	testConfig := ctx.Value(e2e.TestConfigContextKey).(*testConfig)

	return createNetworkConfig(configName, testConfig.E2EConfig.TargetNetworkRPCURL, testConfig.E2EConfig.TargetNetworkPassPhrase, testConfig.E2EConfig)
}

func createMyIdentityStep(ctx context.Context, identityName string) error {

	testConfig := ctx.Value(e2e.TestConfigContextKey).(*testConfig)

	if err := createIdentityConfig(identityName, testConfig.E2EConfig.TargetNetworkSecretKey, testConfig.E2EConfig); err != nil {
		return err
	}
	testConfig.Identities[identityName] = testConfig.E2EConfig.TargetNetworkPublicKey
	return nil
}

func createTestAccountIdentityStep(ctx context.Context, identityName string) error {

	testConfig := ctx.Value(e2e.TestConfigContextKey).(*testConfig)

	if err := createIdentityConfig(identityName, testConfig.TesterAccountPrivateKey, testConfig.E2EConfig); err != nil {
		return err
	}
	testConfig.Identities[identityName] = testConfig.TesterAccountPublicKey
	return nil
}

func newTestConfig(e2eConfig *e2e.E2EConfig) *testConfig {
	return &testConfig{
		E2EConfig:  e2eConfig,
		Identities: make(map[string]string, 0),
	}
}

func theResultShouldBeStep(ctx context.Context, expectedResult string) error {
	testConfig := ctx.Value(e2e.TestConfigContextKey).(*testConfig)

	var t e2e.Asserter
	assert.Equal(&t, expectedResult, testConfig.ContractFunctionResponse, "Expected %v but got %v", expectedResult, testConfig.ContractFunctionResponse)
	return t.Err
}

func queryAccountStep(ctx context.Context) error {
	testConfig := ctx.Value(e2e.TestConfigContextKey).(*testConfig)

	accountInfo, err := e2e.QueryAccount(testConfig.E2EConfig, testConfig.E2EConfig.TargetNetworkPublicKey)

	if err != nil {
		return fmt.Errorf("soroban rpc account retrieval had error %e", err)
	}

	var t e2e.Asserter
	assert.Equal(&t, testConfig.E2EConfig.TargetNetworkPublicKey, accountInfo.ID, "RPC get account, Expected %v but got %v", testConfig.E2EConfig.TargetNetworkPublicKey, accountInfo.ID)
	return t.Err
}

func createTesterAccountStep(ctx context.Context) error {
	testConfig := ctx.Value(e2e.TestConfigContextKey).(*testConfig)

	kp := keypair.MustParseFull(testConfig.E2EConfig.TargetNetworkSecretKey)
	address := kp.Address()

	addressState, err := e2e.QueryAccount(testConfig.E2EConfig, address)

	if err != nil {
		return fmt.Errorf("unable to query latest account state for %v, had error %e", address, err)
	}

	account := txnbuild.NewSimpleAccount(address, addressState.Sequence)

	testerKp, err := keypair.Random()
	if err != nil {
		return fmt.Errorf("unable to generate key pair for tester account had error %e", err)
	}

	tx, err := txnbuild.NewTransaction(txnbuild.TransactionParams{
		SourceAccount:        &account,
		IncrementSequenceNum: true,
		Operations: []txnbuild.Operation{
			&txnbuild.CreateAccount{
				Destination:   testerKp.Address(),
				Amount:        "100",
				SourceAccount: address,
			},
		},
		BaseFee: txnbuild.MinBaseFee,
		Preconditions: txnbuild.Preconditions{
			TimeBounds: txnbuild.NewInfiniteTimeout(),
		},
	})
	if err != nil {
		return fmt.Errorf("building transaction to create tester account had error %e", err)
	}

	tx, err = tx.Sign(testConfig.E2EConfig.TargetNetworkPassPhrase, kp)
	if err != nil {
		return fmt.Errorf("signing transaction to create tester account had error %v, %e", tx, err)
	}

	txStatus, err := e2e.TxSub(testConfig.E2EConfig, tx)
	if err != nil {
		return fmt.Errorf("not able to submit transaction to create tester account %e", err)
	}
	if txStatus.Error != nil {
		return fmt.Errorf("submission of tx to create tester account had error %v", txStatus.Error)
	}

	if testConfig.E2EConfig.VerboseOutput {
		fmt.Fprintf(os.Stdout, "created and funded test accout %v", testerKp.Address())
	}

	testConfig.TesterAccountPublicKey = testerKp.Address()
	testConfig.TesterAccountPrivateKey = testerKp.Seed()
	return nil
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

		scenarioCtx.Step(`^I used cargo to compile example contract ([\S|\s]+)$`, compileContractStep)
		scenarioCtx.Step(`^I used rpc to verify my account is on the network`, queryAccountStep)

		switch scenario.Name {
		case "DApp developer compiles, installs, deploys and invokes a contract":
			scenarioCtx.Step(`^I used cli to install contract ([\S|\s]+) on network using my secret key$`, installContractStep)
			scenarioCtx.Step(`^I used cli to deploy contract ([\S|\s]+) by installed hash using my secret key$`, deployContractStep)
			scenarioCtx.Step(`^I invoke function ([\S|\s]+) on ([\S|\s]+) with request parameter ([\S|\s]*) from ([\S|\s]+) using my secret key$`, invokeNonAuthContractStep)
		case "DApp developer compiles, deploys and invokes a contract":
			scenarioCtx.Step(`^I used cli to deploy contract ([\S|\s]+) using my secret key$`, deployContractStep)
			scenarioCtx.Step(`^I invoke function ([\S|\s]+) on ([\S|\s]+) with request parameter ([\S|\s]*) from ([\S|\s]+) using my secret key$`, invokeNonAuthContractStep)
		case "DApp developer uses config states, compiles, deploys and invokes contract with authorizations":
			scenarioCtx.Step(`^I used rpc to submit transaction to create tester account on the network$`, createTesterAccountStep)
			scenarioCtx.Step(`^I used cli to add Network Config ([\S|\s]+) for rpc and standalone$`, createNetworkConfigStep)
			scenarioCtx.Step(`^I used cli to add Identity ([\S|\s]+) for my secret key$`, createMyIdentityStep)
			scenarioCtx.Step(`^I used cli to add Identity ([\S|\s]+) for tester secret key$`, createTestAccountIdentityStep)
			scenarioCtx.Step(`^I used cli to deploy contract ([\S|\s]+) using my Identity ([\S|\s]+) and Network Config ([\S|\s]+)$`, deployContractUsingConfigParamsStep)
			scenarioCtx.Step(`^I invoke function ([\S|\s]+) on ([\S|\s]+) with request parameters ([\S|\s]*) from ([\S|\s]+) using tester Identity ([\S|\s]+) as invoker and Network Config ([\S|\s]+)$`, invokeInvokerAuthContractStep)
		}

		scenarioCtx.Step(`^the result should be (\S+)$`, theResultShouldBeStep)

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
