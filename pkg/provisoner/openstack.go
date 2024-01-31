package provisoner

import (
	"bufio"
	"encoding/json"
	"fmt"
	ostack "github.com/drewbernetes/baski/pkg/providers/openstack"
	"github.com/drewbernetes/baski/pkg/providers/packer"
	"github.com/drewbernetes/baski/pkg/providers/scanner"
	"github.com/drewbernetes/baski/pkg/s3"
	"github.com/drewbernetes/baski/pkg/trivy"
	"github.com/drewbernetes/baski/pkg/util/data"
	"github.com/drewbernetes/baski/pkg/util/flags"
	"github.com/drewbernetes/baski/pkg/util/sign"
	"github.com/gophercloud/gophercloud/openstack/imageservice/v2/images"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

// OpenStackBuildProvisioner contains the options for the build.
type OpenStackBuildProvisioner struct {
	Opts *flags.BuildOptions
}

// newOpenStackBuilder returns a new OpenStackBuildProvisioner.
func newOpenStackBuilder(o *flags.BuildOptions) *OpenStackBuildProvisioner {
	p := &OpenStackBuildProvisioner{
		Opts: o,
	}

	return p
}

// Init will set an ENV VAR so that the OpenStack builder knows which cloud to use.
func (p *OpenStackBuildProvisioner) Init() error {
	err := os.Setenv("OS_CLOUD", p.Opts.OpenStackFlags.CloudName)
	if err != nil {
		return err
	}
	return nil
}

// GeneratePackerConfig generates a packer vars file for OpenStack builds.
func (p *OpenStackBuildProvisioner) GeneratePackerConfig() *packer.GlobalBuildConfig {
	o := p.Opts
	b, imgName := packer.NewCoreBuildconfig(o)

	b.OpenStackBuildconfig = packer.OpenStackBuildconfig{
		AttachConfigDrive:     strconv.FormatBool(o.OpenStackFlags.AttachConfigDrive),
		Flavor:                o.OpenStackFlags.FlavorName,
		FloatingIpNetwork:     o.OpenStackFlags.FloatingIPNetworkName,
		ImageDiskFormat:       o.OpenStackFlags.ImageDiskFormat,
		ImageVisibility:       o.OpenStackFlags.ImageVisibility,
		ImageName:             imgName,
		Networks:              o.OpenStackFlags.NetworkID,
		SecurityGroup:         o.OpenStackFlags.SecurityGroup,
		SourceImage:           o.OpenStackFlags.SourceImageID,
		UseBlockStorageVolume: o.OpenStackFlags.UseBlockStorageVolume,
		UseFloatingIp:         strconv.FormatBool(o.OpenStackFlags.UseFloatingIP),
		VolumeType:            o.OpenStackFlags.VolumeType,
		VolumeSize:            strconv.Itoa(o.OpenStackFlags.VolumeSize),
	}

	if len(o.OpenStackFlags.SSHPrivateKeyFile) > 0 && len(o.OpenStackFlags.SSHKeypairName) > 0 {
		b.OpenStackBuildconfig.SSHPrivateKeyFile = o.OpenStackFlags.SSHPrivateKeyFile
		b.OpenStackBuildconfig.SSHKeypairName = o.OpenStackFlags.SSHKeypairName
	}

	b.Metadata = generateBuilderMetadata(o)

	if len(o.RootfsUUID) > 0 {
		b.Metadata["rootfs_uuid"] = o.RootfsUUID
	}

	return b
}

// UpdatePackerBuilders will update the builders field with the metadata values as required. This is done this way as passing it in via Packer vars is prone to error or just complete failures.
func (p *OpenStackBuildProvisioner) UpdatePackerBuilders(metadata map[string]string, data []byte) []byte {
	jsonStruct := struct {
		Builders       []map[string]interface{} `json:"builders"`
		PostProcessors []map[string]interface{} `json:"post-processors"`
		Provisioners   []map[string]interface{} `json:"provisioners"`
		Variables      map[string]interface{}   `json:"variables"`
	}{}

	err := json.Unmarshal(data, &jsonStruct)
	if err != nil {
		log.Fatalln(err)
		return nil
	}

	jsonStruct.Builders[0]["metadata"] = metadata

	res, err := json.Marshal(jsonStruct)
	if err != nil {
		log.Fatalln(err)
		return nil
	}

	return res
}

// PostBuildAction retrieves the image ID from the output and stores it into a file.
func (p *OpenStackBuildProvisioner) PostBuildAction() error {

	imgID, err := retrieveNewOpenStackImageID()
	if err != nil {
		return err
	}

	err = saveImageIDToFile(imgID)
	if err != nil {
		return err
	}

	return nil
}

// retrieveNewOpenStackImageID identifies the new ImageID from the output text so that it can be used/retrieved later.
func retrieveNewOpenStackImageID() (string, error) {
	var i string

	//TODO: If the output goes to stdOUT in buildImage,
	// we need to figure out if we can pull this from the openstack instance instead.
	f, err := os.Open("/tmp/out-build.txt")
	if err != nil {
		return "", err
	}
	defer f.Close()

	r := bufio.NewScanner(f)
	re := regexp.MustCompile("An image was created: [0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}")
	for r.Scan() {
		m := re.MatchString(string(r.Bytes()))
		if m {
			//There is likely two outputs here due to how packer outputs, so we need to break on the first find.
			i = strings.Split(r.Text(), ": ")[2]
			break
		}
	}

	return i, nil
}

// OpenStackScanProvisioner
type OpenStackScanProvisioner struct {
	Opts          *flags.ScanOptions
	imageClient   *ostack.ImageClient
	computeClient *ostack.ComputeClient
	networkClient *ostack.NetworkClient
	imageWildCard string
	imageID       string
}

// newOpenStackScanner
func newOpenStackScanner(o *flags.ScanOptions) *OpenStackScanProvisioner {
	p := &OpenStackScanProvisioner{
		Opts: o,
	}

	return p
}

// Prepare
func (s *OpenStackScanProvisioner) Prepare() error {
	var err error
	o := s.Opts

	o.OpenStackFlags.FlavorName = o.FlavorName

	cloudProvider := ostack.NewCloudsProvider(o.OpenStackFlags.CloudName)

	s.imageClient, err = ostack.NewImageClient(cloudProvider)
	if err != nil {
		return err
	}

	s.computeClient, err = ostack.NewComputeClient(cloudProvider)
	if err != nil {
		return err
	}

	s.networkClient, err = ostack.NewNetworkClient(cloudProvider)
	if err != nil {
		return err
	}

	s.imageID = o.ScanSingleOptions.ImageID
	s.imageWildCard = o.ScanMultipleOptions.ImageSearch

	return nil
}

// ScanImages
func (s *OpenStackScanProvisioner) ScanImages() error {
	var err error
	o := s.Opts

	imgs := []images.Image{}

	// Parse the image ID or wildcard and load the images from OpenStack
	if s.imageID != "" {
		var img *images.Image

		img, err = s.imageClient.FetchImage(o.ScanSingleOptions.ImageID)
		if err != nil {
			return err
		}

		imgs = append(imgs, *img)
	} else if s.imageWildCard != "" {
		imgs, err = s.imageClient.FetchAllImages(o.ImageSearch)
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("no image(s) provided")
	}

	severity := trivy.Severity(strings.ToUpper(o.MaxSeverityType))

	var s3Conn *s3.S3

	s3Conn, err = s3.New(o.S3Flags.Endpoint, o.S3Flags.AccessKey, o.S3Flags.SecretKey, o.ScanBucket, o.S3Flags.Region)
	if err != nil {
		log.Println(err)
		return err
	}

	// Let's scan a bunch of images based on the concurrency
	semaphore := make(chan struct{}, o.Concurrency)
	var wg sync.WaitGroup

	for _, img := range imgs {
		wg.Add(1)
		semaphore <- struct{}{}
		go func(image images.Image) {
			defer func() {
				<-semaphore // Release the slot in the semaphore
			}()

			sc := scanner.NewOpenStackScanner(s.computeClient, s.imageClient, s.networkClient, s3Conn, severity, &image)
			err = s.scanServer(sc, &wg)
			if err != nil {
				log.Println(err)
			}

		}(img)
	}
	wg.Wait()

	close(semaphore)

	return nil
}

// scanServer will scan, parse the results and upload them to S3. It's in its own function for the purpose of threading.
func (s *OpenStackScanProvisioner) scanServer(sc *scanner.OpenStackScannerClient, wg *sync.WaitGroup) error {
	defer wg.Done()
	o := s.Opts

	log.Printf("Processing Image with ID: %s\n", sc.Img.ID)

	// Run the scan.
	err := sc.RunScan(o)
	if err != nil {
		return err
	}

	// Fetch the results and write them to a file locally.
	err = sc.FetchScanResults()
	if err != nil {
		return err
	}

	// Read the local results file and parse them into a more consumable json format, then write out to file.
	err = sc.CheckResults()
	if err != nil {
		return err
	}

	// If the image is not set to auto delete, tag the image with the check result.
	if !o.AutoDeleteImage {
		err = sc.TagImage()
		if err != nil {
			return err
		}
	} else {
		if len(sc.Vulns) != 0 {
			// Remove the image if vulns are found and the flag/config item is set.
			err = s.imageClient.RemoveImage(sc.Img.ID)
			if err != nil {
				return err
			}
		}
	}

	// Upload the parsed results file to S3
	err = sc.UploadResultsToS3()
	if err != nil {
		return err
	}

	log.Printf("Finished processing Image ID: %s\n", sc.Img.ID)

	// Check if the CVE checking is being skipped, if not then bail out here.
	if !o.SkipCVECheck {
		errMsg := "vulnerabilities detected above threshold. Please see the possible fixes located at '/tmp/results.json' for further information on this"
		if o.AutoDeleteImage {
			errMsg = fmt.Sprintf("%s - %s", errMsg, ". The image has been removed from the infra.")
		}
		return fmt.Errorf(errMsg)
	}
	return nil
}

type OpenStackSignProvisioner struct {
	Opts *flags.SignOptions
}

// newOpenStackSigner
func newOpenStackSigner(o *flags.SignOptions) *OpenStackSignProvisioner {
	p := &OpenStackSignProvisioner{
		Opts: o,
	}

	return p
}

// SignImage
func (s *OpenStackSignProvisioner) SignImage(digest string) error {
	o := s.Opts
	cloudProvider := ostack.NewCloudsProvider(o.OpenStackCoreFlags.CloudName)

	i, err := ostack.NewImageClient(cloudProvider)
	if err != nil {
		return err
	}

	img, err := i.FetchImage(o.ImageID)
	if err != nil {
		return err
	}

	err = i.TagImage(img.Properties, o.ImageID, digest, "digest")
	if err != nil {
		return err
	}

	return nil
}

// ValidateImage
func (s *OpenStackSignProvisioner) ValidateImage(key []byte) error {
	o := s.Opts
	cloudProvider := ostack.NewCloudsProvider(o.OpenStackCoreFlags.CloudName)

	i, err := ostack.NewImageClient(cloudProvider)
	if err != nil {
		return err
	}

	img, err := i.FetchImage(o.ImageID)
	if err != nil {
		return err
	}

	field, err := data.GetNestedField(img.Properties, "digest")
	if err != nil {
		return err
	}
	if field == nil {
		return fmt.Errorf("the digest field was empty")
	}

	digest := field.(string)

	valid, err := sign.Validate(o.ImageID, key, digest)
	if err != nil {
		return err
	}

	log.Printf("The validation result was: %t", valid)

	return nil
}
