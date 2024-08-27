REGISTRY ?= us-docker.pkg.dev/gcb-catalog-release/catalog
TAG ?= deprecated-public-image-ubuntu24

build:
	@echo "+ building git image"
	cd .. && docker buildx bake git-steps

test:
	container-structure-test test --image ${REGISTRY}/git-steps:${TAG} --config test_config.yaml

.PHONY: build test