#!/usr/bin/env ts-node-script

import { ArgumentParser } from 'argparse';
import * as SorobanClient from 'soroban-client';
const xdr = SorobanClient.xdr;

async function main() {
  const parser = new ArgumentParser({ description: 'Invoke a contract function' })

  const subparsers = parser.add_subparsers();
  parser.add_argument('--id', { dest: 'contractId', required: true, help: 'Contract ID' });
  parser.add_argument('--rpc-url', { dest: 'rpcUrl', required: true, help: 'RPC URL' });
  parser.add_argument('--source', { dest: 'source', required: true, help: 'Secret key' });
  parser.add_argument('--network-passphrase', { dest: 'networkPassphrase', required: true, help: 'Network passphrase' });
  const functionParamParser = subparsers.add_parser('function', { help: 'Function' });
  functionParamParser.add_argument('--name', { dest: 'functionName', help: 'Function Name' });
  functionParamParser.add_argument('--params', { dest: 'functionParams', help: 'Function Params, comma separated' })

  const {
    contractId,
    rpcUrl,
    functionParams,
    networkPassphrase,
    source,
    functionName,
  } = parser.parse_args() as Record<string, string>;

  const contract = new SorobanClient.Contract(contractId);
  const server = new SorobanClient.Server(rpcUrl, { allowHttp: true });
  const secretKey = SorobanClient.Keypair.fromSecret(source);
  const account = secretKey.publicKey();
  const sourceAccount = await server.getAccount(account);

  // Some hacky param-parsing as csv. Generated Typescript bindings would be better.
  const params: SorobanClient.xdr.ScVal[] = [];
  if (functionParams) {
    functionParams.split(",").forEach((param) => {
        params.push(xdr.ScVal.scvSymbol(param));
    });
  }
 
  const txn = await server.prepareTransaction(
    new SorobanClient.TransactionBuilder(sourceAccount, {
      fee: "100",
      networkPassphrase,
    })
    .addOperation(contract.call(functionName, ...params))
    .setTimeout(30)
    .build(),
    networkPassphrase
  );

  txn.sign(secretKey);
  const send = await server.sendTransaction(txn);
  if (send.errorResultXdr) {
    throw new Error(`Transaction failed: ${JSON.stringify(send)}`);
  }
  let response = await server.getTransaction(send.hash);
  for (let i = 0; i < 50; i++) {
    switch (response.status) {
    case "NOT_FOUND": {
      // retry
      await new Promise(resolve => setTimeout(resolve, 100));
      response = await server.getTransaction(send.hash);
      break;
    }
    case "SUCCESS": {
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
      case xdr.ScValType.scvVec(): {
        // Total hack, we just assume the object is a vec. Good enough for now.
        parsed = scval.vec()!.map(v => v.sym().toString());
        break;
      }
      default:
        throw new Error(`Unexpected scval type: ${scval.switch().name}`);
      }
      console.log(JSON.stringify(parsed));
      return;
    }
    case "FAILED": {
      throw new Error(`Transaction failed: ${JSON.stringify(response)}`);
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
