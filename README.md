# System Test

### Runing tests using the docker image:
To run tests, requires two steps: 
  (1) First build the system test docker image with the correct versions of core,  
  horizon, soroban rpc, rust toolchain, soroban cli, this will create a docker image named 
  `stellar/system-test:dev`:
  ```
  make 
       QUICKSTART_GIT_REF=? \
       QUICKSTART_GIT_REPO=?
       CORE_GIT_REF=? \
       CORE_COMPILE_CONFIGURE_FLAGS=? \
       SOROBAN_RPC_GIT_REF=? \
       RUST_TOOLCHAIN_VERSION=? \
       SOROBAN_CLI_CRATE_VERSION=? \
       SOROBAN_CLI_GIT_REF=? build     
  ```  

  example of build using specific git branches, latest in this case, or use tag names for releases:  
  ```
  $ make CORE_GIT_REF=f1dc39f0f146815e5e3a94ed162e2f0639cb433f \
         CORE_COMPILE_CONFIGURE_FLAGS="--disable-tests --enable-next-protocol-version-unsafe-for-production" \
         SOROBAN_RPC_GIT_REF=main \
         RUST_TOOLCHAIN_VERSION=stable \
         SOROBAN_CLI_GIT_REF=main \
         QUICKSTART_GIT_REF=master build
  ```  

  example of build using the existing quickstart:soroban-dev image which has latest released versions:  
  ```
  $ make QUICKSTART_IMAGE=stellar/quickstart:soroban-dev \
         RUST_TOOLCHAIN_VERSION=1.66.0 \
         SOROBAN_CLI_GIT_REF=main build
  ```  


  some settings have defaults pre-set, and optionally be overriden:  
  ```
  SOROBAN_CLI_GIT_REF=main  
  SOROBAN_RPC_GIT_REF=main  
  RUST_TOOLCHAIN_VERSION=stable   
  QUICKSTART_GIT_REF=master
  QUICKSTART_GIT_REPO=https://github.com/stellar/quickstart.git
  ```  

  some are required to be set:
  ```
  CORE_GIT_REF=latest git commit for core with soroban support
  CORE_COMPILE_CONFIGURE_FLAGS="--disable-tests --enable-next-protocol-version-unsafe-for-production"
  ```  

  optional to set:  
  ```
  # this will override SOROBAN_CLI_GIT_REF, and install soroban cli from crates repo instead
  SOROBAN_CLI_CRATE_VERSION=0.4.0  

  # this will skip building core, horion, rpc from source and instead will use the versions already compiled in the existing quickstart image provided: 
  QUICKSTART_IMAGE=stellar/quickstart:soroban-dev
  ```

  (2) Run the system test docker image:
  ```
  docker run --rm -it --name e2e_test stellar/system-test:dev --VerboseOutput false 
  ```


Optional settings to pass when running system-test image, `stellar/system-test:<tag>`:

To specify git version of the smart contract source code used as test fixtures.  
`--SorobanExamplesGitHash {branch, tag, git commit hash}`  
`--SorobanExamplesRepoURL "https://github.com/stellar/soroban-examples.git"` 

To specify which system test feature/scenarios to run, it is a regex of the feature test name and a scenario defined within, each row in example data for a scenario outline is postfixed with '#01', '#02', examples:  
`--TestFilter "^TestDappDevelop$/^DApp developer compiles, deploys and invokes a contract.*$"`  
or  
`--TestFilter "^TestDappDevelop$/^DApp developer compiles, deploys and invokes a contract#01$"`  

The ending wildcard allows for all combinations of example data for a scenario outline, without that it would just run the first example data set in a scenario outline.

