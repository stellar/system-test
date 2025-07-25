SHELL:=/bin/bash
MAKEFILE_DIR:=$(dir $(abspath $(lastword $(MAKEFILE_LIST))))

.EXPORT_ALL_VARIABLES:
.PHONY: build build-quickstart build-core build-friendbot build-horizon build-stellar-rpc

SYSTEM_TEST_SHA=$(shell git rev-parse HEAD)
QUICKSTART_STAGE_IMAGE=stellar/system-test-base:dev
CORE_STAGE_IMAGE=stellar/system-test-core:dev
HORIZON_STAGE_IMAGE=stellar/system-test-horizon:dev
RS_XDR_STAGE_IMAGE=stellar/system-test-rs-xdr:dev
FRIENDBOT_STAGE_IMAGE=stellar/system-test-friendbot:dev
STELLAR_RPC_STAGE_IMAGE=stellar/system-test-stellar-rpc:dev
STELLAR_CLI_STAGE_IMAGE=stellar/system-test-stellar-cli:dev

# The rest of these variables can be set as environment variables to the makefile
# to modify how system test is built.

# The default protocol version that the image should start with. Should
# typically be set to the maximum supported protocol version of all components.
# If not set will default to the core max supported protocol version in quickstart.
PROTOCOL_VERSION_DEFAULT=23

# variables to set for source code, can be any valid docker context url local path github remote repo `https://github.com/repo#<ref>`
CORE_GIT_REF=https://github.com/stellar/stellar-core.git\#master
STELLAR_RPC_GIT_REF=https://github.com/stellar/stellar-rpc.git\#main
STELLAR_CLI_GIT_REF=https://github.com/stellar/stellar-cli.git\#main
GO_GIT_REF=https://github.com/stellar/go.git\#master
RS_XDR_GIT_REPO=https://github.com/stellar/rs-stellar-xdr
RS_XDR_GIT_REF=main
QUICKSTART_GIT_REF=https://github.com/stellar/quickstart.git\#main
# specify the published npm repo version of soroban-client js library, 
# or you can specify gh git ref url as the version
JS_STELLAR_SDK_NPM_VERSION=https://github.com/stellar/js-stellar-sdk.git\#master

# variables to set if wanting to use existing dockerhub images instead of compiling
# image during build. if using this option, the image ref should provide a version for same
# platform arch as the build host is on, i.e. linux/amd64 or linux/arm64.
#
# image must have soroban cli bin at /usr/local/cargo/bin/soroban
STELLAR_CLI_IMAGE=
#
# image must have soroban rpc bin at /bin/stellar-rpc
STELLAR_RPC_IMAGE=
#
# image must have horizon bin at /go/bin/horizon
HORIZON_IMAGE=
#
# image must have friendbot bin at /app/friendbot
FRIENDBOT_IMAGE=
#
# image must have the bin at /usr/local/cargo/bin/stellar-xdr
RS_XDR_IMAGE=
#
# image with stellar-core binary, assumes core bin at /usr/local/bin/stellar-core
CORE_IMAGE=
# define a custom path that core bin is located on CORE_IMAGE, other than /usr/local/bin/stellar-core
CORE_IMAGE_BIN_PATH=
#
# a prebuilt 'soroban-dev' image from the quickstart repo, if this is supplied,
# the other core, rpc, horizon, friendbot config settings are mostly ignored, since the quickstart image
# has them compiled in already. the 'stellar/quickstart' images also support multi-arch, so the build will
# work those images whether the build host is arm64 or amd64.
QUICKSTART_IMAGE=

NODE_VERSION?=18.19.0

# if crate version is set, then it overrides STELLAR_CLI_GIT_REF, cli will be installed from this create instead
STELLAR_CLI_CRATE_VERSION=

# sets the rustc version in the system test image
RUST_TOOLCHAIN_VERSION=stable

# temporarily needed, builds core with soroban enabled features
CORE_COMPILE_CONFIGURE_FLAGS=--disable-tests

# the final image name that is created in local docker images store for system test
SYSTEM_TEST_IMAGE=stellar/system-test:dev

build-friendbot:
	if [ -z "$(QUICKSTART_IMAGE)" ] && [ -z "$(FRIENDBOT_IMAGE)" ]; then \
		SOURCE_URL="$(GO_GIT_REF)"; \
		if [[ ! "$(GO_GIT_REF)" =~ \.git ]]; then \
			pushd "$(GO_GIT_REF)"; \
			SOURCE_URL=.; \
		fi; \
		docker build -t "$(FRIENDBOT_STAGE_IMAGE)" \
		--build-arg BUILDKIT_CONTEXT_KEEP_GIT_DIR=true \
		-f services/friendbot/docker/Dockerfile "$$SOURCE_URL"; \
	fi

