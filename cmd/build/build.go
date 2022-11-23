/*
Copyright 2022 EscherCloud.
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
	"bufio"
	gitRepo "github.com/drew-viles/baskio/pkg/git"
	systemUtils "github.com/drew-viles/baskio/pkg/system"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/google/uuid"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// CreateRepoDirectory create the random directory where the Image repo will be cloned into.
func CreateRepoDirectory() string {
	var tmpDir string
	uuidDir, err := uuid.NewUUID()
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
func FetchBuildRepo(path, imageRepo string, gpuSupport bool) {
	var branch plumbing.ReferenceName
	branch = plumbing.Master

	if gpuSupport {
		branch = "refs/heads/nvidia-driver-support"
	}

	_, err := gitRepo.GitClone(imageRepo, path, branch)
	if err != nil {
		panic(err)
	}
}

// InstallDependencies will run make dep-openstack so that any requirements such as packer, ansible
// and goss will be installed.
func InstallDependencies(repoPath string) {
	log.Printf("fetching dependencies\n")

	w, err := os.Create("/tmp/out-deps.txt")
	if err != nil {
		log.Fatalln(err)
	}
	defer w.Close()

	wr := io.MultiWriter(w, os.Stdout)

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

// BuildImage will run make build-openstack-buildOSFlag which will launch an instance in Openstack,
// add any requirements as defined in the image-builder imageRepo and then create an image from that build.
func BuildImage(capiPath string, buildOS string) error {
	log.Printf("building image\n")

	w, err := os.Create("/tmp/out-build.txt")
	if err != nil {
		return err
	}
	defer w.Close()

	wr := io.MultiWriter(w, os.Stdout)
	//TODO: Maybe fetch from openstack and sort by newest.
	//  Would require some trickery to get new image ID.

	args := strings.Join([]string{"build-openstack", buildOS}, "-")

	env := []string{"PACKER_VAR_FILES=tmp.json"}
	env = append(env, os.Environ()...)
	err = systemUtils.RunMake(args, capiPath, env, wr)
	if err != nil {
		log.Fatalln(err)
	}

	return nil
}

// RetrieveNewImageID fetches the newly create image's ID from the out.txt file
// that is generated during the buildImage() run.
func RetrieveNewImageID() (string, error) {
	var i string

	//TODO: If the output goes to stdOUT in buildImage,
	// we need to figure out if we can pull this from the openstack instance instead.
	f, err := os.Open("/tmp/out-build.txt")
	if err != nil {
		return "", err
	}
	defer f.Close()

	r := bufio.NewScanner(f)
	re := regexp.MustCompile("An image was created: [0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}")
	for r.Scan() {
		m := re.MatchString(string(r.Bytes()))
		if m {
			//There is likely two outputs here due to how packer outputs, so we need to break on the first find.
			i = strings.Split(r.Text(), ": ")[2]
			break
		}
	}

	return i, nil
}

// SaveImageIDToFile exports the image ID to a file so that it can be read later by the scan system - this will generally be used by the gitHub action.
func SaveImageIDToFile(imgID string) error {
	f, err := os.Create("/tmp/imgid.out")
	if err != nil {
		return err
	}
	defer f.Close()
	f.Write([]byte(imgID))

	return nil
}
