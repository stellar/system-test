# System Test

### Running system tests:
  Identify the system-test image you want to use for running tests:
  - Use a prebuilt system test image published as tags under `dockerhub.io/stellar/system-test`
  - Build the system test docker image locally with specific versions of core, horizon, soroban rpc, rust toolchain, soroban cli, this will create a docker image named
  `stellar/system-test:dev`.
  All `GIT_REF` variables can refer to either a fully qualified local path to checked out git repo, or a fully qualified github remote repo url `https://github.com/repo#<ref>`
  ```
  make
       QUICKSTART_GIT_REF=? \
       CORE_GIT_REF=? \
       CORE_COMPILE_CONFIGURE_FLAGS=? \
       STELLAR_RPC_GIT_REF=? \
       SOROBAN_CLI_GIT_REF=? \
       GO_GIT_REF=? \
       RUST_TOOLCHAIN_VERSION=? \
       SOROBAN_CLI_CRATE_VERSION=? \
       JS_STELLAR_SDK_NPM_VERSION=? \
       NODE_VERSION=? \
       PROTOCOL_VERSION_DEFAULT=? \
       build
  ```

  example of build using specific git refs, mainline from repos in this example, or use tags, branches, etc:
  ```
  make CORE_GIT_REF=https://github.com/stellar/stellar-core.git#f1dc39f0f146815e5e3a94ed162e2f0639cb433f \
         CORE_COMPILE_CONFIGURE_FLAGS="--disable-tests --enable-next-protocol-version-unsafe-for-production" \
         STELLAR_RPC_GIT_REF=https://github.com/stellar/soroban-tools.git#main \
         RUST_TOOLCHAIN_VERSION=stable \
         SOROBAN_CLI_GIT_REF=https://github.com/stellar/soroban-tools.git#main \
         QUICKSTART_GIT_REF=https://github.com/stellar/quickstart.git#master \
         JS_STELLAR_SDK_NPM_VERSION=https://github.com/stellar/js-stellar-sdk.git#master \
         build
  ```

  example of build using an existing quickstart image, this can dramatically speed up the build time, as the existing quickstart image will provide the pre-compiled rpc, and core runtimes already:
  ```
  make QUICKSTART_IMAGE=stellar/quickstart:soroban-dev \
         RUST_TOOLCHAIN_VERSION=1.66.0 \
         SOROBAN_CLI_GIT_REF=/Users/user/soroban-tools build
  ```

  some settings have defaults pre-set, and optionally be overriden:
  ```
  SOROBAN_CLI_GIT_REF=https://github.com/stellar/soroban-tools.git#main
  STELLAR_RPC_GIT_REF=https://github.com/stellar/soroban-tools.git#main
  RUST_TOOLCHAIN_VERSION=stable
  QUICKSTART_GIT_REF=https://github.com/stellar/quickstart.git#master
  # the GO_GIT_REF provides the reference on the stellar/go repo from which
  # to build horizon
  GO_GIT_REF=https://github.com/stellar/go.git#master
  CORE_COMPILE_CONFIGURE_FLAGS="--disable-tests"
  CORE_GIT_REF=https://github.com/stellar/stellar-core.git#master
  JS_STELLAR_SDK_NPM_VERSION=https://github.com/stellar/js-stellar-sdk.git#master
  ```

  optional params to set:
  ```
  # this will override SOROBAN_CLI_GIT_REF, and install soroban cli from crates repo instead
  SOROBAN_CLI_CRATE_VERSION=0.4.0

  # this will override the default Node JS vm version used for running the JS code:
  NODE_VERSION=16.20.2

  # js sdk version can be set to a published npm version from https://www.npmjs.com/package/stellar-sdk
  JS_STELLAR_SDK_NPM_VERSION=latest
  # or it can be set to a github git ref of a js-stellar-sdk repo
  JS_STELLAR_SDK_NPM_VERSION=https://github.com/stellar/js-stellar-sdk.git#master

  # Image overrides.
  # If using these, the image ref should provide a manifiest version for same
  # platform arch as the build host is running on, i.e. linux/amd64 or linux/arm64.
  # Otherwise, build will fail if image is not available for matching host platform.
  #
  # this will skip building from source for core(CORE_GIT_REF), rpc(STELLAR_RPC_GIT_REF) and quickstart(QUICKSTART_GIT_REF), instead
  # will use the versions already compiled in the existing quickstart docker image provided:
  QUICKSTART_IMAGE=<docker registry>/<docker image name>:<docker tag>

  # this will skip building core from CORE_GIT_REF and instead
  # will use the `stellar-core` by default at /usr/local/bin/stellar-core in the existing docker image provided:
  CORE_IMAGE=<docker registry>/<docker image name>:<docker tag>

  # define a custom path that `stellar-core` bin is located on CORE_IMAGE, 
  # to override the default of /usr/local/bin/stellar-core
  CORE_IMAGE_BIN_PATH=

  # this will skip building stellar-rpc from STELLAR_RPC_GIT_REF and instead
  # will use the bin already compiled at /bin/stellar-rpc in the existing docker image provided:
  STELLAR_RPC_IMAGE=<docker registry>/<docker image name>:<docker tag>

  # this will skip building soroban-cli from SOROBAN_CLI_GIT_REF and instead
  # will use the bin already compiled at /usr/local/cargo/bin/soroban in the existing docker image provided:
  SOROBAN_CLI_IMAGE=<docker registry>/<docker image name>:<docker tag>

  # this will skip building horizon from GO_GIT_REF and instead
  # will use the bin already compiled at /go/bin/horizon in the existing docker image provided:
  HORIZON_IMAGE=<docker registry>/<docker image name>:<docker tag>

  # this will skip building friendbot from GO_GIT_REF and instead
  # will use the bin already compiled at /app/friendbot in the existing docker image provided:
  FRIENDBOT_IMAGE=<docker registry>/<docker image name>:<docker tag>

  # set the default network protocol version which the internal core runtime built from `CORE_GIT_REF` should start with. 
  # Should typically be set to the maximum supported protocol version of all components.
  # If not set or set to empty, will default to the core max supported protocol version defined in quickstart.
  PROTOCOL_VERSION_DEFAULT=
  ```

