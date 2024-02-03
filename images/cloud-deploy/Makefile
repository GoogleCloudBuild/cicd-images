REGISTRY ?= us-docker.pkg.dev/gcb-catalog-release/catalog
TAG ?= ubuntu22

build:
	@echo "+ building cloud-deploy image"
	cd .. && docker buildx bake cloud-deploy

test:
	# TODO(zhangquan): update tests
	echo 'no tests yet'

.PHONY: build test
