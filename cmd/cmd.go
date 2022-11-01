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

package cmd

import (
	"fmt"
	"github.com/drew-viles/baskio/pkg/constants"
	ostack "github.com/drew-viles/baskio/pkg/openstack"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"log"
	"path/filepath"
	"strings"
	"time"
)

var (
	rootCmd *cobra.Command

	//Openstack specific flags
	osAuthURLFlag, osProjectNameFlag, osProjectIDFlag,
	osUsernameFlag, osPasswordFlag, osRegionNameFlag, osInterfaceFlag,
	osUserDomainNameFlag, osProjectDomainNameFlag, osIdentityAPIVersionFlag,
	osAuthPluginFlag string

	//Additional Openstack flags
	networkIDFlag, openstackBuildConfigPathFlag string
	enableConfigDriveFlag                       bool

	//Build flags
	repoRoot                   = "https://github.com/eschercloudai/image-builder"
	imageRepoFlag, buildOSFlag string

	//gitHub flags
	ghUserFlag, ghProjectFlag, ghTokenFlag, ghPagesBranchFlag string
)

// init prepares the tool with all available flag. It also contains the main program loop which runs the tasks.
func init() {
	rootCmd = &cobra.Command{
		Use:   "baskio",
		Short: "Baskio is a tools for building and scanning Kubernetes images within Openstack.",
		Long: `Build And Scan Kubernetes Images on Openstack
		This tool has been designed to automatically build images for the Openstack potion of the Kubernetes Image Builder.
		It could be extended out to provide images for a variety of other builders however for now it's main goal is to work with Openstack.`,
		Run: func(cmd *cobra.Command, args []string) {
			// Dump all the input vars into here.
			envs := constants.Env{
				AuthURL:                  osAuthURLFlag,
				ProjectID:                osProjectIDFlag,
				ProjectName:              osProjectNameFlag,
				Username:                 osUsernameFlag,
				Password:                 osPasswordFlag,
				Region:                   osRegionNameFlag,
				Interface:                osInterfaceFlag,
				UserDomainName:           osUserDomainNameFlag,
				ProjectDomainName:        osProjectDomainNameFlag,
				IdentityAPIVersion:       osIdentityAPIVersionFlag,
				AuthPlugin:               osAuthPluginFlag,
				NetworkID:                networkIDFlag,
				OpenstackBuildConfigPath: openstackBuildConfigPathFlag,
				EnableConfigDrive:        fmt.Sprintf("%t", enableConfigDriveFlag),
				ImageRepo:                imageRepoFlag,
				BuildOS:                  buildOSFlag,
				GhUser:                   ghUserFlag,
				GhProject:                ghProjectFlag,
				GhToken:                  ghTokenFlag,
				GhPagesBranch:            ghPagesBranchFlag,
			}

			// Now we check to see if any env vars have been passed instead of flags. If so, set the flags to the env vars.
			envs.CheckForEnvVars()

			osClient := &ostack.Client{
				Env: envs,
			}

			//Build image
			buildGitDir := fetchBuildRepo(envs.ImageRepo)
			buildConfig := ostack.ParseBuildConfig(envs.OpenstackBuildConfigPath)
			buildConfig.Networks = envs.NetworkID

			scanDate := fmt.Sprintf("%d-%d-%d--%d-%d-%d", time.Now().Year(), time.Now().Month(), time.Now().Day(), time.Now().Hour(), time.Now().Minute(), time.Now().Second())
			imageUUID, err := uuid.NewUUID()
			if err != nil {
				log.Fatalln(err)
			}
			imageName := fmt.Sprintf("%s-kube-%s-%s-%s", buildConfig.BuildName, buildConfig.KubernetesSemver, scanDate, imageUUID.String())

			generateVariablesFile(buildGitDir, buildConfig)

			capiPath := filepath.Join(buildGitDir, "images/capi")
			fetchDependencies(capiPath)
			err = buildImage(capiPath, envs.BuildOS)
			if err != nil {
				log.Fatalln(err)
			}
			imgID, err := retrieveNewImageID()
			if err != nil {
				log.Fatalln(err)
			}

			//Scan image
			osClient.OpenstackInit()
			kp := osClient.CreateKeypair()
			server, freeIP := osClient.CreateServer(kp, imgID, buildConfig.Flavor, buildConfig.Networks, enableConfigDriveFlag)

			resultsFile, err := fetchResultsFromServer(freeIP, kp)
			if err != nil {
				removeScanningResources(server.ID, osClient)
				log.Fatalln(err.Error())
			}

			defer resultsFile.Close()

			//Cleanup the scanning resources
			removeScanningResources(server.ID, osClient)

			//GitHub pages
			pagesGitDir, pagesRepo, err := fetchPagesRepo(envs.GhUser, envs.GhToken, envs.GhProject, envs.GhPagesBranch)
			if err != nil {
				log.Fatalln(err)
			}

			err = copyResultsFileIntoPages(pagesGitDir, imageName, resultsFile)
			checkErrorPagesWithCleanup(err, pagesGitDir)

			reports, err := fetchExistingReports(pagesGitDir)
			checkErrorPagesWithCleanup(err, pagesGitDir)

			results, err := parseReports(reports)
			checkErrorPagesWithCleanup(err, pagesGitDir)

			err = buildPages(pagesGitDir, results)
			checkErrorPagesWithCleanup(err, pagesGitDir)

			err = publishPages(pagesRepo, pagesGitDir)
			checkErrorPagesWithCleanup(err, pagesGitDir)

			pagesCleanup(pagesGitDir)
		},
	}

	//Openstack specific
	rootCmd.Flags().StringVar(&osAuthURLFlag, "os-auth-url", "", "The Openstack Auth URL (Can also set env OS_AUTH_URL)")
	rootCmd.Flags().StringVar(&osProjectNameFlag, "os-project-name", "", "The Openstack Project Name (Can also set env OS_PROJECT_NAME)")
	rootCmd.Flags().StringVar(&osProjectIDFlag, "os-project-id", "", "The Openstack Project Name (Can also set env OS_PROJECT_ID)")
	rootCmd.Flags().StringVar(&osUsernameFlag, "os-username", "", "The Openstack UserName (Can also set env OS_USERNAME)")
	rootCmd.Flags().StringVar(&osPasswordFlag, "os-password", "", "The Openstack Password (Can also set env OS_PASSWORD)")
	rootCmd.Flags().StringVar(&osRegionNameFlag, "os-region-name", "RegionOne", "The Openstack Region Name (Can also set env OS_REGION_NAME)")
	rootCmd.Flags().StringVar(&osInterfaceFlag, "os-interface", "public", "The Openstack Interface (Can also set env OS_INTERFACE)")
	rootCmd.Flags().StringVar(&osUserDomainNameFlag, "os-user-domain-name", "Default", "The Openstack User Domain Name (Can also set env OS_USER_DOMAIN_NAME)")
	rootCmd.Flags().StringVar(&osProjectDomainNameFlag, "os-project-domain-name", "default", "The Openstack Project Domain Name (Can also set env OS_PROJECT_DOMAIN_NAME)")
	rootCmd.Flags().StringVar(&osIdentityAPIVersionFlag, "os-identity-api-version", "3", "The Openstack Identity API Version (Can also set env OS_IDENTITY_API_VERSION)")
	rootCmd.Flags().StringVar(&osAuthPluginFlag, "os-auth-plugin", "password", "The Openstack Auth Plugin (Can also set env OS_AUTH_PLUGIN)")
	rootCmd.Flags().BoolVarP(&enableConfigDriveFlag, "enable-config-drive", "c", false, "Used to enable a config drive on Openstack. This may be required if using an external network. (Can also set env OS_ENABLE_CONFIG_DRIVE)")
	rootCmd.Flags().StringVarP(&networkIDFlag, "network-id", "n", "", "Network ID to deploy the server onto for scanning. (Can also set env OS_NETWORK_ID)")

	//Configuration items
	rootCmd.Flags().StringVarP(&openstackBuildConfigPathFlag, "build-config", "b", "", strings.Join([]string{"The openstack variables to use to build the image (see ", repoRoot, "/blob/master/images/capi/packer/openstack/openstack-ubuntu-2004.json) (Can also set env OS_BUILD_CONFIG)"}, ""))
	rootCmd.Flags().StringVarP(&imageRepoFlag, "imageRepo", "r", strings.Join([]string{repoRoot, "git"}, "."), "The imageRepo from which the image builder should be deployed. (Can also set env IMAGE_REPO)")
	rootCmd.Flags().StringVarP(&buildOSFlag, "build-os", "o", "ubuntu-2204", "This is the target os to build. Valid values are currently: ubuntu-2004 and ubuntu-2204 (Can also set env BUILD_OS)")

	//GitHub specific
	rootCmd.Flags().StringVarP(&ghUserFlag, "github-user", "u", "", "The user for the GitHub project to which the pages will be pushed. (Can also set env GH_USER)")
	rootCmd.Flags().StringVarP(&ghProjectFlag, "github-project", "p", "", "The GitHub project to which the pages will be pushed. (Can also set env GH_PROJECT)")
	rootCmd.Flags().StringVarP(&ghTokenFlag, "github-token", "t", "", "The token for the GitHub project to which the pages will be pushed. (Can also set env GH_TOKEN)")
	rootCmd.Flags().StringVarP(&ghPagesBranchFlag, "github-pages-branch", "g", "gh-pages", "The branch name for GitHub project to which the pages will be pushed. (Can also set env GH_PAGES_BRANCH)")

	rootCmd.AddCommand(versionCmd())
}

// checkErrorWithCleanup takes an error and if it is not nil, will attempt to run a cleanup to ensure no resources are left lying around.
func checkErrorPagesWithCleanup(err error, dir string) {
	if err != nil {
		pagesCleanup(dir)
		log.Fatalln(err)
	}
}

// Execute runs the execute command for the Cobra library allowing commands & flags to be utilised.
func Execute() error {
	return rootCmd.Execute()
}
