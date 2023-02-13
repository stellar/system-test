SHELL := /bin/bash
.EXPORT_ALL_VARIABLES:
.PHONY: build-docker-driver-for-cache build build-base build-with-registry-cache build-base-with-registry-cache \
		build-core-with-registry-cache build-friendbot-with-registry-cache build-horizon-with-registry-cache build-soroban-rpc-with-registry-cache

SYSTEM_TEST_SHA=$(shell git rev-parse HEAD)

CORE_GIT_REF=
CORE_COMPILE_CONFIGURE_FLAGS?=--disable-tests --enable-next-protocol-version-unsafe-for-production
SOROBAN_RPC_GIT_REF=main
# TODO: remove go ref if quickstart /scripts/soroban_repo_to_horizon_repo.sh can resolve remote repo pr refs
# it currently does not and allowing this to be configured directly.
GO_GIT_REF=soroban-xdr-next
QUICKSTART_IMAGE=
QUICKSTART_GIT_REF=master
QUICKSTART_GIT_REPO?=https://github.com/stellar/quickstart.git
RUST_TOOLCHAIN_VERSION?=stable
SOROBAN_CLI_CRATE_VERSION=
SOROBAN_CLI_GIT_REF=main

DOCKER_CACHE_REGISTRY=
ifeq ($(strip $(DOCKER_CACHE_REGISTRY)),)
	DOCKER_CACHE_REGISTRY_PATH=localhost:5001/stellar
else
	DOCKER_CACHE_REGISTRY_PATH=$(DOCKER_CACHE_REGISTRY)
endif

SYSTEM_TEST_STAGE_IMAGE=$(DOCKER_CACHE_REGISTRY_PATH)/system-test:dev
QUICKSTART_STAGE_IMAGE=$(DOCKER_CACHE_REGISTRY_PATH)/system-test-base:dev
CORE_STAGE_IMAGE=$(DOCKER_CACHE_REGISTRY_PATH)/system-test-core:dev
HORIZON_STAGE_IMAGE=$(DOCKER_CACHE_REGISTRY_PATH)/system-test-horizon:dev
FRIENDBOT_STAGE_IMAGE=$(DOCKER_CACHE_REGISTRY_PATH)/system-test-friendbot:dev
SOROBAN_RPC_STAGE_IMAGE=$(DOCKER_CACHE_REGISTRY_PATH)/system-test-soroban-rpc:dev

SYSTEM_TEST_PUSH_IMAGE=
ifeq ($(strip $(SYSTEM_TEST_PUSH_IMAGE)),)
	SYSTEM_TEST_IMAGE=$(SYSTEM_TEST_STAGE_IMAGE)
	DOCKER_SYSTEM_TEST_OUTPUT_TYPE=--load
else 
	SYSTEM_TEST_IMAGE=$(SYSTEM_TEST_PUSH_IMAGE)
	DOCKER_SYSTEM_TEST_OUTPUT_TYPE=--push
endif	

build-docker-driver-for-cache:
	docker buildx inspect systest_builder > /dev/null 2>&1; \
	if [ $$? -eq 1 ]; then \
		docker buildx create --name systest_builder --driver-opt network=host --driver docker-container --use; \
	else \
		docker buildx use systest_builder; \
	fi

build-soroban-rpc-with-registry-cache:
	if [ -z "$(QUICKSTART_IMAGE)" ]; then \
		docker buildx build -t "$(SOROBAN_RPC_STAGE_IMAGE)" --target build --push \
			--cache-to $(SOROBAN_RPC_STAGE_IMAGE)cache \
			--cache-from $(SOROBAN_RPC_STAGE_IMAGE)cache \
			-f cmd/soroban-rpc/docker/Dockerfile https://github.com/stellar/soroban-tools.git#$(SOROBAN_RPC_GIT_REF); \
	fi 

build-horizon-with-registry-cache:
	if [ -z "$(QUICKSTART_IMAGE)" ]; then \
		docker buildx build -t "$(HORIZON_STAGE_IMAGE)" --target builder --push \
			--cache-to $(HORIZON_STAGE_IMAGE)cache \
			--cache-from $(HORIZON_STAGE_IMAGE)cache \
			--build-arg REF=$$GO_GIT_REF \
			-f Dockerfile.horizon --target builder https://github.com/stellar/quickstart.git#$(QUICKSTART_GIT_REF); \
	fi   