build-rs-xdr: 
	if [ -z "$(QUICKSTART_IMAGE)" ] && [ -z "$(RS_XDR_IMAGE)" ]; then \
		SOURCE_URL="$(QUICKSTART_GIT_REF)"; \
		if [[ ! "$(QUICKSTART_GIT_REF)" =~ \.git ]]; then \
			pushd "$(QUICKSTART_GIT_REF)"; \
			SOURCE_URL=.; \
		fi; \
		docker build -t "$(RS_XDR_STAGE_IMAGE)" --target builder \
		--build-arg BUILDKIT_CONTEXT_KEEP_GIT_DIR=true \
		--build-arg REPO=$$RS_XDR_GIT_REPO \
		--build-arg REF=$$RS_XDR_GIT_REF \
		-f Dockerfile.xdr "$$SOURCE_URL"; \
	fi	

build-stellar-rpc:
	if [ -z "$(QUICKSTART_IMAGE)" ] && [ -z "$(STELLAR_RPC_IMAGE)" ]; then \
		SOURCE_URL="$(STELLAR_RPC_GIT_REF)"; \
		if [[ ! "$(STELLAR_RPC_GIT_REF)" =~ \.git ]]; then \
			pushd "$(STELLAR_RPC_GIT_REF)"; \
			SOURCE_URL=.; \
		fi; \
		docker build -t "$(STELLAR_RPC_STAGE_IMAGE)" --target build \
		--build-arg BUILDKIT_CONTEXT_KEEP_GIT_DIR=true \
		-f cmd/stellar-rpc/docker/Dockerfile "$$SOURCE_URL"; \
	fi

build-stellar-cli:
	if [ -z "$(STELLAR_CLI_IMAGE)" ]; then \
		DOCKERHUB_RUST_VERSION=rust:$$( [ "$(RUST_TOOLCHAIN_VERSION)" = "stable" ] && echo "latest" || echo "$(RUST_TOOLCHAIN_VERSION)"); \
		docker buildx build -t "$(STELLAR_CLI_STAGE_IMAGE)" --target builder \
		--build-arg BUILDKIT_CONTEXT_KEEP_GIT_DIR=true \
		--build-arg DOCKERHUB_RUST_VERSION="$$DOCKERHUB_RUST_VERSION" \
		--build-arg STELLAR_CLI_CRATE_VERSION="$(STELLAR_CLI_CRATE_VERSION)" \
		-f- $(STELLAR_CLI_GIT_REF) < $(MAKEFILE_DIR)Dockerfile.stellar-cli; \
	fi

build-horizon:
	if [ -z "$(QUICKSTART_IMAGE)" ] && [ -z "$(HORIZON_IMAGE)" ]; then \
		SOURCE_URL="$(GO_GIT_REF)"; \
		if [[ ! "$(GO_GIT_REF)" =~ \.git ]]; then \
			pushd "$(GO_GIT_REF)"; \
			SOURCE_URL=.; \
		fi; \
		docker build -t "$(HORIZON_STAGE_IMAGE)" \
		--build-arg BUILDKIT_CONTEXT_KEEP_GIT_DIR=true \
		--target builder -f services/horizon/docker/Dockerfile.dev "$$SOURCE_URL"; \
	fi

build-core:
	if [ -z "$(QUICKSTART_IMAGE)" ] && [ -z "$(CORE_IMAGE)" ]; then \
		SOURCE_URL="$(CORE_GIT_REF)"; \
		if [[ ! "$(CORE_GIT_REF)" =~ \.git ]]; then \
			pushd "$(CORE_GIT_REF)"; \
			SOURCE_URL=.; \
		fi; \
		docker build -t "$(CORE_STAGE_IMAGE)" \
		--build-arg BUILDKIT_CONTEXT_KEEP_GIT_DIR=true \
		--build-arg CONFIGURE_FLAGS="$(CORE_COMPILE_CONFIGURE_FLAGS)" \
		-f docker/Dockerfile.testing "$$SOURCE_URL"; \
	fi; \
	if [ ! -z "$(CORE_IMAGE)" ] && [ ! -z "$(CORE_IMAGE_BIN_PATH)" ]; then \
		docker build -t "$(CORE_STAGE_IMAGE)" \
		--build-arg CORE_IMAGE="$(CORE_IMAGE)" \
		--build-arg CORE_IMAGE_BIN_PATH="$(CORE_IMAGE_BIN_PATH)" \
		-f Dockerfile.core .; \
	fi

