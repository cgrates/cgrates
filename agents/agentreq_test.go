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

package agents

import (
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestAgReqAsNavigableMap(t *testing.T) {
	data, _ := engine.NewMapStorage()
	dm := engine.NewDataManager(data)
	cfg, _ := config.NewDefaultCGRConfig()
	filterS := engine.NewFilterS(cfg, nil, dm)
	agReq := newAgentRequest(nil, nil,
		"cgrates.org", filterS)
	// populate request, emulating the way will be done in HTTPAgent
	agReq.CGRRequest.Set(&engine.NMItem{Path: []string{utils.CGRID},
		Data: utils.Sha1("dsafdsaf", time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC).String())}, false)
	agReq.CGRRequest.Set(&engine.NMItem{Path: []string{utils.ToR}, Data: utils.VOICE}, false)
	agReq.CGRRequest.Set(&engine.NMItem{Path: []string{utils.Account}, Data: "1001"}, false)
	agReq.CGRRequest.Set(&engine.NMItem{Path: []string{utils.Destination}, Data: "1002"}, false)
	agReq.CGRRequest.Set(&engine.NMItem{Path: []string{utils.AnswerTime},
		Data: time.Date(2013, 12, 30, 15, 0, 1, 0, time.UTC)}, false)
	agReq.CGRRequest.Set(&engine.NMItem{Path: []string{utils.RequestType}, Data: utils.META_PREPAID}, false)
	agReq.CGRRequest.Set(&engine.NMItem{Path: []string{utils.Usage}, Data: time.Duration(3 * time.Minute)}, false)

	cgrRply := map[string]interface{}{
		utils.CapAttributes: map[string]interface{}{
			"PaypalAccount": "cgrates@paypal.com",
		},
		utils.CapMaxUsage: time.Duration(120 * time.Second),
		utils.Error:       "",
	}
	agReq.CGRReply = engine.NewNavigableMap(cgrRply)

	tplFlds := []*config.CfgCdrField{
		&config.CfgCdrField{Tag: "Tenant",
			FieldId: utils.Tenant, Type: utils.META_COMPOSED,
			Value: utils.ParseRSRFieldsMustCompile("^cgrates.org", utils.INFIELD_SEP)},
		&config.CfgCdrField{Tag: "Account",
			FieldId: utils.Account, Type: utils.META_COMPOSED,
			Value: utils.ParseRSRFieldsMustCompile("*cgrRequest>Account", utils.INFIELD_SEP)},
		&config.CfgCdrField{Tag: "Destination",
			FieldId: utils.Destination, Type: utils.META_COMPOSED,
			Value: utils.ParseRSRFieldsMustCompile("*cgrRequest>Destination", utils.INFIELD_SEP)},

		&config.CfgCdrField{Tag: "RequestedUsageVoice",
			FieldId: "RequestedUsage", Type: utils.META_COMPOSED,
			Filters: []string{"*string:*cgrRequest>ToR:*voice"},
			Value: utils.ParseRSRFieldsMustCompile(
				"*cgrRequest>Usage{*duration_seconds}", utils.INFIELD_SEP)},
		&config.CfgCdrField{Tag: "RequestedUsageData",
			FieldId: "RequestedUsage", Type: utils.META_COMPOSED,
			Filters: []string{"*string:*cgrRequest>ToR:*data"},
			Value: utils.ParseRSRFieldsMustCompile(
				"*cgrRequest>Usage{*duration_nanoseconds}", utils.INFIELD_SEP)},
		&config.CfgCdrField{Tag: "RequestedUsageSMS",
			FieldId: "RequestedUsage", Type: utils.META_COMPOSED,
			Filters: []string{"*string:*cgrRequest>ToR:*sms"},
			Value: utils.ParseRSRFieldsMustCompile(
				"*cgrRequest>Usage{*duration_nanoseconds}", utils.INFIELD_SEP)},

		&config.CfgCdrField{Tag: "AttrPaypalAccount",
			FieldId: "PaypalAccount", Type: utils.META_COMPOSED,
			Filters: []string{"*string:*cgrReply>Error:"},
			Value: utils.ParseRSRFieldsMustCompile(
				"*cgrReply>Attributes>PaypalAccount", utils.INFIELD_SEP)},
		&config.CfgCdrField{Tag: "MaxUsage",
			FieldId: "MaxUsage", Type: utils.META_COMPOSED,
			Filters: []string{"*string:*cgrReply>Error:"},
			Value: utils.ParseRSRFieldsMustCompile(
				"*cgrReply>MaxUsage{*duration_seconds}", utils.INFIELD_SEP)},
		&config.CfgCdrField{Tag: "Error",
			FieldId: "Error", Type: utils.META_COMPOSED,
			Filters: []string{"*rsr::*cgrReply>Error(!^$)"},
			Value: utils.ParseRSRFieldsMustCompile(
				"*cgrReply>Error", utils.INFIELD_SEP)},
	}
	eMp := engine.NewNavigableMap(nil)
	eMp.Set(&engine.NMItem{Path: []string{utils.Tenant}, Data: "cgrates.org"}, true)
	eMp.Set(&engine.NMItem{Path: []string{utils.Account}, Data: "1001"}, true)
	eMp.Set(&engine.NMItem{Path: []string{utils.Destination}, Data: "1002"}, true)
	eMp.Set(&engine.NMItem{Path: []string{"RequestedUsage"}, Data: "180"}, true)
	eMp.Set(&engine.NMItem{Path: []string{"PaypalAccount"}, Data: "cgrates@paypal.com"}, true)
	eMp.Set(&engine.NMItem{Path: []string{"MaxUsage"}, Data: "120"}, true)
	if mpOut, err := agReq.AsNavigableMap(tplFlds); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eMp, mpOut) {
		t.Errorf("expecting: %+v, received: %+v",
			eMp.AsMapStringInterface(), mpOut.AsMapStringInterface())
	}
}
