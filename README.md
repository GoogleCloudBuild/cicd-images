# CI/CD Images

These images are intended to be used as building blocks for CI/CD tasks.
As such, they include common tools and helper binaries that make it easy to perform
common operations, like building source code or issuing `gcloud` commands.

Preview images are available at `us-docker.pkg.dev/gcb-catalog-release/preview`.
Note that these images are under early development and may change during the preview period without notice.

To file issues and feature requests against these images, [create an issue in this repo](https://github.com/GoogleCloudBuild/cicd-images/issues/new).

# Development

The top-level directory acts as a [standard go module](https://go.dev/doc/modules/layout).
Command line binaries to be added in images are included in the [`cmd`](./cmd) directory.

## Images Directory

The [`images`](./images) directory contians the source for individual images. Each sub-directory
here holds the source for one or more images. At minimum, the sub-directory will have a `Dockerfile`,
though it may also include other files (e.g helper scripts needed on the image).

We use Docker's [bake](https://docs.docker.com/build/bake/) to configure build of images.
This allows us to describe the build for each image declaratively, but also gives us the ability
to [specify that images have dependencies on other images](https://docs.docker.com/build/bake/build-contexts/#using-a-result-of-one-target-as-a-base-image-in-another-target). The latter feature gives us the ability to detect breakages in child images that were actually caused in a parent image.

At minimum, a new directory should include a `Dockerfile` and `Makefile`. You will need to
also add the directory as a `SUBDIR` in the main [`Makefile`](./images/Makefile).

If this image should be released, it must also be added to the [`cloudbuild.yaml`](cloudbuild.yaml). Note there are a number of places the image reference should appear to ensure that it is
built, scanned and pushed appropriately.

Locally, you can `cd` into any image directory and run `make build` to build the images in that directory, and `make test` to test them. In some cases you may want to run `docker buildx bake ...<targets>` directly. This should be done only from the `images` directory.

# Testing

For testing changes to go binaries, you can simply run `go test ./...`, which will run
all the package tests in the repository.

For testing changes to image, each image directory should have a `Makefile` that is
responsible for building and testing the image. There is a simple `./presubmit.sh`
script that looks at changed image directories and runs `make build test` for each one.
Note that this script (1) assumes changes have been committed, and (2) will not
trigger for changes only to go binaries the image includes.

## Known Issues

- Images from `marketplace.gcr.io` require authentication when building locally. Please see [here](https://cloud.google.com/artifact-registry/docs/docker/authentication#gcloud-helper) for details on configuring this authentication. In some cases, you may also need to `export BUILDKIT_NO_CLIENT_TOKEN=true` (though restarting docker appears to also subvert the issue). This is presumably an [issue with buildx](https://github.com/docker/buildx/issues/1613).
- Using a remote cache to improve builds currently is blocked on [this issue](https://github.com/docker/buildx/issues/2144).
- Building all images may exhaust some system resources. If working on cross-cutting changes that require
  re-building multiple images, you may need to periodically run `docker buildx prune`.
