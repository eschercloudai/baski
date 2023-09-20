/*
Copyright 2023 EscherCloudAI.

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

package scans

import (
	"encoding/json"
	"fmt"
	"github.com/eschercloudai/baski/pkg/s3"
	"github.com/eschercloudai/baski/pkg/server/generated"
)

func FetchScanResult(imageID string, s3 *s3.S3) (*[]generated.ScanResult, error) {
	res, err := s3.FetchFromS3(fmt.Sprintf("scans/%s/results.json", imageID))
	if err != nil {
		return nil, err
	}
	scanRes := []generated.ScanResult{}

	fmt.Println(string(res))
	err = json.Unmarshal(res, &scanRes)
	if err != nil {
		return nil, err
	}
	return &scanRes, nil
}
