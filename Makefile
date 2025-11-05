SHELL:=/bin/bash
MAKEFILE_DIR:=$(dir $(abspath $(lastword $(MAKEFILE_LIST))))

.EXPORT_ALL_VARIABLES:
.PHONY: build 

SYSTEM_TEST_SHA=$(shell git rev-parse HEAD)
STELLAR_CLI_STAGE_IMAGE=stellar/system-test-stellar-cli:dev

# The rest of these variables can be set as environment variables to the makefile
# to modify how system test is built.

# variables to set for source code, can be any valid docker context url local
# path github remote repo `https://github.com/repo#<ref>`
STELLAR_CLI_GIT_REF=https://github.com/stellar/stellar-cli.git\#main

# if crate version is set, then it overrides STELLAR_CLI_GIT_REF, cli will be installed from this create instead
STELLAR_CLI_CRATE_VERSION=

# variables to set if wanting to use existing dockerhub images instead of compiling
# image during build. if using this option, the image ref should provide a version for same
# platform arch as the build host is on, i.e. linux/amd64 or linux/arm64.
#
# image must have soroban cli bin at /usr/local/cargo/bin/soroban
STELLAR_CLI_IMAGE=

# defauult node version to install in system test image
NODE_VERSION?=20.19.4

# specify the published npm repo version of soroban-client js library,
# or you can specify gh git ref url as the version
JS_STELLAR_SDK_NPM_VERSION=https://github.com/stellar/js-stellar-sdk.git\#master

# sets the rustc version in the system test image
RUST_TOOLCHAIN_VERSION=stable

# the final image name that is created in local docker images store for system test
SYSTEM_TEST_IMAGE=stellar/system-test:dev

# set to true to enable github actions cache for layer caching during build
USE_GHA_CACHE=false

# github actions cache arguments for docker build
GHA_CACHE_ARGS=--cache-from type=gha,scope=system-test-layer-cache --cache-to type=gha,scope=system-test-layer-cache,mode=max,compression=zstd

# set cache args based on USE_GHA_CACHE flag
ifeq ($(USE_GHA_CACHE),true)
	CACHE_ARGS=$(GHA_CACHE_ARGS)
else
	CACHE_ARGS=
endif

build-stellar-cli:
	if [ -z "$(STELLAR_CLI_IMAGE)" ]; then \
		DOCKERHUB_RUST_VERSION=rust:$$( [ "$(RUST_TOOLCHAIN_VERSION)" = "stable" ] && echo "latest" || echo "$(RUST_TOOLCHAIN_VERSION)"); \
		docker buildx build --progress=plain --load -t "$(STELLAR_CLI_STAGE_IMAGE)" --target builder \
		--build-arg BUILDKIT_CONTEXT_KEEP_GIT_DIR=true \
		--build-arg DOCKERHUB_RUST_VERSION="$$DOCKERHUB_RUST_VERSION" \
		--build-arg STELLAR_CLI_CRATE_VERSION="$(STELLAR_CLI_CRATE_VERSION)" \
		$(CACHE_ARGS) \
		-f- $(STELLAR_CLI_GIT_REF) < $(MAKEFILE_DIR)Dockerfile.stellar-cli; \
	fi

build: build-stellar-cli
	STELLAR_CLI_IMAGE_REF=$$( [ -z "$(STELLAR_CLI_IMAGE)" ] && echo "$(STELLAR_CLI_STAGE_IMAGE)" || echo "$(STELLAR_CLI_IMAGE)"); \
	docker buildx build --progress=plain --load -t "$(SYSTEM_TEST_IMAGE)" -f Dockerfile \
	    --build-arg BUILDKIT_CONTEXT_KEEP_GIT_DIR=true \
		--build-arg STELLAR_CLI_CRATE_VERSION=$(STELLAR_CLI_CRATE_VERSION) \
		--build-arg STELLAR_CLI_IMAGE_REF=$$STELLAR_CLI_IMAGE_REF \
		--build-arg RUST_TOOLCHAIN_VERSION=$(RUST_TOOLCHAIN_VERSION) \
		--build-arg NODE_VERSION=$(NODE_VERSION) \
		--build-arg JS_STELLAR_SDK_NPM_VERSION=$(JS_STELLAR_SDK_NPM_VERSION) \
		--label org.opencontainers.image.revision="$(SYSTEM_TEST_SHA)" \
		$(CACHE_ARGS) .;
