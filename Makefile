SHELL := /bin/bash
.EXPORT_ALL_VARIABLES:

SYSTEM_TEST_IMAGE?=stellar/system-test:dev
SYSTEM_TEST_SHA=$(shell git rev-parse HEAD)
ENABLE_PUSH?=false

CORE_GIT_REF=
CORE_COMPILE_CONFIGURE_FLAGS?=--disable-tests --enable-next-protocol-version-unsafe-for-production
SOROBAN_RPC_GIT_REF=main

RUST_TOOLCHAIN_VERSION?=stable
SOROBAN_CLI_CRATE_VERSION=
SOROBAN_CLI_GIT_REF=main
# TODO: remove go ref when quickstart /scripts/soroban_repo_to_horizon_repo.sh resolves pr refs.
GO_GIT_REF?=soroban-xdr-next

QUICKSTART_IMAGE=
QUICKSTART_GIT_REF=master
QUICKSTART_GIT_REPO?=https://github.com/stellar/quickstart.git

.PHONY: build build-base

build-base: 	
	if [ -z "$(QUICKSTART_IMAGE)" ]; then \
	  set -e ;\
	  rm -rf .quickstart_repo; \
      git clone -q $(QUICKSTART_GIT_REPO) .quickstart_repo; \
      pushd .quickstart_repo; \
      git fetch origin "$(QUICKSTART_GIT_REF)" && git checkout FETCH_HEAD; \
      $(MAKE) CORE_REF=$(CORE_GIT_REF) \
      	   GO_REF=$(GO_GIT_REF) \
      	   CORE_CONFIGURE_FLAGS="$(CORE_COMPILE_CONFIGURE_FLAGS)" \
           SOROBAN_TOOLS_REF=$(SOROBAN_RPC_GIT_REF); \
    fi

build: build-base
	QUICKSTART_IMAGE=$$( [ -z "$(QUICKSTART_IMAGE)" ] && echo "stellar/quickstart:dev" || echo "$(QUICKSTART_IMAGE)"); \
    docker build -t "$(SYSTEM_TEST_IMAGE)" \
      -f Dockerfile \
      --build-arg QUICKSTART_IMAGE_REF="$$QUICKSTART_IMAGE" \
      --build-arg SOROBAN_CLI_CRATE_VERSION="$(SOROBAN_CLI_CRATE_VERSION)" \
      --build-arg SOROBAN_CLI_GIT_REF=$(SOROBAN_CLI_GIT_REF) \
      --build-arg RUST_TOOLCHAIN_VERSION=$(RUST_TOOLCHAIN_VERSION) \
      --label org.opencontainers.image.revision="$(SYSTEM_TEST_SHA)" .
	