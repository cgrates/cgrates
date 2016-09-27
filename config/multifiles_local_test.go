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
package config

import (
	"flag"
	"github.com/cgrates/cgrates/utils"
	"testing"
)

var testLocal = flag.Bool("local", false, "Perform the tests only on local test environment, disabled by default.") // This flag will be passed here via "go test -local" args

var mfCgrCfg *CGRConfig

func TestMfInitConfig(t *testing.T) {
	if !*testLocal {
		return
	}
	var err error
	if mfCgrCfg, err = NewCGRConfigFromFolder("/usr/share/cgrates/conf/samples/multifiles"); err != nil {
		t.Fatal("Got config error: ", err.Error())
	}
}

func TestMfGeneralItems(t *testing.T) {
	if !*testLocal {
		return
	}
	if mfCgrCfg.DefaultReqType != utils.META_PSEUDOPREPAID { // Twice reconfigured
		t.Error("DefaultReqType: ", mfCgrCfg.DefaultReqType)
	}
	if mfCgrCfg.DefaultCategory != "call" { // Not configred, should be inherited from default
		t.Error("DefaultCategory: ", mfCgrCfg.DefaultCategory)
	}
}

func TestMfCdreDefaultInstance(t *testing.T) {
	if !*testLocal {
		return
	}
	for _, prflName := range []string{"*default", "export1"} {
		if _, hasIt := mfCgrCfg.CdreProfiles[prflName]; !hasIt {
			t.Error("Cdre does not contain profile ", prflName)
		}
	}
	prfl := "*default"
	if mfCgrCfg.CdreProfiles[prfl].CdrFormat != "csv" {
		t.Error("Default instance has cdrFormat: ", mfCgrCfg.CdreProfiles[prfl].CdrFormat)
	}
	if mfCgrCfg.CdreProfiles[prfl].DataUsageMultiplyFactor != 1024.0 {
		t.Error("Default instance has cdrFormat: ", mfCgrCfg.CdreProfiles[prfl].DataUsageMultiplyFactor)
	}
	if len(mfCgrCfg.CdreProfiles[prfl].HeaderFields) != 0 {
		t.Error("Default instance has number of header fields: ", len(mfCgrCfg.CdreProfiles[prfl].HeaderFields))
	}
	if len(mfCgrCfg.CdreProfiles[prfl].ContentFields) != 12 {
		t.Error("Default instance has number of content fields: ", len(mfCgrCfg.CdreProfiles[prfl].ContentFields))
	}
	if mfCgrCfg.CdreProfiles[prfl].ContentFields[2].Tag != "Direction" {
		t.Error("Unexpected headerField value: ", mfCgrCfg.CdreProfiles[prfl].ContentFields[2].Tag)
	}
}

func TestMfCdreExport1Instance(t *testing.T) {
	if !*testLocal {
		return
	}
	prfl := "export1"
	if mfCgrCfg.CdreProfiles[prfl].CdrFormat != "csv" {
		t.Error("Export1 instance has cdrFormat: ", mfCgrCfg.CdreProfiles[prfl].CdrFormat)
	}
	if mfCgrCfg.CdreProfiles[prfl].DataUsageMultiplyFactor != 1.0 {
		t.Error("Export1 instance has DataUsageMultiplyFormat: ", mfCgrCfg.CdreProfiles[prfl].DataUsageMultiplyFactor)
	}
	if mfCgrCfg.CdreProfiles[prfl].CostRoundingDecimals != 3.0 {
		t.Error("Export1 instance has cdrFormat: ", mfCgrCfg.CdreProfiles[prfl].CostRoundingDecimals)
	}
	if len(mfCgrCfg.CdreProfiles[prfl].HeaderFields) != 2 {
		t.Error("Export1 instance has number of header fields: ", len(mfCgrCfg.CdreProfiles[prfl].HeaderFields))
	}
	if mfCgrCfg.CdreProfiles[prfl].HeaderFields[1].Tag != "RunId" {
		t.Error("Unexpected headerField value: ", mfCgrCfg.CdreProfiles[prfl].HeaderFields[1].Tag)
	}
	if len(mfCgrCfg.CdreProfiles[prfl].ContentFields) != 9 {
		t.Error("Export1 instance has number of content fields: ", len(mfCgrCfg.CdreProfiles[prfl].ContentFields))
	}
	if mfCgrCfg.CdreProfiles[prfl].ContentFields[2].Tag != "Account" {
		t.Error("Unexpected headerField value: ", mfCgrCfg.CdreProfiles[prfl].ContentFields[2].Tag)
	}
}
