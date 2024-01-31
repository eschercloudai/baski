package scanner

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/drewbernetes/baski/pkg/trivy"
	"github.com/drewbernetes/baski/pkg/util"
	"log"
	"os"
	"time"
)

//TODO: This may not be needed afterall
//type ScannerInterface interface {
//	RunScan(o *flags.ScanOptions) error
//	FetchScanResults() error
//	CheckResults() error
//	TagImage() error
//	UploadResultsToS3() error
//}

type BaseScanner struct {
	ResultsFile string
	MetaTag     string
	Vulns       []trivy.ScanFailedReport
}

// fetchResultsFromServer pulls the results.json from the remote scanning server.
func fetchResultsFromServer(client util.SSHInterface, imgID string) error {
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

	_, err := client.CopyFromRemoteServer("/tmp/results.json", fmt.Sprintf("/tmp/%s.json", imgID))
	if err != nil {
		log.Println(err.Error())
		return err
	}
	return nil
}

func hasScanCompleted(client util.SSHInterface) bool {
	status, err := client.CopyFromRemoteServer("/tmp/finished", "/tmp/finished")
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

// parsingVulnerabilities will read the results file and parse it into a more user-friendly format.
func parsingVulnerabilities(resultsFile string) ([]trivy.ScanFailedReport, error) {
	log.Println("checking results")
	rf, err := os.ReadFile(resultsFile)
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
			vuln := trivy.ScanFailedReport{
				VulnerabilityID:  v.VulnerabilityID,
				Description:      v.Description,
				PkgName:          v.PkgName,
				InstalledVersion: v.InstalledVersion,
				FixedVersion:     v.FixedVersion,
				Severity:         v.Severity,
				Cvss:             v.Cvss,
			}

			vf = append(vf, vuln)
		}
	}
	return vf, nil
}
