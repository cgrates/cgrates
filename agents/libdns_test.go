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
	"net"
	"strings"
	"testing"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/kamevapi"
	"github.com/miekg/dns"
)

func TestAppendDNSAnswerTypeNAPTR(t *testing.T) {
	if a, err := newDNSAnswer(dns.TypeNAPTR, "3.6.9.4.7.1.7.1.5.6.8.9.4.e164.arpa."); err != nil {
		t.Error(err)
	} else if a.Header().Name != "3.6.9.4.7.1.7.1.5.6.8.9.4.e164.arpa." {
		t.Errorf("expecting: <3.6.9.4.7.1.7.1.5.6.8.9.4.e164.arpa.>, received: <%+v>", a.Header().Name)
	} else if a.Header().Rrtype != 35 {
		t.Errorf("expecting: <35>, received: <%+v>", a.Header().Rrtype)
	} else if a.Header().Class != dns.ClassINET {
		t.Errorf("expecting: <%+v>, received: <%+v>", dns.ClassINET, a.Header().Rrtype)
	} else if a.Header().Ttl != 60 {
		t.Errorf("expecting: <60>, received: <%+v>", a.Header().Rrtype)
	}
}

func TestAppendDNSAnswerTypeA(t *testing.T) {
	if a, err := newDNSAnswer(dns.TypeA, "3.6.9.4.7.1.7.1.5.6.8.9.4.e164.arpa."); err != nil {
		t.Error(err)
	} else if a.Header().Name != "3.6.9.4.7.1.7.1.5.6.8.9.4.e164.arpa." {
		t.Errorf("expecting: <3.6.9.4.7.1.7.1.5.6.8.9.4.e164.arpa.>, received: <%+v>", a.Header().Name)
	} else if a.Header().Rrtype != 1 {
		t.Errorf("expecting: <1>, received: <%+v>", a.Header().Rrtype)
	} else if a.Header().Class != dns.ClassINET {
		t.Errorf("expecting: <%+v>, received: <%+v>", dns.ClassINET, a.Header().Rrtype)
	} else if a.Header().Ttl != 60 {
		t.Errorf("expecting: <60>, received: <%+v>", a.Header().Rrtype)
	}
}

func TestAppendDNSAnswerUnexpectedType(t *testing.T) {
	if _, err := newDNSAnswer(dns.TypeAFSDB, "3.6.9.4.7.1.7.1.5.6.8.9.4.e164.arpa."); err == nil || err.Error() != "unsupported DNS type: <AFSDB>" {
		t.Error(err)
	}
}

