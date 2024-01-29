package scanner

import (
	"encoding/json"
	"errors"
	"fmt"
	ostack "github.com/eschercloudai/baski/pkg/providers/openstack"
	sshconnect "github.com/eschercloudai/baski/pkg/remote"
	"github.com/eschercloudai/baski/pkg/s3"
	"github.com/eschercloudai/baski/pkg/trivy"
	"github.com/eschercloudai/baski/pkg/util"
	"github.com/eschercloudai/baski/pkg/util/data"
	"github.com/eschercloudai/baski/pkg/util/flags"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/keypairs"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/servers"
	"github.com/gophercloud/gophercloud/openstack/imageservice/v2/images"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/layer3/floatingips"
	"log"
	"os"
	"time"
)

type ScannerClient struct {
	computeClient *ostack.ComputeClient
	imageClient   *ostack.ImageClient
	networkClient *ostack.NetworkClient
	keyPair       *keypairs.KeyPair
	fip           *floatingips.FloatingIP
	s3Credentials *s3.S3
	server        *servers.Server
	trivyOptions  *trivy.TrivyOptions
}

// fetchResultsFromServer pulls the results.json from the remote scanning server.
func fetchResultsFromServer(client util.SSHInterface) error {
	log.Println("Successfully connected to ssh server")
	log.Println("checking for scan completion")
	retries := 20
	for !hasScanCompleted(client) {
		if retries <= 0 {
			return errors.New("couldn't fetch the results - timed out waiting for condition")
		}
		log.Printf("scan still running... %d retries left\n", retries)
		time.Sleep(10 * time.Second)
		retries -= 1
	}
	log.Println("scan completed, fetching results")

	_, err := client.CopyFromRemoteServer("/tmp/", "/tmp/", "results.json")
	if err != nil {
		log.Println(err.Error())
		return err
	}
	return nil
}

func hasScanCompleted(client util.SSHInterface) bool {
	status, err := client.CopyFromRemoteServer("/tmp/", "/tmp/", "finished")
	if err != nil {
		return false
	}

	fi, err := os.Stat(status.Name())
	if err != nil {
		log.Println(err.Error())
		return false
	}

	if fi.Size() == 0 {
		return false
	}
	return true
}

// removeScanningResources cleans up the server and keypair from Openstack to ensure nothing is left lying around.
func removeScanningResources(serverID, keyName string, fip *floatingips.FloatingIP, c *ostack.ComputeClient, n *ostack.NetworkClient) error {
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

// parsingVulnerabilities will read the results file and parse it into a more user friendly format.
func parsingVulnerabilities() ([]trivy.ScanFailedReport, error) {
	log.Println("checking results")
	rf, err := os.ReadFile("/tmp/results.json")
	if err != nil {
		return nil, err
	}

	report := &trivy.Report{}

	err = json.Unmarshal(rf, report)

	if err != nil {
		return nil, err
	}

	var vf []trivy.ScanFailedReport

	for _, r := range report.Results {
		for _, v := range r.Vulnerabilities {
			//if checkSeverityThresholdPassed(trivy.Severity(v.Severity), v.Cvss, checkScore, trivy.Severity(severityThreshold)) {
			vuln := trivy.ScanFailedReport{
				VulnerabilityID:  v.VulnerabilityID,
				Description:      v.Description,
				PkgName:          v.PkgName,
				InstalledVersion: v.InstalledVersion,
				FixedVersion:     v.FixedVersion,
				Severity:         v.Severity,
				Cvss:             v.Cvss,
			}
			//// We don't need all scores in here, so we just grab the one that triggered the threshold
			//if v.Cvss.Nvd != nil {
			//	if v.Cvss.Nvd.V3Score >= checkScore {
			//		vuln.Cvss = trivy.CVSS{Nvd: &trivy.Score{V3Score: v.Cvss.Nvd.V3Score}}
			//	} else if v.Cvss.Nvd.V2Score >= checkScore {
			//		vuln.Cvss = trivy.CVSS{Nvd: &trivy.Score{V2Score: v.Cvss.Nvd.V2Score}}
			//	}
			//} else if v.Cvss.Redhat != nil {
			//	if v.Cvss.Redhat.V3Score >= checkScore {
			//		vuln.Cvss = trivy.CVSS{Redhat: &trivy.Score{V3Score: v.Cvss.Redhat.V3Score}}
			//	} else if v.Cvss.Redhat.V2Score >= checkScore {
			//		vuln.Cvss = trivy.CVSS{Redhat: &trivy.Score{V2Score: v.Cvss.Redhat.V2Score}}
			//	}
			//} else if v.Cvss.Ghsa != nil {
			//	if v.Cvss.Ghsa.V3Score >= checkScore {
			//		vuln.Cvss = trivy.CVSS{Ghsa: &trivy.Score{V3Score: v.Cvss.Ghsa.V3Score}}
			//	}
			//}

			vf = append(vf, vuln)
		}
		//}
	}
	return vf, nil
}

//// checkSeverityThresholdPassed checks for a score that is >= checkScore and checkSeverity. It will return true if so.
//func checkSeverityThresholdPassed(severity trivy.Severity, cvss trivy.CVSS, checkScore float64, severityThreshold trivy.Severity) bool {
//	if cvss.Nvd != nil {
//		if cvss.Nvd.V3Score >= checkScore && trivy.CheckSeverity(severity, severityThreshold) {
//			return true
//		} else if cvss.Nvd.V2Score >= checkScore && trivy.CheckSeverity(severity, severityThreshold) {
//			return true
//		}
//	}
//	if cvss.Redhat != nil {
//		if cvss.Redhat.V3Score >= checkScore && trivy.CheckSeverity(severity, severityThreshold) {
//			return true
//		} else if cvss.Redhat.V2Score >= checkScore && trivy.CheckSeverity(severity, severityThreshold) {
//			return true
//		}
//	}
//	if cvss.Ghsa != nil {
//		if cvss.Ghsa.V3Score >= checkScore && trivy.CheckSeverity(severity, severityThreshold) {
//			return true
//		}
//	}
//	return false
//}

func (s *ScannerClient) getKeypair(imgID string) error {
	kp, err := s.computeClient.CreateKeypair(imgID)
	if err != nil {
		return err
	}
	s.keyPair = kp
	return nil
}

func (s *ScannerClient) getFip(fipNetworkName string) error {

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

func (s *ScannerClient) buildServer(flavor, networkID, imgID string, attachConfigDrive bool, userData []byte) error {

	server, err := s.computeClient.CreateServer(s.keyPair.Name, flavor, networkID, &attachConfigDrive, userData, imgID)
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
		e := removeScanningResources(server.ID, s.keyPair.Name, s.fip, s.computeClient, s.networkClient)
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
			e := removeScanningResources(server.ID, s.keyPair.Name, s.fip, s.computeClient, s.networkClient)
			if e != nil {
				return e
			}
			return err
		}
		checkLimit++
	}

	err = s.computeClient.AttachIP(server.ID, s.fip.FloatingIP)
	if err != nil {
		e := removeScanningResources(server.ID, s.keyPair.Name, s.fip, s.computeClient, s.networkClient)
		if e != nil {
			return e
		}
		return err
	}

	s.server = server

	return nil
}

