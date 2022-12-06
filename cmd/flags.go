package cmd

var (
	//Root
	baskioConfigFlag string
	cloudsPathFlag   string
	cloudNameFlag    string
	verboseFlag      bool

	//Build & Scan
	imageRepoFlag             string
	buildOSFlag               string
	sourceImageIDFlag         string
	networkIDFlag             string
	flavorNameFlag            string
	userFloatingIPFlag        bool
	floatingIPNetworkNameFlag string
	attachConfigDriveFlag     bool
	imageVisibilityFlag       string
	crictlVersionFlag         string
	kubeVersionFlag           string

	addNvidiaSupportFlag   bool
	nvidiaVersionFlag      string
	nvidiaInstallerURLFlag string
	gridLicenseServerFlag  string
	imageIDFlag            string

	// Publish
	ghUserFlag        string
	ghProjectFlag     string
	ghAccountFlag     string
	ghTokenFlag       string
	ghPagesBranchFlag string
)
