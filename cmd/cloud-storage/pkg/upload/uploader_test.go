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
	"bytes"
	"compress/gzip"
	"context"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"cloud.google.com/go/storage"
	"github.com/google/go-cmp/cmp"
	"github.com/googleapis/google-cloud-go-testing/storage/stiface"
)

type bucketMock struct {
	stiface.BucketHandle
	oMock objectMock
}

type objectMock struct {
	stiface.ObjectHandle
	wMock *mockWriter
}

type mockWriter struct {
	stiface.Writer
	attrs           *storage.ObjectAttrs
	numBytesWritten int
	data            []byte
}

func (m bucketMock) Object(_ string) stiface.ObjectHandle {
	return m.oMock
}

func (m objectMock) NewWriter(_ context.Context) (w stiface.Writer) {
	return m.wMock
}

func (m *mockWriter) Close() error {
	return nil
}
func (m *mockWriter) Write(p []byte) (n int, err error) {
	m.data = append(m.data, p...)
	m.numBytesWritten += len(p)
	return len(p), nil
}

func (m mockWriter) ObjectAttrs() *storage.ObjectAttrs {
	return m.attrs
}

func setup(t *testing.T) (*Uploader, *storage.ObjectAttrs, *bucketMock, *mockWriter, []UploadInput) {
	t.Helper()

	c := Uploader{
		Gzip:        false,
		Concurrency: 1,
		ACL:         "",
		Headers:     map[string]string{},
	}

	a := storage.ObjectAttrs{}
	a.Metadata = map[string]string{}
	w := mockWriter{attrs: &a, numBytesWritten: 0, data: []byte{}}
	o := objectMock{wMock: &w}
	b := bucketMock{oMock: o}

	tempDir := t.TempDir()
	data := []byte("this is some example data")
	path := filepath.Join(tempDir, "testFile.txt")
	e := os.WriteFile(path, data, 0600)
	if e != nil {
		t.Fatalf("os.WriteFile(path, data, 0600) = %v", e)
	}
	ui := UploadInput{}
	ui.FilePath = path
	ui.ObjectName = path

	return &c, &a, &b, &w, []UploadInput{ui}
}

func TestGetKnownHeaders_Success(t *testing.T) {
	var want = map[string]bool{
		"content-type":        true,
		"cache-control":       true,
		"content-disposition": true,
		"content-encoding":    true,
		"content-language":    true,
		"custom-time":         true,
	}
	got := GetKnownHeaders()

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("GetKnownHeaders diff (-want +got):\n%s", diff)
	}
}

func TestUploadObjects_Success(t *testing.T) {
	c, _, b, _, ui := setup(t)

	want := []UploadResults{
		{
			FilePath:   ui[0].FilePath,
			ObjectName: ui[0].ObjectName,
			Success:    true,
			Message:    "",
		},
	}

	got := c.UploadObjects(ui, b)
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("UploadObjects(ui, b) diff (-want +got):\n%s", diff)
	}
}

func TestUploadObjects_Headers_Success(t *testing.T) {
	contentTypeVal := "plain/text"
	cacheControlVal := "no-cache"
	contentDispositionVal := "inline"
	contentEncodingVal := GzipContentEncoding
	contentLanguageVal := "en-US"
	customTimeVal := "2024-03-07T00:11:22Z"
	custom1Val := "foo"
	custom2Val := "bar"

	headers := map[string]string{
		"content-type":        contentTypeVal,
		"cache-control":       cacheControlVal,
		"content-disposition": contentDispositionVal,
		"content-encoding":    contentEncodingVal,
		"content-language":    contentLanguageVal,
		"custom-time":         customTimeVal,
		"x-goog-meta-custom1": custom1Val,
		"x-goog-meta-custom2": custom2Val,
	}

	u, a, b, _, ui := setup(t)
	u.Headers = headers

	want := []UploadResults{
		{
			FilePath:   ui[0].FilePath,
			ObjectName: ui[0].ObjectName,
			Success:    true,
			Message:    "",
		},
	}

	got := u.UploadObjects(ui, b)
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("UploadObjects(ui, b) diff (-want +got):\n%s", diff)
	}

	// Validate all the headers were properly in ObjectAttrs
	if a.ContentType != contentTypeVal {
		t.Errorf("ContentType doesn't match want: %v got: %v", contentTypeVal, a.ContentType)
	}
	if a.CacheControl != cacheControlVal {
		t.Errorf("CacheControl doesn't match want: %v got: %v", cacheControlVal, a.CacheControl)
	}
	if a.ContentDisposition != contentDispositionVal {
		t.Errorf("ContentDisposition doesn't match want: %v got: %v", contentDispositionVal, a.ContentDisposition)
	}
	if a.ContentEncoding != contentEncodingVal {
		t.Errorf("ContentEncoding doesn't match want: %v got: %v", contentEncodingVal, a.ContentEncoding)
	}
	if a.ContentLanguage != contentLanguageVal {
		t.Errorf("ContentLanguage doesn't match want: %v got: %v", contentLanguageVal, a.ContentLanguage)
	}
	customTime, _ := time.Parse(time.RFC3339, customTimeVal)
	if a.CustomTime != customTime {
		t.Errorf("CustomTime doesn't match want: %v got: %v", customTime, a.CustomTime)
	}
	if a.Metadata["x-goog-meta-custom1"] != custom1Val {
		t.Errorf("Metadata['x-goog-meta-custom1'] doesn't match want: %v got: %v", custom1Val, a.Metadata["x-goog-meta-custom1"])
	}
	if a.Metadata["x-goog-meta-custom2"] != custom2Val {
		t.Errorf("Metadata['x-goog-meta-custom2'] doesn't match want: %v got: %v", custom2Val, a.Metadata["x-goog-meta-custom2"])
	}
	if len(a.Metadata) != 2 {
		t.Errorf("len(a.Metadata) want: %d got: %d", 2, len(a.Metadata))
	}
}

