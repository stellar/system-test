#!/usr/bin/env ts-node-script

import * as fs from 'fs';
import { ArgumentParser } from 'argparse';
import {
  Keypair,
  TransactionBuilder,
  SorobanRpc,
  scValToNative,
  xdr,
  Operation,
  OperationOptions,
} from '@stellar/stellar-sdk';

const { Server } = SorobanRpc;

async function main() {
  const parser = new ArgumentParser({ description: 'Install a contract' })

  parser.add_argument('--wasm', { dest: 'wasm', required: true, help: 'Path to wasm binary' });
  parser.add_argument('--rpc-url', { dest: 'rpcUrl', required: true, help: 'RPC URL' });
  parser.add_argument('--source', { dest: 'source', required: true, help: 'Secret key' });
  parser.add_argument('--network-passphrase', { dest: 'networkPassphrase', required: true, help: 'Network passphrase' });

  const {
    wasm,
    rpcUrl,
    networkPassphrase,
    source,
  } = parser.parse_args() as Record<string, string>;


  const server = new Server(rpcUrl, { allowHttp: true });
  const secretKey = Keypair.fromSecret(source);
  const account = secretKey.publicKey();
  const sourceAccount = await server.getAccount(account);
  const wasmBuffer = fs.readFileSync(wasm);
  

  const options: OperationOptions.InvokeHostFunction = {
    "func": xdr.HostFunction.hostFunctionTypeUploadContractWasm(Buffer.from(wasmBuffer)),
    "source": account
  };
  const op = Operation.invokeHostFunction(options);

  const originalTxn = new TransactionBuilder(sourceAccount, {
      fee: "100",
      networkPassphrase
    })
    .addOperation(op)
    .setTimeout(30)
    .build();

  const txn = await server.prepareTransaction(originalTxn);
  txn.sign(secretKey);
  const send = await server.sendTransaction(txn);
  if (send.errorResult) {
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
      if (!response.returnValue) {
        throw new Error(`No invoke host fn return value provided: ${JSON.stringify(response)}`);
      }

      const parsed = scValToNative(response.returnValue);
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
