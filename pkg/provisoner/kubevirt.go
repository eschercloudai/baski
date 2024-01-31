package provisoner

import (
	"context"
	"fmt"
	"github.com/drewbernetes/baski/pkg/k8s"
	"github.com/drewbernetes/baski/pkg/providers/packer"
	"github.com/drewbernetes/baski/pkg/util/flags"
	simple_s3 "github.com/drewbernetes/simple-s3"
	v1 "k8s.io/api/core/v1"
	errorsv1 "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	dv_client "kubevirt.io/client-go/generated/containerized-data-importer/clientset/versioned"
	"kubevirt.io/containerized-data-importer-api/pkg/apis/core/v1beta1"
	"log"
	"os"
)

// KubeVirtBuildProvisioner contains the options for the build.
type KubeVirtBuildProvisioner struct {
	Opts *flags.BuildOptions
}

// newKubeVirtBuilder returns a new KubeVirtBuildProvisioner
func newKubeVirtBuilder(o *flags.BuildOptions) *KubeVirtBuildProvisioner {
	p := &KubeVirtBuildProvisioner{
		Opts: o,
	}

	return p
}

// Init currently has no action
func (p *KubeVirtBuildProvisioner) Init() error {
	return nil
}

// GeneratePackerConfig generates a packer vars file for KubeVirt builds.
func (p *KubeVirtBuildProvisioner) GeneratePackerConfig() *packer.GlobalBuildConfig {
	o := p.Opts
	b, imgName := packer.NewCoreBuildconfig(o)

	// Set the image name here as it's
	o.KubeVirtFlags.ImageName = imgName

	b.KubeVirtBuildConfig = packer.KubeVirtBuildConfig{
		QemuBinary:      o.KubeVirtFlags.QemuBinary,
		DiskSize:        o.KubeVirtFlags.DiskSize,
		OutputDirectory: fmt.Sprintf("%s/%s", o.KubeVirtFlags.OutputDirectory, o.KubeVirtFlags.ImageName),
	}

	b.Metadata = generateBuilderMetadata(o)

	return b
}

// UpdatePackerBuilders currently has no action
func (p *KubeVirtBuildProvisioner) UpdatePackerBuilders(metadata map[string]string, data []byte) []byte {
	return nil
}

// PostBuildAction will upload the image to S3 if the option is enabled. It will then deploy a DataVolume to the cluster referencing the S3 endpoint.
//
// If the S3 endpoint is not enabled then it will just print out the location of the file.
func (p *KubeVirtBuildProvisioner) PostBuildAction() error {
	ctx := context.Background()
	s3ImgPath := fmt.Sprintf("images/%s.qcow2", p.Opts.KubeVirtFlags.ImageName)
	resultImage := fmt.Sprintf("%s/%s/%s", p.Opts.KubeVirtFlags.OutputDirectory, p.Opts.KubeVirtFlags.ImageName, fmt.Sprintf("%s-kube-v%s", p.Opts.BuildOS, p.Opts.KubeVersion))

	if p.Opts.KubeVirtFlags.StoreInS3 {
		////If S3 enabled, upload to S3
		log.Printf("uploading the image: %s to %s in S3... this may take a few minutes\n", p.Opts.KubeVirtFlags.ImageName, s3ImgPath)
		s, err := simple_s3.New(p.Opts.S3Flags.Endpoint, p.Opts.S3Flags.AccessKey, p.Opts.S3Flags.SecretKey, p.Opts.KubeVirtFlags.ImageBucket, p.Opts.S3Flags.Region)
		if err != nil {
			return err
		}

		f, err := os.Open(resultImage)
		if err != nil {
			return err
		}
		defer f.Close()

		err = s.Put(s3ImgPath, f)
		if err != nil {
			return err
		}

		// Create DV
		client, err := k8s.NewClient(p.Opts.KubernetesClusterFlags.KubeconfigPath)
		if err != nil {
			return err
		}

		log.Printf("checking for Namespace: %s - will create if it doesn't exist\n", p.Opts.KubeVirtFlags.ImageNamespace)
		ns, err := createOrGetNamespace(ctx, client.Client, p.Opts.KubeVirtFlags.ImageNamespace)
		if err != nil {
			return err
		}

		log.Printf("checking for Secret: %s - will create if it doesn't exist\n", "baski-image-s3-credentials")
		secret, err := createOrGetS3Secret(ctx, client.Client, ns.ObjectMeta.Name, p.Opts.S3Flags)
		if err != nil {
			return err
		}

		log.Printf("Creating DataVolume: %s\n", p.Opts.KubeVirtFlags.ImageName)
		err = createDataVolume(ctx, client.KubeVirt, ns.ObjectMeta.Name, secret.ObjectMeta.Name, s3ImgPath, p.Opts)
		if err != nil {
			return err
		}

	} else {
		// TODO: Consider options to upload via the CDI proxy.
		//   To mimic the `virtlctl image-upload` approach would require it all recoding here which seems fruitless.
		//   It would be easier to just run virtctl as a direct exec than copy-pasta the code over that handles uploading an image to a PVC/DV.

		// If not tell user where file is
		log.Printf("the image has been built and exists in %s", resultImage)
	}

	return nil
}

