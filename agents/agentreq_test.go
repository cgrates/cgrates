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
	"bufio"
	"bytes"
	"net/http"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/radigo"
	"github.com/fiorix/go-diameter/diam"
	"github.com/fiorix/go-diameter/diam/avp"
	"github.com/fiorix/go-diameter/diam/datatype"
)

func TestAgReqAsNavigableMap(t *testing.T) {
	data, _ := engine.NewMapStorage()
	dm := engine.NewDataManager(data)
	cfg, _ := config.NewDefaultCGRConfig()
	filterS := engine.NewFilterS(cfg, nil, dm)
	agReq := newAgentRequest(nil, nil, nil, nil, "cgrates.org", "", filterS)
	// populate request, emulating the way will be done in HTTPAgent
	agReq.CGRRequest.Set([]string{utils.CGRID},
		utils.Sha1("dsafdsaf", time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC).String()),
		false, false)
	agReq.CGRRequest.Set([]string{utils.ToR}, utils.VOICE, false, false)
	agReq.CGRRequest.Set([]string{utils.Account}, "1001", false, false)
	agReq.CGRRequest.Set([]string{utils.Destination}, "1002", false, false)
	agReq.CGRRequest.Set([]string{utils.AnswerTime},
		time.Date(2013, 12, 30, 15, 0, 1, 0, time.UTC), false, false)
	agReq.CGRRequest.Set([]string{utils.RequestType}, utils.META_PREPAID, false, false)
	agReq.CGRRequest.Set([]string{utils.Usage}, time.Duration(3*time.Minute), false, false)

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
			Value: config.NewRSRParsersMustCompile("cgrates.org", true, utils.INFIELD_SEP)},
		&config.FCTemplate{Tag: "Account",
			FieldId: utils.Account, Type: utils.META_COMPOSED,
			Value: config.NewRSRParsersMustCompile("~*cgreq.Account", true, utils.INFIELD_SEP)},
		&config.FCTemplate{Tag: "Destination",
			FieldId: utils.Destination, Type: utils.META_COMPOSED,
			Value: config.NewRSRParsersMustCompile("~*cgreq.Destination", true, utils.INFIELD_SEP)},

		&config.FCTemplate{Tag: "RequestedUsageVoice",
			FieldId: "RequestedUsage", Type: utils.META_COMPOSED,
			Filters: []string{"*string:*cgreq.ToR:*voice"},
			Value: config.NewRSRParsersMustCompile(
				"~*cgreq.Usage{*duration_seconds}", true, utils.INFIELD_SEP)},
		&config.FCTemplate{Tag: "RequestedUsageData",
			FieldId: "RequestedUsage", Type: utils.META_COMPOSED,
			Filters: []string{"*string:*cgreq.ToR:*data"},
			Value: config.NewRSRParsersMustCompile(
				"~*cgreq.Usage{*duration_nanoseconds}", true, utils.INFIELD_SEP)},
		&config.FCTemplate{Tag: "RequestedUsageSMS",
			FieldId: "RequestedUsage", Type: utils.META_COMPOSED,
			Filters: []string{"*string:*cgreq.ToR:*sms"},
			Value: config.NewRSRParsersMustCompile(
				"~*cgreq.Usage{*duration_nanoseconds}", true, utils.INFIELD_SEP)},

		&config.FCTemplate{Tag: "AttrPaypalAccount",
			FieldId: "PaypalAccount", Type: utils.META_COMPOSED,
			Filters: []string{"*string:*cgrep.Error:"},
			Value: config.NewRSRParsersMustCompile(
				"~*cgrep.Attributes.PaypalAccount", true, utils.INFIELD_SEP)},
		&config.FCTemplate{Tag: "MaxUsage",
			FieldId: "MaxUsage", Type: utils.META_COMPOSED,
			Filters: []string{"*string:*cgrep.Error:"},
			Value: config.NewRSRParsersMustCompile(
				"~*cgrep.MaxUsage{*duration_seconds}", true, utils.INFIELD_SEP)},
		&config.FCTemplate{Tag: "Error",
			FieldId: "Error", Type: utils.META_COMPOSED,
			Filters: []string{"*rsr::~*cgrep.Error(!^$)"},
			Value: config.NewRSRParsersMustCompile(
				"~*cgrep.Error", true, utils.INFIELD_SEP)},
	}
	eMp := config.NewNavigableMap(nil)
	eMp.Set([]string{utils.Tenant}, []*config.NMItem{
		&config.NMItem{Data: "cgrates.org", Path: []string{utils.Tenant},
			Config: tplFlds[0]}}, false, true)
	eMp.Set([]string{utils.Account}, []*config.NMItem{
		&config.NMItem{Data: "1001", Path: []string{utils.Account},
			Config: tplFlds[1]}}, false, true)
	eMp.Set([]string{utils.Destination}, []*config.NMItem{
		&config.NMItem{Data: "1002", Path: []string{utils.Destination},
			Config: tplFlds[2]}}, false, true)
	eMp.Set([]string{"RequestedUsage"}, []*config.NMItem{
		&config.NMItem{Data: "180", Path: []string{"RequestedUsage"},
			Config: tplFlds[3]}}, false, true)
	eMp.Set([]string{"PaypalAccount"}, []*config.NMItem{
		&config.NMItem{Data: "cgrates@paypal.com", Path: []string{"PaypalAccount"},
			Config: tplFlds[6]}}, false, true)
	eMp.Set([]string{"MaxUsage"}, []*config.NMItem{
		&config.NMItem{Data: "120", Path: []string{"MaxUsage"},
			Config: tplFlds[7]}}, false, true)
	if mpOut, err := agReq.AsNavigableMap(tplFlds); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eMp, mpOut) {
		t.Errorf("expecting: %+v, received: %+v", eMp, mpOut)
	}
}

