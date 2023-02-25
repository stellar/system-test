SHELL:=/bin/bash
MAKEFILE_DIR:=$(dir $(abspath $(lastword $(MAKEFILE_LIST))))

.EXPORT_ALL_VARIABLES:
.PHONY: build build-quickstart build-core build-friendbot build-horizon build-soroban-rpc

SYSTEM_TEST_SHA=$(shell git rev-parse HEAD)
QUICKSTART_STAGE_IMAGE=stellar/system-test-base:dev
CORE_STAGE_IMAGE=stellar/system-test-core:dev
HORIZON_STAGE_IMAGE=stellar/system-test-horizon:dev
FRIENDBOT_STAGE_IMAGE=stellar/system-test-friendbot:dev
SOROBAN_RPC_STAGE_IMAGE=stellar/system-test-soroban-rpc:dev
SOROBAN_CLI_STAGE_IMAGE=stellar/system-test-soroban-cli:dev

# The rest of these variables can be set as environment variables to the makefile
# to modify how system test is built.

# variables to set for source code, can be any valid docker context url local path github remote repo `https://github.com/repo#<ref>`
CORE_GIT_REF=https://github.com/stellar/stellar-core.git\#master
SOROBAN_RPC_GIT_REF=https://github.com/stellar/soroban-tools.git\#main
SOROBAN_CLI_GIT_REF=https://github.com/stellar/soroban-tools.git\#main
GO_GIT_REF=https://github.com/stellar/go.git\#soroban-xdr-next
QUICKSTART_GIT_REF=https://github.com/stellar/quickstart.git\#master

NON_AMD_ARCH=false 
ifneq ($(shell uname -p),x86_64)
	NON_AMD_ARCH=true
endif

# variables to set if wanting to use existing dockerhub images instead of compiling
# image during build. if using this option, the image ref should provide a version for same 
# platform arch as the build host is on, i.e. linux/amd64 or linux/arm64.
#
# image must have soroban cli bin at /usr/local/cargo/bin/soroban
SOROBAN_CLI_IMAGE=
#
# image must have soroban rpc bin at /bin/soroban-rpc
SOROBAN_RPC_IMAGE=
#
# image must have horizon bin at /go/bin/horizon
HORIZON_IMAGE=
#
# image must have friendbot bin at /app/friendbot
FRIENDBOT_IMAGE=
#
# image must have core bin at /usr/local/bin/stellar-core
CORE_IMAGE=
#
# a prebuilt 'soroban-dev' image from the quickstart repo, if this is supplied, 
# the other core, rpc, horizon, friendbot config settings are mostly ignored, since the quickstart image
# has them compiled in already. the 'stellar/quickstart' images also support multi-arch, so the build will
# work those images whether the build host is arm64 or amd64.
QUICKSTART_IMAGE=

# if crate version is set, then it overrides SOROBAN_CLI_GIT_REF, cli will be installed from this create instead
SOROBAN_CLI_CRATE_VERSION=

# sets the rustc version in the system test image 
RUST_TOOLCHAIN_VERSION=stable

# temporarily needed, builds core with soroban enabled features
CORE_COMPILE_CONFIGURE_FLAGS=--disable-tests --enable-next-protocol-version-unsafe-for-production

# the final image name that is created in local docker images store for system test
SYSTEM_TEST_IMAGE=stellar/system-test:dev

build-friendbot:
	if [ -z "$(QUICKSTART_IMAGE)" ] && [ -z "$(FRIENDBOT_IMAGE)" ]; then \
		SOURCE_URL="$(GO_GIT_REF)"; \
		if [[ ! "$(GO_GIT_REF)" =~ \.git ]]; then \
			pushd "$(GO_GIT_REF)"; \
			SOURCE_URL=.; \
		fi; \
		docker build -t "$(FRIENDBOT_STAGE_IMAGE)" -f services/friendbot/docker/Dockerfile "$$SOURCE_URL"; \
	fi	

build-soroban-rpc:
	if [ -z "$(QUICKSTART_IMAGE)" ] && [ -z "$(SOROBAN_RPC_IMAGE)" ]; then \
		SOURCE_URL="$(SOROBAN_RPC_GIT_REF)"; \
		if [[ ! "$(SOROBAN_RPC_GIT_REF)" =~ \.git ]]; then \
			pushd "$(SOROBAN_RPC_GIT_REF)"; \
			SOURCE_URL=.; \
		fi; \
		docker build -t "$(SOROBAN_RPC_STAGE_IMAGE)" --target build -f cmd/soroban-rpc/docker/Dockerfile "$$SOURCE_URL"; \
	fi

