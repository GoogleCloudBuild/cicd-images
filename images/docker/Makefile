REGISTRY ?= us-docker.pkg.dev/gcb-catalog-release/catalog
TAG ?= ubuntu24

build:
	@echo "+ building docker images"
	cd .. && docker buildx bake docker-dind

test:
	container-structure-test test --image ${REGISTRY}/docker/dind:${TAG} --config test_config_dind.yaml

.PHONY: build test