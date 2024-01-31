# syntax=docker/dockerfile:1

FROM google-go.pkg.dev/golang:1.21.5 as builder

RUN go install github.com/google/go-licenses@latest

WORKDIR /app
COPY . .

RUN go mod download && \
    go test ./... -timeout 30s && \
    go build -o google-cloud-auth

# Scan and save licenses.
RUN go-licenses check . && \
    go-licenses save . --save_path=/THIRD_PARTY_NOTICES/google-cloud-auth && \
    go-licenses report . > /THIRD_PARTY_NOTICES/google-cloud-auth/licenses.csv

# hadolint ignore=DL3007
FROM marketplace.gcr.io/google/ubuntu2204:latest
WORKDIR /app
COPY --from=builder /app/google-cloud-auth /usr/local/bin/google-cloud-auth
CMD ["google-cloud-auth"]