func TestUpdateDNSMsgFromNM(t *testing.T) {
	m := new(dns.Msg)
	m.SetQuestion("3.6.9.4.7.1.7.1.5.6.8.9.4.e164.arpa.", dns.TypeNAPTR)

	nM := utils.NewOrderedNavigableMap()
	path := []string{utils.DNSRcode}
	itm := &utils.DataNode{Type: utils.NMDataType, Value: &utils.DataLeaf{
		Data: 10,
	}}
	nM.SetAsSlice(&utils.FullPath{
		Path:      strings.Join(path, utils.NestingSep),
		PathSlice: path,
	}, []*utils.DataNode{itm})
	if err := updateDNSMsgFromNM(m, nM, m.Question[0].Qtype, m.Question[0].Name); err != nil {
		t.Fatal(err)
	}
	if m.Rcode != 10 {
		t.Errorf("expecting: <10>, received: <%+v>", m.Rcode)
	}

	nM = utils.NewOrderedNavigableMap()
	path = []string{utils.DNSRcode}
	itm = &utils.DataNode{Type: utils.NMDataType, Value: &utils.DataLeaf{
		Data: "RandomValue",
	}}
	nM.SetAsSlice(&utils.FullPath{
		Path:      strings.Join(path, utils.NestingSep),
		PathSlice: path,
	}, []*utils.DataNode{itm})
	if err := updateDNSMsgFromNM(m, nM, m.Question[0].Qtype, m.Question[0].Name); err == nil ||
		err.Error() != `item: <Rcode>, err: strconv.ParseInt: parsing "RandomValue": invalid syntax` {
		t.Error(err)
	}

	nM = utils.NewOrderedNavigableMap()
	path = []string{utils.DNSAnswer, utils.Order}
	itm = &utils.DataNode{Type: utils.NMDataType, Value: &utils.DataLeaf{
		Data: "RandomValue",
	}}
	nM.SetAsSlice(&utils.FullPath{
		Path:      strings.Join(path, utils.NestingSep),
		PathSlice: path,
	}, []*utils.DataNode{itm})
	if err := updateDNSMsgFromNM(m, nM, m.Question[0].Qtype, m.Question[0].Name); err == nil ||
		err.Error() != `item: <[Answer Order]>, err: strconv.ParseInt: parsing "RandomValue": invalid syntax` {
		t.Error(err)
	}

	nM = utils.NewOrderedNavigableMap()
	path = []string{utils.DNSAnswer, utils.Preference}
	itm = &utils.DataNode{Type: utils.NMDataType, Value: &utils.DataLeaf{
		Data: "RandomValue",
	}}
	nM.SetAsSlice(&utils.FullPath{
		Path:      strings.Join(path, utils.NestingSep),
		PathSlice: path,
	}, []*utils.DataNode{itm})
	if err := updateDNSMsgFromNM(m, nM, m.Question[0].Qtype, m.Question[0].Name); err == nil ||
		err.Error() != `item: <[Answer Preference]>, err: strconv.ParseInt: parsing "RandomValue": invalid syntax` {
		t.Error(err)
	}

	m = new(dns.Msg)
	m.SetQuestion("3.6.9.4.7.1.7.1.5.6.8.9.4.e164.arpa.", dns.TypeAFSDB)
	nM = utils.NewOrderedNavigableMap()
	path = []string{utils.DNSAnswer, utils.Order}
	itm = &utils.DataNode{Type: utils.NMDataType, Value: &utils.DataLeaf{
		Data: 10,
	}}
	nM.SetAsSlice(&utils.FullPath{
		Path:      strings.Join(path, utils.NestingSep),
		PathSlice: path,
	}, []*utils.DataNode{itm})
	if err := updateDNSMsgFromNM(m, nM, m.Question[0].Qtype, m.Question[0].Name); err == nil ||
		err.Error() != `item: <[Answer Order]>, err: unsupported DNS type: <AFSDB>` {
		t.Error(err)
	}

	nM = utils.NewOrderedNavigableMap()
	path = []string{utils.DNSAnswer, utils.Preference}
	itm = &utils.DataNode{Type: utils.NMDataType, Value: &utils.DataLeaf{
		Data: 10,
	}}
	nM.SetAsSlice(&utils.FullPath{
		Path:      strings.Join(path, utils.NestingSep),
		PathSlice: path,
	}, []*utils.DataNode{itm})
	if err := updateDNSMsgFromNM(m, nM, m.Question[0].Qtype, m.Question[0].Name); err == nil ||
		err.Error() != `item: <[Answer Preference]>, err: unsupported DNS type: <AFSDB>` {
		t.Error(err)
	}

	nM = utils.NewOrderedNavigableMap()
	path = []string{utils.DNSAnswer, utils.Flags}
	itm = &utils.DataNode{Type: utils.NMDataType, Value: &utils.DataLeaf{
		Data: 10,
	}}
	nM.SetAsSlice(&utils.FullPath{
		Path:      strings.Join(path, utils.NestingSep),
		PathSlice: path,
	}, []*utils.DataNode{itm})
	if err := updateDNSMsgFromNM(m, nM, m.Question[0].Qtype, m.Question[0].Name); err == nil ||
		err.Error() != `item: <[Answer Flags]>, err: unsupported DNS type: <AFSDB>` {
		t.Error(err)
	}

	nM = utils.NewOrderedNavigableMap()
	path = []string{utils.DNSAnswer, utils.Service}
	itm = &utils.DataNode{Type: utils.NMDataType, Value: &utils.DataLeaf{
		Data: 10,
	}}
	nM.SetAsSlice(&utils.FullPath{
		Path:      strings.Join(path, utils.NestingSep),
		PathSlice: path,
	}, []*utils.DataNode{itm})
	if err := updateDNSMsgFromNM(m, nM, m.Question[0].Qtype, m.Question[0].Name); err == nil ||
		err.Error() != `item: <[Answer Service]>, err: unsupported DNS type: <AFSDB>` {
		t.Error(err)
	}

	nM = utils.NewOrderedNavigableMap()
	path = []string{utils.DNSAnswer, utils.Regexp}
	itm = &utils.DataNode{Type: utils.NMDataType, Value: &utils.DataLeaf{
		Data: 10,
	}}
	nM.SetAsSlice(&utils.FullPath{
		Path:      strings.Join(path, utils.NestingSep),
		PathSlice: path,
	}, []*utils.DataNode{itm})
	if err := updateDNSMsgFromNM(m, nM, m.Question[0].Qtype, m.Question[0].Name); err == nil ||
		err.Error() != `item: <[Answer Regexp]>, err: unsupported DNS type: <AFSDB>` {
		t.Error(err)
	}

	nM = utils.NewOrderedNavigableMap()
	path = []string{utils.DNSAnswer, utils.Replacement}
	itm = &utils.DataNode{Type: utils.NMDataType, Value: &utils.DataLeaf{
		Data: 10,
	}}
	nM.SetAsSlice(&utils.FullPath{
		Path:      strings.Join(path, utils.NestingSep),
		PathSlice: path,
	}, []*utils.DataNode{itm})
	if err := updateDNSMsgFromNM(m, nM, m.Question[0].Qtype, m.Question[0].Name); err == nil ||
		err.Error() != `item: <[Answer Replacement]>, err: unsupported DNS type: <AFSDB>` {
		t.Error(err)
	}

}

