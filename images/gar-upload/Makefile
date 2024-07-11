REGISTRY ?= us-docker.pkg.dev/gcb-catalog-release/catalog
TAG ?= ubuntu24

build:
	@echo "+ building gar-upload image"
	cd .. && docker buildx bake gar-upload

test:
	# TODO(yongxuanzhang): update tests
	echo 'no tests yet'

.PHONY: build test
