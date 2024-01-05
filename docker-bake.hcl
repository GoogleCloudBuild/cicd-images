variable "REGISTRY" {
  default = "us-docker.pkg.dev/gcb-catalog-release/catalog"
}
variable "TAG" {
  default = "ubuntu22"
}

group "default" {
  targets = ["base"]
}

target "base" {
  dockerfile = "Dockerfile"
  context = "base"
  tags = ["${REGISTRY}/gcb-base:${TAG}"]
}

