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
      "gar-upload",
      "gcloud",
      "git",
      "gke-deploy",
      "syft",
      "cloud-deploy",
      "cloud-storage",
      "cloud-run"
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
      src = "../"
    }
    tags = [
      "${REGISTRY}/git:${TAG}",
      "${REGISTRY}/git:latest",
    ]
}

target "gke-deploy" {
    dockerfile = "Dockerfile"
    context = "gke-deploy"
    contexts = {
      base = "target:base"
      src = "../"
    }
    tags = [
      "${REGISTRY}/gke-deploy:${TAG}",
      "${REGISTRY}/gke-deploy:latest"
    ]
}

target "syft" {
    dockerfile = "Dockerfile"
    context = "syft"
    contexts = {
      base = "target:base"
      src = "../"
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

target "cloud-storage" {
  dockerfile = "Dockerfile"
  context = "cloud-storage"
  contexts = {
    base = "target:base"
    src = "../"
  }
  tags = [
    "${REGISTRY}/cloud-storage:${TAG}",
    "${REGISTRY}/cloud-storage:latest"
  ]
}

target "cloud-run" {
  dockerfile = "Dockerfile"
  context = "cloud-run"
  contexts = {
    base = "target:base"
    src = "../"
  }
  tags = [
    "${REGISTRY}/cloud-run:${TAG}",
    "${REGISTRY}/cloud-run:latest"
  ]
}

target "gar-upload" {
  dockerfile = "Dockerfile"
  context = "gar-upload"
  contexts = {
    base = "target:base"
    src = "../"
  }
  tags = [
    "${REGISTRY}/gar-upload:${TAG}",
    "${REGISTRY}/gar-upload:latest"
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

target "builder" {
  dockerfile = "Dockerfile"
  context = "builder"
  contexts = {
    base = "target:base"
    src = "../"
  }
  tags = [
    "${REGISTRY}/builder:${TAG}",
    "${REGISTRY}/builder:latest"
  ]
}