func TestLibdnsNewDnsReply(t *testing.T) {
	req := new(dns.Msg)
	req.SetQuestion("cgrates.org", dns.TypeA)
	rply := newDnsReply(req)
	if len(rply.Question) != len(req.Question) {
		t.Errorf("Expected %d questions, got %d", len(req.Question), len(rply.Question))
	}
	for i, q := range rply.Question {
		if q.Name != req.Question[i].Name {
			t.Errorf("Expected question name %s, got %s", req.Question[i].Name, q.Name)
		}
		if q.Qtype != req.Question[i].Qtype {
			t.Errorf("Expected question type %d, got %d", req.Question[i].Qtype, q.Qtype)
		}
		if q.Qclass != req.Question[i].Qclass {
			t.Errorf("Expected question class %d, got %d", req.Question[i].Qclass, q.Qclass)
		}
	}
	rplyOpts := rply.IsEdns0()
	if rplyOpts == nil {
		rply.Extra = append(rply.Extra, &dns.OPT{
			Hdr: dns.RR_Header{Name: ".", Rrtype: dns.TypeOPT},
			Option: []dns.EDNS0{
				&dns.EDNS0_NSID{Code: dns.EDNS0NSID, Nsid: "test"},
			},
		})
	} else {
		if rplyOpts.UDPSize() != 4096 {
			t.Errorf("Expected EDNS0 UDP size 4096, got %d", rplyOpts.UDPSize())
		}
	}
}

func TestLibDnsDPString(t *testing.T) {
	want := "{\"key\":\"test\"}"
	dp := dnsDP{
		req: utils.MapStorage{
			"key": "test",
		},
	}
	got := dp.String()
	if got != want {
		t.Errorf("Expected String() to return %q, got %q", want, got)
	}
}

func TestLibdnsUpdateDnsRRHeader(t *testing.T) {
	type testCase struct {
		name     string
		rrHeader dns.RR_Header
		path     []string
		value    any
		wantErr  bool
		expected dns.RR_Header
	}
	testCases := []testCase{
		{
			name:     "Update Rrtype (invalid value)",
			rrHeader: dns.RR_Header{},
			path:     []string{utils.DNSRrtype},
			value:    "invalid",
			wantErr:  true,
		},
		{
			name:     "Update invalid path",
			rrHeader: dns.RR_Header{},
			path:     []string{"invalid_path"},
			value:    "test",
			wantErr:  true,
		},
		{
			name:     "Empty path",
			rrHeader: dns.RR_Header{},
			path:     []string{},
			value:    "test",
			wantErr:  true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			originalHeader := tc.rrHeader
			err := updateDnsRRHeader(&tc.rrHeader, tc.path, tc.value)
			if tc.rrHeader != originalHeader {
				t.Errorf("Expected original header to remain unchanged, got changed header")
			}
			if (err != nil) != tc.wantErr {
				t.Errorf("Unexpected error: %v", err)
			}

		})
	}
}

