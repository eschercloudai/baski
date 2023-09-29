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

package systemUtils

import (
	"os"
	"testing"
)

func TestRunMake(t *testing.T) {
	makeFile := "/tmp/makefile"
	//Make a random makefile for testing
	w, err := os.Create(makeFile)
	if err != nil {
		t.Error(err)
		return
	}
	defer w.Close()
	_, err = w.WriteString(`
# Define the target to echo the environment variable
echo-env-var:
	@echo "The value of $(TEST) is: $$(echo $$$(TEST))"
`)
	if err != nil {
		t.Error(err)
		return
	}

	o, err := os.Create("/tmp/output")
	if err != nil {
		t.Error(err)
		return
	}
	defer o.Close()

	args := "echo-env-var"
	env := []string{"TEST=test-run"}
	env = append(env, os.Environ()...)
	err = RunMake(args, "/tmp", env, w)
	if err != nil {
		t.Error(err)
	}
}
