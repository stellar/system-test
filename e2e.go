package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-cmd/cmd"
	"github.com/stellar/go/txnbuild"
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
}

type AccountInfo struct {
	ID       string `json:"id"`
	Sequence int64  `json:"sequence,string"`
}

type RPCAccountResponse struct {
	Result AccountInfo `json:"result"`
}

type TransactionResponseError struct {
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Data    map[string]interface{} `json:"data"`
}

type TransactionResponse struct {
	ID     string `json:"id"`
	Status string `json:"status"`
	// Error will be nil unless Status is equal to "error"
	Error *TransactionResponseError `json:"error"`
}

type RPCTransactionResponse struct {
	Result TransactionResponse `json:"result"`
}

type TransactionStatusResponseError struct {
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Data    map[string]interface{} `json:"data"`
}

type TransactionStatusResponse struct {
	ID            string `json:"id"`
	Status        string `json:"status"`
	EnvelopeXdr   string `json:"envelopeXdr,omitempty"`
	ResultXdr     string `json:"resultXdr,omitempty"`
	ResultMetaXdr string `json:"resultMetaXdr,omitempty"`
	// Error will be nil unless Status is equal to "error"
	Error *TransactionStatusResponseError `json:"error,omitempty"`
}

type RPCTransactionStatusResponse struct {
	Result TransactionStatusResponse `json:"result"`
}

const TestTmpDirectory = "test_tmp_workspace"

func InitEnvironment() (*E2EConfig, error) {
	var flagConfig = &E2EConfig{}
	var err error

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

	return flagConfig, nil
}

type TestContextKey string

var TestConfigContextKey = TestContextKey("TestConfig")

func RunCommand(testCmd *cmd.Cmd, config *E2EConfig) (int, []string, error) {
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
	<-envCmd.Start()

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

func QueryAccount(e2eConfig *E2EConfig, publicKey string) (*AccountInfo, error) {
	getAccountRequest := []byte(`{
           "jsonrpc": "2.0",
           "id": 10235,
           "method": "getAccount",
           "params": { 
               "address": "` + publicKey + `"
            }
        }`)

	resp, err := http.Post(e2eConfig.TargetNetworkRPCURL, "application/json", bytes.NewBuffer(getAccountRequest))
	if err != nil {
		return nil, fmt.Errorf("soroban rpc get account had error %e", err)
	}

	var rpcResponse RPCAccountResponse
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&rpcResponse)
	if err != nil {
		return nil, fmt.Errorf("soroban rpc get account, not able to parse response, %v, %e", resp.Body, err)
	}

	return &rpcResponse.Result, nil
}

func QueryTxStatus(e2eConfig *E2EConfig, txHashId string) (*TransactionStatusResponse, error) {
	getTxStatusRequest := []byte(`{
           "jsonrpc": "2.0",
           "id": 10235,
           "method": "getTransactionStatus",
           "params": { 
               "hash": "` + txHashId + `"
            }
        }`)

	resp, err := http.Post(e2eConfig.TargetNetworkRPCURL, "application/json", bytes.NewBuffer(getTxStatusRequest))
	if err != nil {
		return nil, fmt.Errorf("soroban rpc get tx status had error %e", err)
	}

	var rpcResponse RPCTransactionStatusResponse
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&rpcResponse)
	if err != nil {
		return nil, fmt.Errorf("soroban rpc get tx status, not able to parse response, %v, %e", resp.Body, err)
	}

	return &rpcResponse.Result, nil
}

func TxSub(e2eConfig *E2EConfig, tx *txnbuild.Transaction) (*TransactionStatusResponse, error) {
	b64, err := tx.Base64()
	if err != nil {
		return nil, fmt.Errorf("soroban rpc tx sub, not able to serialize tx, %v, %e", tx, err)
	}

	txsubRequest := []byte(`{
           "jsonrpc": "2.0",
           "id": 10235,
           "method": "sendTransaction",
           "params": { 
               "transaction": "` + b64 + `"
            }
        }`)

	resp, err := http.Post(e2eConfig.TargetNetworkRPCURL, "application/json", bytes.NewBuffer(txsubRequest))
	if err != nil {
		return nil, fmt.Errorf("soroban rpc tx sub had error %e", err)
	}

	var rpcResponse RPCTransactionResponse
	decoder := json.NewDecoder(resp.Body)

	err = decoder.Decode(&rpcResponse)
	if err != nil {
		return nil, fmt.Errorf("soroban rpc tx sub, not able to parse response, %v, %e", resp.Body, err)
	}

	if rpcResponse.Result.Error != nil {
		return nil, fmt.Errorf("soroban rpc tx sub, got bad submission response, %v", rpcResponse)
	}

	start := time.Now().Unix()
	for x := range time.NewTicker(3 * time.Second).C {
		if x.Unix()-start > 30 {
			break
		}
		transactionStatusResponse, err := QueryTxStatus(e2eConfig, rpcResponse.Result.ID)
		if err != nil {
			return nil, fmt.Errorf("soroban rpc tx sub, unable to call tx status check, %v, %e", rpcResponse, err)
		}

		if transactionStatusResponse.Status == "success" {
			return transactionStatusResponse, nil
		}
		if transactionStatusResponse.Status != "pending" {
			return nil, fmt.Errorf("soroban rpc tx sub, got bad response on tx status check, %v, %v, %e", rpcResponse, transactionStatusResponse, err)
		}
	}

	return nil, fmt.Errorf("soroban rpc tx sub, timeout after 30 seconds on tx status check, %v", rpcResponse)
}

func getEnv(key string) (string, error) {
	if value, ok := os.LookupEnv(key); ok {
		return value, nil
	}
	return "", fmt.Errorf("missing required env variable %s", key)
}
