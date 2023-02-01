# System Test Releases



#### 1.0.2

* Modified tests to follow the new dynamic args format on cli [soroban-tools, #307](https://github.com/stellar/soroban-tools/pull/307)

This version of tests is based on [Soroban Preview 6](https://soroban.stellar.org/docs/releases#preview-6-january-9th-2023) system interfaces, combined with the additional change applied on top of dynamic args in cli `contract invoke` 


#### 1.0.1

* Modified tests to follow the new `contract` sub-command on cli [soroban-tools, #319](https://github.com/stellar/soroban-tools/pull/319)

This version of tests executes the [Soroban Preview 6](https://soroban.stellar.org/docs/releases#preview-6-january-9th-2023) system interfaces only.


#### 1.0.0
First release of packaged system tests. Initial focus is on Soroban e2e cases using cli,rpc,core:

* DApp developer compiles, installs, deploys and invokes a contract
* DApp developer compiles, deploys and invokes a contract

This version of tests execute the [Soroban Preview 5](https://soroban.stellar.org/docs/releases#preview-5-december-8th-2022) system interfaces only.

