REGISTRY ?= us-docker.pkg.dev/gcb-catalog-release/catalog
TAG ?= ubuntu22

build:
	@echo "+ building generic builder image"
	cd .. && docker buildx bake builder

test:
	container-structure-test test --image ${REGISTRY}/builder:${TAG} --config test_config.yaml

.PHONY: build test