build-friendbot-with-registry-cache:
	if [ -z "$(QUICKSTART_IMAGE)" ]; then \
		docker buildx build -t "$(FRIENDBOT_STAGE_IMAGE)" --push \
			--cache-to $(FRIENDBOT_STAGE_IMAGE)cache \
			--cache-from $(FRIENDBOT_STAGE_IMAGE)cache \
			-f services/friendbot/docker/Dockerfile https://github.com/stellar/go.git#$(GO_GIT_REF); \
	fi

build-core-with-registry-cache: 
	if [ -z "$(QUICKSTART_IMAGE)" ]; then \
		docker buildx build -t "$(CORE_STAGE_IMAGE)" --push \
			--cache-to $(CORE_STAGE_IMAGE)cache \
			--cache-from $(CORE_STAGE_IMAGE)cache \
			-f docker/Dockerfile.testing https://github.com/stellar/stellar-core.git#$(CORE_GIT_REF) \
			--build-arg BUILDKIT_CONTEXT_KEEP_GIT_DIR=true \
			--build-arg CONFIGURE_FLAGS="$(CORE_COMPILE_CONFIGURE_FLAGS)"; \
	fi

build-base-with-registry-cache: build-docker-driver-for-cache build-core-with-registry-cache build-friendbot-with-registry-cache build-horizon-with-registry-cache build-soroban-rpc-with-registry-cache
	if [ -z "$(QUICKSTART_IMAGE)" ]; then \
		docker buildx build -f Dockerfile https://github.com/stellar/quickstart.git#$(QUICKSTART_GIT_REF) \
			-t "$(QUICKSTART_STAGE_IMAGE)" --push \
			--cache-to $(QUICKSTART_STAGE_IMAGE)cache \
			--cache-from $(QUICKSTART_STAGE_IMAGE)cache \
			--label org.opencontainers.image.revision="$(QUICKSTART_GIT_REF)" \
			--build-arg STELLAR_CORE_IMAGE_REF=$(CORE_STAGE_IMAGE) \
			--build-arg HORIZON_IMAGE_REF=$(HORIZON_STAGE_IMAGE) \
			--build-arg FRIENDBOT_IMAGE_REF=$(FRIENDBOT_STAGE_IMAGE) \
			--build-arg SOROBAN_RPC_IMAGE_REF=$(SOROBAN_RPC_STAGE_IMAGE); \
	fi

build-with-registry-cache: build-base-with-registry-cache
	QUICKSTART_IMAGE=$$( [ -z "$(QUICKSTART_IMAGE)" ] && echo "$(QUICKSTART_STAGE_IMAGE)" || echo "$(QUICKSTART_IMAGE)"); \
	docker buildx build -t "$(SYSTEM_TEST_IMAGE)" $(DOCKER_SYSTEM_TEST_OUTPUT_TYPE) \
		--cache-from $(SYSTEM_TEST_STAGE_IMAGE)cache \
		--cache-to $(SYSTEM_TEST_STAGE_IMAGE)cache \
		-f Dockerfile \
		--build-arg QUICKSTART_IMAGE_REF=$$QUICKSTART_IMAGE \
		--build-arg SOROBAN_CLI_CRATE_VERSION=$(SOROBAN_CLI_CRATE_VERSION) \
		--build-arg SOROBAN_CLI_GIT_REF=$(SOROBAN_CLI_GIT_REF) \
		--build-arg RUST_TOOLCHAIN_VERSION=$(RUST_TOOLCHAIN_VERSION) \
		--label org.opencontainers.image.revision="$(SYSTEM_TEST_SHA)" .;

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
    docker build -t stellar/system-test:dev \
      -f Dockerfile \
      --build-arg QUICKSTART_IMAGE_REF="$$QUICKSTART_IMAGE" \
      --build-arg SOROBAN_CLI_CRATE_VERSION="$(SOROBAN_CLI_CRATE_VERSION)" \
      --build-arg SOROBAN_CLI_GIT_REF=$(SOROBAN_CLI_GIT_REF) \
      --build-arg RUST_TOOLCHAIN_VERSION=$(RUST_TOOLCHAIN_VERSION) \
      --label org.opencontainers.image.revision="$(SYSTEM_TEST_SHA)" .
	



