//go:build tools
// +build tools

package tools

import (
	_ "github.com/GoogleCloudPlatform/docker-credential-gcr/v2"
	_ "github.com/anchore/syft/cmd/syft"
	_ "github.com/google/go-licenses"
)
