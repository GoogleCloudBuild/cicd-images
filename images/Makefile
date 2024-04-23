REGISTRY ?= us-docker.pkg.dev/gcb-catalog-release/catalog
TAG ?= ubuntu22
SUBDIRS = base cloud-deploy cloud-storage cloud-run docker gar-upload gcloud git gke-deploy go nodejs openjdk python syft builder

build:
	docker buildx bake

test: $(SUBDIRS)

$(SUBDIRS):
	$(MAKE) -C $@ test

.PHONY: build test $(SUBDIRS)

