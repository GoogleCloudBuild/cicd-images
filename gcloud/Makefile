REGISTRY ?= us-docker.pkg.dev/gcb-catalog-release/catalog
TAG ?= ubuntu22

build:
	@echo "+ building gcloud image"
	cd .. && docker buildx bake gcloud

test:
	container-structure-test test --image ${REGISTRY}/gcloud:${TAG} --config test_config.yaml

.PHONY: build test $(SUBDIRS)