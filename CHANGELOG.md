# System Test Releases

#### 1.0.19

* js verification of contract invoke results, compare as strings. [system-test, #68](https://github.com/stellar/system-test/pull/68)

#### 1.0.18

* verify js contract invoke return value as native arrays. [system-test, #67](https://github.com/stellar/system-test/pull/67)

#### 1.0.15

* hex-ify the contract id for getEvents. [system-test, #61](https://github.com/stellar/system-test/pull/61)

#### 1.0.14

* removed usage of accelerate flag on quickstart options, it's been deprecated [system-test, #60](https://github.com/stellar/system-test/pull/60)

#### 1.0.13

* tests use new output paths compiled wasm in soroban-examples. [system-test, #58](https://github.com/stellar/system-test/pull/58)
* replaces instances of getLedgerEntry with getLedgerEntries. [system-test, #54](https://github.com/stellar/system-test/pull/54)


#### 1.0.12

* removed installation of phantomjs in the runtime image, it wasn't used and was triggering build errors due to no debian package available on 22.04. [system-test, #55](https://github.com/stellar/system-test/pull/55)


#### 1.0.11

* updated dapp scenario for verifying diagnostic events retrieval from cli and js. [system-test, #49](https://github.com/stellar/system-test/pull/49)


#### 1.0.10

* verify no value overhaul issues on test coverage. [system-test, #47](https://github.com/stellar/system-test/pull/47)


#### 1.0.9

* Fix bug in NODEJS test invocation. [system-test, #46](https://github.com/stellar/system-test/pull/46) 


#### 1.0.8

* Enable some NODEJS tests. [system-test, #44](https://github.com/stellar/system-test/pull/44) and [system-test, #42](https://github.com/stellar/system-test/pull/42)
* Use new `--source` flag for cli. [system-test, #43](https://github.com/stellar/system-test/pull/43)


#### 1.0.7

* Reorg the dockerfile for better cache-ability. [system-test, #31](https://github.com/stellar/system-test/pull/31).
* soroban-rpc: do not wait for Horizon since it does no longer depend on it. [system-test, #36](https://github.com/stellar/system-test/pull/36).
* local source path for git refs and/or image overrides during build. [system-test, #37](https://github.com/stellar/system-test/pull/37).
* Update system tests for new contract invoke. [system-test, #38](https://github.com/stellar/system-test/pull/38).
* Added Auth Next Scenario Test. [system-test, #34](https://github.com/stellar/system-test/pull/34).


#### 1.0.6

* Remove usage of soroban-rpc method getAccount, because it is deprecated, and will be removed in the next release. Use getLedgerEntry instead. [system-test, #30](https://github.com/stellar/system-test/pull/30).


#### 1.0.5

* Fixed `--TargetNetwork futurenet`, was incorrectly trying to configure artificial acceleration on core config also, which is only allowed on `standalone`. [system-test, #25](https://github.com/stellar/system-test/pull/25).  

This version of tests is based on [Soroban Preview 7](https://soroban.stellar.org/docs/releases) system interfaces. 

#### 1.0.4

* Fixed git ref resolution to detached head state when compiling soroban cli during image build [system-test, #23](https://github.com/stellar/system-test/pull/23), to build [Stellar Quickstart](https://github.com/stellar/quickstart).  

This version of tests is based on [Soroban Preview 7](https://soroban.stellar.org/docs/releases) system interfaces. 

#### 1.0.3

* Modified test framework, [system-test, #21](https://github.com/stellar/system-test/pull/21), to build [Stellar Quickstart](https://github.com/stellar/quickstart) as the base image.  
Refer to [README.md](https://github.com/stellar/system-test#readme) for new two step process of running tests:  
(1) run make with server versions, creates docker image  
(2) run the docker image to run tests.  


This version of tests is based on [Soroban Preview 7](https://soroban.stellar.org/docs/releases) system interfaces. 

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

