/*
Copyright 2023 EscherCloud.

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
	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/gophercloud/gophercloud/openstack/imageservice/v2/images"
	"log"
	"strings"
)

// createImageClient will generate the image client required for updating image metadata.
func createImageClient(client *Client) *gophercloud.ServiceClient {
	c, err := openstack.NewImageServiceV2(client.Provider, *client.EndpointOptions)
	if err != nil {
		panic(err)
	}

	return c
}

// ModifyImageMetadata allows image metadata to be added, updated or removed.
func (c *Client) ModifyImageMetadata(imgID string, key, value string, operation images.UpdateOp) (*images.Image, error) {
	client := createImageClient(c)
	client.Microversion = "2.2"

	updateOpts := images.UpdateOpts{
		images.UpdateImageProperty{
			Op:    operation,
			Name:  fmt.Sprintf("/%s", key),
			Value: value,
		},
	}

	img, err := images.Update(client, imgID, updateOpts).Extract()

	if err != nil {
		return nil, err
	}

	return img, nil
}

// FetchAllImages Fetches all the images from Openstack so that they can parsed after.
// Because silly GopherCloud - or maybe OpenStack itself doesn't support wildcard search on names
// and the tag search is limited to an id+tag :facepalm:
// This probably can be improved though to prevent fetching billions of images.
func (c *Client) FetchAllImages(wildcard string) ([]images.Image, error) {

	client, err := openstack.NewComputeV2(c.Provider, *c.EndpointOptions)
	if err != nil {
		return nil, err
	}

	i, err := images.List(client, images.ListOpts{}).AllPages()
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
func (c *Client) FetchImage(imgID string) (*images.Image, error) {

	client, err := openstack.NewComputeV2(c.Provider, *c.EndpointOptions)
	if err != nil {
		return nil, err
	}

	i, err := images.List(client, images.ListOpts{
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

// RemoveImage will delete an image from Openstack.
func (c *Client) RemoveImage(imgID string) {
	log.Println("removing image")
	client := createImageClient(c)
	res := images.Delete(client, imgID)
	if res.Err != nil {
		log.Println(res.Err)
	}
}
