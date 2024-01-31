# CI/CD Images

These images are intended to be used as building blocks for CI/CD tasks.
As such, they include common tools and helper binaries that make it easy to perform
common operations, like building source code or issuing `gcloud` commands.

Preview images are available at `us-docker.pkg.dev/gcb-catalog-release/preview`.
Note that these images are under early development and may change during the preview period without notice.

To file issues and feature requests against these images, [create an issue in this repo](https://github.com/GoogleCloudBuild/cicd-images/issues/new).

# Development

Each directory in this repository hold the source for one or more images. At minimum, the
directory will have a `Dockerfile`, though it may also include other files (e.g `go` source files for additional helper binaries included on the image).

We use Docker's [bake](https://docs.docker.com/build/bake/) to configure build of images.
This allows us to describe the build for each image declaratively, but also gives us the ability
to [specify that images have dependencies on other images](https://docs.docker.com/build/bake/build-contexts/#using-a-result-of-one-target-as-a-base-image-in-another-target). The latter feature gives us the ability to detect breakages in child images that were actually caused
in a parent image.

At minimum, a new directory should include a `Dockerfile` and `Makefile`. You will need to
also add the directory as a `SUBDIR` in the [top-level `Makefile`](./Makefile).

If this image should be released, it must also be added to the [top-level `cloudbuild.yaml`](./cloudbuild.yaml). Note there are a number of places the image reference should appear.

Locally, you can `cd` into any directory and run `make build` to build the images in that directory, and `make test` to test them. In some cases you may want to run `docker buildx bake ...<targets>` directly. This should be done only from the top-level directory.

There is a simple `./presubmit.sh` script that looks at changed files in the current commit, and iterates over directories covering these files
to run `make build test`.

## Known Issues

- Images from `marketplace.gcr.io` require authentication when building locally. Please see [here](https://cloud.google.com/artifact-registry/docs/docker/authentication#gcloud-helper) for details on configuring this authentication. In some cases, you may also need to `export BUILDKIT_NO_CLIENT_TOKEN=true` (though restarting docker appears to also subvert the issue). This is presumably an [issue with buildx](https://github.com/docker/buildx/issues/1613).
- Using a remote cache to improve builds currently is blocked on [this issue](https://github.com/docker/buildx/issues/2144).
