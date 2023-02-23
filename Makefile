SHELL := /bin/bash
.EXPORT_ALL_VARIABLES:
.PHONY: build-docker-driver-for-cache build build-base build-with-cache build-base-with-cache docker-stop-local-reg docker-start-local-reg \
		build-core-with-cache build-friendbot-with-cache build-horizon-with-cache build-soroban-rpc-with-cache build-system-test-with-cache

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

DOCKER_REGISTRY_PATH?=localhost:5001/
SYSTEM_TEST_STAGE_IMAGE=$(DOCKER_REGISTRY_PATH)stellar/system-test:dev
QUICKSTART_STAGE_IMAGE=$(DOCKER_REGISTRY_PATH)stellar/system-test-base:dev
CORE_STAGE_IMAGE=$(DOCKER_REGISTRY_PATH)stellar/system-test-core:dev
HORIZON_STAGE_IMAGE=$(DOCKER_REGISTRY_PATH)stellar/system-test-horizon:dev
FRIENDBOT_STAGE_IMAGE=$(DOCKER_REGISTRY_PATH)stellar/system-test-friendbot:dev
SOROBAN_RPC_STAGE_IMAGE=$(DOCKER_REGISTRY_PATH)stellar/system-test-soroban-rpc:dev

ifeq ($(strip $(DOCKER_CACHE)),gha)
	DOCKER_CACHE_TO=type=gha,mode=max
	DOCKER_CACHE_FROM=type=gha
else
	DOCKER_CACHE_TO=type=registry,ref=$(DOCKER_REGISTRY_PATH)stellar/system-test:cache
	DOCKER_CACHE_FROM=type=registry,ref=$(DOCKER_REGISTRY_PATH)stellar/system-test:cache
endif

SYSTEM_TEST_PUSH_IMAGE=
ifeq ($(strip $(SYSTEM_TEST_PUSH_IMAGE)),)
	SYSTEM_TEST_IMAGE=$(SYSTEM_TEST_STAGE_IMAGE)
	DOCKER_SYSTEM_TEST_OUTPUT_TYPE=--load
else 
	SYSTEM_TEST_IMAGE=$(SYSTEM_TEST_PUSH_IMAGE)
	DOCKER_SYSTEM_TEST_OUTPUT_TYPE=--push
endif	

docker-start-local-reg: docker-stop-local-reg
	docker run -d -p 5001:5000 --name registry registry:2 > /dev/null 2>&1;

docker-stop-local-reg:
	docker container stop registry || true > /dev/null 2>&1; \
	docker rm registry || true > /dev/null 2>&1; \
	docker buildx use default || true > /dev/null 2>&1;

build-docker-driver-for-cache:
	docker buildx inspect systest_builder > /dev/null 2>&1; \
	if [ $$? -eq 1 ]; then \
		docker buildx create --name systest_builder --driver-opt network=host --driver docker-container --use; \
	else \
		docker buildx use systest_builder; \
	fi

build-soroban-rpc-with-cache:
	if [ -z "$(QUICKSTART_IMAGE)" ]; then \
		docker buildx build -t "$(SOROBAN_RPC_STAGE_IMAGE)" --target build --push \
			--cache-to "$(subst <SCOPE>,soroban-rpc,$(DOCKER_CACHE_TO))" \
			--cache-from "$(subst <SCOPE>,soroban-rpc,$(DOCKER_CACHE_FROM))" \
			-f cmd/soroban-rpc/docker/Dockerfile https://github.com/stellar/soroban-tools.git#$(SOROBAN_RPC_GIT_REF); \
	fi 

build-horizon-with-cache:
	if [ -z "$(QUICKSTART_IMAGE)" ]; then \
		docker buildx build -t "$(HORIZON_STAGE_IMAGE)" --target builder --push \
			--cache-to "$(subst <SCOPE>,horizon,$(DOCKER_CACHE_TO))" \
			--cache-from "$(subst <SCOPE>,horizon,$(DOCKER_CACHE_FROM))" \
			--build-arg REF=$$GO_GIT_REF \
			-f Dockerfile.horizon --target builder https://github.com/stellar/quickstart.git#$(QUICKSTART_GIT_REF); \
	fi   

