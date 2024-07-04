REGISTRY ?= us-docker.pkg.dev/gcb-catalog-release/catalog
TAG ?= ubuntu22

build:
	@echo "+ building python images"
	cd .. && docker buildx bake python-steps

test:
	container-structure-test test --image ${REGISTRY}/python-steps:${TAG} --config test_config.yaml

.PHONY: build test