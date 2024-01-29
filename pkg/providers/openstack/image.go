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

package ostack

import (
	"fmt"
	"github.com/drewbernetes/baski/pkg/util/flags"
	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/gophercloud/gophercloud/openstack/imageservice/v2/images"
	"log"
	"strings"
	"time"
)

type ImageClient struct {
	client *gophercloud.ServiceClient
}

func NewImageClient(provider Provider) (*ImageClient, error) {
	p, err := provider.Client()
	if err != nil {
		return nil, err
	}
	client, err := openstack.NewImageServiceV2(p, gophercloud.EndpointOpts{})
	if err != nil {
		return nil, err
	}
	return &ImageClient{
		client: client,
	}, nil
}

// GenerateBuilderMetadata generates some glance metadata for the image.
func GenerateBuilderMetadata(o *flags.BuildOptions) map[string]string {
	gpu := "no_gpu"
	if o.AddNvidiaSupport {
		gpu = o.NvidiaVersion
	}
	return map[string]string{
		"os":          o.BuildOS,
		"k8s":         o.KubeVersion,
		"gpu":         gpu,
		"date":        time.Now().Format(time.RFC3339),
		"rootfs_uuid": o.RootfsUUID,
	}
}

// ModifyImageMetadata allows image metadata to be added, updated or removed.
func (c *ImageClient) ModifyImageMetadata(imgID string, key, value string, operation images.UpdateOp) (*images.Image, error) {
	c.client.Microversion = "2.2"

	updateOpts := images.UpdateOpts{
		images.UpdateImageProperty{
			Op:    operation,
			Name:  fmt.Sprintf("/%s", key),
			Value: value,
		},
	}

	img, err := images.Update(c.client, imgID, updateOpts).Extract()

	if err != nil {
		return nil, err
	}

	return img, nil
}

// RemoveImage will delete an image from Openstack.
func (c *ImageClient) RemoveImage(imgID string) error {
	log.Println("removing image")
	res := images.Delete(c.client, imgID)
	if res.Err != nil {
		return res.Err
	}

	return nil
}

// FetchAllImages Fetches all the images from Openstack so that they can parsed after.
// Because silly GopherCloud - or maybe OpenStack itself doesn't support wildcard search on names
// and the tag search is limited to an id+tag :facepalm:
// This probably can be improved though to prevent fetching billions of images.
func (c *ImageClient) FetchAllImages(wildcard string) ([]images.Image, error) {
	i, err := images.List(c.client, images.ListOpts{}).AllPages()
	if err != nil {
		return nil, err
	}

	imageInfo, err := images.ExtractImages(i)
	if err != nil {
		return nil, err
	}

	imgs := []images.Image{}
	for _, im := range imageInfo {
		if strings.Contains(im.Name, wildcard) {
			imgs = append(imgs, im)
		}
	}

	return imgs, nil
}

// FetchImage allows us to fetch a single image by the id.
func (c *ImageClient) FetchImage(imgID string) (*images.Image, error) {
	i, err := images.List(c.client, images.ListOpts{
		ID: imgID,
	}).AllPages()
	if err != nil {
		return nil, err
	}

	imageInfo, err := images.ExtractImages(i)
	if err != nil {
		return nil, err
	}

	for _, im := range imageInfo {
		if im.ID == imgID {
			return &im, nil
		}
	}

	return nil, nil
}
