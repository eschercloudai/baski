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

package sign

import (
	"github.com/drewbernetes/baski/pkg/mock"
	"github.com/drewbernetes/baski/pkg/util"
	"go.uber.org/mock/gomock"
	"testing"
)

func TestFetch(t *testing.T) {
	c := gomock.NewController(t)
	defer c.Finish()
	m := mock.NewMockVaultInterface(c)

	m.EXPECT().Fetch(gomock.Eq("kv/eso"), gomock.Eq("some/path"), gomock.Eq("key")).Return([]byte("some data"), nil)
	if _, err := fetch(m); err != nil {
		t.Error(err)
	}
}

func fetch(v util.VaultInterface) ([]byte, error) {
	return v.Fetch("kv/eso", "some/path", "key")
}
