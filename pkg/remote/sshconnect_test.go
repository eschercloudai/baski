/*
Copyright 2024 Drewbernetes.

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

package remote

import (
	"github.com/drewbernetes/baski/pkg/mock"
	"go.uber.org/mock/gomock"
	"os"
	"testing"
)

func generateTestData(t *testing.T, from string) *os.File {
	f, err := os.Create(from)
	if err != nil {
		t.Error(err)
		return nil
	}

	return f
}

// TestCopyFromRemoteServer uses sftp to copy a file from a remotes server to a local directory.
func TestCopyFromRemoteServer(t *testing.T) {
	testFile := generateTestData(t, "/tmp/some-file.json")

	c := gomock.NewController(t)
	defer c.Finish()

	m := mock.NewMockSSHInterface(c)
	m.EXPECT().CopyFromRemoteServer("/tmp/some-file.json", "/tmp/another-file.json").Return(testFile, nil).AnyTimes()
	m.EXPECT().SSHClose().Return(nil).AnyTimes()
	m.EXPECT().SFTPClose().Return(nil).AnyTimes()

	// Define test cases
	testCases := []struct {
		name     string
		ssh      *mock.MockSSHInterface
		from     string
		to       string
		filename string
	}{
		{
			name: "Test case 1: Copy file from remote location",
			ssh:  m,
			from: "/tmp/some-file.json",
			to:   "/tmp/another-file.json",
		},
	}

	// RunScan the test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := tc.ssh.CopyFromRemoteServer(tc.from, tc.to)
			if err != nil {
				t.Error(err)
				return
			}

			if testFile != result {
				t.Errorf("Expected data %+v, got: %+v\n", testFile, result)
			}

			if err = tc.ssh.SSHClose(); err != nil {
				t.Error(err)
				return
			}

			if err = tc.ssh.SFTPClose(); err != nil {
				t.Error(err)
				return
			}
		})
	}
}
