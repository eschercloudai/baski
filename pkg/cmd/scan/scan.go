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

package scan

import (
	ostack "github.com/eschercloudai/baskio/pkg/openstack"
	sshconnect "github.com/eschercloudai/baskio/pkg/ssh"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/keypairs"
	"github.com/pkg/sftp"
	"log"
	"os"
	"time"
)

// FetchResultsFromServer pulls the results.json from the remote scanning server.
func FetchResultsFromServer(freeIP string, kp *keypairs.KeyPair) (*os.File, error) {
	client, err := sshconnect.NewClient(kp, freeIP)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	log.Println("Successfully connected to ssh server.")
	log.Println("waiting 2 minutes for the results of the scan to become available.")
	time.Sleep(2 * time.Minute)

	remoteCommand := "while [ ! -f /tmp/results.json ] && [ -s /tmp/results.json ] ; do echo \"results not ready\"; sleep 5; done;"
	err = sshconnect.RunRemoteCommand(client, remoteCommand)
	if err != nil {
		return nil, err
	}

	// open an SFTP session over an existing ssh connection.
	log.Println("pulling results.")
	sftpConnection, err := sftp.NewClient(client)
	if err != nil {
		return nil, err
	}
	defer sftpConnection.Close()

	resultsFile, err := sshconnect.CopyFromRemoteServer(sftpConnection, "/tmp/", "/tmp/", "results.json")

	//Check there is data in the file in case it was pulled early.
	fi, err := resultsFile.Stat()
	if err != nil {
		log.Println(err.Error())
	}

	for fi.Size() == 0 {
		resultsFile, err = sshconnect.CopyFromRemoteServer(sftpConnection, "/tmp/", "/tmp/", "results.json")

		//Check there is data in the file in case it was pulled early.
		fi, err = resultsFile.Stat()
		if err != nil {
			log.Println(err.Error())
		}
	}

	return resultsFile, err
}

// RemoveScanningResources cleans up the server and keypair from Openstack to ensure nothing is left lying around.
func RemoveScanningResources(serverID, keyName string, os *ostack.Client) {
	os.RemoveServer(serverID)
	os.RemoveKeypair(keyName)
}
