REGISTRY ?= us-docker.pkg.dev/gcb-catalog-release/catalog
TAG ?= ubuntu24

build:
	@echo "+ building cloud-run image"
	cd .. && docker buildx bake cloud-run

test:
	echo 'no tests yet'

.PHONY: build test