func TestAgReqMaxCost(t *testing.T) {
	data, _ := engine.NewMapStorage()
	dm := engine.NewDataManager(data)
	cfg, _ := config.NewDefaultCGRConfig()
	filterS := engine.NewFilterS(cfg, nil, dm)
	agReq := newAgentRequest(nil, nil, nil, nil, "cgrates.org", "", filterS)
	// populate request, emulating the way will be done in HTTPAgent
	agReq.CGRRequest.Set([]string{utils.CapMaxUsage}, "120s", false, false)

	cgrRply := map[string]interface{}{
		utils.CapMaxUsage: time.Duration(120 * time.Second),
	}
	agReq.CGRReply = config.NewNavigableMap(cgrRply)

	tplFlds := []*config.FCTemplate{
		&config.FCTemplate{Tag: "MaxUsage",
			FieldId: "MaxUsage", Type: utils.META_COMPOSED,
			Filters: []string{"*rsr::~*cgrep.MaxUsage(>0s)"},
			Value: config.NewRSRParsersMustCompile(
				"~*cgrep.MaxUsage{*duration_seconds}", true, utils.INFIELD_SEP)},
	}
	eMp := config.NewNavigableMap(nil)

	eMp.Set([]string{"MaxUsage"}, []*config.NMItem{
		&config.NMItem{Data: "120", Path: []string{"MaxUsage"},
			Config: tplFlds[0]}}, false, true)
	if mpOut, err := agReq.AsNavigableMap(tplFlds); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eMp, mpOut) {
		t.Errorf("expecting: %+v, received: %+v", eMp, mpOut)
	}
}

