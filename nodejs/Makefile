REGISTRY ?= us-docker.pkg.dev/gcb-catalog-release/catalog
TAG ?= ubuntu22

build:
	@echo "+ building nodejs images"
	cd .. && docker buildx bake nodejs

test:
	container-structure-test test --image ${REGISTRY}/nodejs:${TAG} --config test_config.yaml

.PHONY: build test $(SUBDIRS)