// createOrGetNamespace fetch a Namespace or create it if it doesn't exist
func createOrGetNamespace(ctx context.Context, client *kubernetes.Clientset, namespace string) (*v1.Namespace, error) {

	ns, err := client.CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{})
	if err != nil {
		if errorsv1.IsNotFound(err) {
			ns, err = client.CoreV1().Namespaces().Create(ctx, &v1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: namespace,
				},
			}, metav1.CreateOptions{})
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}
	return ns, nil
}

// createOrGetS3Secret fetch the image-s3 credentials Secret or create it if it doesn't exist
func createOrGetS3Secret(ctx context.Context, client *kubernetes.Clientset, namespace string, s3Opts flags.S3Flags) (*v1.Secret, error) {
	secretName := "baski-image-s3-credentials"

	data := map[string][]byte{
		"accessKeyId": []byte(s3Opts.AccessKey),
		"secretKey":   []byte(s3Opts.SecretKey),
	}

	secret, err := client.CoreV1().Secrets(namespace).Get(ctx, secretName, metav1.GetOptions{})
	if err != nil {
		if errorsv1.IsNotFound(err) {
			secret, err = client.CoreV1().Secrets(namespace).Create(ctx, &v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      secretName,
					Namespace: namespace,
				},
				Data: data,
				Type: "Opaque",
			}, metav1.CreateOptions{})
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}
	return secret, nil
}

// createDataVolume the DataVolume to enable the image to be pulled from S3 and stored as a PVC
func createDataVolume(ctx context.Context, client *dv_client.Clientset, namespace, secret, s3ImgPath string, opts *flags.BuildOptions) error {
	var err error

	dv := &v1beta1.DataVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name:      opts.KubeVirtFlags.ImageName,
			Namespace: namespace,
			Labels: map[string]string{
				"builder": "baski",
			},
		},

		Spec: v1beta1.DataVolumeSpec{
			Source: &v1beta1.DataVolumeSource{
				S3: &v1beta1.DataVolumeSourceS3{
					URL:       fmt.Sprintf("%s/%s", opts.S3Flags.Endpoint, s3ImgPath),
					SecretRef: secret,
				},
			},
			PVC: &v1.PersistentVolumeClaimSpec{
				AccessModes: []v1.PersistentVolumeAccessMode{
					v1.ReadWriteOnce,
				},
				Resources: v1.VolumeResourceRequirements{
					Requests: map[v1.ResourceName]resource.Quantity{
						v1.ResourceStorage: resource.MustParse("10Gi"),
					},
				},
			},
		},
	}

	dv, err = client.CdiV1beta1().DataVolumes(namespace).Create(ctx, dv, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	return nil
}

// KubeVirtScanProvisioner
type KubeVirtScanProvisioner struct {
	Opts *flags.ScanOptions
}

// newKubeVirtScanner
func newKubeVirtScanner(o *flags.ScanOptions) *KubeVirtScanProvisioner {
	p := &KubeVirtScanProvisioner{
		Opts: o,
	}

	return p
}

// Prepare
func (s *KubeVirtScanProvisioner) Prepare() error {

	return nil
}

// ScanImages
func (s *KubeVirtScanProvisioner) ScanImages() error {
	return nil
}

// KubeVirtSignProvisioner
type KubeVirtSignProvisioner struct {
	Opts *flags.SignOptions
}

// newKubeVirtSigner
func newKubeVirtSigner(o *flags.SignOptions) *KubeVirtSignProvisioner {
	p := &KubeVirtSignProvisioner{
		Opts: o,
	}

	return p
}

// SignImage
func (s *KubeVirtSignProvisioner) SignImage(digest string) error {

	return nil
}

// ValidateImage
func (s *KubeVirtSignProvisioner) ValidateImage(key []byte) error {

	return nil
}