func TestAgReqParseFieldDiameter(t *testing.T) {
	//creater diameter message
	m := diam.NewRequest(diam.CreditControl, 4, nil)
	m.NewAVP("Session-Id", avp.Mbit, 0, datatype.UTF8String("simuhuawei;1449573472;00002"))
	m.NewAVP("Subscription-Id", avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(450, avp.Mbit, 0, datatype.Enumerated(2)),              // Subscription-Id-Type
			diam.NewAVP(444, avp.Mbit, 0, datatype.UTF8String("208708000004")), // Subscription-Id-Data
			diam.NewAVP(avp.ValueDigits, avp.Mbit, 0, datatype.Integer64(20000)),
		}})
	//create diameterDataProvider
	dP := newDADataProvider(nil, m)
	data, _ := engine.NewMapStorage()
	dm := engine.NewDataManager(data)
	cfg, _ := config.NewDefaultCGRConfig()
	filterS := engine.NewFilterS(cfg, nil, dm)
	//pass the data provider to agent request
	agReq := newAgentRequest(dP, nil, nil, nil, "cgrates.org", "", filterS)

	tplFlds := []*config.FCTemplate{
		&config.FCTemplate{Tag: "MandatoryFalse",
			FieldId: "MandatoryFalse", Type: utils.META_COMPOSED,
			Value:     config.NewRSRParsersMustCompile("~*req.MandatoryFalse", true, utils.INFIELD_SEP),
			Mandatory: false},
		&config.FCTemplate{Tag: "MandatoryTrue",
			FieldId: "MandatoryTrue", Type: utils.META_COMPOSED,
			Value:     config.NewRSRParsersMustCompile("~*req.MandatoryTrue", true, utils.INFIELD_SEP),
			Mandatory: true},
		&config.FCTemplate{Tag: "Session-Id", Filters: []string{},
			FieldId: "Session-Id", Type: utils.META_COMPOSED,
			Value:     config.NewRSRParsersMustCompile("~*req.Session-Id", true, utils.INFIELD_SEP),
			Mandatory: true},
	}
	expected := ""
	if out, err := agReq.ParseField(tplFlds[0]); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(out, expected) {
		t.Errorf("expecting: <%+v>, received: <%+v>", expected, out)
	}
	if _, err := agReq.ParseField(tplFlds[1]); err == nil ||
		err.Error() != "Empty source value for fieldID: <MandatoryTrue>" {
		t.Error(err)
	}
	expected = "simuhuawei;1449573472;00002"
	if out, err := agReq.ParseField(tplFlds[2]); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(out, expected) {
		t.Errorf("expecting: <%+v>, received: <%+v>", expected, out)
	}
}

func TestAgReqParseFieldRadius(t *testing.T) {
	//creater radius message
	pkt := radigo.NewPacket(radigo.AccountingRequest, 1, dictRad, coder, "CGRateS.org")
	if err := pkt.AddAVPWithName("User-Name", "flopsy", ""); err != nil {
		t.Error(err)
	}
	if err := pkt.AddAVPWithName("Cisco-NAS-Port", "CGR1", "Cisco"); err != nil {
		t.Error(err)
	}
	//create radiusDataProvider
	dP := newRADataProvider(pkt)
	data, _ := engine.NewMapStorage()
	dm := engine.NewDataManager(data)
	cfg, _ := config.NewDefaultCGRConfig()
	filterS := engine.NewFilterS(cfg, nil, dm)
	//pass the data provider to agent request
	agReq := newAgentRequest(dP, nil, nil, nil, "cgrates.org", "", filterS)
	tplFlds := []*config.FCTemplate{
		&config.FCTemplate{Tag: "MandatoryFalse",
			FieldId: "MandatoryFalse", Type: utils.META_COMPOSED,
			Value:     config.NewRSRParsersMustCompile("~*req.MandatoryFalse", true, utils.INFIELD_SEP),
			Mandatory: false},
		&config.FCTemplate{Tag: "MandatoryTrue",
			FieldId: "MandatoryTrue", Type: utils.META_COMPOSED,
			Value:     config.NewRSRParsersMustCompile("~*req.MandatoryTrue", true, utils.INFIELD_SEP),
			Mandatory: true},
	}
	expected := ""
	if out, err := agReq.ParseField(tplFlds[0]); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(out, expected) {
		t.Errorf("expecting: <%+v>, received: <%+v>", expected, out)
	}
	if _, err := agReq.ParseField(tplFlds[1]); err == nil ||
		err.Error() != "Empty source value for fieldID: <MandatoryTrue>" {
		t.Error(err)
	}
}

