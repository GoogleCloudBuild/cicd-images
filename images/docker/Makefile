REGISTRY ?= us-docker.pkg.dev/gcb-catalog-release/catalog
TAG ?= debian12

build:
	@echo "+ building docker images"
	cd .. && docker buildx bake docker-cli docker-dind

test:
	container-structure-test test --image ${REGISTRY}/docker/dind:${TAG} --config test_config_dind.yaml
	container-structure-test test --image ${REGISTRY}/docker/cli:${TAG} --config test_config_cli.yaml

.PHONY: build test