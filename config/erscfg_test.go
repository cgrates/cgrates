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
	"reflect"
	"testing"

	"github.com/cgrates/cgrates/utils"
)

func TestEventRedearClone(t *testing.T) {
	orig := &EventReaderCfg{
		ID:            utils.MetaDefault,
		Type:          "RandomType",
		FieldSep:      ",",
		SourceID:      "RandomSource",
		Filters:       []string{"Filter1", "Filter2"},
		Tenant:        NewRSRParsersMustCompile("cgrates.org", true, utils.INFIELD_SEP),
		Header_fields: []*FCTemplate{},
		Content_fields: []*FCTemplate{
			{
				Tag:       "TOR",
				FieldId:   "ToR",
				Type:      "*composed",
				Value:     NewRSRParsersMustCompile("~2", true, utils.INFIELD_SEP),
				Mandatory: true,
			},
			{
				Tag:       "RandomField",
				FieldId:   "RandomField",
				Type:      "*composed",
				Value:     NewRSRParsersMustCompile("Test", true, utils.INFIELD_SEP),
				Mandatory: true,
			},
		},
		Trailer_fields: []*FCTemplate{},
		Continue:       true,
	}
	cloned := orig.Clone()
	if !reflect.DeepEqual(cloned, orig) {
		t.Errorf("expected: %s \n,received: %s", utils.ToJSON(orig), utils.ToJSON(cloned))
	}
	initialOrig := &EventReaderCfg{
		ID:            utils.MetaDefault,
		Type:          "RandomType",
		FieldSep:      ",",
		SourceID:      "RandomSource",
		Filters:       []string{"Filter1", "Filter2"},
		Tenant:        NewRSRParsersMustCompile("cgrates.org", true, utils.INFIELD_SEP),
		Header_fields: []*FCTemplate{},
		Content_fields: []*FCTemplate{
			{
				Tag:       "TOR",
				FieldId:   "ToR",
				Type:      "*composed",
				Value:     NewRSRParsersMustCompile("~2", true, utils.INFIELD_SEP),
				Mandatory: true,
			},
			{
				Tag:       "RandomField",
				FieldId:   "RandomField",
				Type:      "*composed",
				Value:     NewRSRParsersMustCompile("Test", true, utils.INFIELD_SEP),
				Mandatory: true,
			},
		},
		Trailer_fields: []*FCTemplate{},
		Continue:       true,
	}
	orig.Filters = []string{"SingleFilter"}
	orig.Content_fields = []*FCTemplate{
		{
			Tag:       "TOR",
			FieldId:   "ToR",
			Type:      "*composed",
			Value:     NewRSRParsersMustCompile("~2", true, utils.INFIELD_SEP),
			Mandatory: true,
		},
	}
	if !reflect.DeepEqual(cloned, initialOrig) {
		t.Errorf("expected: %s \n,received: %s", utils.ToJSON(initialOrig), utils.ToJSON(cloned))
	}
}
