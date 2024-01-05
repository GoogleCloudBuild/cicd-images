REGISTRY ?= us-docker.pkg.dev/gcb-catalog-release/catalog
TAG ?= ubuntu22

build:
	@echo "+ building git image"
	cd .. && docker buildx bake git

test:
	container-structure-test test --image ${REGISTRY}/git:${TAG} --config test_config.yaml

.PHONY: build test $(SUBDIRS)