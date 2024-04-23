# Omnibus `builder` image

This dir contains source for a generic builder image which contains almost all
the tools that are supported by the Google Cloud Build.

Following tools and versions are pre-installed on the image. A number of
additional versions for some tools are also supported.

Language Runtime:
  - Java (17)
    * Available versions: 11, 17, 21
  - Python (3.10)
    * Available versions: 3.9, 3.11, 3.12
  - NodeJS (18)
      * Available versions: 16, 18, 20
  - Go (1.21)
      * Available versions: 1.19, 1.20, 1.21, 1.22

Build tools:
  - Docker (25.x)
    * containerd (1.x)
    * docker-buildx (0.x)
    * docker-compose (2.x)
  - Git (2.x) along with git-lfs (3.x)
  - Maven (3.9)
    * version 3.8 is available
  - Gradle (8.7)
    * version 8.6 is available
  - Github CLI
  - Google Cloud CLI (gcloud)
    * latest version (released weekly()

Custom `go` binaries:
  - gke-deploy
  - gcs-uploader
  - gcs-fetcher
  - go-license
  - [yq (v4)](github.com/mikefarah/yq)
  - [syft](github.com/anchore/syft)

In addition following OS packages are also installed from Ubuntu 22.04:
  - bash
  - gnupg
  - curl
  - wget
  - openssh-client
  - gawk
  - curl
  - jq
  - wget
  - sudo
  - gnupg-agent
  - ca-certificates
  - software-properties-common
  - apt-transport-https
  - autoconf
  - automake
  - dpkg
  - fakeroot
  - gnupg
  - binutils
  - coreutils
  - file
  - findutils
  - iproute2
  - netcat
  - shellcheck
  - sudo
  - parallel
  - net-tools
  - bind9-dnsutils
  - libyaml-0-2
  - zstd
  - zip
  - unzip
  - xz-utils
  - bzip2
  - iputils-ping
  - locales
  - p7zip-rar
  - texinfo
  - tzdata
  - xz-utils
  - zsync
  - aria2
  - brotli
  - haveged
  - lz4
  - m4
  - mediainfo
  - p7zip-full
  - pigz
  - pollinate
  - rsync
  - time
