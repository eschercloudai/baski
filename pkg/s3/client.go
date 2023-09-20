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
	s3Conn    *simples3.S3
}

// FetchFromS3 Downloads a file from an S3 bucket and returns its contents as a byte array.
func (s *S3) FetchFromS3(fileName string) ([]byte, error) {
	s.s3Conn = simples3.New("us-east-1", s.AccessKey, s.SecretKey)
	s.s3Conn.SetEndpoint(s.Endpoint)

	input := simples3.DownloadInput{
		Bucket:    s.Bucket,
		ObjectKey: fileName,
	}
	// Download the file.
	file, err := s.download(input)

	if err != nil {
		return nil, fmt.Errorf("failed to download file from S3: %v\n", err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file contents: %v\n", err)
	}

	return data, nil
}

// fileDownload Downloads the file from an S3 bucket
func (s *S3) download(input simples3.DownloadInput) (io.ReadCloser, error) {
	return s.s3Conn.FileDownload(input)
}

// PutToS3 Pushes a file to an S3 bucket.
func (s *S3) PutToS3(contentType, key, fileName string, body io.ReadSeeker) error {
	s3Conn := simples3.New("us-east-1", s.AccessKey, s.SecretKey)
	s3Conn.SetEndpoint(s.Endpoint)

	// Put the file into S3.
	input := simples3.UploadInput{
		Bucket:      s.Bucket,
		ObjectKey:   key,
		ContentType: contentType,
		FileName:    fileName,
		Body:        body,
	}
	_, err := s.upload(input)
	if err != nil {
		return fmt.Errorf("failed to push file to S3: %v\n", err)
	}

	return nil
}

// PutToS3 Pushes a file to an S3 bucket.
func (s *S3) upload(input simples3.UploadInput) (simples3.PutResponse, error) {
	return s.s3Conn.FilePut(input)
}
