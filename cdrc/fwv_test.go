/*
Real-time Charging System for Telecom & ISP environments
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

package cdrc

import (
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"testing"
)

func TestFwvValue(t *testing.T) {
	cdrLine := "CDR0000010  0 20120708181506000123451234         0040123123120                  004                                            000018009980010001ISDN  ABC   10Buiten uw regio                         EHV 00000009190000000009"
	if val := fwvValue(cdrLine, 30, 19, "right"); val != "0123451234" {
		t.Errorf("Received: <%s>", val)
	}
	if val := fwvValue(cdrLine, 14, 16, "right"); val != "2012070818150600" { // SetupTime
		t.Errorf("Received: <%s>", val)
	}
	if val := fwvValue(cdrLine, 127, 8, "right"); val != "00001800" { // Usage
		t.Errorf("Received: <%s>", val)
	}
	cdrLine = "HDR0001DDB     ABC                                     Some Connect A.B.                       DDB-Some-10022-20120711-309.CDR         00030920120711100255                                                "
	if val := fwvValue(cdrLine, 135, 6, "zeroleft"); val != "309" {
		t.Errorf("Received: <%s>", val)
	}

}

func TestFwvRecordPassesCfgFilter(t *testing.T) {
	//record, configKey string) bool {
	cgrConfig, _ := config.NewDefaultCGRConfig()
	cdrcConfig := cgrConfig.CdrcProfiles["/var/log/cgrates/cdrc/in"][0] // We don't really care that is for .csv since all we want to test are the filters
	cdrcConfig.CdrFilter = utils.ParseRSRFieldsMustCompile(`~52:s/^0(\d{9})/+49${1}/(^+49123123120)`, utils.INFIELD_SEP)
	fwvRp := &FwvRecordsProcessor{cdrcCfgs: cgrConfig.CdrcProfiles["/var/log/cgrates/cdrc/in"]}
	cdrLine := "CDR0000010  0 20120708181506000123451234         0040123123120                  004                                            000018009980010001ISDN  ABC   10Buiten uw regio                         EHV 00000009190000000009"
	if passesFilter := fwvRp.recordPassesCfgFilter(cdrLine, cdrcConfig); !passesFilter {
		t.Error("Not passes filter")
	}
}
