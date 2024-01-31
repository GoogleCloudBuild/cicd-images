# syntax=docker/dockerfile:1

# hadolint ignore=DL3006
FROM base as auth-base

FROM google-go.pkg.dev/golang:1.21.5 as builder
WORKDIR /app
COPY . .
RUN go build -o cloud-deploy

# TODO(@zhangquan): use shared Catalog base image
# hadolint ignore=DL3007
FROM marketplace.gcr.io/google/ubuntu2204:latest
WORKDIR /app
COPY --from=builder /app/cloud-deploy /usr/local/bin/cloud-deploy
COPY --from=auth-base /usr/local/bin/google-cloud-auth /usr/local/bin/google-cloud-auth
CMD ["cloud-deploy"]