func TestAgReqParseFieldHttpUrl(t *testing.T) {
	//creater radius message
	br := bufio.NewReader(strings.NewReader(`GET /cdr?request_type=MOSMS_CDR&timestamp=2008-08-15%2017:49:21&message_date=2008-08-15%2017:49:21&transactionid=100744&CDR_ID=123456&carrierid=1&mcc=222&mnc=10&imsi=235180000000000&msisdn=%2B4977000000000&destination=%2B497700000001&message_status=0&IOT=0&service_id=1 HTTP/1.1
Host: api.cgrates.org

`))
	req, err := http.ReadRequest(br)
	if err != nil {
		t.Error(err)
	}
	//create radiusDataProvider
	dP, _ := newHTTPUrlDP(req)
	data, _ := engine.NewMapStorage()
	dm := engine.NewDataManager(data)
	cfg, _ := config.NewDefaultCGRConfig()
	filterS := engine.NewFilterS(cfg, nil, dm)
	//pass the data provider to agent request
	agReq := newAgentRequest(dP, nil, nil, nil, "cgrates.org", "", filterS)
	tplFlds := []*config.FCTemplate{
		&config.FCTemplate{Tag: "MandatoryFalse",
			FieldId: "MandatoryFalse", Type: utils.META_COMPOSED,
			Value:     config.NewRSRParsersMustCompile("~*req.MandatoryFalse", true, utils.INFIELD_SEP),
			Mandatory: false},
		&config.FCTemplate{Tag: "MandatoryTrue",
			FieldId: "MandatoryTrue", Type: utils.META_COMPOSED,
			Value:     config.NewRSRParsersMustCompile("~*req.MandatoryTrue", true, utils.INFIELD_SEP),
			Mandatory: true},
	}
	expected := ""
	if out, err := agReq.ParseField(tplFlds[0]); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(out, expected) {
		t.Errorf("expecting: <%+v>, received: <%+v>", expected, out)
	}

	if _, err := agReq.ParseField(tplFlds[1]); err == nil ||
		err.Error() != "Empty source value for fieldID: <MandatoryTrue>" {
		t.Error(err)
	}
}

func TestAgReqParseFieldHttpXml(t *testing.T) {
	//creater radius message
	body := `<complete-success-notification callid="109870">
	<createtime>2005-08-26T14:16:42</createtime>
	<connecttime>2005-08-26T14:16:56</connecttime>
	<endtime>2005-08-26T14:17:34</endtime>
	<reference>My Call Reference</reference>
	<userid>386</userid>
	<username>sampleusername</username>
	<customerid>1</customerid>
	<companyname>Conecto LLC</companyname>
	<totalcost amount="0.21" currency="USD">US$0.21</totalcost>
	<hasrecording>yes</hasrecording>
	<hasvoicemail>no</hasvoicemail>
	<agenttotalcost amount="0.13" currency="USD">US$0.13</agenttotalcost>
	<agentid>44</agentid>
	<callleg calllegid="222146">
		<number>+441624828505</number>
		<description>Isle of Man</description>
		<seconds>38</seconds>
		<perminuterate amount="0.0200" currency="USD">US$0.0200</perminuterate>
		<cost amount="0.0140" currency="USD">US$0.0140</cost>
		<agentperminuterate amount="0.0130" currency="USD">US$0.0130</agentperminuterate>
		<agentcost amount="0.0082" currency="USD">US$0.0082</agentcost>
	</callleg>
	<callleg calllegid="222147">
		<number>+44 7624 494075</number>
		<description>Isle of Man</description>
		<seconds>37</seconds>
		<perminuterate amount="0.2700" currency="USD">US$0.2700</perminuterate>
		<cost amount="0.1890" currency="USD">US$0.1890</cost>
		<agentperminuterate amount="0.1880" currency="USD">US$0.1880</agentperminuterate>
		<agentcost amount="0.1159" currency="USD">US$0.1159</agentcost>
	</callleg>
</complete-success-notification>
`
	req, err := http.NewRequest("POST", "http://localhost:8080/", bytes.NewBuffer([]byte(body)))
	if err != nil {
		t.Error(err)
	}
	//create radiusDataProvider
	dP, _ := newHTTPXmlDP(req)
	data, _ := engine.NewMapStorage()
	dm := engine.NewDataManager(data)
	cfg, _ := config.NewDefaultCGRConfig()
	filterS := engine.NewFilterS(cfg, nil, dm)
	//pass the data provider to agent request
	agReq := newAgentRequest(dP, nil, nil, nil, "cgrates.org", "", filterS)
	tplFlds := []*config.FCTemplate{
		&config.FCTemplate{Tag: "MandatoryFalse",
			FieldId: "MandatoryFalse", Type: utils.META_COMPOSED,
			Value:     config.NewRSRParsersMustCompile("~*req.MandatoryFalse", true, utils.INFIELD_SEP),
			Mandatory: false},
		&config.FCTemplate{Tag: "MandatoryTrue",
			FieldId: "MandatoryTrue", Type: utils.META_COMPOSED,
			Value:     config.NewRSRParsersMustCompile("~*req.MandatoryTrue", true, utils.INFIELD_SEP),
			Mandatory: true},
	}
	expected := ""
	if out, err := agReq.ParseField(tplFlds[0]); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(out, expected) {
		t.Errorf("expecting: <%+v>, received: <%+v>", expected, out)
	}
	if _, err := agReq.ParseField(tplFlds[1]); err == nil ||
		err.Error() != "Empty source value for fieldID: <MandatoryTrue>" {
		t.Error(err)
	}
}

