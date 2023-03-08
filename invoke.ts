import { ArgumentParser } from 'argparse';
import * as SorobanClient from 'soroban-client';
const xdr = SorobanClient.xdr;

async function main() {
  const parser = new ArgumentParser({ description: 'Invoke a contract function' })

  parser.add_argument('--id', { dest: 'contractId', required: true, help: 'Contract ID' });
  parser.add_argument('--rpc-url', { dest: 'rpcUrl', required: true, help: 'RPC URL' });
  parser.add_argument('--account', { dest: 'account', required: true, help: 'Account ID' });
  parser.add_argument('--secret-key', { dest: 'secretKey', required: true, help: 'Secret key' });
  parser.add_argument('--network-passphrase', { dest: 'networkPassphrase', required: true, help: 'Network passphrase' });
  parser.add_argument('--fn', { dest: 'functionName', required: true, help: 'Function name' });
  parser.add_argument('--param1', { dest: 'param1', help: 'Param 1' });

  const {
    contractId,
    rpcUrl,
    account,
    param1,
    networkPassphrase,
    secretKey,
    functionName,
  } = parser.parse_args() as Record<string, string>;

  const contract = new SorobanClient.Contract(contractId);
  const server = new SorobanClient.Server(rpcUrl, { allowHttp: true });
  const source = await server.getAccount(account)

  // Some hacky param-parsing. Generated Typescript bindings would be better
  // here. But those don't exist yet as I'm writing this.
  const params = param1 ? [xdr.ScVal.scvSymbol(param1)] : [];

  let txn = new SorobanClient.TransactionBuilder(source, {
      fee: "100",
      networkPassphrase,
    })
    .addOperation(contract.call(functionName, ...params))
    .setTimeout(30)
    .build();

  // TODO: This is a workaround for
  // https://github.com/stellar/js-soroban-client/pull/57, which should be
  // fixed in the next release.
  // Once that is fixed, we could simply do:
  // txn = await server.prepareTransaction(txn, networkPassphrase);
  const sim = await server.simulateTransaction(txn);
  if (!sim.results || sim.results.length !== 1) {
    throw new Error(`Simulation failed: ${JSON.stringify(sim)}`);
  }
  const { footprint } = sim.results[0];
  txn = SorobanClient.assembleTransaction(txn, networkPassphrase, [{auth: [], footprint}]);

  txn.sign(SorobanClient.Keypair.fromSecret(secretKey));
  const send = await server.sendTransaction(txn);
  if (send.error) {
    throw new Error(`Transaction failed: ${send.error}`);
  }
  let response = await server.getTransactionStatus(send.id);
  for (let i = 0; i < 50; i++) {
    switch (response.status) {
    case "pending": {
      // retry
      await new Promise(resolve => setTimeout(resolve, 100));
      response = await server.getTransactionStatus(response.id);
      break;
    }
    case "success": {
      // parse and print the response (assuming it is a vec)
      // TODO: Move this scval serializing stuff to stellar-base
      if (!response.resultXdr) {
          throw new Error(`No result XDR: ${JSON.stringify(response)}`);
      }
      const result = xdr.TransactionResult.fromXDR(response.resultXdr, "base64");
      const scval = result.result().results()[0].tr().invokeHostFunctionResult().success();

      // Hacky result parsing. We should have some helpers from the
      // js-stellar-base, or the generated Typescript bindings. But we don't yet as
      // I'm writing this.
      let parsed = null;
      switch (scval.switch()) {
      case xdr.ScValType.scvU32(): {
        parsed = scval.u32();
        break;
      }
      case xdr.ScValType.scvI32(): {
        parsed = scval.i32();
        break;
      }
      case xdr.ScValType.scvObject(): {
        // Total hack, we just assume the object is a vec. Good enough for now.
        parsed = scval.obj()!.vec()!.map(v => v.sym().toString());
        break;
      }
      default:
        throw new Error(`Unexpected scval type: ${scval.switch().name}`);
      }
      console.log(JSON.stringify(parsed));
      return;
    }
    case "error": {
      throw new Error(`Transaction failed: ${response}`);
    }
    default:
      throw new Error(`Unknown transaction status: ${response.status}`);
    }
  }
  throw new Error("Transaction timed out");
}

main().catch(err => {
  console.error(JSON.stringify(err));
  throw err;
});
