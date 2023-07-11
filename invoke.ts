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
  const params: xdr.ScVal[] = [];
  if (functionParams) {
    functionParams.split(",").forEach((param) => {
        params.push(xdr.ScVal.scvSymbol(param));
    });
  }

  const originalTxn = new SorobanClient.TransactionBuilder(sourceAccount, {
      fee: "100",
      networkPassphrase,
    })
    .addOperation(contract.call(functionName, ...params))
    .setTimeout(30)
    .build();

  const txn = await server.prepareTransaction(originalTxn,networkPassphrase);
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
      const result = xdr.TransactionResultMeta.fromXDR(response.resultMetaXdr!, "base64");
      const scval = result.txApplyProcessing().v3().sorobanMeta()?.returnValue()!;

      const parsed = SorobanClient.scValToNative(scval);
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
