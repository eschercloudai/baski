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
	"github.com/eschercloudai/baski/pkg/mock"
	"github.com/eschercloudai/baski/pkg/util"
	"go.uber.org/mock/gomock"
	"os"
	"testing"
)

func put(s util.S3Interface) error {
	f := createFile()
	t := s.PutToS3("text/plain", "path/results.json", "results.json", f)
	removeFile(f)
	return t
}

func TestFetchFromS3(t *testing.T) {
	c := gomock.NewController(t)
	defer c.Finish()
	m := mock.NewMockS3Interface(c)

	m.EXPECT().FetchFromS3(gomock.Eq("trivyignore")).Return([]byte("some data"), nil)
	if _, err := m.FetchFromS3("trivyignore"); err != nil {
		t.Error(err)
	}
}

func TestPutToS3(t *testing.T) {
	c := gomock.NewController(t)
	defer c.Finish()
	m := mock.NewMockS3Interface(c)

	f := createFile()

	m.EXPECT().PutToS3(gomock.Eq("text/plain"), gomock.Eq("path/results.json"), gomock.Eq("results.json"), gomock.Eq(f)).Return(nil)
	if err := put(m); err != nil {
		t.Error(err)
	}

	removeFile(f)
}

func removeFile(f *os.File) {
	err := os.Remove(f.Name())
	if err != nil {
		fmt.Println(err)
	}
}

func createFile() *os.File {
	f, err := os.Create("/tmp/test.txt")
	if err != nil {
		fmt.Println(err)
	}
	_, err = f.Write([]byte("test"))
	if err != nil {
		fmt.Println(err)
	}
	defer f.Close()

	return f
}
