REGISTRY ?= us-docker.pkg.dev/gcb-catalog-release/catalog
TAG ?= deprecated-public-image-ubuntu24

build:
	@echo "+ building go images"
	cd .. && docker buildx bake go-steps

test:
	container-structure-test test --image ${REGISTRY}/go-steps:${TAG} --config test_config.yaml

.PHONY: build test