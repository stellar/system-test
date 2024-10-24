#!/usr/bin/env ts-node-script

import { ArgumentParser } from 'argparse';
import { Contract, rpc } from '@stellar/stellar-sdk';

async function main() {
  const parser = new ArgumentParser({ description: 'Get contract events' })
  parser.add_argument('--id', { dest: 'contractId', required: true, help: 'Contract ID' });
  parser.add_argument('--rpc-url', { dest: 'rpcUrl', required: true, help: 'RPC URL' });
  parser.add_argument('--size', { dest: 'size', required: true, help: 'Max Number of events to fetch' });
  parser.add_argument('--ledgerFrom', { dest: 'ledgerFrom', help: 'Ledget Start' });

  const {
    contractId,
    rpcUrl,
    size,
    ledgerFrom,
  } = parser.parse_args() as Record<string, string>;

  const server = new rpc.Server(rpcUrl, { allowHttp: true });

  let filters: rpc.Api.EventFilter[] = [];

  if (contractId != null) {
    filters.push({
      contractIds: [ new Contract(contractId).contractId() ]
    });
  }

  let response = await server.getEvents({
    startLedger: Number(ledgerFrom),
    filters: filters,
    limit: Number(size)
  });

  if (!response.events) {
      throw new Error(`No events in response: ${JSON.stringify(response)}`);
  }

  console.log(JSON.stringify(response.events));
}

main().catch(err => {
  console.error(JSON.stringify(err));
  throw err;
});