func TestUploadObjects_PredefinedACL_Success(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "authenticatedRead"},
		{name: "bucketOwnerFullControl"},
		{name: "bucketOwnerRead"},
		{name: "private"},
		{name: "projectPrivate"},
		{name: "publicRead"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			u, a, b, _, ui := setup(t)
			u.ACL = test.name
			want := []UploadResults{
				{
					FilePath:   ui[0].FilePath,
					ObjectName: ui[0].ObjectName,
					Success:    true,
					Message:    "",
				},
			}

			got := u.UploadObjects(ui, b)
			if diff := cmp.Diff(want, got); diff != "" {
				t.Errorf("UploadObjects(ui, b) diff (-want +got):\n%s", diff)
			}

			if a.PredefinedACL != test.name {
				t.Errorf("a.PredefinedACL() = %v, want: %v", a.PredefinedACL, test.name)
			}
		})
	}
}

func TestUploadObjects_Gzip_Success(t *testing.T) {
	u, a, b, w, ui := setup(t)
	u.Gzip = true
	want := []UploadResults{
		{
			FilePath:   ui[0].FilePath,
			ObjectName: ui[0].ObjectName,
			Success:    true,
			Message:    "",
		},
	}

	got := u.UploadObjects(ui, b)
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("UploadObjects(ui, b) diff (-want +got):\n%s", diff)
	}

	f, _ := os.Open(ui[0].FilePath)

	// gzip the file ourselves into a byte buffer
	// use this to validate the bytes written to our mockwriter at end of test
	buf := &bytes.Buffer{}
	gw := gzip.NewWriter(buf)
	_, err := io.Copy(gw, f)
	if err != nil {
		t.Errorf("failed to write file %v", err)
	}
	f.Close()
	gw.Close()

	if got, want := w.data, buf.Bytes(); !bytes.Equal(got, want) {
		t.Errorf("data was not gzipped. got = %v, want = %v", got, want)
	}
	if a.ContentEncoding != GzipContentEncoding {
		t.Errorf("ContentDisposition doesn't match want: %v got: %v", GzipContentEncoding, a.ContentEncoding)
	}
}

func TestUploadObjects_GzipOverridesContentEncodingHeader_Success(t *testing.T) {
	u, a, b, w, ui := setup(t)
	u.Gzip = true
	u.Headers = map[string]string{"content-encoding": "myValue"}
	want := []UploadResults{
		{
			FilePath:   ui[0].FilePath,
			ObjectName: ui[0].ObjectName,
			Success:    true,
			Message:    "",
		},
	}

	got := u.UploadObjects(ui, b)
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("UploadObjects(ui, b) diff (-want +got):\n%s", diff)
	}

	f, _ := os.Open(ui[0].FilePath)

	// gzip the file ourselves into a byte buffer
	// use this to validate the bytes written to our mockwriter at end of test
	buf := &bytes.Buffer{}
	gw := gzip.NewWriter(buf)
	_, err := io.Copy(gw, f)
	if err != nil {
		t.Errorf("failed to write file %v", err)
	}
	f.Close()
	gw.Close()

	if got, want := w.data, buf.Bytes(); !bytes.Equal(got, want) {
		t.Errorf("data was not gzipped. got = %v, want = %v", got, want)
	}
	if a.ContentEncoding != GzipContentEncoding {
		t.Errorf("ContentDisposition doesn't match want: %v got: %v", GzipContentEncoding, a.ContentEncoding)
	}
}
