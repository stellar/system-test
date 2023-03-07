package dapp_develop

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/go-cmd/cmd"
	"github.com/pkg/errors"

	e2e "github.com/stellar/system-test"
)

const script = `
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
			.addOperation(contract.call("{{js .functionName}}", xdr.ScVal.scvSymbol("{{js .param1}}")))
			.setTimeout(30)
			.build();

		// TODO: This is a workaround for
		// https://github.com/stellar/js-soroban-client/pull/57, which should be
		// fixed in the next release.
		// Once that is fixed, we could simply do:
		// txn = await server.prepareTransaction(txn, "{{js .networkPassphrase}}");
		const sim = await server.simulateTransaction(txn, "{{js .networkPassphrase}}");
		const { footprint } = sim.results[0];
		txn = SorobanClient.assembleTransaction(txn, "{{js .networkPassphrase}}", [{auth: [], footprint}]);

		txn.sign(SorobanClient.Keypair.fromSecret("{{js .secretKey}}"));
		let response = await server.sendTransaction(txn);
		let i = 0;
		while (response.status === "pending") {
			i += 1;
			if (i > 50) {
				throw new Error("Transaction timed out");
			}
			await new Promise(resolve => setTimeout(resolve, 100));
			response = await server.getTransactionStatus(response.id);
			switch (response.status) {
			case "pending": {
				// noop
				break;
			}
			case "success": {
				// parse and print the response (assuming it is a vec)
				// TODO: Move this scval serializing stuff to stellar-base
				const result = xdr.TransactionResult.fromXDR(response.resultXdr, "base64");
				const scval = result.result().results()[0].tr().invokeHostFunctionResult().success();
				const vec = scval.obj().vec().map(v => v.sym().toString());
				console.log(JSON.stringify(vec));
				return;
			}
			case "error": {
				throw new Error("Transaction failed:", response);
			}
			default:
				throw new Error("Unknown transaction status: " + response.status);
			}
		}
		throw new Error("Transaction failed:", response);
	})();
`

// return the fn response as a serialized string
// uses secret-key and network-passphrase directly on command
func invokeContractFromNodeJSTool(deployedContractId, contractName, functionName, param1 string, e2eConfig *e2e.E2EConfig) (string, error) {
	stdin := &bytes.Buffer{}
	tmpl, err := template.New("script").Parse(script)
	if err != nil {
		return "", errors.Wrap(err, "failed to parse javascript template")
	}
	err = tmpl.Execute(stdin, map[string]string{
		"contractId":        deployedContractId,
		"rpcUrl":            e2eConfig.TargetNetworkRPCURL,
		"account":           e2eConfig.TargetNetworkPublicKey,
		"secretKey":         e2eConfig.TargetNetworkSecretKey,
		"networkPassphrase": e2eConfig.TargetNetworkPassPhrase,
		"functionName":      functionName,
		"param1":            param1,
	})
	if err != nil {
		return "", errors.Wrap(err, "failed to execute javascript template")
	}

	envCmd := cmd.NewCmd("node")
	status, stdOut, err := e2e.RunCommandWithStdin(envCmd, e2eConfig, stdin)

	if status != 0 || err != nil {
		return "", fmt.Errorf("nodejs invoke of example contract %s had error %v, %v", contractName, status, err)
	}

	if len(stdOut) < 1 {
		return "", fmt.Errorf("nodejs invoke of example contract %s did not print any response", contractName)
	}

	return stdOut[0], nil
}

// invokes the contract using identities and network from prior setup of config state in cli
func invokeContractFromNodeJSToolWithConfig(deployedContractId string, contractName string, functionName string, parameters string, identity string, networkConfig string, e2eConfig *e2e.E2EConfig) (string, error) {
	return "", fmt.Errorf("invoke with named identity not supported for NODEJS tool")
}
