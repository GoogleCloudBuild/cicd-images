REGISTRY ?= us-docker.pkg.dev/gcb-catalog-release/catalog
TAG ?= ubuntu22

build:
	@echo "+ building go images"
	cd .. && docker buildx bake go

test:
	container-structure-test test --image ${REGISTRY}/go:${TAG} --config test_config.yaml

.PHONY: build test $(SUBDIRS)