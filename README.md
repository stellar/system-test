# System Test

### How to build the testing docker image
`docker build --platform linux/amd64 --no-cache -t stellar/system-test -f Dockerfile .`

### Run tests from the docker image:
In short term, running tests on the `stellar/stellar-system-test` docker image is only supported on hosts that are on x86/amd cpu platforms. Arm cpu platforms are not supported for docker usage yet, this includes any Apple M1 device. If you are in the latter group, then can still run the tests but will need to refer to running tests directly from checked out repo instead.

Running docker image, this is an example of command line args, please change these to settings that make sense for you expectations:
```
docker run --platform linux/amd64 --rm -it --name e2e_test stellar/system-test:latest \
--SorobanCLICrateVersion 0.2.1 \
--CoreDebianVersion "19.5.1-1137.b3a6bc281.focal~soroban" \
--HorizonDebianVersion "2.22.0~soroban-318" \
--SorobanRPCDebianVersion "0.3.1-32" \
--VerboseOutput false 
```

Optional settings:

To compile soroban tools components from local source on host machine instead of package versions. Usage of either of these two requires adding a docker volume mount to your local cloned copy of soroban-tools git repo `-v /full/path/to/soroban-tools:/soroban-tools`

to compile local copy of cli source code, include this parameter, if empty it uses `--SorobanCLICrateVersion` instead. 
`--SorobanCLISourceVolume=/soroban-tools`

to compile local copy of rpc source code, set this paramter, if empty it uses `--SorobanRPCDebianVersion` instead.
`--SorobanRPCSourceVolume=/soroban-tools`

To specify git version of the smart contract source code used as test fixtures. 
`--SorobanExamplesGitHash {branch, tag, git commit hash}` 
`--SorobanExamplesRepoURL "https://github.com/stellar/soroban-examples.git"` 

Default rust toolchain version in system test image was 1.65.0, to override and install different:
`--RustToolchainVersion {version}`

To specify which system test feature/scenarios to run, it is a regex of the feature test name and a scenario defined within, each row in example data for a scenario outline is postfixed with '#01', '#02', examples:
`--TestFilter "^TestDappDevelop$/^DApp developer compiles, deploys and invokes a contract.*$"`
or
`--TestFilter "^TestDappDevelop$/^DApp developer compiles, deploys and invokes a contract#01$"`

The ending wildcard allows for all combinations of example data for a scenario outline, without that it would just run the first example data set in a scenario outline.

Default target network for system tests is new/empty instance of standalone network, and tests will use the default root account already seeded into standalone network, can override here.
`--TargetNetwork {standalone|futurenet}`
`--TargetNetworkPassphrase "{passphrase}"`
`--TargetNetworkTestAccountSecret "{your test account key pair info}"`
`--TargetNetworkTestAccountPublic "{your test account key pair info}"`

Debug mode, the docker container will exit with error code when any pre-setup or test fails to pass,
you can enable DEBUG_MODE flag, and the container will stay running, prompting you for enter key before shutting down, make sure you invoke docker with `-it` so the prompt will reach your command line. While container is kept running, you can shell into it and view log files of core, rpc, horizon all of which are located in container at `/var/log/supervisor`.
`--DebugMode=true`


The docker run follows standard exit code conventions, so if all tests pass in the container run, exit code from command line execution will be 0, otherwise, if any failures in container or tests, then exit code will be greater than 0.


### Run tests from locally checked out repo.
This approach allows to run the tests directly on host as go tests. It allows to configure more aspects directly, like target network to use, and whether to try to use pre-existing cli on the host if desired but does require more environment setup.

#### Prerequisites:

 1. go 1.18 or above - https://go.dev/doc/install
 2. rust toolchain(cargo and rustc), install the version per testing requirements or stable, - use rustup - https://www.rust-lang.org/tools/install 
 3. target network stack for the tests - need a soroban-rpc instance connected to horizon and core. This will usually be a standalone instance of the network for testing purposes. You can reference an existing network or can use docker image `stellar/stellar-system-test` with `--RunTargetStackOnly true` to spin up just the target network stack, specifying the versions of each component to launch:
     ```
     docker run --platform linux/amd64 --rm -it --name e2e_test -p "8000:8000" stellar/system-test:latest --CoreDebianVersion "19.5.1-1111.eba1d3de9.focal~soroban" --HorizonDebianVersion "2.22.0~soroban-304" --SorobanRPCDebianVersion "0.0.1~alpha-2" --RunTargetStackOnly true
     ```
 4. locally checkout stellar/system-test GH repo and go into top folder - `git clone https://github.com/stellar/system-test.git;cd system-test`

#### Running tests (and a test/scenario filter as example)
```
system-test $ SorobanCLICrateVersion=0.2.1 \
 SorobanExamplesGitHash="main" \
 SorobanExamplesRepoURL="https://github.com/stellar/soroban-examples.git" \
 TargetNetworkPassPhrase="Standalone Network ; February 2017" \
 TargetNetworkSecretKey="SC5O7VZUXDJ6JBDSZ74DSERXL7W3Y5LTOAMRF7RQRL3TAGAPS7LUVG3L" \
 TargetNetworkPublicKey="GBZXN7PIRZGNMHGA7MUUUF4GWPY5AYPV6LY4UV2GL6VJGIQRXFDNMADI" \
 TargetNetworkRPCURL="http://localhost:8000/soroban/rpc" \
 VerboseOutput=false \
 go test -v --run "^TestDappDevelop$/^DApp developer compiles, deploys and invokes a contract.*$" ./features/dapp_develop/...
```

This follows standard go test conventions, so if all tests pass, exit code from command line execution will be 0, otherwise, if any tests fail, then exit code will be greater than 0.

* SorobanCLICrateVersion is optional, if not defined, test will attempt to run soroban as provided from your operating system PATH, i.e. you install soroban cli manually on your machine first. Otherwise, the test will install this soroban cli version onto the o/s.

* the color coded output of BDD scenerio results for tests is dependent on go's testing verbose output rules, need to specify -v and a directory with single package, if multiple packages detected on directory location, then go won't print verbose output for each package, i.e. you wont see the BDD scenerio summaries printed, just the standard one liner for summary of package pass/fail status.