func NewScanner(c *ostack.ComputeClient, i *ostack.ImageClient, n *ostack.NetworkClient, s3Conn *s3.S3) *ScannerClient {
	return &ScannerClient{
		computeClient: c,
		imageClient:   i,
		networkClient: n,
		s3Credentials: s3Conn,
	}
}

func (s *ScannerClient) RunScan(o *flags.ScanOptions, severity trivy.Severity, img *images.Image) error {
	s.trivyOptions = trivy.New(o.TrivyignorePath, o.TrivyignoreFilename, o.TrivyignoreList, severity)
	err := s.getKeypair(img.ID)
	if err != nil {
		return err
	}
	err = s.getFip(o.FloatingIPNetworkName)
	if err != nil {
		return err
	}

	userData, err := s.trivyOptions.GenerateTrivyCommand(s.s3Credentials)
	if err != nil {
		return err
	}

	err = s.buildServer(o.FlavorName, o.NetworkID, img.ID, o.AttachConfigDrive, userData)
	if err != nil {
		return err
	}
	return nil
}

func (s *ScannerClient) FetchScanResults() error {
	//TODO: We need to capture-ctl c and cleanup resources if it's hit.
	client, err := sshconnect.NewSSHClient(s.keyPair, s.fip.FloatingIP)
	if err != nil {
		return err
	}

	err = fetchResultsFromServer(client)
	if err != nil {
		e := removeScanningResources(s.server.ID, s.keyPair.Name, s.fip, s.computeClient, s.networkClient)
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
	e := removeScanningResources(s.server.ID, s.keyPair.Name, s.fip, s.computeClient, s.networkClient)
	if e != nil {
		return e
	}
	return nil
}

//TODO split this out - it's horrid

// CheckResultsTagImageAndUploadToS3 checks the results file for vulns and parses it into a more friendly format. Then it tags the image with the passed or failed property, and then it uploads to S3 - horrible.
func (s *ScannerClient) CheckResultsTagImageAndUploadToS3(img *images.Image, autoDelete, skipCVECheck bool) error {
	// Default to replace unless the field isn't found below
	operation := images.ReplaceOp

	field, err := data.GetNestedField(img.Properties, "image", "metadata", "security_scan")
	if err != nil || field == nil {
		operation = images.AddOp
	}
	metaValue := "passed"
	resultsFile := fmt.Sprintf("/tmp/%s.json", img.ID)

	vulns, err := parsingVulnerabilities()
	if err != nil {
		return err
	}
	if len(vulns) != 0 {
		if autoDelete {
			err = s.imageClient.RemoveImage(img.ID)
			if err != nil {
				return err
			}
		}
		var j []byte
		j, err = json.Marshal(vulns)
		if err != nil {
			return errors.New("couldn't marshall vulnerability trivyIgnoreFile: " + err.Error())
		}

		// write the vulnerabilities into the results file
		err = os.WriteFile(resultsFile, j, os.FileMode(0644))
		if err != nil {
			return errors.New("couldn't write vulnerability trivyIgnoreFile to file: " + err.Error())
		}

		metaValue = "failed"
	} else {
		err = os.WriteFile(resultsFile, []byte("{}"), os.FileMode(0644))
		if err != nil {
			return errors.New("couldn't write vulnerability trivyIgnoreFile to file: " + err.Error())
		}
	}
	if !autoDelete {
		_, err = s.imageClient.ModifyImageMetadata(img.ID, "security_scan", metaValue, operation)
		if err != nil {
			return err
		}
	}

	//Upload results to S3
	f, err := os.Open(resultsFile)
	if err != nil {
		return err
	}
	defer f.Close()

	err = s.s3Credentials.Put("text/plain", fmt.Sprintf("scans/%s/%s", img.ID, "results.json"), f)
	if err != nil {
		return err
	}

	if !skipCVECheck {
		errMsg := "vulnerabilities detected above threshold. Please see the possible fixes located at '/tmp/results.json' for further information on this"
		if autoDelete {
			errMsg = fmt.Sprintf("%s - %s", errMsg, ". The image has been removed from the infra.")
		}
		return fmt.Errorf(errMsg)
	}
	return nil
}
