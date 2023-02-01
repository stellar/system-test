ARG QUICKSTART_IMAGE_REF=stellar/quickstart:soroban-dev

FROM golang:1.19 as go

RUN ["mkdir", "-p", "/test"] 
RUN ["mkdir", "-p", "/test/bin"] 

WORKDIR /test
ADD go.mod go.sum e2e.go ./
ADD features ./features
RUN go mod download

# build each feature folder with go test module.
# compiles each feature to a binary to be executed, 
# and copies the .feature file with it for runtime.
RUN go test -c -o ./bin/dapp_develop_test ./features/dapp_develop/...
ADD features/dapp_develop/dapp_develop.feature ./bin

FROM $QUICKSTART_IMAGE_REF as base
ARG SOROBAN_CLI_CRATE_VERSION
ARG SOROBAN_CLI_GIT_REF
ARG RUST_TOOLCHAIN_VERSION
 
RUN ["mkdir", "-p", "/rust"] 
ENV CARGO_HOME=/rust/.cargo
ENV RUSTUP_HOME=/rust/.rust
ENV SOROBAN_CLI_CRATE_VERSION=$SOROBAN_CLI_CRATE_VERSION
ENV SOROBAN_CLI_GIT_REF=$SOROBAN_CLI_GIT_REF
ENV RUST_TOOLCHAIN_VERSION=$RUST_TOOLCHAIN_VERSION
ENV PATH="/usr/local/go/bin:$CARGO_HOME/bin:${PATH}"

ADD install /
RUN ["chmod", "+x", "install"]
RUN /install 

FROM base as build
RUN ["mkdir", "-p", "/opt/test"] 
ADD start /opt/test
COPY --from=go /test/bin/ /opt/test/bin

RUN ["chmod", "+x", "/opt/test/start"]

ENTRYPOINT ["/opt/test/start"]
