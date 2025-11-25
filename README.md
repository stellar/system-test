# System Test

### Running system tests:

Identify the quickstart image you want to use as target for running tests. Quickstart contains the server stack of core, rpc:

- Use a prebuilt system test image published as tags under
  `dockerhub.io/stellar/quickstart`
- Or build the system test docker image locally with specific versions of cli and stellar-js-sdk,
  this will create a docker image named `stellar/system-test:dev`. All `GIT_REF` variables can refer to
  either a fully qualified local path to checked out git repo, or a fully
  qualified github remote repo url `https://github.com/repo#<ref>`

```
make
     STELLAR_CLI_GIT_REF=? \
     RUST_TOOLCHAIN_VERSION=? \
     STELLAR_CLI_CRATE_VERSION=? \
     JS_STELLAR_SDK_NPM_VERSION=? \
     NODE_VERSION=? \
     build
```

all settings have defaults pre-set, and optionally be overriden, refer to the Makefile for the defaulted values.

#### Optional Build Params

```
# cli will be compiled from this git ref
STELLAR_CLI_GIT_REF=https://github.com/stellar/stellar-cli.git#main

# this will override STELLAR_CLI_GIT_REF, and install stellar cli from crates repo instead
STELLAR_CLI_CRATE_VERSION=0.4.0

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

# this will skip building stellar-cli from STELLAR_CLI_GIT_REF and instead
# will use the bin already compiled at /usr/local/cargo/bin/soroban in the existing docker image provided:
STELLAR_CLI_IMAGE=<docker registry>/<docker image name>:<docker tag>
```

#### Required runtime params

Set the target network to test. This just needs to be the RPC server of the network. It can be from a running quickstart instance or testnet or pubnet:
  `--TargetNetworkRPCURL {http://<rpc_host:rpc_port>/soroban/rpc}`

Note - If you are running quickstart as a docker container on the same host machine, then specify the rpc url as `--TargetNetworkRPCURL http://host.docker.internal:8000/rpc` to use Docker's convention to reference host network. 

#### Optional runtime params

System test will by default use the network settings for `local` network from quickstart. 
If `TargetNetworkRPCURL` is pointed at any stellar network other than a `local` network instance provided from quickstart, then you'll need to provide the network specifics. 
`--TargetNetworkPassphrase "{passphrase}"`
`--TargetNetworkTestAccountSecret "{your test account key pair info}"`
`--TargetNetworkTestAccountPublic "{your test account key pair info}"`

To specify git version of the smart contract source code used in soroban test
fixtures. `--SorobanExamplesGitHash {branch, tag, git commit hash}`
`--SorobanExamplesRepoURL "https://github.com/stellar/soroban-examples.git"`

To specify which system test feature/scenarios to run, it is a regex of the
feature test name and a scenario defined within, each row in example data for a
scenario outline is postfixed with '#01', '#02', examples:
`--TestFilter "^TestDappDevelop$/^DApp developer compiles, deploys and invokes a contract.*$"`
or
`--TestFilter "^TestDappDevelop$/^DApp developer compiles, deploys and invokes a contract#01$"`

Set verbose logging output `--VerboseOutput true` 

#### Running Tests

- Run tests against a remote instance of rpc hosted on a quickstart configured for testnet. 
  The tests requires you provide a key pair of an account that is funded with Lumens on the target
  network for the tests to use as source account on transactions it will submit
  to target network:

  ```
  docker run --rm -t --name e2e_test stellar/system-test:<tag> \
  --VerboseOutput true \
  --TargetNetworkRPCURL https://<rpc host url> \
  --TargetNetworkPassphrase "Test SDF Network ; September 2015" \
  --TargetNetworkTestAccountSecret <your test account key pair info> \
  --TargetNetworkTestAccountPublic <your test account key pair info> \
  --SorobanExamplesGitHash v22.0.1
  ```

#### Debug test failures
Use `--VerboseOutput true` and may need to check the lops of the rpc server instance if you have access to those at same time.

The docker container will exit with error code when any pre-setup or
test fails to pass, you can enable DEBUG_MODE flag, and the container will stay
running, prompting you for depressing enter key before shutting down, make sure you invoke
docker with `-it` so the prompt will reach your command line. While container is
kept running, you can shell into it via `docker exec -it <container id or name>`
and manually re-run tests in container, check local outputs. 

The docker run follows standard exit code conventions, so if all tests pass in
the container run, exit code from command line execution will be 0, otherwise,
if any failures in container or tests, then exit code will be greater than 0.

### Development mode and running tests directly from checked out system-test repo.

This approach allows to run the tests from source code directly on host as go
tests, no docker image is used.

#### Prerequisites:

1.  go 1.18 or above - https://go.dev/doc/install
2.  rust toolchain(cargo and rustc), install the version per testing
    requirements or stable, - use rustup -
    https://www.rust-lang.org/tools/install
3.  `stellar` cli, compile or install via cargo crate a version of stellar cli
    onto your machine and accessible from PATH.
4.  run an instance of RPC locally by running quickstart on a local network such as `stellar/quickstart:latest`.
    ```
    docker run --rm -it -p 8000:8000 --name stellar stellar/quickstart:latest --local 
    ```
5.  locally checkout stellar/system-test 
    `git clone https://github.com/stellar/system-test.git;cd system-test`

#### Running tests locally as go programs

```
system-test $ SorobanExamplesGitHash="main" \
SorobanExamplesRepoURL="https://github.com/stellar/soroban-examples.git" \
TargetNetworkPassPhrase="Standalone Network ; February 2017" \
TargetNetworkSecretKey="SC5O7VZUXDJ6JBDSZ74DSERXL7W3Y5LTOAMRF7RQRL3TAGAPS7LUVG3L" \
TargetNetworkPublicKey="GBZXN7PIRZGNMHGA7MUUUF4GWPY5AYPV6LY4UV2GL6VJGIQRXFDNMADI" \
TargetNetworkRPCURL="http://localhost:8000/rpc" \
VerboseOutput=false \
go test -v --run "^TestDappDevelop$/^DApp developer compiles, deploys and invokes a contract.*$" ./features/dapp_develop/...
```

This follows standard go test conventions, so if all tests pass, exit code from
command line execution will be 0, otherwise, if any tests fail, then exit code
will be greater than 0.

This example uses a feature/scenario filter also to limit which tests are run.

- Tests will attempt to run `stellar` as the cli as provided from your operating
  system PATH.

- the verbose output of BDD scenerio results for tests is dependent on go's
  testing verbose output rules, need to specify -v and a directory with single
  package, if multiple packages detected on directory location, then go won't
  print verbose output for each package, i.e. you wont see the BDD scenerio
  summaries printed, just the standard one liner for summary of package
  pass/fail status.

#### Debugging tests

A debug config [launch.json](.vscode/launch.json) is provided for example
reference on how to run a test with the go/dlv debugger.
