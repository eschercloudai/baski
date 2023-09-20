/*
Copyright 2023 EscherCloud.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package s3

import (
	"fmt"
	"github.com/rhnvrm/simples3"
	"io"
)

type S3 struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
}

type InterfaceS3 interface {
	FetchFromS3(string) ([]byte, error)
	PutToS3(string, string, string, io.ReadSeeker) error
}

// FetchFromS3 Downloads a file from an S3 bucket and returns its contents as a byte array.
func (s *S3) FetchFromS3(fileName string) ([]byte, error) {
	s3Conn := simples3.New("us-east-1", s.AccessKey, s.SecretKey)
	s3Conn.SetEndpoint(s.Endpoint)

	// Download the file.
	file, err := s3Conn.FileDownload(simples3.DownloadInput{
		Bucket:    s.Bucket,
		ObjectKey: fileName,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to download file from S3: %v", err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file contents: %v", err)
	}

	return data, nil
}

// PutToS3 Pushes a file to an S3 bucket.
func (s *S3) PutToS3(contentType, key, fileName string, body io.ReadSeeker) error {
	s3Conn := simples3.New("us-east-1", s.AccessKey, s.SecretKey)
	s3Conn.SetEndpoint(s.Endpoint)

	// Put the file into S3.
	opts := simples3.UploadInput{
		Bucket:      s.Bucket,
		ObjectKey:   key,
		ContentType: contentType,
		FileName:    fileName,
		Body:        body,
	}
	_, err := s3Conn.FilePut(opts)
	if err != nil {
		return fmt.Errorf("failed to push file to S3: %v", err)
	}

	return nil
}
