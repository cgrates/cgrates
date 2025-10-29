/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/

package engine

import (
	"testing"

	"github.com/cgrates/cgrates/config"
)

func TestCDRENewCDRExporter(t *testing.T) {
	_, err := NewCDRExporter([]*CDR{}, &config.CdreCfg{}, "test", "test", "test", "test", false, 1, 'a', false, []string{}, &FilterS{})
	if err != nil {
		t.Error(err)
	}
}

func TestCDREmetaHandler(t *testing.T) {
	cdre := CDRExporter{}

	rcv, err := cdre.metaHandler("test", "test")

	if err != nil {
		if err.Error() != "Unsupported METATAG: test" {
			t.Fatal(err)
		}
	}

	if rcv != "" {
		t.Error(rcv)
	}
}

func TestCDREcomposeHeader(t *testing.T) {
	cdre := CDRExporter{
		exportTemplate: &config.CdreCfg{
			Fields: []*config.FCTemplate{
				{
					Type:  "*filler",
					Value: config.RSRParsers{{Rules: "test()"}},
				},
			},
		},
	}

	err := cdre.composeHeader()

	if err != nil {
		if err.Error() != "err" {
			t.Error(err)
		}
	}
}
