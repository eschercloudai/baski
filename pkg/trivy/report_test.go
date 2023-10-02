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

package trivy

import (
	"strings"
	"testing"
)

// CheckSeverity compares two severities to see if a threshold has been met. IE: is sev: HIGH >= check: MEDIUM.
func TestCheckSeverity(t *testing.T) {
	testCases := []struct {
		Name      string
		Severity  Severity
		Threshold Severity
		Expected  bool
	}{
		{
			Name:      "Test MEDIUM is below CRITICAL",
			Severity:  MEDIUM,
			Threshold: CRITICAL,
			Expected:  false,
		},
		{
			Name:      "Test CRITICAL is above MEDIUM",
			Severity:  CRITICAL,
			Threshold: MEDIUM,
			Expected:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			res := CheckSeverity(tc.Severity, tc.Threshold)
			if res != tc.Expected {
				t.Errorf("Expected data %t, got: %t\n", tc.Expected, res)
			}
		})
	}
}

// ValidSeverity confirms that the supplied value is a valid severity value.
func TestValidSeverity(t *testing.T) {
	testCases := []struct {
		Name     string
		Severity string
		Expected bool
	}{
		{
			Name:     "Test Medium is valid",
			Severity: "MEDIUM",
			Expected: true,
		},
		{
			Name:     "Test HIGH is valid",
			Severity: "HIGH",
			Expected: true,
		},
		{
			Name:     "Test NOTHING is invalid",
			Severity: "NOTHING",
			Expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			res := ValidSeverity(Severity(strings.ToUpper(tc.Severity)))
			if res != tc.Expected {
				t.Errorf("Expected data %t, got: %t\n", tc.Expected, res)
			}
		})
	}
}
