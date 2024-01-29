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

package remote

import (
	"encoding/base64"
	"fmt"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/keypairs"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type SSHClient struct {
	SSH  *ssh.Client
	SFTP *sftp.Client
}

// keyString is used to generate key that will be used to validate a remote host.
func keyString(k ssh.PublicKey) string {
	return k.Type() + "" + base64.StdEncoding.EncodeToString(k.Marshal())
}

// trustedHostKeyCallback checks a hosts key to ensure it is valid.
// If a blank one is passed, it is presumed this is the first time connecting to a server.
// If it doesn't match, then a warning is returned and the SSH connection will fail as a result.
func trustedHostKeyCallback(key string) ssh.HostKeyCallback {
	if key == "" {
		return func(_ string, _ net.Addr, k ssh.PublicKey) error {
			log.Println("WARNING: SSH-key verification is not in effect")
			return nil
		}
	}

	return func(_ string, _ net.Addr, k ssh.PublicKey) error {
		ks := keyString(k)
		if key != ks {
			return fmt.Errorf("SSH-key verification: expected %q but got %q\n", key, ks)
		}

		return nil
	}
}

// NewSSHClient creates a new ssh connection to a remote server.
// It will attempt to connect up to 10 times with a 10-second gap to prevent a crash
// should the first attempt fail while the server is booting.
func NewSSHClient(kp *keypairs.KeyPair, ip string) (*SSHClient, error) {
	var hostKey string
	var client *ssh.Client

	// Create the Signer for this private key.
	signer, err := ssh.ParsePrivateKey([]byte(kp.PrivateKey))
	if err != nil {
		return nil, err
	}

	config := &ssh.ClientConfig{
		User: "ubuntu",
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: trustedHostKeyCallback(hostKey),
	}

	log.Println("waiting for server to boot.")
	for i := 10; i > 0; i-- {
		// Connect to the remote server and perform the SSH handshake.
		client, err = ssh.Dial("tcp", strings.Join([]string{ip, "22"}, ":"), config)
		if err != nil {
			if i > 0 {
				log.Printf("ssh unavailable - server is probably still booting. %d retires left\n", i)
				time.Sleep(10 * time.Second)
			} else {
				return nil, err
			}
		} else {
			break
		}
	}

	c := &SSHClient{SSH: client}

	// open an SFTP session over an existing ssh connection.
	c.SFTP, err = sftp.NewClient(c.SSH)
	if err != nil {
		return nil, err
	}
	return c, err
}

// CopyFromRemoteServer uses sftp to copy a file from a remotes server to a local directory.
func (s *SSHClient) CopyFromRemoteServer(srcPath, dstPath, filename string) (*os.File, error) {
	src := filepath.Join(srcPath, filename)
	dst := filepath.Join(dstPath, filename)

	// Open the source file
	srcFile, err := s.SFTP.Open(src)
	if err != nil {
		return nil, err
	}
	defer srcFile.Close()

	// Create the destination file
	dstFile, err := os.Create(dst)
	if err != nil {
		return nil, err
	}
	defer dstFile.Close()

	// Copy the file
	_, err = srcFile.WriteTo(dstFile)
	if err != nil {
		return nil, err
	}

	return dstFile, nil
}

// SSHClose just runs client.SSHClose()
func (s *SSHClient) SSHClose() error {
	return s.SSH.Close()
}

// SFTPClose just runs client.SFTPClose()
func (s *SSHClient) SFTPClose() error {
	return s.SFTP.Close()
}
