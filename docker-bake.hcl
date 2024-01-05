variable "REGISTRY" {
  default = "us-docker.pkg.dev/gcb-catalog-release/catalog"
}
variable "TAG" {
  default = "ubuntu22"
}

group "default" {
  targets = ["base", "tool-images", "toolchain-images"]
}

target "base" {
  dockerfile = "Dockerfile"
  context = "base"
  tags = ["${REGISTRY}/gcb-base:${TAG}"]
}

group "tool-images" {
    targets = [
      "docker-cli",
      "docker-dind",
      "gcloud",
      "git",
    ]
}

target "docker-cli" {
  dockerfile = "Dockerfile.cli"
  context = "docker"
  contexts = {
    base = "target:base"
  }
  tags = ["${REGISTRY}/docker/cli:${TAG}"]
}

target "docker-dind" {
  dockerfile = "Dockerfile.dind"
  context = "docker"
  contexts = {
    base = "target:docker-cli"
  }
  tags = ["${REGISTRY}/docker/dind:${TAG}"]
}

target "gcloud" {
    dockerfile = "Dockerfile"
    context = "gcloud"
    contexts = {
      base = "target:base"
    }
    tags = ["${REGISTRY}/gcloud:${TAG}"]
}

target "git" {
    dockerfile = "Dockerfile"
    context = "git"
    contexts = {
      base = "target:base"
    }
    tags = ["${REGISTRY}/git:${TAG}"]
}

target "go-base" {
  dockerfile = "Dockerfile.base"
  context = "go"
  contexts = {
    base = "target:base"
  }
  output = ["type=cacheonly"]
}

group "toolchain-images" {
    targets = [
      "go",
    ]
}

target "go" {
  dockerfile = "Dockerfile"
  context = "go"
  contexts = {
    base = "target:go-base"
  }
  tags = ["${REGISTRY}/go:${TAG}"]
}