Optional parameters to pass when running the system-test image, `stellar/system-test:<tag>`:

To specify git version of the smart contract source code used in soroban test fixtures.
`--SorobanExamplesGitHash {branch, tag, git commit hash}`
`--SorobanExamplesRepoURL "https://github.com/stellar/soroban-examples.git"`

To specify which system test feature/scenarios to run, it is a regex of the feature test name and a scenario defined within, each row in example data for a scenario outline is postfixed with '#01', '#02', examples:
`--TestFilter "^TestDappDevelop$/^DApp developer compiles, deploys and invokes a contract.*$"`
or
`--TestFilter "^TestDappDevelop$/^DApp developer compiles, deploys and invokes a contract#01$"`

The default target network for system tests is a new/empty instance of local network hosted inside the docker container, tests will use the default root account already seeded into local network. Alternatively, can override the network settings for local and remote usages:
- Tests will use an internally hosted core watcher node:
`--TargetNetwork {standalone|futurenet|testnet}`
- Tests will use an external rpc instance and the container will not run core, horizon, rpc services internally:
`--TargetNetworkRPCURL {http://<rpc_host:rpc_port>/soroban/rpc}`
- Tests use these settings in either target network mode, and these are by default set to work with local:
`--TargetNetworkPassphrase "{passphrase}"`
`--TargetNetworkTestAccountSecret "{your test account key pair info}"`
`--TargetNetworkTestAccountPublic "{your test account key pair info}"`

Debug mode, the docker container will exit with error code when any pre-setup or test fails to pass,
you can enable DEBUG_MODE flag, and the container will stay running, prompting you for enter key before shutting down, make sure you invoke docker with `-it` so the prompt will reach your command line. While container is kept running, you can shell into it via `docker exec -it <container id or name>` and view log files of services in the stack such as core, rpc located in container at `/var/log/supervisor`.
`--DebugMode=true`

The docker run follows standard exit code conventions, so if all tests pass in the container run, exit code from command line execution will be 0, otherwise, if any failures in container or tests, then exit code will be greater than 0.

#### Running Test Examples
- Run tests against an instance of core and rpc on a local network all running in the test container:


  ```
  docker run --rm -t --name e2e_test stellar/system-test:<tag> \
  --VerboseOutput true
  ```

- Run tests against a remote instance of rpc configured for testnet, this will not run core or rpc instances locally in the test container. It requires you provide a key pair of an account that is funded with Lumens on the target network for the tests to use as source account on transactions it will submit to target network:


  ```
  docker run --rm -t --name e2e_test stellar/system-test:<tag> \
  --VerboseOutput true \
  --TargetNetworkRPCURL https://<rpc host url> \
  --TargetNetworkPassphrase "Test SDF Network ; September 2015" \
  --TargetNetworkTestAccountSecret <your test account key pair info> \
  --TargetNetworkTestAccountPublic <your test account key pair info> \
  --SorobanExamplesGitHash v20.0.0-rc2
  ```


### Development mode and running tests directly from checked out system-test repo.
This approach allows to run the tests from source code directly on host as go tests, no docker image is used.

#### Prerequisites:

 1. go 1.18 or above - https://go.dev/doc/install
 2. rust toolchain(cargo and rustc), install the version per testing requirements or stable, - use rustup - https://www.rust-lang.org/tools/install
 3. `soroban` cli, compile or install via cargo crate a version of soroban cli onto your machine and accessible from PATH.
 4. target network stack for the tests to access stellar-rpc instance. You can use an existing/running instance if reachable or can use the quickstart image `stellar/quickstart:soroban-dev` from dockerhub to run the latest stable target network stack locally, or build quickstart with specific versions of core, horizon and soroban rpc first [following these instructions](https://github.com/stellar/quickstart#building-custom-images) and run `stellar/quickstart:dev` locally.
     ```
     docker run --rm -it -p 8000:8000 --name stellar stellar/quickstart:dev --standalone --enable rpc
     ```
 5. locally checkout stellar/system-test GH repo and go into top folder - `git clone https://github.com/stellar/system-test.git;cd system-test`


#### Running tests locally as go programs
```
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

#### Debugging tests

A debug config [launch.json](.vscode/launch.json) is provided for example reference on how to run a test with the go/dlv debugger.
