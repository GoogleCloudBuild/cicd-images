REGISTRY ?= us-docker.pkg.dev/gcb-catalog-release/catalog
TAG ?= ubuntu22
SUBDIRS = base app-engine cloud-deploy cloud-storage cloud-run cloud-function docker gar-upload git-steps gke-deploy go nodejs-steps maven-steps python-steps builder

build:
	docker buildx bake

test: $(SUBDIRS)

$(SUBDIRS):
	$(MAKE) -C $@ test

.PHONY: build test $(SUBDIRS)
