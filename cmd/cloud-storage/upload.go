// Copyright 2024 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"github.com/GoogleCloudBuild/cicd-images/cmd/cloud-storage/pkg/upload"
	"github.com/googleapis/google-cloud-go-testing/storage/stiface"
	"github.com/spf13/cobra"
	"google.golang.org/api/option"
)

var (
	path          string
	destination   string
	projID        string
	headers       string
	parsedHeaders map[string]string
	predefinedACL string
	glob          string
	useIgnoreList bool
	useGzip       bool
	concurrency   int
	includeParent bool
	userAgent     string
)

// uploadCmd represents the upload command
var uploadCmd = &cobra.Command{
	Use:   "upload",
	Short: "Upload a file to GCS",
	PreRunE: func(_ *cobra.Command, _ []string) error {
		var err error
		parsedHeaders, err = convertHeaderStringToMap(headers)
		if err != nil {
			return err
		}

		err = validateHeaders(parsedHeaders)
		if err != nil {
			return err
		}

		err = validatePredefinedACL(predefinedACL)

		return err
	},
	RunE: func(_ *cobra.Command, _ []string) error {
		ctx := context.Background()
		client, err := storage.NewClient(ctx, option.WithUserAgent(userAgent), option.WithQuotaProject(projID))
		if err != nil {
			return err
		}
		results, err := Upload(stiface.AdaptClient(client), destination, path, glob, predefinedACL, useIgnoreList,
			includeParent, useGzip, parsedHeaders, concurrency)
		if err != nil {
			return err
		}

		// print out a summary of successful/failed uploads and a unique list of errors encountered
		successCount, failCount, uniqueErrors := summarizeResults(results)
		fmt.Printf("Attempted upload of %d files\n", len(results))
		fmt.Printf("Successful upload count: %d\n", successCount)
		if failCount > 0 {
			fmt.Printf("Failed upload count: %d\n", failCount)
			fmt.Println("Unique errors encountered:")
			for k := range uniqueErrors {
				fmt.Println(k)
			}
		}
		return nil
	},
}

func Upload(client stiface.Client, destination, path, glob, predefinedACL string, useIgnoreList, includeParent,
	useGzip bool, parsedHeaders map[string]string, concurrency int) ([]upload.UploadResults, error) {
	bucket, prefix, _ := strings.Cut(destination, "/")
	fmt.Printf("Parsed destination into bucketName: %v and prefix: %v\n", bucket, prefix)
	ui, err := upload.ProcessPath(path, prefix, glob, useIgnoreList, includeParent)
	if err != nil {
		return nil, err
	}
	u := upload.Uploader{
		Gzip:        useGzip,
		Concurrency: concurrency,
		ACL:         predefinedACL,
		Headers:     parsedHeaders,
	}

	results := u.UploadObjects(ui, client.Bucket(bucket))
	return results, nil
}

const ConcurrencyDefault = 100
const PredefinedACLList = "'authenticatedRead', 'bucketOwnerFullControl', 'bucketOwnerRead', 'private', 'projectPrivate', 'publicRead'"

