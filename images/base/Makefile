REGISTRY ?= us-docker.pkg.dev/gcb-catalog-release/catalog
TAG ?= ubuntu24

build:
	@echo "+ building base image"
	cd .. && docker buildx bake base

test:
	container-structure-test test --image ${REGISTRY}/gcb-base:${TAG} --config test_config.yaml

.PHONY: build test