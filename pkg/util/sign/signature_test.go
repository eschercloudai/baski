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

package sign

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"strings"
	"testing"
)

var (
	prik, pubk []byte
	id         = "123456abc"
	digest     string
)

func TestEncode(t *testing.T) {
	pk, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	if err != nil {
		t.Error(err)
		return
	}

	prik, pubk = EncodeKeys(pk)
	if ok := strings.Contains(string(prik), "BEGIN PRIVATE KEY"); !ok {
		t.Errorf("Expected private key to contain %s but the test returned %t\n", "BEGIN PUBLIC KEY", ok)
	}
	if ok := strings.Contains(string(pubk), "BEGIN PUBLIC KEY"); !ok {
		t.Errorf("Expected public key to contain %s but the test returned %t\n", "BEGIN PUBLIC KEY", ok)
	}
}

func TestSign(t *testing.T) {
	var err error
	digest, err = Sign(id, prik)
	if err != nil {
		t.Error(err)
	}
}

func TestValidate(t *testing.T) {
	valid, err := Validate(id, pubk, digest)
	if err != nil {
		t.Error(err)
		return
	}
	if !valid {
		t.Errorf("Expected %t but the test returned %t\n", true, valid)
	}
}
