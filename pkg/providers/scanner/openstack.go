package scanner

import (
	"encoding/json"
	"errors"
	"fmt"
	ostack "github.com/drewbernetes/baski/pkg/providers/openstack"
	sshconnect "github.com/drewbernetes/baski/pkg/remote"
	"github.com/drewbernetes/baski/pkg/s3"
	"github.com/drewbernetes/baski/pkg/trivy"
	"github.com/drewbernetes/baski/pkg/util/flags"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/keypairs"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/servers"
	"github.com/gophercloud/gophercloud/openstack/imageservice/v2/images"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/layer3/floatingips"
	"log"
	"os"
	"time"
)

type OpenStackScannerClient struct {
	BaseScanner
	Img *images.Image

	computeClient *ostack.ComputeClient
	imageClient   *ostack.ImageClient
	networkClient *ostack.NetworkClient
	keyPair       *keypairs.KeyPair
	fip           *floatingips.FloatingIP
	s3Credentials *s3.S3
	server        *servers.Server
	severity      trivy.Severity
}

// NewOpenStackScanner returns new scanner client.
func NewOpenStackScanner(c *ostack.ComputeClient, i *ostack.ImageClient, n *ostack.NetworkClient, s3Conn *s3.S3, severity trivy.Severity, img *images.Image) *OpenStackScannerClient {
	return &OpenStackScannerClient{
		computeClient: c,
		imageClient:   i,
		networkClient: n,
		s3Credentials: s3Conn,
		severity:      severity,
		Img:           img,
	}
}

// RunScan builds the server for scanning and starts the scan
func (s *OpenStackScannerClient) RunScan(o *flags.ScanOptions) error {
	trivyOptions := trivy.New(o.TrivyignorePath, o.TrivyignoreFilename, o.TrivyignoreList, s.severity)
	err := s.getKeypair(s.Img.ID)
	if err != nil {
		return err
	}
	err = s.getFip(o.FloatingIPNetworkName)
	if err != nil {
		return err
	}

	userData, err := trivyOptions.GenerateTrivyCommand(s.s3Credentials)
	if err != nil {
		return err
	}

	err = s.buildServer(o.FlavorName, o.NetworkID, s.Img.ID, o.AttachConfigDrive, userData, []string{o.OpenStackFlags.SecurityGroup})
	if err != nil {
		return err
	}
	return nil
}

func (s *OpenStackScannerClient) FetchScanResults() error {
	//TODO: We need to capture-ctl c and cleanup resources if it's hit.
	client, err := sshconnect.NewSSHClient("ubuntu", s.keyPair.PrivateKey, s.fip.FloatingIP, "22")
	if err != nil {
		return err
	}

	err = fetchResultsFromServer(client, s.Img.ID)
	if err != nil {
		e := removeOpenStackResources(s.server.ID, s.keyPair.Name, s.fip, s.computeClient, s.networkClient)
		if e != nil {
			return e
		}
		return err
	}

	//Close SSH & SFTP connection
	err = client.SFTPClose()
	if err != nil {
		return err
	}
	err = client.SSHClose()
	if err != nil {
		return err
	}

	// Cleanup the scanning resources
	e := removeOpenStackResources(s.server.ID, s.keyPair.Name, s.fip, s.computeClient, s.networkClient)
	if e != nil {
		return e
	}
	return nil
}

//TODO split this out - it's horrid

// CheckResults checks the results file for vulns and parses it into a more friendly format.
func (s *OpenStackScannerClient) CheckResults() error {
	var err error
	j := []byte("{}")

	s.MetaTag = "passed"
	s.ResultsFile = fmt.Sprintf("/tmp/%s.json", s.Img.ID)
	s.Vulns, err = parsingVulnerabilities(s.ResultsFile)
	if err != nil {
		return err
	}
	if len(s.Vulns) != 0 {
		j, err = json.Marshal(s.Vulns)
		if err != nil {
			return errors.New("couldn't marshall vulnerability trivyIgnoreFile: " + err.Error())
		}
		s.MetaTag = "failed"
	}

	// write the vulnerabilities into the results file
	err = os.WriteFile(s.ResultsFile, j, os.FileMode(0644))
	if err != nil {
		return errors.New("couldn't write vulnerability trivyIgnoreFile to file: " + err.Error())
	}

	return nil
}

