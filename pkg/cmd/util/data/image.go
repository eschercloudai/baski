package data

import (
	"bufio"
	"os"
	"regexp"
	"strings"
)

// RetrieveNewImageID fetches the newly create image's ID from the out.txt file
// that is generated during the buildImage() run.
func RetrieveNewImageID() (string, error) {
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
