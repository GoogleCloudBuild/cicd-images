variable "REGISTRY" {
  default = "us-docker.pkg.dev/gcb-catalog-release/catalog"
}
variable "TAG" {
  default = "ubuntu22"
}

group "default" {
  targets = ["base", "tool-images"]
}

target "base" {
  dockerfile = "Dockerfile"
  context = "base"
  tags = ["${REGISTRY}/gcb-base:${TAG}"]
}

group "tool-images" {
    targets = [
      "docker-cli", 
      "docker-dind"
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

