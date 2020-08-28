/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>
*/

package v1

import (
	"reflect"
	"testing"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestComposeArgsReload(t *testing.T) {
	apv1 := &APIerSv1{DataManager: &engine.DataManager{}}
	expArgs := utils.AttrReloadCacheWithOpts{
		Opts:      make(map[string]interface{}),
		TenantArg: utils.TenantArg{Tenant: "cgrates.org"},
		ArgsCache: map[string][]string{
			utils.AttributeProfileIDs: {"cgrates.org:ATTR1"},
		},
	}

	if rply, err := apv1.composeArgsReload("cgrates.org", utils.CacheAttributeProfiles,
		"cgrates.org:ATTR1", nil, nil, make(map[string]interface{})); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(expArgs, rply) {
		t.Errorf("Expected %s ,received: %s", utils.ToJSON(expArgs), utils.ToJSON(rply))
	}

	expArgs.ArgsCache[utils.AttributeFilterIndexIDs] = []string{"cgrates.org:*cdrs:*none:*any:*any"}

	if rply, err := apv1.composeArgsReload("cgrates.org", utils.CacheAttributeProfiles,
		"cgrates.org:ATTR1", &[]string{}, []string{utils.MetaCDRs}, make(map[string]interface{})); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(expArgs, rply) {
		t.Errorf("Expected %s ,received: %s", utils.ToJSON(expArgs), utils.ToJSON(rply))
	}

	expArgs.ArgsCache[utils.AttributeFilterIndexIDs] = []string{
		"cgrates.org:*cdrs:*string:*req.Account:1001",
		"cgrates.org:*cdrs:*prefix:*req.Destination:1001",
	}

	if rply, err := apv1.composeArgsReload("cgrates.org", utils.CacheAttributeProfiles,
		"cgrates.org:ATTR1", &[]string{"*string:~*req.Account:1001;~req.Subject", "*prefix:1001:~*req.Destination"}, []string{utils.MetaCDRs}, make(map[string]interface{})); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(expArgs, rply) {
		t.Errorf("Expected %s ,received: %s", utils.ToJSON(expArgs), utils.ToJSON(rply))
	}
}
