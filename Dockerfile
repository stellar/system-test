ARG QUICKSTART_IMAGE_REF=stellar/quickstart:soroban-dev
ARG STELLAR_CLI_IMAGE_REF=stellar/system-test-soroban-cli:dev

FROM golang:1.24 AS go

SHELL ["/bin/bash", "-c"]
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
RUN go test -c -o ./bin/dapp_develop_test.bin ./features/dapp_develop/...
ADD features/dapp_develop/dapp_develop.feature ./bin
# copy over a dapp develop test specific file, used for expect/tty usage
ADD features/dapp_develop/soroban_config.exp ./bin

FROM $STELLAR_CLI_IMAGE_REF AS stellar-cli
FROM $QUICKSTART_IMAGE_REF AS base

ARG RUST_TOOLCHAIN_VERSION
ARG NODE_VERSION

ENV DEBIAN_FRONTEND=noninteractive
RUN apt-get update && apt-get install -y build-essential expect git libdbus-1-dev libudev-dev && apt-get clean

# Install Rust
RUN ["mkdir", "-p", "/rust"]
ENV CARGO_HOME=/rust/.cargo
ENV RUSTUP_HOME=/rust/.rust
ENV RUST_TOOLCHAIN_VERSION=$RUST_TOOLCHAIN_VERSION
ENV PATH="/usr/local/go/bin:$CARGO_HOME/bin:${PATH}"
RUN curl https://sh.rustup.rs -sSf | sh -s -- -y --default-toolchain "$RUST_TOOLCHAIN_VERSION"
RUN rustup show active-toolchain || rustup toolchain install
# Older toolchain to compile soroban examples
RUN rustup toolchain install 1.81-x86_64-unknown-linux-gnu
# Wasm toolchain to compile contracts
RUN rustup target add wasm32v1-none

# Use a non-root user
ARG USERNAME=tester
ARG USER_UID=1000
ARG USER_GID=$USER_UID
RUN groupadd --gid $USER_GID $USERNAME \
    && useradd --uid $USER_UID --gid $USER_GID -m $USERNAME \
    && apt-get update \
    && apt-get install -y sudo \
    && echo $USERNAME ALL=\(root\) NOPASSWD:ALL > /etc/sudoers.d/$USERNAME \
    && chmod 0440 /etc/sudoers.d/$USERNAME
RUN ["mkdir", "-p", "/home/tester"]
USER tester
WORKDIR /home/tester
RUN mkdir -p ~/.ssh
RUN chmod 700 ~/.ssh
RUN echo "HOST *" > ~/.ssh/config
RUN echo "StrictHostKeyChecking no" >> ~/.ssh/config

# Install Node.js
RUN curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.3/install.sh | bash
ENV NVM_DIR=/home/tester/.nvm
RUN . "$NVM_DIR/nvm.sh" && \
  nvm install ${NODE_VERSION} && \
  nvm alias default v${NODE_VERSION} && \
  nvm use default
ENV PATH "$NVM_DIR/versions/node/v${NODE_VERSION}/bin/:$PATH"
RUN echo $PATH; node --version; npm --version
RUN npm install -g ts-node yarn

# Install js-stellar-sdk
ARG JS_STELLAR_SDK_NPM_VERSION
ADD package.json /home/tester/
ADD js-stellar-sdk /home/tester/js-stellar-sdk
RUN sudo chown -R tester:tester /home/tester
RUN yarn cache clean && yarn install --network-concurrency 1
RUN if echo "$JS_STELLAR_SDK_NPM_VERSION" | grep -q '.*file:.*'; then \
    cd /home/tester/js-stellar-sdk; \
    yarn cache clean; \
    yarn install --network-concurrency 1; \
    cd /home/tester; \
    yarn add ${JS_STELLAR_SDK_NPM_VERSION} --network-concurrency 1; \
  else \
    yarn add "@stellar/stellar-sdk@${JS_STELLAR_SDK_NPM_VERSION}" --network-concurrency 1; \
  fi

ADD *.ts /home/tester/bin/
RUN ["sudo", "chmod", "+x", "/home/tester/bin/invoke.ts"]

FROM base AS build

# Tests expect to be run as root so they can launch stuff
USER root

ADD start /home/tester
COPY --from=stellar-cli /usr/local/cargo/bin/stellar $CARGO_HOME/bin/
COPY --from=go /test/bin/ /home/tester/bin

ENTRYPOINT ["/home/tester/start"]
