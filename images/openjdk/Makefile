REGISTRY ?= us-docker.pkg.dev/gcb-catalog-release/catalog
TAG ?= ubuntu22

build:
	@echo "+ building openjdk images"
	cd .. && docker buildx bake openjdk

test:
	container-structure-test test --image ${REGISTRY}/openjdk:${TAG} --config test_config.yaml

.PHONY: build test