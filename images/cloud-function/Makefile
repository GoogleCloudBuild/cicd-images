REGISTRY ?= us-docker.pkg.dev/gcb-catalog-release/catalog
TAG ?= ubuntu24

build:
	@echo "+ building cloud-function image"
	cd .. && docker buildx bake cloud-function

test:
	echo 'no tests yet'

.PHONY: build test
