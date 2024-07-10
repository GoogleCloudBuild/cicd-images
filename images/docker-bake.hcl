variable "REGISTRY" {
  default = "us-docker.pkg.dev/gcb-catalog-release/catalog"
}
variable "TAG" {
  default = "ubuntu22"
}

group "default" {
    targets = [
      "base",
      "app-engine",
      "docker-cli",
      "docker-dind",
      "gar-upload",
      "git-steps",
      "gke-deploy",
      "syft",
      "cloud-deploy",
      "cloud-function",
      "cloud-storage",
      "cloud-run",
      "go",
      "nodejs-steps",
      "python-steps",
      "maven-steps",
      "builder"
    ]
}

target "base" {
  dockerfile = "Dockerfile"
  context = "base"
  tags = [
    "${REGISTRY}/gcb-base:${TAG}",
    "${REGISTRY}/gcb-base:latest"
  ]
}

target "docker-cli" {
  dockerfile = "Dockerfile.cli"
  context = "docker"
  tags = [
    "${REGISTRY}/docker/cli:debian12",
    "${REGISTRY}/docker/cli:latest",
  ]
}

target "docker-dind" {
  dockerfile = "Dockerfile.dind"
  context = "docker"
  tags = [
    "${REGISTRY}/docker/dind:debian12",
    "${REGISTRY}/docker/dind:latest"
  ]
}

target "git-steps" {
    dockerfile = "Dockerfile"
    context = "git-steps"
    contexts = {
      base = "target:base"
      src = "../"
    }
    tags = [
      "${REGISTRY}/git-steps:${TAG}",
      "${REGISTRY}/git-steps:latest",
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

target "app-engine" {
  dockerfile = "Dockerfile"
  context = "app-engine"
  contexts = {
    base = "target:base"
    src = "../"
  }
  tags = [
    "${REGISTRY}/app-engine:${TAG}",
    "${REGISTRY}/app-engine:latest"
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

target "cloud-function" {
  dockerfile = "Dockerfile"
  context = "cloud-function"
  contexts = {
    base = "target:base"
    src = "../"
  }
  tags = [
    "${REGISTRY}/cloud-function:${TAG}",
    "${REGISTRY}/cloud-function:latest"
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

target "nodejs-steps" {
  dockerfile = "Dockerfile"
  context = "nodejs-steps"
  contexts = {
    base = "target:base"
    src = "../"
  }
  tags = [
    "${REGISTRY}/nodejs-steps:${TAG}",
    "${REGISTRY}/nodejs-steps:latest"
  ]
}

target "maven-steps" {
  dockerfile = "Dockerfile"
  context = "maven-steps"
  contexts = {
    base = "target:base"
    src = "../"
  }
  tags = [
    "${REGISTRY}/maven-steps:${TAG}",
    "${REGISTRY}/maven-steps:latest"
  ]
}

target "python-steps" {
  dockerfile = "Dockerfile"
  context = "python-steps"
  contexts = {
    base = "target:base"
    src = "../"
  }
  tags = [
    "${REGISTRY}/python-steps:${TAG}",
    "${REGISTRY}/python-steps:latest"
  ]
}

target "builder" {
  dockerfile = "Dockerfile"
  context = "builder"
  contexts = {
    src = "../"
  }
  tags = [
    "${REGISTRY}/builder:debian12",
    "${REGISTRY}/builder:latest"
  ]
}
