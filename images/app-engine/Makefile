REGISTRY ?= us-docker.pkg.dev/gcb-catalog-release/catalog
TAG ?= ubuntu24

build:
	@echo "+ building app-engine image"
	cd .. && docker buildx bake app-engine

test:
	echo 'no tests yet'

.PHONY: build test
