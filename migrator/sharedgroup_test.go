// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package migrator

import (
	"reflect"
	"testing"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestV1SharedGroupAsSharedGroup(t *testing.T) {
	v1sg := &v1SharedGroup{
		Id: "Test",
		AccountParameters: map[string]*engine.SharingParameters{
			"test": {Strategy: "*highest"},
		},
		MemberIds: []string{"1", "2", "3"},
	}
	sg := &engine.SharedGroup{
		Id: "Test",
		AccountParameters: map[string]*engine.SharingParameters{
			"test": {Strategy: "*highest"},
		},
		MemberIds: utils.NewStringMap("1", "2", "3"),
	}
	newsg := v1sg.AsSharedGroup()
	if !reflect.DeepEqual(sg, newsg) {
		t.Errorf("Expecting: %+v, received: %+v", sg, newsg)
	}

}
