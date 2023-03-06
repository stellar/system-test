Feature: DApp Contract Development

Scenario Outline: DApp developer compiles, installs, deploys and invokes a contract
Given I used cargo to compile example contract <ContractExampleSubPath>
  And I used rpc to verify my account is on the network
  And I used cli to install contract <ContractCompiledFileName> on network using my secret key
  And I used cli to deploy contract <ContractCompiledFileName> by installed hash using my secret key
  When I invoke function <FunctionName> on <ContractName> with request parameter <Param1> from <Tool> using my secret key
  Then the result should be <Result>

  Examples: 
        | Tool         | ContractExampleSubPath | ContractName                  | ContractCompiledFileName             | FunctionName | Param1         | Result             |
#       | NODEJS       | hello_world            | soroban-hello-world-contract  | soroban_hello_world_contract.wasm    | hello        | --to=Aloha     | ["Hello","Aloha"]  |
        | CLI          | hello_world            | soroban-hello-world-contract  | soroban_hello_world_contract.wasm    | hello        | --to=Aloha     | ["Hello","Aloha"]  |
#       | NODEJS       | increment              | soroban-increment-contract    | soroban_increment_contract.wasm      | increment    |                | 1                  | 
        | CLI          | increment              | soroban-increment-contract    | soroban_increment_contract.wasm      | increment    |                | 1                  |


Scenario Outline: DApp developer compiles, deploys and invokes a contract
  Given I used cargo to compile example contract <ContractExampleSubPath>
  And I used rpc to verify my account is on the network
  And I used cli to deploy contract <ContractCompiledFileName> using my secret key
  When I invoke function <FunctionName> on <ContractName> with request parameter <Param1> from <Tool> using my secret key
  Then the result should be <Result>

  Examples: 
        | Tool         | ContractExampleSubPath | ContractName                  | ContractCompiledFileName             | FunctionName | Param1     | Result             |
#       | NODEJS       | hello_world            | soroban-hello-world-contract  | soroban_hello_world_contract.wasm    | hello        | --to=Aloha | ["Hello","Aloha"]  |
        | CLI          | hello_world            | soroban-hello-world-contract  | soroban_hello_world_contract.wasm    | hello        | --to=Aloha | ["Hello","Aloha"]  |
#       | NODEJS       | increment              | soroban-increment-contract    | soroban_increment_contract.wasm      | increment    |            | 1                  |
        | CLI          | increment              | soroban-increment-contract    | soroban_increment_contract.wasm      | increment    |            | 1                  |


Scenario Outline: DApp developer uses config states, compiles, deploys and invokes contract with authorizations
  Given I used cargo to compile example contract <ContractExampleSubPath> 
  And I used rpc to verify my account is on the network
  And I used rpc to submit transaction to create tester account on the network
  And I used cli to add Network Config <NetworkConfigName> for rpc and standalone
  And I used cli to add Identity <RootIdentityName> for my secret key
  And I used cli to add Identity <TesterIdentityName> for tester secret key
  And I used cli to deploy contract <ContractCompiledFileName> using my Identity <RootIdentityName> and Network Config <NetworkConfigName>
  When I invoke function <FunctionName> on <ContractName> with request parameters <FunctionParams> from <Tool> using tester Identity <TesterIdentityName> as invoker and Network Config <NetworkConfigName>
  Then the result should be <Result>

  Examples: 
        | Tool         | ContractExampleSubPath | ContractName                  | ContractCompiledFileName      | FunctionName     | FunctionParams                              | RootIdentityName  | TesterIdentityName  | NetworkConfigName   | Result |
#       | NODEJS       | auth                   | soroban-auth-contract         | soroban_auth_contract.wasm    | increment        | --user <tester_identity_pub_key> --value 2  | r1                | t1                  | standalone          | 2      |
        | CLI          | auth                   | soroban-auth-contract         | soroban_auth_contract.wasm    | increment        | --user <tester_identity_pub_key> --value 2  | r1                | t1                  | standalone          | 2      |
