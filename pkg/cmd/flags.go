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

package cmd

var (
	//Root
	baskioConfigFlag string
	cloudsPathFlag   string
	cloudNameFlag    string
	verboseFlag      bool

	//Build & Scan
	imageRepoFlag             string
	buildOSFlag               string
	sourceImageIDFlag         string
	networkIDFlag             string
	flavorNameFlag            string
	userFloatingIPFlag        bool
	floatingIPNetworkNameFlag string
	attachConfigDriveFlag     bool
	imageVisibilityFlag       string
	cniVersionFlag            string
	crictlVersionFlag         string
	kubeVersionFlag           string
	extraDebsFlag             string

	addNvidiaSupportFlag   bool
	nvidiaVersionFlag      string
	nvidiaInstallerURLFlag string
	gridLicenseServerFlag  string
	imageIDFlag            string

	// Publish
	ghUserFlag        string
	ghProjectFlag     string
	ghAccountFlag     string
	ghTokenFlag       string
	ghPagesBranchFlag string
)
