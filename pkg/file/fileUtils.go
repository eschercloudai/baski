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

package file

import (
	"os"
)

// CopyFile does what it says, it copies a file from location A to location B.
func CopyFile(from, to string) (*os.File, error) {
	f, err := os.ReadFile(from)
	if err != nil {
		return nil, err
	}

	w, err := os.Create(to)
	if err != nil {
		return nil, err
	}
	defer w.Close()

	_, err = w.Write(f)
	if err != nil {
		return nil, err
	}

	return w, nil
}
