package publish

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/GoogleCloudBuild/cicd-images/cmd/maven-steps/publish/xmlmodules"
	"github.com/GoogleCloudBuild/cicd-images/internal/helper"
	"github.com/GoogleCloudBuild/cicd-images/internal/logger"
	"github.com/spf13/pflag"
)

type Arguments struct {
	Repository      string
	ArtifactPath    string
	ArtifactID      string
	GroupID         string
	Version         string
	Verbose         bool
	IsBuildArtifact string
	ResultsPath     string
}

type Provenance struct {
	URI             string `json:"uri"`
	Digest          string `json:"digest"`
	IsBuildArtifact string `json:"isBuildArtifact"`
}

const (
	repoID = "remote-repository"
)

// ParseArgs parses the arguments passed to the publish command.
func ParseArgs(f *pflag.FlagSet) (Arguments, error) {
	repository, err := f.GetString("repository")
	if err != nil {
		return Arguments{}, fmt.Errorf("failed to get repository: %w", err)
	}
	artifactPath, err := f.GetString("artifactPath")
	if err != nil {
		return Arguments{}, fmt.Errorf("failed to get artifactPath: %w", err)
	}
	artifactID, err := f.GetString("artifactId")
	if err != nil {
		return Arguments{}, fmt.Errorf("failed to get artifactId: %w", err)
	}
	groupID, err := f.GetString("groupId")
	if err != nil {
		return Arguments{}, fmt.Errorf("failed to get groupId: %w", err)
	}
	version, err := f.GetString("version")
	if err != nil {
		return Arguments{}, fmt.Errorf("failed to get version: %w", err)
	}
	verbose, err := f.GetBool("verbose")
	if err != nil {
		return Arguments{}, fmt.Errorf("failed to get verbose: %w", err)
	}
	isBuildArtifact, err := f.GetString("isBuildArtifact")
	if err != nil {
		return Arguments{}, fmt.Errorf("failed to get isBuildArtifact: %w", err)
	}
	resultsPath, err := f.GetString("resultsPath")
	if err != nil {
		return Arguments{}, fmt.Errorf("failed to get resultsPath: %w", err)
	}

	return Arguments{
		Repository:      repository,
		ArtifactPath:    artifactPath,
		ArtifactID:      artifactID,
		GroupID:         groupID,
		Version:         version,
		Verbose:         verbose,
		IsBuildArtifact: isBuildArtifact,
		ResultsPath:     resultsPath,
	}, nil
}

// Execute is the entrypoint for the publish command execution.
func Execute(ctx context.Context, args Arguments) error {
	logger.SetupLogger(args.Verbose)

	slog.Info("Generating settings.xml file")
	token, err := helper.GetAccessToken(ctx)
	if err != nil {
		return fmt.Errorf("error getting authentication token: %w", err)
	}

	file, err := os.CreateTemp("", "settings*.xml")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(file.Name())

	if err := writeSettingsXML(token, repoID, file.Name()); err != nil {
		return err
	}
	slog.Info("Successfully generated settings.xml file")

	slog.Info("Publishing maven artifact")
	if err := publishArtifact(file.Name(), args); err != nil {
		return err
	}
	slog.Info("Successfully published maven artifact")

	slog.Info("Generating provenance")
	if err := generateProvenance(ctx, args); err != nil {
		return fmt.Errorf("error generating provenance: %w", err)
	}
	slog.Info("Successfully generated provenance")

	return nil
}

// Generate provenance for the maven artifact published to Artifact Registry.
func generateProvenance(ctx context.Context, args Arguments) error {
	artifactGroupID := strings.ReplaceAll(args.GroupID, ".", "/")
	artifactName := filepath.Base(args.ArtifactPath)

	uri := fmt.Sprintf("%s/%s/%s/%s/%s", args.Repository, artifactGroupID, args.ArtifactID, args.Version, artifactName)

	digest, err := getCheckSum(ctx, uri)
	if err != nil {
		return fmt.Errorf("error generating checksum: %w", err)
	}

	provenance := &Provenance{
		URI:             strings.TrimSpace(uri),
		Digest:          strings.TrimSpace(digest),
		IsBuildArtifact: strings.TrimSpace(args.IsBuildArtifact),
	}

	file, err := json.Marshal(provenance)
	if err != nil {
		return fmt.Errorf("error marshaling json %v: %w", provenance, err)
	}
	// Write provenance as json file in path
	if err := os.WriteFile(args.ResultsPath, file, 0o600); err != nil {
		return fmt.Errorf("error writing results into %s: %w", args.ResultsPath, err)
	}
	return nil
}

// Publish maven artifact to Artifact Registry.
func publishArtifact(settingsPath string, args Arguments) error {
	runner := &helper.DefaultCommandRunner{}

	command := fmt.Sprintf(`deploy:deploy-file --settings %s -DrepositoryId=%s -Durl=%s -Dfile=%s -DartifactId=%s -DgroupId=%s -Dversion=%s`,
		settingsPath, repoID, args.Repository, args.ArtifactPath, args.ArtifactID, args.GroupID, args.Version)

	if err := runner.Run("mvn", strings.Fields(command)...); err != nil {
		return fmt.Errorf("failed to deploy maven artifacts: %w", err)
	}

	return nil
}

// Create settings.xml with authentication token.
func writeSettingsXML(token, repoID, settingsPath string) error {
	settings := xmlmodules.Settings{
		Servers: xmlmodules.Servers{},
	}

	server := xmlmodules.Server{
		ID: repoID,
		Configuration: xmlmodules.Configuration{
			HTTPConfiguration: xmlmodules.HTTPConfiguration{
				Get:  true,
				Head: true,
				Put: xmlmodules.PutParams{
					Property: []xmlmodules.Property{
						{
							Name:  "http.protocol.expect-continue",
							Value: "false",
						},
					},
				},
			},
		},
		// Authentication token reference: https://cloud.google.com/artifact-registry/docs/helm/authentication#token
		Username: "oauth2accesstoken",
		Password: token,
	}
	settings.Servers.Server = append(settings.Servers.Server, server)

	file, err := os.OpenFile(settingsPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return fmt.Errorf("error opening settings.xml: %w", err)
	}
	defer file.Close()
	// Write the XML data to a file
	encoder := xml.NewEncoder(file)
	encoder.Indent("", "  ")
	err = encoder.Encode(settings)
	if err != nil {
		return fmt.Errorf("error writing to settings.xml file: %w", err)
	}

	return nil
}

// Get SHA1 checksum of remote artifact in Artifact Registry.
func getCheckSum(ctx context.Context, artifactRegistryURL string) (string, error) {
	token, err := helper.GetAccessToken(ctx)
	if err != nil {
		return "", fmt.Errorf("error getting auth token: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s.sha1", artifactRegistryURL), nil)
	if err != nil {
		return "", fmt.Errorf("error in GET request: %w", err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("error downloading checksum: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response body: %w", err)
	}

	checksum := fmt.Sprintf("sha1:%s", string(body))
	return checksum, nil
}