func TestCreateDnsOption(t *testing.T) {
	tests := []struct {
		name     string
		field    string
		value    any
		wantErr  bool
		wantType dns.EDNS0
	}{
		{name: "valid_nsid", field: utils.DNSNsid, value: "1234", wantType: &dns.EDNS0_NSID{Nsid: "1234"}},
		{name: "valid_family", field: utils.DNSFamily, value: 16, wantType: &dns.EDNS0_SUBNET{Family: 16}},
		{name: "invalid_family_type", field: utils.DNSFamily, value: "invalid", wantErr: true},
		{name: "valid_source_netmask", field: utils.DNSSourceNetmask, value: 24, wantType: &dns.EDNS0_SUBNET{SourceNetmask: 24}},
		{name: "valid_address", field: utils.Address, value: "1.2.3.4", wantType: &dns.EDNS0_SUBNET{Address: net.ParseIP("1.2.3.4")}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotOption, gotErr := createDnsOption(tt.field, tt.value)
			if (gotErr != nil) != tt.wantErr {
				t.Errorf("createDnsOption() error = %v, wantErr = %v", gotErr, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if gotOption == nil {
				t.Errorf("createDnsOption() returned nil option")
				return
			}
			switch gotOption := gotOption.(type) {
			case *dns.EDNS0_NSID:
			case *dns.EDNS0_SUBNET:
			default:
				t.Errorf("Unexpected option type returned from createDnsOption: %T", gotOption)
			}
		})
	}
}

func equalDNSQuestionsTest(q1, q2 []dns.Question) bool {
	if len(q1) != len(q2) {
		return false
	}
	for i := range q1 {
		if q1[i] != q2[i] {
			return false
		}
	}
	return true
}

func TestLibdnsUpdateDnsQuestions(t *testing.T) {
	testQ := dns.Question{Name: "cgrates.org", Qtype: dns.TypeA, Qclass: dns.ClassINET}
	tests := []struct {
		name      string
		q         []dns.Question
		path      []string
		value     any
		newBranch bool
		wantErr   bool
		wantQ     []dns.Question
	}{
		{
			"update_name_existing",
			[]dns.Question{testQ},
			[]string{utils.DNSName},
			"cgrates.org",
			false,
			false,
			[]dns.Question{{Name: "cgrates.org", Qtype: dns.TypeA, Qclass: dns.ClassINET}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotQ, err := updateDnsQuestions(tt.q, tt.path, tt.value, tt.newBranch)
			if (err != nil) != tt.wantErr {
				t.Errorf("updateDnsQuestions() error = %v, wantErr = %v", err, tt.wantErr)
				return
			}
			if !equalDNSQuestionsTest(gotQ, tt.wantQ) {
				t.Errorf("updateDnsQuestions() gotQ = %v, wantQ = %v", gotQ, tt.wantQ)
			}
		})
	}

}

func TestKamailioAgentCall(t *testing.T) {
	cfg := &config.KamAgentCfg{}
	connMgr := &engine.ConnManager{}
	conns := []*kamevapi.KamEvapi{}
	activeSessionIDs := make(chan []*sessions.SessionID)
	ctx := &context.Context{}
	ka := &KamailioAgent{
		cfg:              cfg,
		connMgr:          connMgr,
		timezone:         "UTC",
		conns:            conns,
		activeSessionIDs: activeSessionIDs,
		ctx:              ctx,
	}
	args := struct {
		Message string
	}{
		Message: "message",
	}
	var reply string
	err := ka.Call("UNSUPPORTED_SERVICE_METHOD", args, &reply)
	if err == nil {
		t.Errorf("UNSUPPORTED_SERVICE_METHOD %v", err)
	}
	expectedReply := ""
	if reply != expectedReply {
		t.Errorf("Expected reply %q, got %q", expectedReply, reply)
	}
}

func TestLibDnsUpdateDnsOption(t *testing.T) {

	ednsOptions := []dns.EDNS0{
		&dns.EDNS0_NSID{},
		&dns.EDNS0_SUBNET{},
		&dns.EDNS0_COOKIE{},
		&dns.EDNS0_UL{},
		&dns.EDNS0_LLQ{},
		&dns.EDNS0_DAU{},
		&dns.EDNS0_DHU{},
		&dns.EDNS0_N3U{},
		&dns.EDNS0_EXPIRE{},
		&dns.EDNS0_TCP_KEEPALIVE{},
		&dns.EDNS0_PADDING{},
		&dns.EDNS0_EDE{},
		&dns.EDNS0_ESU{},
		&dns.EDNS0_LOCAL{},
	}
	path := []string{"0", utils.DNSNsid}
	value := "test-nsid"
	newBranch := false
	updatedOptions, err := updateDnsOption(ednsOptions, path, value, newBranch)
	if err != nil {
		t.Errorf("Update EDNS0_NSID's NSID field returned unexpected error: %v", err)
	}
	if nsidOption, ok := updatedOptions[0].(*dns.EDNS0_NSID); ok {
		if nsidOption.Nsid != value {
			t.Errorf("Expected NSID %s, got %s", value, nsidOption.Nsid)
		}
	} else {
		t.Errorf("Expected EDNS0_NSID option, got %T", updatedOptions[0])
	}
	path = []string{"1", utils.DNSFamily}
	valueInt := 1
	newBranch = true
	updatedOptions, err = updateDnsOption(ednsOptions, path, valueInt, newBranch)
	if err != nil {
		t.Errorf("Update EDNS0_SUBNET's Family field returned unexpected error: %v", err)
	}
	if subnetOption, ok := updatedOptions[1].(*dns.EDNS0_SUBNET); ok {
		if subnetOption.Family != uint16(valueInt) {
			t.Errorf("Expected Family %d, got %d", valueInt, subnetOption.Family)
		}
	} else {
		t.Errorf("Expected EDNS0_SUBNET option, got %T", updatedOptions[1])
	}
	path = []string{"0", utils.DNSNsid, "extra"}
	value = "value"
	newBranch = false
	_, err = updateDnsOption(ednsOptions, path, value, newBranch)
	expectedErrMsg := "WRONG_PATH"
	if err == nil || err.Error() != expectedErrMsg {
		t.Errorf("Expected error '%s', got '%v'", expectedErrMsg, err)
	}

}
