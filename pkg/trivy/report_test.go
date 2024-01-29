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

package trivy

import (
	"strings"
	"testing"
)

// CheckSeverity compares two severities to see if a threshold has been met. IE: is sev: HIGH >= check: MEDIUM.
func TestCheckSeverity(t *testing.T) {
	testCases := []struct {
		Name     string
		Severity Severity
		Expected []string
	}{
		{
			Name:     "Test UNKOWN entry",
			Severity: UNKNOWN,
			Expected: []string{"UNKNOWN", "LOW", "MEDIUM", "HIGH", "CRITICAL"},
		},
		{
			Name:     "Test LOW entry",
			Severity: LOW,
			Expected: []string{"LOW", "MEDIUM", "HIGH", "CRITICAL"},
		},
		{
			Name:     "Test MEDIUM entry",
			Severity: MEDIUM,
			Expected: []string{"MEDIUM", "HIGH", "CRITICAL"},
		},
		{
			Name:     "Test HIGH entry",
			Severity: HIGH,
			Expected: []string{"HIGH", "CRITICAL"},
		},
		{
			Name:     "Test CRITICAL entry",
			Severity: CRITICAL,
			Expected: []string{"CRITICAL"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			res := ParseSeverity(tc.Severity)

			rt := strings.Join(res, ",")
			et := strings.Join(res, ",")
			if rt != et {
				t.Errorf("Expected data %s, got: %s\n", strings.Join(tc.Expected, ","), res)
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
