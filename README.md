# System Test

### How to build the testing docker image 
`docker build --platform linux/amd64 --no-cache -t stellar/system-test -f Dockerfile .`


### How to run tests from the docker image

In short term, the docker image is only supported on hosts that are on x86/amd cpu platforms. Arm cpu platforms are not supported for docker usage yet, this includes any Apple M1 device. If you are in the latter group, then can still run the tests but will need to refer to running tests directly from checked out repo instead.

Running docker image, the following command line args can be passed, example values used, change to your settings:
```
docker run --platform linux/amd64 --rm -it --name e2e_test stellar/stellar-system-test:latest \
--SorobanCLICrateVersion 0.2.1 \
--SorobanExamplesGitHash master \
--SorobanExamplesRepoURL "https://github.com/stellar/soroban-examples.git" \
--RustToolchainVersion "stable" \
--CoreDebianVersion "19.5.1-1111.eba1d3de9.focal~soroban" \
--HorizonDebianVersion "2.22.0~soroban-304" \
--SorobanRPCDebianVersion "0.0.1~alpha-2" \
--TestFilter "<filter>" \
--VerboseOutput false \
```

* TestFilter is optional, regex of the feature test name and a scenario defined within, each row in example data for a scenario outline is postfixed with '#01', '#02', example:
"^TestDappDevelop$/^DApp developer compiles, deploys and invokes a contract.*$"
"^TestDappDevelop$/^DApp developer compiles, deploys and invokes a contract#01$"

The ending wildcard allows for all combonations of example data for a scenario outline, without that it would just run the first example data set in a scenario outline.
* the target network under test in this usage is an internal instance of standalone network launched from the quickstart soroban-dev image. This references static account key pair which was seeded in the quickstart standalone network. At some point, may allow this to reference external network services if requirements come up, maybe pass friendbot url in that case, etc, if doing that external targets, then CoreDebianVersion, HorizonDebianVersion, SorobanRPCDebianVersion would be ignored. 


### How to run tests from the locally checked out repo.

#### pre-requirements:

 1. go 1.18 or above - https://go.dev/doc/install
 2. rust toolchain(cargo and rustc), install the version per testing requirements or stable, - use rustup - https://www.rust-lang.org/tools/install 
 3. in a terminal, run the target network stack to test against using the stellar/stellar-system-test docker image:
     ```
     docker run --platform linux/amd64 --rm -it --name e2e_test -p "8000:8000" stellar/stellar-system-test:latest --CoreDebianVersion "19.5.1-1111.eba1d3de9.focal~soroban" --HorizonDebianVersion "2.22.0~soroban-304" --SorobanRPCDebianVersion "0.0.1~alpha-2" --RunTargetStackOnly true
     ```
 4. locally checkout stellar/system-test GH repo and go into top folder - `git clone https://github.com/stellar/system-test.git;cd system-test`

#### running the test (and a test/scenario filter as example)
```
system-test $ SorobanCLICrateVersion=0.2.1 SorobanExamplesGitHash="master" SorobanExamplesRepoURL="https://github.com/stellar/soroban-examples.git" TargetNetworkPassPhrase="Standalone Network ; February 2017" TargetNetworkSecretKey="SC5O7VZUXDJ6JBDSZ74DSERXL7W3Y5LTOAMRF7RQRL3TAGAPS7LUVG3L" TargetNetworkPublicKey="GBZXN7PIRZGNMHGA7MUUUF4GWPY5AYPV6LY4UV2GL6VJGIQRXFDNMADI" TargetNetworkRPCURL="http://localhost:8000/soroban/rpc" VerboseOutput=true go test -v --run "^TestDappDevelop$/^DApp developer compiles, deploys and invokes a contract.*$" ./features/dapp_develop/...
```

* SorobanCLICrateVersion is optional, if not defined, test will attempt to run soroban as provided from your operating system PATH, i.e. you install soroban cli manually on your machine first. Otherwise, the test will install this soroban cli version onto the o/s.

* the color coded output of scenerio results for tests is dependent on go's test verbose output rules, need to specify -v and a directory with single package, if multiple packages detected on directory location, then go won't print verbose output for each package.

