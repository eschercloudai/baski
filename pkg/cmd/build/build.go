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
	"fmt"
	"github.com/eschercloudai/baski/pkg/cmd/util/data"
	"github.com/eschercloudai/baski/pkg/cmd/util/flags"
	"github.com/eschercloudai/baski/pkg/constants"
	ostack "github.com/eschercloudai/baski/pkg/openstack"
	"github.com/spf13/cobra"
	"path/filepath"
)

var (
	imageBuilderRepo = "https://github.com/kubernetes-sigs/image-builder"
)

// NewBuildCommand creates a command that allows the building of an image.
func NewBuildCommand() *cobra.Command {
	o := &flags.BuildOptions{}

	cmd := &cobra.Command{
		Use:   "build",
		Short: "Build image",
		Long: `Build image.

Building images requires a set of commands to be run on the terminal however this is tedious and time consuming.
By using this, the time is reduced and automation can be enabled.

Overtime this will become more dynamic to allow for build customised
images such as ones with GPU/HPC drivers/tools.

To use baski to build an image, an Openstack cluster is required.`,
		TraverseChildren: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			o.SetOptionsFromViper()

			cloudsConfig := ostack.InitOpenstack(o.CloudsPath)
			packerBuildConfig := ostack.InitPackerConfig(o)
			if !checkValidOSSelected(o.BuildOS) {
				return fmt.Errorf("an unsupported OS has been entered. Please select a valid OS: %s\n", constants.SupportedOS)
			}

			buildGitDir := CreateRepoDirectory()
			FetchBuildRepo(buildGitDir, o)

			metadata := ostack.GenerateBuilderMetadata(o)
			err := ostack.UpdatePackerBuildersJson(buildGitDir, metadata)
			if err != nil {
				return err
			}
			capiPath := filepath.Join(buildGitDir, "images", "capi")
			packerBuildConfig.GenerateVariablesFile(capiPath)

			InstallDependencies(capiPath, o.Verbose)

			cloudsConfig.SetOpenstackEnvs(o.CloudName)

			err = BuildImage(capiPath, o.BuildOS, o.Verbose)
			if err != nil {
				return err
			}

			imgID, err := data.RetrieveNewImageID()
			if err != nil {
				return err
			}

			err = SaveImageIDToFile(imgID)
			if err != nil {
				return err
			}
			return nil
		},
	}

	o.AddFlags(cmd, imageBuilderRepo)

	return cmd
}

// checkValidOSSelected checks that the build os provided is a valid one.
func checkValidOSSelected(buildOS string) bool {
	for _, v := range constants.SupportedOS {
		if buildOS == v {
			return true
		}
	}
	return false
}
