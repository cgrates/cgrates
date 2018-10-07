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
	agReq := newAgentRequest(nil, nil, nil, nil, "cgrates.org", "", filterS)
	// populate request, emulating the way will be done in HTTPAgent
	agReq.CGRRequest.Set([]string{utils.CGRID},
		utils.Sha1("dsafdsaf", time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC).String()), false)
	agReq.CGRRequest.Set([]string{utils.ToR}, utils.VOICE, false)
	agReq.CGRRequest.Set([]string{utils.Account}, "1001", false)
	agReq.CGRRequest.Set([]string{utils.Destination}, "1002", false)
	agReq.CGRRequest.Set([]string{utils.AnswerTime},
		time.Date(2013, 12, 30, 15, 0, 1, 0, time.UTC), false)
	agReq.CGRRequest.Set([]string{utils.RequestType}, utils.META_PREPAID, false)
	agReq.CGRRequest.Set([]string{utils.Usage}, time.Duration(3*time.Minute), false)

	cgrRply := map[string]interface{}{
		utils.CapAttributes: map[string]interface{}{
			"PaypalAccount": "cgrates@paypal.com",
		},
		utils.CapMaxUsage: time.Duration(120 * time.Second),
		utils.Error:       "",
	}
	agReq.CGRReply = config.NewNavigableMap(cgrRply)

	tplFlds := []*config.FCTemplate{
		&config.FCTemplate{Tag: "Tenant",
			FieldId: utils.Tenant, Type: utils.META_COMPOSED,
			Value: config.NewRSRParsersMustCompile("cgrates.org", true)},
		&config.FCTemplate{Tag: "Account",
			FieldId: utils.Account, Type: utils.META_COMPOSED,
			Value: config.NewRSRParsersMustCompile("~*cgreq.Account", true)},
		&config.FCTemplate{Tag: "Destination",
			FieldId: utils.Destination, Type: utils.META_COMPOSED,
			Value: config.NewRSRParsersMustCompile("~*cgreq.Destination", true)},

		&config.FCTemplate{Tag: "RequestedUsageVoice",
			FieldId: "RequestedUsage", Type: utils.META_COMPOSED,
			Filters: []string{"*string:*cgreq.ToR:*voice"},
			Value: config.NewRSRParsersMustCompile(
				"~*cgreq.Usage{*duration_seconds}", true)},
		&config.FCTemplate{Tag: "RequestedUsageData",
			FieldId: "RequestedUsage", Type: utils.META_COMPOSED,
			Filters: []string{"*string:*cgreq.ToR:*data"},
			Value: config.NewRSRParsersMustCompile(
				"~*cgreq.Usage{*duration_nanoseconds}", true)},
		&config.FCTemplate{Tag: "RequestedUsageSMS",
			FieldId: "RequestedUsage", Type: utils.META_COMPOSED,
			Filters: []string{"*string:*cgreq.ToR:*sms"},
			Value: config.NewRSRParsersMustCompile(
				"~*cgreq.Usage{*duration_nanoseconds}", true)},

		&config.FCTemplate{Tag: "AttrPaypalAccount",
			FieldId: "PaypalAccount", Type: utils.META_COMPOSED,
			Filters: []string{"*string:*cgrep.Error:"},
			Value: config.NewRSRParsersMustCompile(
				"~*cgrep.Attributes.PaypalAccount", true)},
		&config.FCTemplate{Tag: "MaxUsage",
			FieldId: "MaxUsage", Type: utils.META_COMPOSED,
			Filters: []string{"*string:*cgrep.Error:"},
			Value: config.NewRSRParsersMustCompile(
				"~*cgrep.MaxUsage{*duration_seconds}", true)},
		&config.FCTemplate{Tag: "Error",
			FieldId: "Error", Type: utils.META_COMPOSED,
			Filters: []string{"*rsr::~*cgrep.Error(!^$)"},
			Value: config.NewRSRParsersMustCompile(
				"~*cgrep.Error", true)},
	}
	eMp := config.NewNavigableMap(nil)
	eMp.Set([]string{utils.Tenant}, []*config.NMItem{
		&config.NMItem{Data: "cgrates.org", Path: []string{utils.Tenant},
			Config: tplFlds[0]}}, true)
	eMp.Set([]string{utils.Account}, []*config.NMItem{
		&config.NMItem{Data: "1001", Path: []string{utils.Account},
			Config: tplFlds[1]}}, true)
	eMp.Set([]string{utils.Destination}, []*config.NMItem{
		&config.NMItem{Data: "1002", Path: []string{utils.Destination},
			Config: tplFlds[2]}}, true)
	eMp.Set([]string{"RequestedUsage"}, []*config.NMItem{
		&config.NMItem{Data: "180", Path: []string{"RequestedUsage"},
			Config: tplFlds[3]}}, true)
	eMp.Set([]string{"PaypalAccount"}, []*config.NMItem{
		&config.NMItem{Data: "cgrates@paypal.com", Path: []string{"PaypalAccount"},
			Config: tplFlds[6]}}, true)
	eMp.Set([]string{"MaxUsage"}, []*config.NMItem{
		&config.NMItem{Data: "120", Path: []string{"MaxUsage"},
			Config: tplFlds[7]}}, true)
	if mpOut, err := agReq.AsNavigableMap(tplFlds); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eMp, mpOut) {
		t.Errorf("expecting: %+v, received: %+v", eMp, mpOut)
	}
}
