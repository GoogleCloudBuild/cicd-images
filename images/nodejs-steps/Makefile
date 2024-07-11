REGISTRY ?= us-docker.pkg.dev/gcb-catalog-release/catalog
TAG ?= ubuntu24

build:
	@echo "+ building nodejs images"
	cd .. && docker buildx bake nodejs-steps

test:
	container-structure-test test --image ${REGISTRY}/nodejs-steps:${TAG} --config test_config.yaml

.PHONY: build test