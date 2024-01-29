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

package gitRepo

import (
	"github.com/go-git/go-git/v5/plumbing"
	"os"
	"testing"
)

// TestGitClone Tests the cloning by cloning the image builder repo.
func TestGitClone(t *testing.T) {
	repo := "https://github.com/drew-viles/image-builder.git"
	cloneLocation := "/tmp/test"
	err := os.RemoveAll(cloneLocation)
	if err != nil {
		t.Error(err)
		return
	}
	ref := plumbing.ReferenceName("refs/heads/main")
	_, err = GitClone(repo, cloneLocation, ref)
	if err != nil {
		t.Error(err)
		return
	}

	f, err := os.Stat(cloneLocation)
	if err != nil {
		t.Error(err)
		return
	}

	if !f.IsDir() {
		t.Error("expected directory, didn't get a directory")
	}

}
