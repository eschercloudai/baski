package provisoner

import (
	"os"
	"testing"
)

func TestInit(t *testing.T) {
	//	err := os.Setenv("OS_CLOUD", p.CloudName)
	//	if err != nil {
	//		return err
	//	}
	//	return nil
}

// TestGeneratePackerConfig generates some glance metadata for the image.
func TestGeneratePackerConfig(t *testing.T) {
	//	th.SetupPersistentPortHTTP(t, th.Port)
	//	defer th.TeardownHTTP()
	//
	//	tests := []struct {
	//		name     string
	//		options  *flags.BuildOptions
	//		expected map[string]string
	//	}{
	//		{
	//			name: "Test with GPU",
	//			options: &flags.BuildOptions{
	//				AddNvidiaSupport: true,
	//				NvidiaVersion:    "1.2.3",
	//				BuildOS:          "ubuntu",
	//				KubeVersion:      "1.28",
	//				OpenStackFlags: flags.OpenStackFlags{
	//					RootfsUUID: "123456",
	//				},
	//			},
	//			expected: map[string]string{
	//				"os":          "ubuntu",
	//				"k8s":         "1.28",
	//				"gpu":         "1.2.3",
	//				"date":        "2006-01-02T15:04:05Z07:00",
	//				"rootfs_uuid": "123456",
	//			},
	//		},
	//		{
	//			name: "Test without GPU",
	//			options: &flags.BuildOptions{
	//				AddNvidiaSupport: false,
	//				BuildOS:          "ubuntu",
	//				KubeVersion:      "1.28",
	//				OpenStackFlags: flags.OpenStackFlags{
	//					RootfsUUID: "123456",
	//				},
	//			},
	//			expected: map[string]string{
	//				"os":          "ubuntu",
	//				"k8s":         "1.28",
	//				"gpu":         "no_gpu",
	//				"date":        "2006-01-02T15:04:05Z07:00",
	//				"rootfs_uuid": "123456",
	//			},
	//		},
	//	}
	//	for _, tc := range tests {
	//		t.Run(tc.name, func(t *testing.T) {
	//			meta := GenerateBuilderMetadata(tc.options)
	//			//We override the dat here as it's based off of time.Now()
	//			meta["date"] = "2006-01-02T15:04:05Z07:00"
	//
	//			if !reflect.DeepEqual(meta, tc.expected) {
	//				t.Errorf("Expected %+v, got %+v", tc.expected, meta)
	//			}
	//		})
	//	}
	//
}

// TestUpdatePackerBuilders generates some glance metadata for the image.
func TestUpdatePackerBuilders(t *testing.T) {

}

// TestPostBuildAction fetches the newly create image's ID from the out.txt file
// that is generated during the buildImage() run.
func TestPostBuildAction(t *testing.T) {

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

			id, err := retrieveNewOpenStackImageID()
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
