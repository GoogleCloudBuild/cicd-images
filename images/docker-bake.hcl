variable "REGISTRY" {
  default = "us-docker.pkg.dev/gcb-catalog-release/catalog"
}
variable "TAG" {
  default = "ubuntu24"
}
variable "TAG_DEBIAN" {
  default = "debian12"
}

group "default" {
    targets = [
      "base",
      "app-engine",
      "docker-dind",
      "gar-upload",
      "gcb-internal",
      "git-steps",
      "gke-deploy",
      "cloud-deploy",
      "cloud-function",
      "cloud-storage",
      "cloud-run",
      "go-steps",
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

target "gcb-internal" {
  dockerfile= "Dockerfile"
  context = "gcb-internal"
  tags = [
    "${REGISTRY}/gcb-internal:${TAG_DEBIAN}",
    "${REGISTRY}/gcb-internal:latest"
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

target "docker-dind" {
  dockerfile = "Dockerfile"
  context = "docker"
  contexts = {
    base = "target:base"
  }
  tags = [
    "${REGISTRY}/docker/dind:deprecated-public-image-${TAG}",
    "${REGISTRY}/docker/dind:deprecated-public-image-latest"
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
      "${REGISTRY}/git-steps:deprecated-public-image-${TAG}",
      "${REGISTRY}/git-steps:deprecated-public-image-latest",
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

target "go-steps" {
  dockerfile = "Dockerfile"
  context = "go-steps"
  contexts = {
    base = "target:base"
    src = "../"
  }
  tags = [
    "${REGISTRY}/go-steps:deprecated-public-image-${TAG}",
    "${REGISTRY}/go-steps:deprecated-public-image-latest"
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
    "${REGISTRY}/nodejs-steps:deprecated-public-image-${TAG}",
    "${REGISTRY}/nodejs-steps:deprecated-public-image-latest"
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
    "${REGISTRY}/maven-steps:deprecated-public-image-${TAG}",
    "${REGISTRY}/maven-steps:deprecated-public-image-latest"
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
    "${REGISTRY}/python-steps:deprecated-public-image-${TAG}",
    "${REGISTRY}/python-steps:deprecated-public-image-latest"
  ]
}