// TagImage Tags the image with the passed or failed property.
func (s *OpenStackScannerClient) TagImage() error {
	err := s.imageClient.TagImage(s.Img.Properties, s.Img.ID, s.MetaTag, "security_scan")
	if err != nil {
		return err
	}

	return nil
}

// UploadResultsToS3 uploads the scan results to S3.
func (s *OpenStackScannerClient) UploadResultsToS3() error {
	//Upload results to S3
	f, err := os.Open(s.ResultsFile)
	if err != nil {
		return err
	}
	defer f.Close()

	err = s.s3Credentials.Put("text/plain", fmt.Sprintf("scans/%s/%s", s.Img.ID, "results.json"), f)
	if err != nil {
		return err
	}

	return nil
}

func (s *OpenStackScannerClient) getKeypair(imgID string) error {
	kp, err := s.computeClient.CreateKeypair(imgID)
	if err != nil {
		return err
	}
	s.keyPair = kp
	return nil
}

func (s *OpenStackScannerClient) getFip(fipNetworkName string) error {

	fip, err := s.networkClient.GetFloatingIP(fipNetworkName)
	if err != nil {
		e := s.computeClient.RemoveKeypair(s.keyPair.Name)
		if e != nil {
			return e
		}
		return err
	}
	s.fip = fip
	return nil
}

// buildServer is responsible for building the server
func (s *OpenStackScannerClient) buildServer(flavor, networkID, imgID string, attachConfigDrive bool, userData []byte, securityGroups []string) error {
	server, err := s.computeClient.CreateServer(s.keyPair.Name, flavor, networkID, &attachConfigDrive, userData, imgID, securityGroups)
	if err != nil {
		e := s.computeClient.RemoveKeypair(s.keyPair.Name)
		if e != nil {
			return e
		}
		e = s.networkClient.RemoveFIP(s.fip.ID)
		if e != nil {
			return e
		}
		return err
	}

	state, err := s.computeClient.GetServerStatus(server.ID)
	if err != nil {
		e := removeOpenStackResources(server.ID, s.keyPair.Name, s.fip, s.computeClient, s.networkClient)
		if e != nil {
			return e
		}
		return err
	}
	checkLimit := 0
	for !state {
		if checkLimit > 100 {
			panic(errors.New("server failed to com online after 500 seconds"))
		}
		log.Println("server not active, waiting 5 seconds and then checking again")
		time.Sleep(5 * time.Second)
		state, err = s.computeClient.GetServerStatus(server.ID)
		if err != nil {
			e := removeOpenStackResources(server.ID, s.keyPair.Name, s.fip, s.computeClient, s.networkClient)
			if e != nil {
				return e
			}
			return err
		}
		checkLimit++
	}

	err = s.computeClient.AttachIP(server.ID, s.fip.FloatingIP)
	if err != nil {
		e := removeOpenStackResources(server.ID, s.keyPair.Name, s.fip, s.computeClient, s.networkClient)
		if e != nil {
			return e
		}
		return err
	}

	s.server = server

	return nil
}

// removeOpenStackResources cleans up the server and keypair from Openstack to ensure nothing is left lying around.
func removeOpenStackResources(serverID, keyName string, fip *floatingips.FloatingIP, c *ostack.ComputeClient, n *ostack.NetworkClient) error {
	err := c.RemoveServer(serverID)
	if err != nil {
		return err
	}
	err = c.RemoveKeypair(keyName)
	if err != nil {
		return err
	}
	err = n.RemoveFIP(fip.ID)
	if err != nil {
		return err
	}
	return nil
}
