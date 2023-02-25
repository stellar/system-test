ARG QUICKSTART_IMAGE_REF=stellar/quickstart:soroban-dev
ARG SOROBAN_CLI_IMAGE_REF=stellar/system-test-soroban-cli:dev

FROM golang:1.19 as go

RUN ["mkdir", "-p", "/test"] 
RUN ["mkdir", "-p", "/test/bin"] 

WORKDIR /test
ADD go.mod go.sum ./
RUN go mod download
ADD e2e.go ./
ADD features ./features

# build each feature folder with go test module.
# compiles each feature to a binary to be executed, 
# and copies the .feature file with it for runtime.
RUN go test -c -o ./bin/dapp_develop_test ./features/dapp_develop/...
ADD features/dapp_develop/dapp_develop.feature ./bin

FROM $SOROBAN_CLI_IMAGE_REF as soroban-cli

FROM $QUICKSTART_IMAGE_REF as base
ARG RUST_TOOLCHAIN_VERSION

ENV DEBIAN_FRONTEND=noninteractive
RUN apt-get update && apt-get install -y build-essential && apt-get clean

# Install Rust
RUN ["mkdir", "-p", "/rust"] 
ENV CARGO_HOME=/rust/.cargo
ENV RUSTUP_HOME=/rust/.rust
ENV RUST_TOOLCHAIN_VERSION=$RUST_TOOLCHAIN_VERSION
ENV PATH="/usr/local/go/bin:$CARGO_HOME/bin:${PATH}"
RUN curl https://sh.rustup.rs -sSf | sh -s -- -y --default-toolchain "$RUST_TOOLCHAIN_VERSION"

# Install soroban-cli
COPY --from=soroban-cli /usr/local/cargo/bin/soroban $CARGO_HOME/bin/

FROM base as build
RUN ["mkdir", "-p", "/opt/test"] 
ADD start /opt/test
COPY --from=go /test/bin/ /opt/test/bin

RUN ["chmod", "+x", "/opt/test/start"]

ENTRYPOINT ["/opt/test/start"]
