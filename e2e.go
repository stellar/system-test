package e2e

import (
	"context"
	"fmt"
	"io"
	"os"
	"strconv"
	"time"

	"github.com/go-cmd/cmd"
	"github.com/stellar/go/txnbuild"
	"github.com/stellar/go/xdr"
	"github.com/stellar/stellar-rpc/client"
	"github.com/stellar/stellar-rpc/protocol"
)

type E2EConfig struct {
	// e2e settings
	SorobanExamplesGitHash string
	SorobanExamplesRepoURL string
	VerboseOutput          bool

	// target network that test will use
	TargetNetworkRPCURL     string
	TargetNetworkPassPhrase string
	TargetNetworkSecretKey  string
	TargetNetworkPublicKey  string
	// if true, means the core is running in same container as tests
	LocalCore bool
	// the relative feature file path
	FeaturePath string
}

type AccountInfo struct {
	ID       string `json:"id"`
	Sequence int64  `json:"sequence,string"`
}

const TestTmpDirectory = "test_tmp_workspace"

func InitEnvironment() (*E2EConfig, error) {
	var flagConfig = &E2EConfig{}
	var err error

	if flagConfig.FeaturePath, err = getEnv("FeaturePath"); err != nil {
		return nil, err
	}
	if flagConfig.SorobanExamplesGitHash, err = getEnv("SorobanExamplesGitHash"); err != nil {
		return nil, err
	}
	if flagConfig.SorobanExamplesRepoURL, err = getEnv("SorobanExamplesRepoURL"); err != nil {
		return nil, err
	}
	if flagConfig.TargetNetworkRPCURL, err = getEnv("TargetNetworkRPCURL"); err != nil {
		return nil, err
	}
	if flagConfig.TargetNetworkPassPhrase, err = getEnv("TargetNetworkPassPhrase"); err != nil {
		return nil, err
	}
	if flagConfig.TargetNetworkSecretKey, err = getEnv("TargetNetworkSecretKey"); err != nil {
		return nil, err
	}
	if flagConfig.TargetNetworkPublicKey, err = getEnv("TargetNetworkPublicKey"); err != nil {
		return nil, err
	}
	if verboseOutput, err := getEnv("VerboseOutput"); err == nil {
		flagConfig.VerboseOutput, _ = strconv.ParseBool(verboseOutput)
	}
	if LocalCore, err := getEnv("LocalCore"); err == nil {
		flagConfig.LocalCore, _ = strconv.ParseBool(LocalCore)
	}

	return flagConfig, nil
}

type TestContextKey string

var TestConfigContextKey = TestContextKey("TestConfig")

func RunCommand(testCmd *cmd.Cmd, config *E2EConfig) (int, []string, error) {
	return RunCommandWithStdin(testCmd, config, nil)
}

func RunCommandWithStdin(testCmd *cmd.Cmd, config *E2EConfig, stdin io.Reader) (int, []string, error) {
	// Run, stream output, and wait for Cmd to return Status
	if config.VerboseOutput {
		fmt.Printf("running command %s %v \n\n", testCmd.Name, testCmd.Args)
	}

	output := []string{}

	cmdOptions := cmd.Options{
		Buffered:  false,
		Streaming: true,
	}
	envCmd := cmd.NewCmdOptions(cmdOptions, testCmd.Name, testCmd.Args...)
	envCmd.Dir = testCmd.Dir

	doneChan := make(chan struct{})
	go func() {
		defer close(doneChan)
		for envCmd.Stdout != nil || envCmd.Stderr != nil {
			select {
			case line, open := <-envCmd.Stdout:
				if !open {
					envCmd.Stdout = nil
					continue
				}
				if config.VerboseOutput {
					fmt.Fprintln(os.Stdout, line)
				}
				output = append(output, line)
			case line, open := <-envCmd.Stderr:
				if !open {
					envCmd.Stderr = nil
					continue
				}
				if config.VerboseOutput {
					fmt.Fprintln(os.Stderr, line)
				}
			}
		}
	}()

	// Run and wait for Cmd to return, discard Status
	if stdin != nil {
		<-envCmd.StartWithStdin(stdin)
	} else {
		<-envCmd.Start()
	}

	// Wait for goroutine to print everything
	<-doneChan

	return envCmd.Status().Exit, output, envCmd.Status().Error
}

