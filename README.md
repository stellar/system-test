# System Test

### How to build the testing docker image 
`docker build --platform linux/amd64 --no-cache -t stellar/system-test -f Dockerfile .`


### How to run tests from the docker image

In short term, the docker image is only supported on hosts that are on x86/amd cpu platforms. Arm cpu platforms are not supported for docker usage yet, this includes any Apple M1. If you are in the latter group, then can still run the tests but will need to refer to running tests from repo instead.

Running docker image requires the following command line args to be passed, example values used, change to your settings:
```
docker run --platform linux/amd64 --rm -it --name e2e_test stellar/system-test:latest \
--SorobanCLICrateVersion 0.2.1 \
--SorobanExamplesGitHash "fc5ef23277f4c032aa7102e4787b406ac2b2f6cd" \
--SorobanExamplesRepoURL "https://github.com/stellar/soroban-examples.git" \
--RustToolchainVersion "stable" \
--CoreDebianVersion "19.5.1-1111.eba1d3de9.focal~soroban" \
--HorizonDebianVersion "2.22.0~soroban-304" \
--SorobanRPCDebianVersion "0.0.1~alpha-2" \
--TestFilter "<filter>" \
--VerboseOutput "true|false" \
```
<filter> is an optional regex of the feature test name and a scenario defined within, example:

"^TestDappDevelop$/^DApp developer compiles, deploys and invokes a contract.*$"

the ending wildcard allows for all combonations of example data for a scenario outline, without that it would just run the first example data set in a scenario outline.

### How to run tests from the locally checked out repo

#### pre-requirements:

 1. go 1.18 or above - https://go.dev/doc/install
 2. rust toolchain(cargo and rustc), latest stable, - use rustup - https://www.rust-lang.org/tools/install 
 3. in a terminal, run the target stack versions to test against in a container:
     ```
     docker run --platform linux/amd64 --rm -it --name e2e_test -p "8000:8000" stellar/system-test:latest --CoreDebianVersion "19.5.1-1111.eba1d3de9.focal~soroban" --HorizonDebianVersion "2.22.0~soroban-304" --SorobanRPCDebianVersion "0.0.1~alpha-2" --RunTargetStackOnly true
     ```
 4. locally checkout e2e test repo and go into top folder - `git clone https://github.com/stellar/system-test.git;cd system-test`

#### running the test (and a test/scenario filter as example)
```
system-test $ SorobanCLICrateVersion=0.2.1 SorobanExamplesGitHash="fc5ef23277f4c032aa7102e4787b406ac2b2f6cd" SorobanExamplesRepoURL="https://github.com/stellar/soroban-examples.git" TargetNetworkPassPhrase="Standalone Network ; February 2017" TargetNetworkSecretKey="SC5O7VZUXDJ6JBDSZ74DSERXL7W3Y5LTOAMRF7RQRL3TAGAPS7LUVG3L" TargetNetworkPublicKey="GBZXN7PIRZGNMHGA7MUUUF4GWPY5AYPV6LY4UV2GL6VJGIQRXFDNMADI" TargetNetworkRPCURL="http://localhost:8000/soroban/rpc" VerboseOutput=true go test -v --run "^TestDappDevelop$/^DApp developer compiles, deploys and invokes a contract.*$" ./...
```

Note: The test will install soroban cli onto the o/s at the version specified in SorobanCLICrateVersion, to make it skip this step, no install, and default to using whatever you have on path, do not specify SorobanCLICrateVersion.

