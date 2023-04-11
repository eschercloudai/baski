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

package constants

import (
	"github.com/eschercloudai/baski/pkg/trivy"
)

var (
	Version     = "v0.1.0-beta.1"
	SupportedOS = []string{
		"ubuntu-2004",
		"ubuntu-2204",
	}
	TrivyVersion = "0.38.3"
)

// Year is used in reports parsing. It is the top level and contains multiple Month(s).
type Year struct {
	Months map[string]Month
}

// Month is used in reports parsing. It is contained within a Year and contains multiple trivy.Report(s).
type Month struct {
	Reports map[string]trivy.Report
}
