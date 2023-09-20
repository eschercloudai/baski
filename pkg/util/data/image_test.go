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

package data

import (
	"os"
	"testing"
)

// TestRetrieveNewImageID fetches the newly create image's ID from the out.txt file
// that is generated during the buildImage() run.
func TestRetrieveNewImageID(t *testing.T) {

	tc := []struct {
		Name     string
		Msg      string
		Expected string
	}{
		{
			Name: "Test valid ID",
			Msg: `==> Wait completed after 10 minutes 45 seconds

==> Builds finished. The artifacts of successful builds are:
--> openstack: An image was created: 42cb1fd0-61aa-4b76-a66d-d0c377cc9c22
--> openstack: An image was created: 42cb1fd0-61aa-4b76-a66d-d0c377cc9c22`,
			Expected: "42cb1fd0-61aa-4b76-a66d-d0c377cc9c22",
		},
		{
			Name: "Test invalid ID",
			Msg: `Just some random text in a file.
This should not be parsed or return anything in any way.
But for good measure, here is an imageID: 42cb1fd0-61aa-4b76-a66d-d0c377cc9c22`,
			Expected: "",
		},
	}

	for _, test := range tc {
		t.Run(test.Name, func(t *testing.T) {
			filename := "/tmp/out-build.txt"
			f, err := os.Create(filename)
			if err != nil {
				t.Error(err)
				return
			}
			_, err = f.Write([]byte(test.Msg))
			if err != nil {
				t.Error(err)
				return
			}
			err = f.Close()
			if err != nil {
				t.Error(err)
				return
			}

			id, err := RetrieveNewImageID()
			if err != nil {
				t.Error(err)
				return
			}

			if id != test.Expected {
				t.Errorf("got %s, expected %s", id, test.Expected)
			}

			err = os.Remove(filename)
			if err != nil {
				t.Error(err)
			}
		})
	}
}