build-soroban-cli:
	if [ -z "$(SOROBAN_CLI_IMAGE)" ]; then \
		DOCKERHUB_RUST_VERSION=rust:$$( [ "$(RUST_TOOLCHAIN_VERSION)" = "stable" ] && echo "latest" || echo "$(RUST_TOOLCHAIN_VERSION)"); \
		docker buildx build -t "$(SOROBAN_CLI_STAGE_IMAGE)" --target builder \
		--build-arg DOCKERHUB_RUST_VERSION="$$DOCKERHUB_RUST_VERSION" \
		--build-arg SOROBAN_CLI_CRATE_VERSION="$(SOROBAN_CLI_CRATE_VERSION)" \
		-f- $(SOROBAN_CLI_GIT_REF) < $(MAKEFILE_DIR)Dockerfile.soroban-cli; \
	fi

build-horizon:
	if [ -z "$(QUICKSTART_IMAGE)" ] && [ -z "$(HORIZON_IMAGE)" ]; then \
		SOURCE_URL="$(GO_GIT_REF)"; \
		if [[ ! "$(GO_GIT_REF)" =~ \.git ]]; then \
			pushd "$(GO_GIT_REF)"; \
			SOURCE_URL=.; \
		fi; \
		docker build -t "$(HORIZON_STAGE_IMAGE)" --target builder -f services/horizon/docker/Dockerfile.dev "$$SOURCE_URL"; \
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
	fi

build-quickstart: build-core build-friendbot build-horizon build-soroban-rpc
	if [ -z "$(QUICKSTART_IMAGE)" ]; then \
		CORE_IMAGE_REF=$$( [ -z "$(CORE_IMAGE)" ] && echo "$(CORE_STAGE_IMAGE)" || echo "$(CORE_IMAGE)"); \
		HORIZON_IMAGE_REF=$$( [ -z "$(HORIZON_IMAGE)" ] && echo "$(HORIZON_STAGE_IMAGE)" || echo "$(HORIZON_IMAGE)"); \
		FRIENDBOT_IMAGE_REF=$$( [ -z "$(FRIENDBOT_IMAGE)" ] && echo "$(FRIENDBOT_STAGE_IMAGE)" || echo "$(FRIENDBOT_IMAGE)"); \
		SOROBAN_RPC_IMAGE_REF=$$( [ -z "$(SOROBAN_RPC_IMAGE)" ] && echo "$(SOROBAN_RPC_STAGE_IMAGE)" || echo "$(SOROBAN_RPC_IMAGE)"); \
		SOURCE_URL="$(QUICKSTART_GIT_REF)"; \
		if [[ ! "$(QUICKSTART_GIT_REF)" =~ \.git ]]; then \
			pushd "$(QUICKSTART_GIT_REF)"; \
			SOURCE_URL=.; \
		fi; \
		docker build -t "$(QUICKSTART_STAGE_IMAGE)" \
		--build-arg STELLAR_CORE_IMAGE_REF=$$CORE_IMAGE_REF \
		--build-arg HORIZON_IMAGE_REF=$$HORIZON_IMAGE_REF \
		--build-arg FRIENDBOT_IMAGE_REF=$$FRIENDBOT_IMAGE_REF \
		--build-arg SOROBAN_RPC_IMAGE_REF=$$SOROBAN_RPC_IMAGE_REF \
		-f Dockerfile "$$SOURCE_URL"; \
	fi

build: build-quickstart build-soroban-cli
	QUICKSTART_IMAGE_REF=$$( [ -z "$(QUICKSTART_IMAGE)" ] && echo "$(QUICKSTART_STAGE_IMAGE)" || echo "$(QUICKSTART_IMAGE)"); \
	SOROBAN_CLI_IMAGE_REF=$$( [ -z "$(SOROBAN_CLI_IMAGE)" ] && echo "$(SOROBAN_CLI_STAGE_IMAGE)" || echo "$(SOROBAN_CLI_IMAGE)"); \
	docker build -t "$(SYSTEM_TEST_IMAGE)" -f Dockerfile \
		--build-arg QUICKSTART_IMAGE_REF=$$QUICKSTART_IMAGE_REF \
		--build-arg SOROBAN_CLI_CRATE_VERSION=$(SOROBAN_CLI_CRATE_VERSION) \
		--build-arg SOROBAN_CLI_IMAGE_REF=$$SOROBAN_CLI_IMAGE_REF \
		--build-arg RUST_TOOLCHAIN_VERSION=$(RUST_TOOLCHAIN_VERSION) \
		--label org.opencontainers.image.revision="$(SYSTEM_TEST_SHA)" .;
	
