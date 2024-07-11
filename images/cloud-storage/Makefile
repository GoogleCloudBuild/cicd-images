REGISTRY ?= us-docker.pkg.dev/gcb-catalog-release/catalog
TAG ?= ubuntu24

build:
	@echo "+ building cloud-storage image"
	cd .. && docker buildx bake cloud-storage

test:
	# TODO(nschnarr): update tests
	echo 'no tests yet'

.PHONY: build test