build-friendbot-with-cache:
	if [ -z "$(QUICKSTART_IMAGE)" ]; then \
		docker buildx build -t "$(FRIENDBOT_STAGE_IMAGE)" --push \
			--cache-to "$(subst <SCOPE>,friendbot,$(DOCKER_CACHE_TO))" \
			--cache-from "$(subst <SCOPE>,friendbot,$(DOCKER_CACHE_FROM))" \
			-f services/friendbot/docker/Dockerfile https://github.com/stellar/go.git#$(GO_GIT_REF); \
	fi

build-core-with-cache: 
	if [ -z "$(QUICKSTART_IMAGE)" ]; then \
		docker buildx build -t "$(CORE_STAGE_IMAGE)" --push \
			--cache-to "$(subst <SCOPE>,core,$(DOCKER_CACHE_TO))" \
			--cache-from "$(subst <SCOPE>,core,$(DOCKER_CACHE_FROM))" \
			-f docker/Dockerfile.testing https://github.com/stellar/stellar-core.git#$(CORE_GIT_REF) \
			--build-arg BUILDKIT_CONTEXT_KEEP_GIT_DIR=true \
			--build-arg CONFIGURE_FLAGS="$(CORE_COMPILE_CONFIGURE_FLAGS)"; \
	fi

build-base-with-cache: build-docker-driver-for-cache build-core-with-cache build-friendbot-with-cache build-horizon-with-cache build-soroban-rpc-with-cache
	if [ -z "$(QUICKSTART_IMAGE)" ]; then \
		docker buildx build -f Dockerfile https://github.com/stellar/quickstart.git#$(QUICKSTART_GIT_REF) \
			-t "$(QUICKSTART_STAGE_IMAGE)" --push \
			--cache-to "$(subst <SCOPE>,quickstart,$(DOCKER_CACHE_TO))" \
			--cache-from "$(subst <SCOPE>,quickstart,$(DOCKER_CACHE_FROM))" \
			--label org.opencontainers.image.revision="$(QUICKSTART_GIT_REF)" \
			--build-arg STELLAR_CORE_IMAGE_REF=$(CORE_STAGE_IMAGE) \
			--build-arg HORIZON_IMAGE_REF=$(HORIZON_STAGE_IMAGE) \
			--build-arg FRIENDBOT_IMAGE_REF=$(FRIENDBOT_STAGE_IMAGE) \
			--build-arg SOROBAN_RPC_IMAGE_REF=$(SOROBAN_RPC_STAGE_IMAGE); \
	fi

build-system-test-with-cache: build-base-with-cache
	QUICKSTART_IMAGE=$$( [ -z "$(QUICKSTART_IMAGE)" ] && echo "$(QUICKSTART_STAGE_IMAGE)" || echo "$(QUICKSTART_IMAGE)"); \
	docker buildx build -t "$(SYSTEM_TEST_IMAGE)" $(DOCKER_SYSTEM_TEST_OUTPUT_TYPE) \
		--cache-from "$(subst <SCOPE>,systemtest,$(DOCKER_CACHE_FROM))" \
		--cache-to "$(subst <SCOPE>,systemtest,$(DOCKER_CACHE_TO))" \
		-f Dockerfile \
		--build-arg QUICKSTART_IMAGE_REF=$$QUICKSTART_IMAGE \
		--build-arg SOROBAN_CLI_CRATE_VERSION=$(SOROBAN_CLI_CRATE_VERSION) \
		--build-arg SOROBAN_CLI_GIT_REF=$(SOROBAN_CLI_GIT_REF) \
		--build-arg RUST_TOOLCHAIN_VERSION=$(RUST_TOOLCHAIN_VERSION) \
		--label org.opencontainers.image.revision="$(SYSTEM_TEST_SHA)" .;\
	$(MAKE) docker-stop-local-reg

build-with-cache: docker-start-local-reg
	$(MAKE) build-system-test-with-cache || $(MAKE) docker-stop-local-reg

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
	