// asserter is used to be able to retrieve the error reported by the called assertion
type Asserter struct {
	Err error
}

// Errorf is used by the called assertion to report an error
func (a *Asserter) Errorf(format string, args ...interface{}) {
	a.Err = fmt.Errorf(format, args...)
}

func QueryNetworkState(e2eConfig *E2EConfig) (protocol.GetLatestLedgerResponse, error) {
	cli := client.NewClient(e2eConfig.TargetNetworkRPCURL, nil)
	return cli.GetLatestLedger(context.Background())
}

func QueryAccount(e2eConfig *E2EConfig, publicKey string) (*AccountInfo, error) {
	accountId := xdr.MustAddress(publicKey)
	key, err := accountId.LedgerKey()
	if err != nil {
		return nil, fmt.Errorf("error transforming %s into LedgerKey: %v", publicKey, err)
	}

	keyXdr, err := key.MarshalBinaryBase64()
	if err != nil {
		return nil, fmt.Errorf("error encoding account ledger key xdr: %v", err)
	}

	cli := client.NewClient(e2eConfig.TargetNetworkRPCURL, nil)
	resp, err := cli.GetLedgerEntries(context.Background(), protocol.GetLedgerEntriesRequest{
		Keys: []string{keyXdr},
	})
	if err != nil {
		return nil, fmt.Errorf("getLedgerEntries failed: %w", err)
	}

	var entry xdr.LedgerEntryData
	if len(resp.Entries) == 0 {
		return nil, fmt.Errorf("unable to find account for key %v, %e", keyXdr, err)
	}
	err = xdr.SafeUnmarshalBase64(resp.Entries[0].DataXDR, &entry)
	if err != nil {
		return nil, fmt.Errorf("failed to parse LedgerEntryData from getLedgerEntries: %w: %v", err, resp.Entries[0].DataXDR)
	}

	return &AccountInfo{
		ID:       entry.Account.AccountId.Address(),
		Sequence: int64(entry.Account.SeqNum),
	}, nil
}

func QueryTxStatus(e2eConfig *E2EConfig, txHashId string) (protocol.GetTransactionResponse, error) {
	c := client.NewClient(e2eConfig.TargetNetworkRPCURL, nil)
	return c.GetTransaction(context.Background(), protocol.GetTransactionRequest{
		Hash: txHashId,
	})
}

func TxSub(e2eConfig *E2EConfig, tx *txnbuild.Transaction) (protocol.GetTransactionResponse, error) {
	b64, err := tx.Base64()
	if err != nil {
		return protocol.GetTransactionResponse{}, fmt.Errorf(
			"sendTransaction: failed to serialize tx (%v): %e", tx, err)
	}

	c := client.NewClient(e2eConfig.TargetNetworkRPCURL, nil)
	resp, err := c.SendTransaction(context.Background(), protocol.SendTransactionRequest{
		Transaction: b64,
	})

	start := time.Now().Unix()
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()
	for x := range ticker.C {
		if x.Unix()-start > 30 {
			break
		}

		txStatus, err := QueryTxStatus(e2eConfig, resp.Hash)
		if err != nil {
			return txStatus, fmt.Errorf("getTransaction failed: %v, %e", txStatus, err)
		}

		switch txStatus.Status {
		case protocol.TransactionStatusSuccess:
			return txStatus, nil
		case protocol.TransactionStatusNotFound:
			// no-op. Retry.
		default:
			return txStatus, fmt.Errorf("bad response to getTransaction: %v", txStatus)
		}
	}

	return protocol.GetTransactionResponse{}, fmt.Errorf("sendTransaction failed: timeout after 30s: %v", resp)
}

func getEnv(key string) (string, error) {
	if value, ok := os.LookupEnv(key); ok {
		return value, nil
	}
	return "", fmt.Errorf("missing required env variable %s", key)
}
