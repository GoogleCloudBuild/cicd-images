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

package upload

import (
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/googleapis/google-cloud-go-testing/storage/stiface"
)

type Uploader struct {
	Gzip        bool
	Concurrency int
	ACL         string
	Headers     map[string]string
}

type UploadInput struct {
	FilePath   string
	ObjectName string
}

type UploadResults struct {
	FilePath   string `json:"file_path"`
	ObjectName string `json:"object_name"`
	Success    bool   `json:"success"`
	Message    string `json:"message,omitempty"`
}

// UploadObjects will upload each UploadInput into the provided bucket. It will perform the uploads
// concurrently up to the Uploader.Concurrency limit. Results are accumulated into a UploadResults
// array and returned when everything completes. It's possible that individual uploads can have an
// error, and the caller can check UploadResults.Success and UploadResults.Message for status and
// any error message
func (u Uploader) UploadObjects(inputs []UploadInput, bucket stiface.BucketHandle) []UploadResults {
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, u.Concurrency)
	results := make([]UploadResults, len(inputs))

	for i, input := range inputs {
		semaphore <- struct{}{}
		wg.Add(1)

		go func(i int, input UploadInput, bucket stiface.BucketHandle) {
			defer wg.Done()
			defer func() { <-semaphore }()
			results[i] = u.uploadFile(input, bucket)
		}(i, input, bucket)
	}

	wg.Wait()
	return results
}

const Timeout = 50
const GzipContentEncoding = "gzip"

// uploadFile will upload the file from the UploadInput into the provided bucket.
// The ObjectName from within the UploadInput will be used as the ObjectName
// Headers, PredefinedACL and whether to gzip the file or not are all pulled
// from the Uploader struct
func (u Uploader) uploadFile(input UploadInput, bucket stiface.BucketHandle) UploadResults {
	result := UploadResults{FilePath: input.FilePath}
	ctx := context.Background()

	// open local file.
	f, err := os.Open(input.FilePath)
	if err != nil {
		result.Message = fmt.Sprintf("os.Open: %v", err)
		return result
	}
	defer f.Close()

	ctx, cancel := context.WithTimeout(ctx, time.Second*Timeout)
	defer cancel()

	o := bucket.Object(input.ObjectName)
	result.ObjectName = input.ObjectName

	// upload an object with storage.Writer.
	wc := o.NewWriter(ctx)

	// apply all provided headers and acl to the object
	applyHeaders(wc, u.Headers)
	applyPredefinedACL(wc, u.ACL)

	if u.Gzip {
		gw := gzip.NewWriter(wc)
		// override any provided content-encoding header to be gzip
		wc.ObjectAttrs().ContentEncoding = GzipContentEncoding

		if _, err := io.Copy(gw, f); err != nil {
			result.Message = fmt.Sprintf("io.Copy: %v", err)
			return result
		}

		if err := gw.Close(); err != nil {
			result.Message = fmt.Sprintf("gzip.Writer.Close: %v", err)
			return result
		}
	} else {
		if _, err := io.Copy(wc, f); err != nil {
			result.Message = fmt.Sprintf("io.Copy: %v", err)
			return result
		}
	}

	if err := wc.Close(); err != nil {
		result.Message = fmt.Sprintf("Writer.Close: %v", err)
		return result
	}
	result.Success = true
	return result
}

func applyPredefinedACL(wc stiface.Writer, acl string) {
	if acl == "" {
		return
	}

	wc.ObjectAttrs().PredefinedACL = acl
}

var knownHeaders = map[string]bool{
	"content-type":        true,
	"cache-control":       true,
	"content-disposition": true,
	"content-encoding":    true,
	"content-language":    true,
	"custom-time":         true,
}

func GetKnownHeaders() map[string]bool {
	return knownHeaders
}

func applyHeaders(wc stiface.Writer, headers map[string]string) {
	// initialize map if it is nil
	if wc.ObjectAttrs().Metadata == nil {
		wc.ObjectAttrs().Metadata = make(map[string]string)
	}

	if contentType, ok := headers["content-type"]; ok {
		wc.ObjectAttrs().ContentType = contentType
	}
	if cacheControl, ok := headers["cache-control"]; ok {
		wc.ObjectAttrs().CacheControl = cacheControl
	}
	if contentDisposition, ok := headers["content-disposition"]; ok {
		wc.ObjectAttrs().ContentDisposition = contentDisposition
	}
	if contentEncoding, ok := headers["content-encoding"]; ok {
		wc.ObjectAttrs().ContentEncoding = contentEncoding
	}
	if contentLanguage, ok := headers["content-language"]; ok {
		wc.ObjectAttrs().ContentLanguage = contentLanguage
	}
	if customTime, ok := headers["custom-time"]; ok {
		t, _ := time.Parse(time.RFC3339, customTime)
		wc.ObjectAttrs().CustomTime = t
	}

	// iterate over all the headers, skipping known headers and applying all metadata headers
	for key, value := range headers {
		if _, ok := knownHeaders[key]; ok {
			continue
		}

		wc.ObjectAttrs().Metadata[key] = value
	}
}
