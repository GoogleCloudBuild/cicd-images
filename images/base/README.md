# GCB Base Image

The GCB Base image is meant to provide a secure base image for tools that require external dependencies e.g. git, docker, maven. This is meant to be used by both Cloud Build V1 builder images and Cloud Build V2 Catalog Task images.

The Catalog base image adopts minimalist image practices from [the Kubernetes debian-base image](https://github.com/kubernetes/release/blob/master/images/build/debian-base/).

This image is based on the latest Ubuntu image from [GoogleContainerTools/base-images-docker](https://github.com/GoogleContainerTools/base-images-docker). It modifies its base image to remove a number of packages and files that are not essential for running a container. The resulting image is small and only contains core packages required for a container.

This image also enables building final images in a secure manner and makes it easy to follow best practices. It includes a `clean-install` script that simplifies installing `apt` packages that auto-clean after installation.
Following packages are removed from base image:

```
bash
e2fsprogs
libmount1
libsmartcols1
libblkid1
libss2
ncurses-base
ncurses-bin
```

A non-root user is also included in this image. Any image that uses Catalog Base image as their base image will automatically run as a non-root user named `non-root` with id of `1000`. If inheriting Dockerfile requires root privileges during image build (e.g. installing new packages) you will need to set the user to root before running commands that require root privileges.

```
USER root
...
# run commands that require root access
...
USER $USER # set container user to non-root
```