build-quickstart: build-core build-friendbot build-horizon build-rs-xdr build-stellar-rpc
	if [ -z "$(QUICKSTART_IMAGE)" ]; then \
		CORE_IMAGE_REF=$$( [[ -z "$(CORE_IMAGE)"  ||  ! -z "$(CORE_IMAGE_BIN_PATH)" ]] && echo "$(CORE_STAGE_IMAGE)" || echo "$(CORE_IMAGE)"); \
		HORIZON_IMAGE_REF=$$( [ -z "$(HORIZON_IMAGE)" ] && echo "$(HORIZON_STAGE_IMAGE)" || echo "$(HORIZON_IMAGE)"); \
		FRIENDBOT_IMAGE_REF=$$( [ -z "$(FRIENDBOT_IMAGE)" ] && echo "$(FRIENDBOT_STAGE_IMAGE)" || echo "$(FRIENDBOT_IMAGE)"); \
		STELLAR_RPC_IMAGE_REF=$$( [ -z "$(STELLAR_RPC_IMAGE)" ] && echo "$(STELLAR_RPC_STAGE_IMAGE)" || echo "$(STELLAR_RPC_IMAGE)"); \
		RS_XDR_IMAGE_REF=$$( [ -z "$(RS_XDR_IMAGE)" ] && echo "$(RS_XDR_STAGE_IMAGE)" || echo "$(RS_XDR_IMAGE)"); \
		SOURCE_URL="$(QUICKSTART_GIT_REF)"; \
		if [[ ! "$(QUICKSTART_GIT_REF)" =~ \.git ]]; then \
			pushd "$(QUICKSTART_GIT_REF)"; \
			SOURCE_URL=.; \
		fi; \
		docker build -t "$(QUICKSTART_STAGE_IMAGE)" \
		--build-arg PROTOCOL_VERSION_DEFAULT=$$PROTOCOL_VERSION_DEFAULT \
		--build-arg BUILDKIT_CONTEXT_KEEP_GIT_DIR=true \
		--build-arg STELLAR_CORE_IMAGE_REF=$$CORE_IMAGE_REF \
		--build-arg STELLAR_XDR_IMAGE_REF=$$RS_XDR_IMAGE_REF \
		--build-arg CORE_SUPPORTS_ENABLE_SOROBAN_DIAGNOSTIC_EVENTS=true \
		--build-arg CORE_SUPPORTS_TESTING_SOROBAN_HIGH_LIMIT_OVERRIDE=true \
		--build-arg HORIZON_IMAGE_REF=$$HORIZON_IMAGE_REF \
		--build-arg FRIENDBOT_IMAGE_REF=$$FRIENDBOT_IMAGE_REF \
		--build-arg STELLAR_RPC_IMAGE_REF=$$STELLAR_RPC_IMAGE_REF \
		-f Dockerfile "$$SOURCE_URL"; \
	fi

build: build-quickstart build-stellar-cli
	QUICKSTART_IMAGE_REF=$$( [ -z "$(QUICKSTART_IMAGE)" ] && echo "$(QUICKSTART_STAGE_IMAGE)" || echo "$(QUICKSTART_IMAGE)"); \
	STELLAR_CLI_IMAGE_REF=$$( [ -z "$(STELLAR_CLI_IMAGE)" ] && echo "$(STELLAR_CLI_STAGE_IMAGE)" || echo "$(STELLAR_CLI_IMAGE)"); \
	docker build -t "$(SYSTEM_TEST_IMAGE)" -f Dockerfile \
	    --build-arg BUILDKIT_CONTEXT_KEEP_GIT_DIR=true \
		--build-arg QUICKSTART_IMAGE_REF=$$QUICKSTART_IMAGE_REF \
		--build-arg STELLAR_CLI_CRATE_VERSION=$(STELLAR_CLI_CRATE_VERSION) \
		--build-arg STELLAR_CLI_IMAGE_REF=$$STELLAR_CLI_IMAGE_REF \
		--build-arg RUST_TOOLCHAIN_VERSION=$(RUST_TOOLCHAIN_VERSION) \
		--build-arg NODE_VERSION=$(NODE_VERSION) \
		--build-arg JS_STELLAR_SDK_NPM_VERSION=$(JS_STELLAR_SDK_NPM_VERSION) \
		--label org.opencontainers.image.revision="$(SYSTEM_TEST_SHA)" .;
