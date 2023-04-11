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

package build

import (
	"github.com/eschercloudai/baski/pkg/cmd/util/flags"
	gitRepo "github.com/eschercloudai/baski/pkg/git"
	systemUtils "github.com/eschercloudai/baski/pkg/system"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/google/uuid"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// CreateRepoDirectory create the random directory where the Image repo will be cloned into.
func CreateRepoDirectory() string {
	var tmpDir string
	uuidDir, err := uuid.NewRandom()
	if err != nil {
		tmpDir = "aaaaaaaa-1111-2222-3333-bbbbbbbbbbbb"
	} else {
		tmpDir = uuidDir.String()
	}

	dir := filepath.Join("/tmp", tmpDir)
	err = os.MkdirAll(dir, 0750)
	if err != nil {
		panic(err)
	}
	return dir
}

// FetchBuildRepo simply pulls the contents of the imageRepo to the specified path
func FetchBuildRepo(path string, o *flags.BuildOptions) {
	var branch plumbing.ReferenceName
	branch = plumbing.Master
	imageRepo := o.ImageRepo

	//FIXME: This check is in place until the nvidia branch and security branch in this repo go upstream.
	// Until it has been added, we must force users over to this repo as it's the only one that has these new additions.
	if o.AddNvidiaSupport || o.AddTrivy || o.AddFalco {
		log.Println("the kubernetes sigs project doesn't currently support nvidia, falco or trivy. Using https://github.com/eschercloudai/image-builder.git until it's pushed upstream")
		imageRepo = "https://github.com/eschercloudai/image-builder.git"
		branch = "refs/heads/nvidia-security"
	}

	_, err := gitRepo.GitClone(imageRepo, path, branch)
	if err != nil {
		panic(err)
	}
}

// InstallDependencies will run make dep-openstack so that any requirements such as packer, ansible
// and goss will be installed.
func InstallDependencies(repoPath string, verbose bool) {
	log.Printf("fetching dependencies\n")

	w, err := os.Create("/tmp/out-deps.txt")
	if err != nil {
		log.Fatalln(err)
	}
	defer w.Close()

	var wr io.Writer
	if verbose {
		wr = io.MultiWriter(w, os.Stdout)
	} else {
		wr = w
	}

	err = systemUtils.RunMake("deps-openstack", repoPath, nil, wr)
	if err != nil {
		log.Fatalln(err)
	}

	newPath := filepath.Join(repoPath, ".local/bin")
	path := strings.Join([]string{os.Getenv("PATH"), newPath}, ":")
	err = os.Setenv("PATH", path)
	if err != nil {
		log.Fatalln(err)
	}
}

// BuildImage will run make build-openstack-buildOS which will launch an instance in Openstack,
// add any requirements as defined in the image-builder imageRepo and then create an image from that build.
func BuildImage(capiPath string, buildOS string, verbose bool) error {
	log.Printf("building image\n")

	w, err := os.Create("/tmp/out-build.txt")
	if err != nil {
		return err
	}
	defer w.Close()

	var wr io.Writer
	if verbose {
		wr = io.MultiWriter(w, os.Stdout)
	} else {
		wr = w
	}

	args := strings.Join([]string{"build-openstack", buildOS}, "-")

	env := []string{"PACKER_VAR_FILES=tmp.json"}
	env = append(env, os.Environ()...)
	err = systemUtils.RunMake(args, capiPath, env, wr)
	if err != nil {
		log.Fatalln(err)
	}

	return nil
}

// SaveImageIDToFile exports the image ID to a file so that it can be read later by the scan system - this will generally be used by the gitHub action.
func SaveImageIDToFile(imgID string) error {
	f, err := os.Create("/tmp/imgid.out")
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write([]byte(imgID))
	if err != nil {
		return err
	}

	return nil
}
