# CI/CD Images

These images can be used as building blocks for your Google Cloud CI/CD tasks.
Common tools and helper binaries that make it easy to perform
common operations like building source code, or issuing `gcloud` commands are included.

Preview images are available at `us-docker.pkg.dev/gcb-catalog-release/preview`.
Preview images are under early development and may change during the preview period without notice.

To file issues and feature requests against these images, [create an issue in this repo](https://github.com/GoogleCloudBuild/cicd-images/issues/new).

# Development

The top-level directory acts as a [standard Go module](https://go.dev/doc/modules/layout).
Command line binaries that are added in images are included in the [`cmd`](./cmd) directory.

## Images Directory

The [`images`](./images) directory contains the source for individual images. Each sub-directory
within the images directory holds the source for one or more images. At minimum, the sub-directory has a `Dockerfile`,
though it may also include other files (e.g helper scripts needed on the image).

We use Docker's [Bake](https://docs.docker.com/build/bake/) command to configure build of images.
Bake lets us describe the build for each image declaratively, and 
[specify that images have dependencies on other images](https://docs.docker.com/build/bake/build-contexts/#using-a-result-of-one-target-as-a-base-image-in-another-target).
Specifying dependencies with Bake gives us the ability to detect breakages in child images caused by a parent image.

A new directory must include a `Dockerfile` and a `Makefile`. You must
also add the new directory as a `SUBDIR` in the main images directory [`Makefile`](./images/Makefile).

To release an image, add it to the [`cloudbuild.yaml`](cloudbuild.yaml) file.

**Note**: there are a number of places a released image reference must appear to ensure that it is
built, scanned, and pushed appropriately.

To build locally, you can `cd` into any image directory and run `make build` to build the images in that directory,
and `make test` to test them. If you want to run `docker buildx bake ...<targets>` directly,
then only run it from the `images` directory.

# Testing

For testing changes to Go binaries, you can simply run `go test ./...`, which will run
all the package tests in the repository.

For testing changes to image, each image directory should have a `Makefile` that is
responsible for building and testing the image. There is a simple `./presubmit.sh`
script that looks at changed image directories and runs `make build test` for each one.

**Note**: the `./presubmit.sh` script assumes changes have been committed, and won't
trigger for changes only to go binaries the image includes.

## Known Issues

- Images from `marketplace.gcr.io` require authentication when building locally. For details on how
  to configure authentication, see
  [gcloud CLI credential helper](https://cloud.google.com/artifact-registry/docs/docker/authentication#gcloud-helper).
 
  In some cases, you might also need to run `export BUILDKIT_NO_CLIENT_TOKEN=true` (though restarting docker appears to also subvert the issue).
  This is presumably an [issue with buildx](https://github.com/docker/buildx/issues/1613).
- Using a remote cache to improve builds is blocked on [GitHub Docker/buildx issue 2144](https://github.com/docker/buildx/issues/2144).
- Building all images might exhaust some system resources. If working on cross-cutting changes that require
  re-building multiple images, you might need to periodically run `docker buildx prune`.
