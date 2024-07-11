REGISTRY ?= us-docker.pkg.dev/gcb-catalog-release/catalog
TAG ?= ubuntu24

build:
	@echo "+ building gke-deploy image for component"
	cd .. && docker buildx bake gke-deploy

test:
	# TODO (chitrangpatel): update tests
	echo 'no tests yet'

.PHONY: build test
