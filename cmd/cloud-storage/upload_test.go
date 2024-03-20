package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"cloud.google.com/go/storage"
	"github.com/GoogleCloudBuild/cicd-images/cmd/cloud-storage/pkg/upload"
	"github.com/google/go-cmp/cmp"
	"github.com/googleapis/google-cloud-go-testing/storage/stiface"
)

// mocks for testing Upload()
type clientMock struct {
	stiface.Client
	bMock bucketMock
}

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

func (c clientMock) Bucket(_ string) stiface.BucketHandle {
	return c.bMock
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

func setup(t *testing.T) (c *clientMock, path string) {
	t.Helper()

	a := storage.ObjectAttrs{}
	a.Metadata = map[string]string{}
	w := mockWriter{attrs: &a, numBytesWritten: 0, data: []byte{}}
	o := objectMock{wMock: &w}
	b := bucketMock{oMock: o}
	c = &clientMock{bMock: b}

	tempDir := t.TempDir()
	data := []byte("this is some example data")
	path = filepath.Join(tempDir, "testFile.txt")
	e := os.WriteFile(path, data, 0600)
	if e != nil {
		t.Fatalf("os.WriteFile(path, data, 0600) = %v", e)
	}

	return c, path
}

func TestUpload_Success(t *testing.T) {
	c, path := setup(t)

	want := []upload.UploadResults{
		{
			FilePath:   path,
			ObjectName: filepath.Base(path),
			Success:    true,
		},
	}
	got, err := Upload(c, "bucketName", path, "", "", false, false, false, map[string]string{}, 100)
	if err != nil {
		t.Fatalf("unexpected err: %s", err)
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("Upload(c, 'bucketName', path, '', '', false, false, false, map[string]string{}, 100) diff (-want +got):\n%s", diff)
	}
}

func TestSummarizeResults_Success(t *testing.T) {
	const errorString1 = "Error writing to GCS"
	const errorString2 = "Unable to open file"

	tests := []struct {
		name                 string
		input                []upload.UploadResults
		expectedSuccessCount int
		expectedFailureCount int
		expectedErr          map[string]bool
	}{
		{
			name: "AllSuccessful",
			input: []upload.UploadResults{
				{
					Success: true,
				},
				{
					Success: true,
				},
			},
			expectedSuccessCount: 2,
			expectedFailureCount: 0,
			expectedErr:          map[string]bool{},
		},
		{
			name: "AllFailure",
			input: []upload.UploadResults{
				{
					Success: false,
					Message: errorString1,
				},
				{
					Success: false,
					Message: errorString1,
				},
			},
			expectedSuccessCount: 0,
			expectedFailureCount: 2,
			expectedErr:          map[string]bool{errorString1: true},
		},
		{
			name: "MixedSuccessAndFailure",
			input: []upload.UploadResults{
				{
					Success: false,
					Message: errorString1,
				},
				{
					Success: true,
				},
				{
					Success: true,
				},
				{
					Success: false,
					Message: errorString1,
				},
				{
					Success: false,
					Message: errorString2,
				},
			},
			expectedSuccessCount: 2,
			expectedFailureCount: 3,
			expectedErr:          map[string]bool{errorString1: true, errorString2: true},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			successCount, failureCount, errorMap := summarizeResults(test.input)

			if diff := cmp.Diff(test.expectedSuccessCount, successCount); diff != "" {
				t.Errorf("summarizeResults(test.input): successCount diff (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(test.expectedFailureCount, failureCount); diff != "" {
				t.Errorf("summarizeResults(test.input): failureCount diff (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(test.expectedErr, errorMap); diff != "" {
				t.Errorf("summarizeResults(test.input): error list diff (-want +got):\n%s", diff)
			}
		})
	}
}

func TestConvertHeaderStringToMap(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedResult map[string]string
		expectedErr    error
	}{
		{
			name:           "singlePair",
			input:          "myKey=myValue",
			expectedResult: map[string]string{"myKey": "myValue"},
		},
		{
			name:           "multiPair",
			input:          "myKey=myValue,FooKey=FooValue",
			expectedResult: map[string]string{"myKey": "myValue", "FooKey": "FooValue"},
		},
		{
			name:           "emptyInput",
			input:          "",
			expectedResult: map[string]string{},
		},
		{
			name:        "invalidInput",
			input:       "MyKey",
			expectedErr: fmt.Errorf("MyKey is not formatted as key=value"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := convertHeaderStringToMap(test.input)

			// diff both the expected result and expected error
			if diff := cmp.Diff(test.expectedResult, got); diff != "" {
				t.Errorf("convertHeaderStringToMap(test.input): result diff (-want +got):\n%s", diff)
			}
			if err != nil || test.expectedErr != nil {
				if diff := cmp.Diff(test.expectedErr.Error(), err.Error()); diff != "" {
					t.Errorf("convertHeaderStringToMap(test.input): error diff (-want +got):\n%s", diff)
				}
			}
		})
	}
}

func TestValidateHeaders_Success(t *testing.T) {
	headers := map[string]string{
		"content-type":        "plain/text",
		"cache-control":       "no-cache",
		"content-disposition": "inline",
		"content-encoding":    "gzip",
		"content-language":    "en-US",
		"custom-time":         "2024-03-07T00:11:22Z",
		"x-goog-meta-custom1": "foo",
		"x-goog-meta-custom2": "bar",
	}
	got := validateHeaders(headers)
	if got != nil {
		t.Errorf("unexpected error validating valid headers: %s", got.Error())
	}
}

func TestValidateHeaders_Fail(t *testing.T) {
	tests := []struct {
		name        string
		headers     map[string]string
		expectedErr error
	}{
		{
			name:    "UnknownHeader",
			headers: map[string]string{"made-up-header": "plain/text"},
			expectedErr: fmt.Errorf("invalid header provided: made-up-header=plain/text. Must be either a known header with " +
				"keys 'content-type', 'cache-control', 'content-disposition', 'content-encoding', 'content-language', " +
				"'custom-time', or it must be a header with a key prefix of 'x-goog-meta-'"),
		},
		{
			name:    "InvalidCustomTime",
			headers: map[string]string{"custom-time": "2024-03"},
			expectedErr: fmt.Errorf(`failed to parse custom-time: parsing time "2024-03" as "2006-01-02T15:04:05Z07:00": ` +
				`cannot parse "" as "-"`),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := validateHeaders(test.headers)

			if diff := cmp.Diff(test.expectedErr.Error(), got.Error()); diff != "" {
				t.Errorf("mismatched error: %s", diff)
			}
		})
	}
}

func TestInvalidPredefinedACL(t *testing.T) {
	tests := []struct {
		name        string
		aclName     string
		expectedErr error
	}{
		{
			name:        "InvalidPredefinedACL",
			aclName:     "madeUpAcl",
			expectedErr: fmt.Errorf("unknown predefined acl provided: madeUpAcl. Must be one of " + PredefinedACLList),
		},
		{
			name:        "InvalidPredefinedACL_AllCaps",
			aclName:     "PRIVATE",
			expectedErr: fmt.Errorf("unknown predefined acl provided: PRIVATE. Must be one of " + PredefinedACLList),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := validatePredefinedACL(test.aclName)

			if diff := cmp.Diff(test.expectedErr.Error(), got.Error()); diff != "" {
				t.Errorf("mismatched error: %s", diff)
			}
		})
	}
}
