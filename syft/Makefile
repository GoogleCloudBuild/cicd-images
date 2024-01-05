REGISTRY ?= us-docker.pkg.dev/gcb-catalog-release/catalog
TAG ?= ubuntu22

build:
	@echo "+ building syft image"
	cd .. && docker buildx bake syft

test:
	container-structure-test test --image ${REGISTRY}/syft:${TAG} --config test_config.yaml

.PHONY: build test $(SUBDIRS)