REGISTRY ?= us-docker.pkg.dev/gcb-catalog-release/catalog
TAG ?= ubuntu22

build:
	@echo "+ building google-cloud-auth image"
	cd .. && docker buildx bake google-cloud-auth

test:
	# TODO(zhangquan): update tests
	echo 'no tests yet'

.PHONY: build test
