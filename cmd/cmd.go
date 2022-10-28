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
	"github.com/spf13/cobra"
	"log"
	"path/filepath"
	"strings"
)

var (
	rootCmd *cobra.Command

	//Openstack specific flags
	osAuthURLFlag, osProjectNameFlag, osProjectIDFlag,
	osUsernameFlag, osPasswordFlag, osRegionNameFlag, osInterfaceFlag,
	osUserDomainNameFlag, osProjectDomainNameFlag, osIdentityAPIVersionFlag,
	osAuthPluginFlag string

	//Additional Openstack flags
	networkIDFlag, serverFlavorIDFlag, openstackBuildConfigPathFlag string
	enableConfigDriveFlag                                           bool

	//Build flags
	repoRoot                   = "https://github.com/eschercloudai/image-builder"
	imageRepoFlag, buildOSFlag string

	//gitHub flags
	ghUserFlag, ghProjectFlag, ghTokenFlag, ghPagesBranchFlag string
)

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
				ServerFlavorID:           serverFlavorIDFlag,
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
			copyVariablesFile(buildGitDir, envs.BuildOS, envs.OpenstackBuildConfigPath)
			capiPath := filepath.Join(buildGitDir, "images/capi")
			fetchDependencies(capiPath)
			err := buildImage(capiPath, envs.BuildOS)
			if err != nil {
				log.Fatalln(err)
			}
			imgID, err := retrieveNewImageID()
			if err != nil {
				log.Fatalln(err)
			}

			//Scan image
			osClient.OpenstackInit()
			kp := osClient.CreateNewKeypair()
			server, freeIP := osClient.CreateServerFromImageForScanning(kp, imgID, serverFlavorIDFlag, networkIDFlag, enableConfigDriveFlag)

			resultsFile, err := fetchResultsFromServer(freeIP, kp)
			if err != nil {
				removeScanningResources(server.ID, osClient)
				log.Fatalln(err.Error())
			}

			defer resultsFile.Close()

			removeScanningResources(server.ID, osClient)

			pagesGitDir, pagesRepo, err := fetchPagesRepo(envs.GhUser, envs.GhToken, envs.GhProject, envs.GhPagesBranch)
			if err != nil {
				log.Fatalln(err)
			}

			err = copyResultsFileIntoPages(pagesGitDir, resultsFile)
			if err != nil {
				pagesCleanup(pagesGitDir)
				log.Fatalln(err)
			}

			reports, err := fetchExistingReports(pagesGitDir)
			if err != nil {
				pagesCleanup(pagesGitDir)
				log.Fatalln(err)
			}

			results, err := parseReports(reports)
			if err != nil {
				pagesCleanup(pagesGitDir)
				log.Fatalln(err)
			}

			err = buildPages(pagesGitDir, results)
			if err != nil {
				pagesCleanup(pagesGitDir)
				log.Fatalln(err)
			}

			err = publishPages(pagesRepo, pagesGitDir)
			if err != nil {
				pagesCleanup(pagesGitDir)
				log.Fatalln(err)
			}

			pagesCleanup(pagesGitDir)
		},
	}

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

	//Commented the required flags out for now as this is going into a Docker container.

	rootCmd.Flags().StringVarP(&networkIDFlag, "network-id", "n", "", "Network ID to deploy the server onto for scanning. (Can also set env OS_NETWORK_ID)")
	//err := rootCmd.MarkFlagRequired("network-id")
	//if err != nil {
	//	log.Fatalf("%s\n", err.Error())
	//}
	rootCmd.Flags().StringVarP(&serverFlavorIDFlag, "server-flavor-id", "s", "", "ID of the server flavor to create for the scan. (Can also set env OS_SERVER_FLAVOR_ID)")
	//err = rootCmd.MarkFlagRequired("server-flavor-id")
	//if err != nil {
	//	log.Fatalf("%s\n", err.Error())
	//}
	rootCmd.Flags().StringVarP(&openstackBuildConfigPathFlag, "build-config", "b", "", strings.Join([]string{"The openstack variables to use to build the image (see ", repoRoot, "/blob/master/images/capi/packer/openstack/openstack-ubuntu-2004.json) (Can also set env OS_BUILD_CONFIG)"}, ""))
	//err = rootCmd.MarkFlagRequired("build-config")
	//if err != nil {
	//	log.Fatalf("%s\n", err.Error())
	//}
	rootCmd.Flags().BoolVarP(&enableConfigDriveFlag, "enable-config-drive", "c", false, "Used to enable a config drive on Openstack. This may be required if using an external network. (Can also set env OS_ENABLE_CONFIG_DRIVE)")

	rootCmd.Flags().StringVarP(&imageRepoFlag, "imageRepo", "r", strings.Join([]string{repoRoot, "git"}, "."), "The imageRepo from which the image builder should be deployed. (Can also set env IMAGE_REPO)")
	rootCmd.Flags().StringVarP(&buildOSFlag, "build-os", "o", "ubuntu-2204", "This is the target os to build. Valid values are currently: ubuntu-2004 and ubuntu-2204 (Can also set env BUILD_OS)")
	//err = rootCmd.MarkFlagRequired("build-os")
	//if err != nil {
	//	log.Fatalf("%s\n", err.Error())
	//}

	rootCmd.Flags().StringVarP(&ghUserFlag, "github-user", "u", "", "The user for the GitHub project to which the pages will be pushed. (Can also set env GH_USER)")
	//err = rootCmd.MarkFlagRequired("github-user")
	//if err != nil {
	//	log.Fatalf("%s\n", err.Error())
	//}
	rootCmd.Flags().StringVarP(&ghProjectFlag, "github-project", "p", "", "The GitHub project to which the pages will be pushed. (Can also set env GH_PROJECT)")
	//err = rootCmd.MarkFlagRequired("github-project")
	//if err != nil {
	//	log.Fatalf("%s\n", err.Error())
	//}
	rootCmd.Flags().StringVarP(&ghTokenFlag, "github-token", "t", "", "The token for the GitHub project to which the pages will be pushed. (Can also set env GH_TOKEN)")
	//err = rootCmd.MarkFlagRequired("github-token")
	//if err != nil {
	//	log.Fatalf("%s\n", err.Error())
	//}
	rootCmd.Flags().StringVarP(&ghPagesBranchFlag, "github-pages-branch", "g", "gh-pages", "The branch name for GitHub project to which the pages will be pushed. (Can also set env GH_PAGES_BRANCH)")
	//err = rootCmd.MarkFlagRequired("github-token")
	//if err != nil {
	//	log.Fatalf("%s\n", err.Error())
	//}

	rootCmd.AddCommand(versionCmd())
}

func Execute() error {
	return rootCmd.Execute()
}
