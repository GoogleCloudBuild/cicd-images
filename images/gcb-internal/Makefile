REGISTRY ?= us-docker.pkg.dev/gcb-catalog-release/catalog
TAG ?= debian12

build:
	@echo "+ building gcb-internal image"
	cd .. && docker buildx bake gcb-internal

test:
	container-structure-test test --image ${REGISTRY}/gcb-internal:${TAG} --config test_config.yaml

.PHONY: build test