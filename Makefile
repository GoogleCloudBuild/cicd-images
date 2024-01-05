REGISTRY ?= us-docker.pkg.dev/gcb-catalog-release/catalog
TAG ?= ubuntu22
SUBDIRS = base docker gcloud git go nodejs openjdk python syft

build:
	docker buildx bake

test: $(SUBDIRS)

$(SUBDIRS):
	$(MAKE) -C $@ test

.PHONY: build test $(SUBDIRS)