func TestAgReqEmptyFilter(t *testing.T) {
	data, _ := engine.NewMapStorage()
	dm := engine.NewDataManager(data)
	cfg, _ := config.NewDefaultCGRConfig()
	filterS := engine.NewFilterS(cfg, nil, dm)
	agReq := newAgentRequest(nil, nil, nil, nil, "cgrates.org", "", filterS)
	// populate request, emulating the way will be done in HTTPAgent
	agReq.CGRRequest.Set([]string{utils.CGRID},
		utils.Sha1("dsafdsaf", time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC).String()),
		false, false)
	agReq.CGRRequest.Set([]string{utils.Account}, "1001", false, false)
	agReq.CGRRequest.Set([]string{utils.Destination}, "1002", false, false)

	tplFlds := []*config.FCTemplate{
		&config.FCTemplate{Tag: "Tenant", Filters: []string{},
			FieldId: utils.Tenant, Type: utils.META_COMPOSED,
			Value: config.NewRSRParsersMustCompile("cgrates.org", true, utils.INFIELD_SEP)},

		&config.FCTemplate{Tag: "Account", Filters: []string{},
			FieldId: utils.Account, Type: utils.META_COMPOSED,
			Value: config.NewRSRParsersMustCompile("~*cgreq.Account", true, utils.INFIELD_SEP)},
		&config.FCTemplate{Tag: "Destination", Filters: []string{},
			FieldId: utils.Destination, Type: utils.META_COMPOSED,
			Value: config.NewRSRParsersMustCompile("~*cgreq.Destination", true, utils.INFIELD_SEP)},
	}
	eMp := config.NewNavigableMap(nil)
	eMp.Set([]string{utils.Tenant}, []*config.NMItem{
		&config.NMItem{Data: "cgrates.org", Path: []string{utils.Tenant},
			Config: tplFlds[0]}}, false, true)
	eMp.Set([]string{utils.Account}, []*config.NMItem{
		&config.NMItem{Data: "1001", Path: []string{utils.Account},
			Config: tplFlds[1]}}, false, true)
	eMp.Set([]string{utils.Destination}, []*config.NMItem{
		&config.NMItem{Data: "1002", Path: []string{utils.Destination},
			Config: tplFlds[2]}}, false, true)

	if mpOut, err := agReq.AsNavigableMap(tplFlds); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eMp, mpOut) {
		t.Errorf("expecting: %+v, received: %+v", eMp, mpOut)
	}
}
