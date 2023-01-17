Feature: DApp Contract Development

Scenario Outline: DApp developer compiles, installs, deploys and invokes a contract
Given I used cli to compile example contract <ContractExampleSubPath>
  And I used rpc to verify my account is on the network
  And I used cli to install contract <ContractCompiledFileName> on ledger using my account to network
  And I used cli to deploy contract <ContractCompiledFileName> by installed hash using my account to network
  When I invoke function <FunctionName> on <ContractName> with request parameter <Param1> from <Tool>
  Then the result should be <Result>

  Examples: 
        | Tool         | ContractExampleSubPath | ContractName                  | ContractCompiledFileName             |FunctionName  | Param1    | Result             |
#       | JSSDK        | hello_world            | soroban-hello-world-contract  | soroban_hello_world_contract.wasm    | hello        | --to=Aloha     | ["Hello","Aloha"]  |
        | CLI          | hello_world            | soroban-hello-world-contract  | soroban_hello_world_contract.wasm    | hello        | --to=Aloha     | ["Hello","Aloha"]  |
#       | JSSDK        | increment              | soroban-increment-contract    | soroban_increment_contract.wasm      | increment    |           | 1                  | 
        | CLI          | increment              | soroban-increment-contract    | soroban_increment_contract.wasm      | increment    |           | 1                  |


Scenario Outline: DApp developer compiles, deploys and invokes a contract
  Given I used cli to compile example contract <ContractExampleSubPath>
  And I used rpc to verify my account is on the network
  And I used cli to deploy contract <ContractCompiledFileName> using my account to network
  When I invoke function <FunctionName> on <ContractName> with request parameter <Param1> from <Tool>
  Then the result should be <Result>

  Examples: 
        | Tool         | ContractExampleSubPath | ContractName                  | ContractCompiledFileName             |FunctionName  | Param1     | Result             |
#       | JSSDK        | hello_world            | soroban-hello-world-contract  | soroban_hello_world_contract.wasm    | hello        | --to=Aloha | ["Hello","Aloha"]  |
        | CLI          | hello_world            | soroban-hello-world-contract  | soroban_hello_world_contract.wasm    | hello        | --to=Aloha | ["Hello","Aloha"]  |
#       | JSSDK        | increment              | soroban-increment-contract    | soroban_increment_contract.wasm      | increment    |            | 1                  | 
        | CLI          | increment              | soroban-increment-contract    | soroban_increment_contract.wasm      | increment    |            | 1                  |
