package ostack

import (
	"os"
)

func generateCloudsFile() error {
	var err error
	f, err := os.Create(cloudPath)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write([]byte(cloud))
	if err != nil {
		return err
	}

	err = os.Setenv("OS_CLIENT_CONFIG_FILE", cloudPath)
	if err != nil {
		return err
	}

	return err
}
