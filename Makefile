.DEFAULT_GOAL     := all

PROJECT_ROOT      := github.com/benshields/messagebox
BUILD_PATH        := bin
DOCKERFILE_PATH   := $(CURDIR)

# configuration for building on host machine
GO_TEST_FLAGS     ?= -v -cover
GO_PACKAGES       ?= $(shell go list ./... | grep -v vendor)

# configuration for image names
IMAGE_REGISTRY    ?= bendshields
IMAGE_NAME        ?= messagebox
GIT_COMMIT        := $(shell git describe --dirty=-unsupported --always --tags || echo pre-commit)
IMAGE_VERSION     ?= $(GIT_COMMIT)

# configuration for server binary and image
SERVER_BINARY     := $(BUILD_PATH)/server
SERVER_PATH       := $(PROJECT_ROOT)/cmd/server
SERVER_IMAGE      := $(IMAGE_REGISTRY)/$(IMAGE_NAME)
SERVER_DOCKERFILE := $(DOCKERFILE_PATH)/Dockerfile

.PHONY: all
all: fmt test

.PHONY: fmt
fmt:
	go fmt $(GO_PACKAGES)

.PHONY: test
test: vendor
	go test $(GO_TEST_FLAGS) $(GO_PACKAGES)

.PHONY: vendor
vendor:
	go mod tidy
	go mod vendor

# docker-build docker-push docker-run docker-up docker-down
.PHONY: docker-build
docker-build:
	docker build -f $(SERVER_DOCKERFILE) -t $(IMAGE_NAME):$(IMAGE_VERSION) .
	docker tag $(IMAGE_NAME):$(IMAGE_VERSION) $(SERVER_IMAGE):$(IMAGE_VERSION)

.docker-$(IMAGE_NAME)-$(IMAGE_VERSION):
	$(MAKE) docker-build
	touch $@

.PHONY: docker
docker: .docker-$(IMAGE_NAME)-$(IMAGE_VERSION)

docker-push-reg: docker-build
ifndef IMAGE_REGISTRY
	@(echo "Please set IMAGE_REGISTRY variable in Makefile to use push command"; exit 1)
else
	docker push $(SERVER_IMAGE):$(IMAGE_VERSION)
endif

.docker-push-$(IMAGE_NAME)-$(IMAGE_VERSION):
	$(MAKE) docker-push-reg
	touch $@

.PHONY: docker-push
docker-push: .docker-push-$(IMAGE_NAME)-$(IMAGE_VERSION)

.PHONY: docker-run
docker-run: docker
	docker run -d --name messagebox -p 8080:8080 --volume $(shell pwd)/config:/config ${IMAGE_NAME}:$(IMAGE_VERSION)

.PHONY: docker-up
docker-up:
	docker compose -f ./docker-compose.yaml up -d

.PHONY: docker-down
docker-down:
	if docker inspect messagebox &>/dev/null; then \
		docker rm messagebox -fv; \
	fi

.PHONY: docker-clean
docker-clean:
	rm -f ./.docker-$(IMAGE_NAME)-*
	rm -f ./.docker-push-$(IMAGE_NAME)-*
