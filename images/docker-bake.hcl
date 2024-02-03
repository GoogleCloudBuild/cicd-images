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
  tags = [
    "${REGISTRY}/gcb-base:${TAG}",
    "${REGISTRY}/gcb-base:latest"
  ]
}

group "tool-images" {
    targets = [
      "docker-cli",
      "docker-dind",
      "gcloud",
      "git",
      "syft",
      "cloud-deploy",
    ]
}

target "docker-cli" {
  dockerfile = "Dockerfile.cli"
  context = "docker"
  contexts = {
    base = "target:base"
  }
  tags = [
    "${REGISTRY}/docker/cli:${TAG}",
    "${REGISTRY}/docker/cli:latest",
  ]
}

target "docker-dind" {
  dockerfile = "Dockerfile.dind"
  context = "docker"
  contexts = {
    base = "target:docker-cli"
  }
  tags = [
    "${REGISTRY}/docker/dind:${TAG}",
    "${REGISTRY}/docker/dind:latest"
  ]
}

target "gcloud" {
    dockerfile = "Dockerfile"
    context = "gcloud"
    contexts = {
      base = "target:base"
    }
    tags = [
      "${REGISTRY}/gcloud:${TAG}",
      "${REGISTRY}/gcloud:latest"
    ]
}

target "git" {
    dockerfile = "Dockerfile"
    context = "git"
    contexts = {
      base = "target:base"
    }
    tags = [
      "${REGISTRY}/git:${TAG}",
      "${REGISTRY}/git:latest",
    ]
}

target "syft" {
    dockerfile = "Dockerfile"
    context = "syft"
    contexts = {
      base = "target:base"
    }
    tags = [
      "${REGISTRY}/syft:${TAG}",
      "${REGISTRY}/syft:latest"
    ]
}

group "toolchain-images" {
    targets = [
      "go",
      "nodejs",
      "python",
      "openjdk",
    ]
}

target "go-base" {
  dockerfile = "Dockerfile.base"
  context = "go"
  contexts = {
    base = "target:base"
  }
  output = ["type=cacheonly"]
}

target "go" {
  dockerfile = "Dockerfile"
  context = "go"
  contexts = {
    base = "target:go-base"
    src = "../"
  }
  tags = [
    "${REGISTRY}/go:${TAG}",
    "${REGISTRY}/go:latest"
  ]
}

target "cloud-deploy" {
  dockerfile = "Dockerfile"
  context = "cloud-deploy"
  contexts = {
    base = "target:base"
    src = "../"
  }
  tags = [
    "${REGISTRY}/cloud-deploy:${TAG}",
    "${REGISTRY}/cloud-deploy:latest"
  ]
}

target "nodejs-base" {
  dockerfile = "Dockerfile.base"
  context = "nodejs"
  contexts = {
    base = "target:base"
  }
  output = ["type=cacheonly"]
}

target "nodejs" {
  dockerfile = "Dockerfile"
  context = "nodejs"
  contexts = {
    base = "target:nodejs-base"
    src = "../"
  }
  tags = [
    "${REGISTRY}/nodejs:${TAG}",
    "${REGISTRY}/nodejs:latest"
  ]
}

target "openjdk-base" {
  dockerfile = "Dockerfile.base"
  context = "openjdk"
  contexts = {
    base = "target:base"
  }
  output = ["type=cacheonly"]
}

target "openjdk" {
  dockerfile = "Dockerfile"
  context = "openjdk"
  contexts = {
    base = "target:openjdk-base"
    src = "../"
  }
  tags = [
    "${REGISTRY}/openjdk:${TAG}",
    "${REGISTRY}/openjdk:latest"
  ]
}

target "python-base" {
  dockerfile = "Dockerfile.base"
  context = "python"
  contexts = {
    base = "target:base"
  } 
  output = ["type=cacheonly"]
}

target "python" {
  dockerfile = "Dockerfile"
  context = "python"
  contexts = {
    base = "target:python-base"
    src = "../"
  }
  tags = [
    "${REGISTRY}/python:${TAG}",
    "${REGISTRY}/python:latest"
  ]
}