The default target network for system tests is a new/empty instance of standalone network hosted inside the docker container, tests will use the default root account already seeded into standalone network. Alternatively, can override the network settings here:  
* Tests will use an internally hosted core node connected to standalone or futurenet network:  
`--TargetNetwork {standalone|futurenet}`  
* Tests will use an external rpc instance and the container will not run core, horizon, rpc services internally:  
`--TargetNetworkRPCURL {http://<rpc_host:rpc_port>/soroban/rpc}`  
* Tests use these settings in either target network mode, and these are by default set to work with standalone:  
`--TargetNetworkPassphrase "{passphrase}"`  
`--TargetNetworkTestAccountSecret "{your test account key pair info}"`  
`--TargetNetworkTestAccountPublic "{your test account key pair info}"`  

Debug mode, the docker container will exit with error code when any pre-setup or test fails to pass,
you can enable DEBUG_MODE flag, and the container will stay running, prompting you for enter key before shutting down, make sure you invoke docker with `-it` so the prompt will reach your command line. While container is kept running, you can shell into it via `docker exec -it <container id or name>` and view log files of core, rpc, horizon all of which are located in container at `/var/log/supervisor`.  
`--DebugMode=true`


The docker run follows standard exit code conventions, so if all tests pass in the container run, exit code from command line execution will be 0, otherwise, if any failures in container or tests, then exit code will be greater than 0.


### Development mode and running tests directly from checked out system-test repo.
This approach allows to run the tests from source code directly on host as go tests, no docker image.  
Tests will use `soroban` cli tool from the host path. You need to set `TargetNetworkRPCURL` to a running instance of soroban rpc.

#### Prerequisites:

 1. go 1.18 or above - https://go.dev/doc/install
 2. rust toolchain(cargo and rustc), install the version per testing requirements or stable, - use rustup - https://www.rust-lang.org/tools/install 
 3. target network stack for the tests to execute against - need a soroban-rpc instance. You can use an existing/running instance if reachable or can use the quickstart image `stellar/quickstart:soroban-dev` from dockerhub to run the latest stable target network stack locally, or build quickstart with specific versions of core, horizon and soroban rpc first [following these instructions](https://github.com/stellar/quickstart#building-custom-images) and run `stellar/quickstart:dev` locally.
     ```
     docker run --rm -it -p 8000:8000 --name stellar stellar/quickstart:dev --standalone --enable-soroban-rpc --enable-core-artificially-accelerate-time-for-testing
     ```
 4. locally checkout stellar/system-test GH repo and go into top folder - `git clone https://github.com/stellar/system-test.git;cd system-test`
 5. compile or install via cargo crate a version of soroban cli onto your machine and accessible from PATH.

#### Running tests 
```
# example values used here are for when running quickstart:soroban-dev standalone locally
system-test $ SorobanExamplesGitHash="main" \  
SorobanExamplesRepoURL="https://github.com/stellar/soroban-examples.git" \  
TargetNetworkPassPhrase="Standalone Network ; February 2017" \  
TargetNetworkSecretKey="SC5O7VZUXDJ6JBDSZ74DSERXL7W3Y5LTOAMRF7RQRL3TAGAPS7LUVG3L" \  
TargetNetworkPublicKey="GBZXN7PIRZGNMHGA7MUUUF4GWPY5AYPV6LY4UV2GL6VJGIQRXFDNMADI" \  
TargetNetworkRPCURL="http://localhost:8000/soroban/rpc" \  
VerboseOutput=false \  
go test -v --run "^TestDappDevelop$/^DApp developer compiles, deploys and invokes a contract.*$" ./features/dapp_develop/...
```

This follows standard go test conventions, so if all tests pass, exit code from command line execution will be 0, otherwise, if any tests fail, then exit code will be greater than 0.

This example uses a feature/scenario filter also to limit which tests are run.

* Tests will attempt to run `soroban` as the cli as provided from your operating system PATH.

* the verbose output of BDD scenerio results for tests is dependent on go's testing verbose output rules, need to specify -v and a directory with single package, if multiple packages detected on directory location, then go won't print verbose output for each package, i.e. you wont see the BDD scenerio summaries printed, just the standard one liner for summary of package pass/fail status.

