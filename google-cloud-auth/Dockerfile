# syntax=docker/dockerfile:1

# TODO(@zhangquan): use shared Catalog base image
FROM google-go.pkg.dev/golang:1.21.5 as builder
WORKDIR /app
COPY . .
RUN go build -o google-cloud-auth

# hadolint ignore=DL3007
FROM marketplace.gcr.io/google/ubuntu2204:latest
WORKDIR /app
COPY --from=builder /app/google-cloud-auth /usr/local/bin/google-cloud-auth
CMD ["google-cloud-auth"]
