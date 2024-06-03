// Copyright 2024 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package utils

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/GoogleCloudBuild/cicd-images/cmd/cloud-function/config"

	functions "cloud.google.com/go/functions/apiv2"
	"cloud.google.com/go/functions/apiv2/functionspb"
)

const (
	// Regex to validate Google Storage reference.
	// Shoud be in the format: gs://{bucket}/{path to object}.
	ccsSourceRegex        = "gs://(?P<bucket>[^/]+)/(?P<path>.+\\.zip)"
	ccsSourceErrorMessage = "invalid Cloud Storage URL provided"

	// Regex to validate Google Source Repository reference.
	// The minimal source repository URL is:
	// https://source.developers.google.com/projects/${PROJECT}/repos/${REPO}
	// To deploy from a revision different from master, the following three
	// sources can be added to the URL:
	// - /revisions/${REVISION},
	// - /moveable-aliases/${MOVEABLE_ALIAS},
	// - /fixed-aliases/${FIXED_ALIAS}.
	// To deploy sources from a directory different from the root, a revision,
	// a moveable alias, or a fixed alias must be specified, as above, and
	// appended with /paths/${PATH_TO_SOURCES_DIRECTORY}.
	csrSourceRegex        = "https://source\\.developers\\.google\\.com/projects/(?P<project_id>[^/]+)/repos/(?P<repo_name>[^/]+)(((/revisions/(?P<commit>[^/]+))|(/moveable-aliases/(?P<branch>[^/]+))|(/fixed-aliases/(?P<tag>[^/]+)))(/paths/(?P<path>[^/]+))?)?/?$"
	csrSourceErrorMessage = "invalid Cloud Source Repository URL provided"
)

// BuildSourceDefinition creates a new Source object for a cloud function based on the
// provided DeployOptions.
func BuildSourceDefinition(ctx context.Context, client *functions.FunctionClient, fileUtil FileUtil, opts config.DeployOptions) (*functionspb.Source, error) {
	if strings.HasPrefix(opts.Source, "gs://") || strings.HasPrefix(opts.Source, "https://") {
		return nil, fmt.Errorf("source must be a local directory")
	}
	storageSource, err := packAndUploadSourceCodeToCloudBucket(ctx, client, fileUtil, opts)
	if err != nil {
		return nil, err
	}
	return &functionspb.Source{
		Source: &functionspb.Source_StorageSource{
			StorageSource: storageSource,
		},
	}, nil
}

// packAndUploadSourceCodeToCloudBucket creates a zip archive from the folder in the local filesystem
// specified in the source path, sends this archive to the Cloud Storage bucket and returns
// storage source object pointing to this archive in the bucket.
func packAndUploadSourceCodeToCloudBucket(ctx context.Context, client *functions.FunctionClient, fileUtil FileUtil, opts config.DeployOptions) (*functionspb.StorageSource, error) {
	source := opts.Source
	if source == "" {
		source = "."
	}
	zipFile, err := fileUtil.ArchiveDirectoryContentIntoZip(source)
	if err != nil {
		return nil, err
	}
	parent := fmt.Sprintf("projects/%s/locations/%s", opts.ProjectID, opts.Region)
	req := &functionspb.GenerateUploadUrlRequest{
		Parent: parent,
	}
	uploadURLResp, err := client.GenerateUploadUrl(ctx, req)
	if err != nil {
		return nil, err
	}
	uploadURL := uploadURLResp.GetUploadUrl()
	err = fileUtil.UploadFileToSignedURL(ctx, uploadURL, zipFile)
	if err != nil {
		return nil, err
	}
	err = fileUtil.CleanUp(zipFile)
	if err != nil {
		return nil, err
	}
	return uploadURLResp.GetStorageSource(), nil
}

// buildSourceGCS creates a storage source object from the source path provided,
// which must be in the format of gs://{bucket}/{path}.
func buildSourceGCS(source string) (*functionspb.Source_StorageSource, error) {
	regex := regexp.MustCompile(ccsSourceRegex)
	if !regex.MatchString(source) {
		return nil, fmt.Errorf(ccsSourceErrorMessage)
	}
	parts := regex.FindStringSubmatch(source)
	storageSource := &functionspb.StorageSource{
		Bucket: parts[regex.SubexpIndex("bucket")],
		Object: parts[regex.SubexpIndex("path")],
	}
	return &functionspb.Source_StorageSource{
		StorageSource: storageSource,
	}, nil
}

// buildSourceCSR creates a repo source object from the source path provided,
// which must be in the format of https://source.developers.google.com/projects/{project_id}/repos/{repo_name}/...
func buildSourceCSR(source string) (*functionspb.Source_RepoSource, error) {
	regex := regexp.MustCompile(csrSourceRegex)
	if !regex.MatchString(source) {
		return nil, fmt.Errorf(csrSourceErrorMessage)
	}
	parts := regex.FindStringSubmatch(source)
	params := make(map[string]string)
	for i, name := range regex.SubexpNames() {
		params[name] = parts[i]
	}
	repoSource := &functionspb.RepoSource{
		ProjectId: params["project_id"],
		RepoName:  params["repo_name"],
		Dir:       params["path"],
	}
	switch {
	case params["commit"] != "":
		repoSource.Revision = &functionspb.RepoSource_CommitSha{
			CommitSha: params["commit"],
		}
	case params["branch"] != "":
		repoSource.Revision = &functionspb.RepoSource_BranchName{
			BranchName: params["branch"],
		}
	case params["tag"] != "":
		repoSource.Revision = &functionspb.RepoSource_TagName{
			TagName: params["tag"],
		}
	default:
		repoSource.Revision = &functionspb.RepoSource_BranchName{
			BranchName: "master",
		}
	}
	return &functionspb.Source_RepoSource{
		RepoSource: repoSource,
	}, nil
}
