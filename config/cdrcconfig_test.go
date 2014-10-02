/*
Real-time Charging System for Telecom & ISP environments
Copyright (C) 2012-2014 ITsysCOM GmbH

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
	"time"

	"github.com/cgrates/cgrates/utils"
)

func TestNewDefaultCdrcConfig(t *testing.T) {
	eDfCdrcConfig := &CdrcConfig{
		Id:             utils.META_DEFAULT,
		Enabled:        false,
		CdrsAddress:    "",
		CdrType:        utils.CSV,
		FieldSeparator: utils.FIELDS_SEP,
		RunDelay:       time.Duration(0),
		CdrInDir:       "/var/log/cgrates/cdrc/in",
		CdrOutDir:      "/var/log/cgrates/cdrc/out",
		CdrSourceId:    utils.CSV,
		CdrFields: []*CfgCdrField{
			&CfgCdrField{
				Tag:        utils.TOR,
				Type:       utils.CDRFIELD,
				CdrFieldId: utils.TOR,
				Value: []*utils.RSRField{
					&utils.RSRField{Id: "2"}},
				Mandatory: true,
			},
			&CfgCdrField{
				Tag:        utils.ACCID,
				Type:       utils.CDRFIELD,
				CdrFieldId: utils.ACCID,
				Value: []*utils.RSRField{
					&utils.RSRField{Id: "3"}},
				Mandatory: true,
			},
			&CfgCdrField{
				Tag:        utils.REQTYPE,
				Type:       utils.CDRFIELD,
				CdrFieldId: utils.REQTYPE,
				Value: []*utils.RSRField{
					&utils.RSRField{Id: "4"}},
				Mandatory: true,
			},
			&CfgCdrField{
				Tag:        utils.DIRECTION,
				Type:       utils.CDRFIELD,
				CdrFieldId: utils.DIRECTION,
				Value: []*utils.RSRField{
					&utils.RSRField{Id: "5"}},
				Mandatory: true,
			},
			&CfgCdrField{
				Tag:        utils.TENANT,
				Type:       utils.CDRFIELD,
				CdrFieldId: utils.TENANT,
				Value: []*utils.RSRField{
					&utils.RSRField{Id: "6"}},
				Mandatory: true},
			&CfgCdrField{
				Tag:        utils.CATEGORY,
				Type:       utils.CDRFIELD,
				CdrFieldId: utils.CATEGORY,
				Value: []*utils.RSRField{
					&utils.RSRField{Id: "7"}},
				Mandatory: true,
			},
			&CfgCdrField{
				Tag:        utils.ACCOUNT,
				Type:       utils.CDRFIELD,
				CdrFieldId: utils.ACCOUNT,
				Value: []*utils.RSRField{
					&utils.RSRField{Id: "8"}},
				Mandatory: true,
			},
			&CfgCdrField{
				Tag:        utils.SUBJECT,
				Type:       utils.CDRFIELD,
				CdrFieldId: utils.SUBJECT,
				Value: []*utils.RSRField{
					&utils.RSRField{Id: "9"}},
				Mandatory: true,
			},
			&CfgCdrField{
				Tag:        utils.DESTINATION,
				Type:       utils.CDRFIELD,
				CdrFieldId: utils.DESTINATION,
				Value: []*utils.RSRField{
					&utils.RSRField{Id: "10"}},
				Mandatory: true,
			},
			&CfgCdrField{
				Tag:        utils.SETUP_TIME,
				Type:       utils.CDRFIELD,
				CdrFieldId: utils.SETUP_TIME,
				Value: []*utils.RSRField{
					&utils.RSRField{Id: "11"}},
				Layout:    "2006-01-02T15:04:05Z07:00",
				Mandatory: true,
			},
			&CfgCdrField{
				Tag:        utils.ANSWER_TIME,
				Type:       utils.CDRFIELD,
				CdrFieldId: utils.ANSWER_TIME,
				Value: []*utils.RSRField{
					&utils.RSRField{Id: "12"}},
				Layout:    "2006-01-02T15:04:05Z07:00",
				Mandatory: true,
			},
			&CfgCdrField{
				Tag:        utils.USAGE,
				Type:       utils.CDRFIELD,
				CdrFieldId: utils.USAGE,
				Value: []*utils.RSRField{
					&utils.RSRField{Id: "13"}},
				Mandatory: true,
			},
		},
	}
	if dfCdrcCfg := NewDefaultCdrcConfig(); !reflect.DeepEqual(eDfCdrcConfig, dfCdrcCfg) {
		t.Errorf("Expected: %+v, received: %+v", eDfCdrcConfig, dfCdrcCfg)
	}
}