// nolint: gochecknoinits
func init() {
	rootCmd.AddCommand(uploadCmd)

	uploadCmd.PersistentFlags().StringVarP(&path, "path", "f", "", "Path to file or folder for upload. If the path is a "+
		"folder, an optional glob param can also be passed in.")
	uploadCmd.PersistentFlags().StringVarP(&destination, "destination", "d", "", "Name of the bucket destination. Can be "+
		"in the format of either <bucketName> or <bucketName>/<prefix>. If a prefix is provided all objects uploaded "+
		"will use the prefix in the object name.")
	uploadCmd.PersistentFlags().StringVarP(&projID, "project-id", "p", "", "The google cloud project ID that will be used "+
		"for quota or billing purposes. If set the caller must have 'serviceusage.services.use' permissions.")
	uploadCmd.PersistentFlags().StringVarP(&headers, "metadata-headers", "m", "", "Metadata headers to include with "+
		"every GCS Object upload. Headers must be provided as a list of key/value pairs (i.e. 'x-goog-meta-foo=bar,"+
		"x-goog-meta-me=foobar'. Settable fields are: cache-control, content-disposition, content-encoding, "+
		"content-language, content-type, custom-time, or a custom metadata field which must be prefixed with "+
		"x-goog-meta-.")
	uploadCmd.PersistentFlags().StringVarP(&predefinedACL, "acl", "a", "", "Apply a predefined set of access "+
		"controls to the file(s). Acceptable values are one of: "+PredefinedACLList)
	uploadCmd.PersistentFlags().StringVarP(&glob, "glob", "g", "", "Glob pattern to search for within the "+
		"path parameter when path is a folder")
	uploadCmd.PersistentFlags().BoolVarP(&useIgnoreList, "ignore-list", "i", true, "Processes the .gcloudignore "+
		"file present in the top-level of the repository. If true, the file is parsed and any filepaths that match are "+
		"not uploaded to the storage bucket. Defaults to true.")
	uploadCmd.PersistentFlags().BoolVarP(&useGzip, "gzip", "z", true, "Gzip files uploaded, defaults to true. This "+
		"will override the 'content-encoding' header to have the value of gzip, and will leave all other user provided "+
		"headers as-is.")
	uploadCmd.PersistentFlags().IntVarP(&concurrency, "concurrency", "c", ConcurrencyDefault, "Number of files to "+
		"simultaneously upload, defaults to 100.")
	uploadCmd.PersistentFlags().BoolVarP(&includeParent, "parent", "x", true, "Whether the base dir of the path "+
		"parameter is included in the GCS object name.")
	uploadCmd.PersistentFlags().StringVarP(&userAgent, "google-apis-user-agent", "u", "", "The user-agent to be "+
		"applied when calling Google APIs")

	_ = uploadCmd.MarkPersistentFlagRequired("destination")
	_ = uploadCmd.MarkPersistentFlagRequired("path")
}

func summarizeResults(results []upload.UploadResults) (successCount, failCount int, uniqueErrors map[string]bool) {
	uniqueErrors = map[string]bool{}
	for _, result := range results {
		if result.Success {
			successCount += 1
		} else {
			failCount += 1
			uniqueErrors[result.Message] = true
		}
	}
	return successCount, failCount, uniqueErrors
}

func convertHeaderStringToMap(val string) (map[string]string, error) {
	out := map[string]string{}
	r := csv.NewReader(strings.NewReader(val))
	ss, err := r.Read()
	switch {
	case errors.Is(err, io.EOF):
		return out, nil
	case err != nil:
		return nil, err
	}
	for _, pair := range ss {
		kv := strings.SplitN(pair, "=", 2)
		if len(kv) != 2 {
			return nil, fmt.Errorf("%s is not formatted as key=value", pair)
		}
		out[kv[0]] = kv[1]
	}
	return out, nil
}

func validatePredefinedACL(acl string) error {
	if acl == "" {
		return nil
	}

	knownAcls := map[string]bool{
		"authenticatedRead":      true,
		"bucketOwnerFullControl": true,
		"bucketOwnerRead":        true,
		"private":                true,
		"projectPrivate":         true,
		"publicRead":             true,
	}

	if _, ok := knownAcls[acl]; !ok {
		return fmt.Errorf("unknown predefined acl provided: %v. Must be one of "+PredefinedACLList, acl)
	}

	return nil
}

func validateHeaders(headers map[string]string) error {
	knownHeaders := upload.GetKnownHeaders()

	// validate custom time is in RFC3339 format
	if customTime, ok := headers["custom-time"]; ok {
		_, err := time.Parse(time.RFC3339, customTime)
		if err != nil {
			return fmt.Errorf("failed to parse custom-time: %w", err)
		}
	}

	// validate each of the provided headers
	// must either be in the acceptableHeaders map, or be prefixed with 'x-goog-meta-'
	for key, value := range headers {
		if _, ok := knownHeaders[key]; ok {
			continue
		}

		if !strings.HasPrefix(key, "x-goog-meta-") {
			return fmt.Errorf("invalid header provided: %s=%s. Must be either a known header with keys 'content-type'"+
				", 'cache-control', 'content-disposition', 'content-encoding', 'content-language', 'custom-time', or it "+
				"must be a header with a key prefix of 'x-goog-meta-'", key, value)
		}
	}
	return nil
}
