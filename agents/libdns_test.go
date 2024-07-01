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

func TestLibDnsUpdateDnsSRVAnswer(t *testing.T) {
	tests := []struct {
		name    string
		path    []string
		value   interface{}
		expect  func(v *dns.SRV) bool
		wantErr bool
	}{
		{
			name:  "update Priority",
			path:  []string{utils.DNSPriority},
			value: int64(10),
			expect: func(v *dns.SRV) bool {
				return v.Priority == 10
			},
			wantErr: false,
		},
		{
			name:  "update Weight",
			path:  []string{utils.Weight},
			value: int64(20),
			expect: func(v *dns.SRV) bool {
				return v.Weight == 20
			},
			wantErr: false,
		},
		{
			name:  "update Port",
			path:  []string{utils.DNSPort},
			value: int64(2012),
			expect: func(v *dns.SRV) bool {
				return v.Port == 2012
			},
			wantErr: false,
		},
		{
			name:  "update Target",
			path:  []string{utils.DNSTarget},
			value: "cgrates.com",
			expect: func(v *dns.SRV) bool {
				return v.Target == "cgrates.com"
			},
			wantErr: false,
		},
		{
			name:    "invalid path length",
			path:    []string{},
			value:   int64(10),
			expect:  nil,
			wantErr: true,
		},
		{
			name:    "invalid path value",
			path:    []string{"invalid"},
			value:   int64(10),
			expect:  nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := &dns.SRV{}
			err := updateDnsSRVAnswer(srv, tt.path, tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("updateDnsSRVAnswer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.expect != nil && !tt.expect(srv) {
				t.Errorf("updateDnsSRVAnswer() unexpected result for %v", srv)
			}
		})
	}
}

func TestLibDnsUpdateDnsSRVAnswerDNSHdr(t *testing.T) {
	srv := &dns.SRV{}

	err := updateDnsSRVAnswer(srv, []string{utils.DNSHdr, utils.DNSName}, "cgrates.com.")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if srv.Hdr.Name != "cgrates.com." {
		t.Errorf("expected Name to be 'cgrates.com.', got %s", srv.Hdr.Name)
	}
}

func TestLibDnsUpdateDnsRRHeaderDNSClass(t *testing.T) {
	rrHeader := new(dns.RR_Header)
	path := []string{utils.DNSClass}
	value := int64(1)
	err := updateDnsRRHeader(rrHeader, path, value)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	expectedClass := uint16(1)
	if rrHeader.Class != expectedClass {
		t.Errorf("Expected rrHeader.Class to be %d, got %d", expectedClass, rrHeader.Class)
	}
}

func TestLibDnsUpdateDnsRRHeaderDNSRdlength(t *testing.T) {
	rrHeader := new(dns.RR_Header)
	path := []string{utils.DNSRdlength}
	value := int64(256)
	err := updateDnsRRHeader(rrHeader, path, value)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	expectedRdlength := uint16(256)
	if rrHeader.Rdlength != expectedRdlength {
		t.Errorf("Expected rrHeader.Rdlength to be %d, got %d", expectedRdlength, rrHeader.Rdlength)
	}
}

func TestLibDnsUpdateDnsRRHeaderDNSTtl(t *testing.T) {
	rrHeader := new(dns.RR_Header)
	path := []string{utils.DNSTtl}
	value := int64(3600)
	err := updateDnsRRHeader(rrHeader, path, value)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	expectedTtl := uint32(3600)
	if rrHeader.Ttl != expectedTtl {
		t.Errorf("Expected rrHeader.Ttl to be %d, got %d", expectedTtl, rrHeader.Ttl)
	}
}

func TestLibDnsUpdateDnsNAPTRAnswerDefaultCase(t *testing.T) {
	naptr := new(dns.NAPTR)
	path := []string{"unsupported_path"}
	value := "value"
	err := updateDnsNAPTRAnswer(naptr, path, value)
	if err != utils.ErrWrongPath {
		t.Errorf("Expected error: %v, got: %v", utils.ErrWrongPath, err)
	}
}

func TestLibDnsUpdateDnsNAPTRAnswerReplacementCase(t *testing.T) {
	naptr := new(dns.NAPTR)
	path := []string{utils.Replacement}
	value := "value"
	err := updateDnsNAPTRAnswer(naptr, path, value)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if naptr.Replacement != value {
		t.Errorf("Expected v.Replacement to be %q, got %q", value, naptr.Replacement)
	}
}

func TestLibDnsUpdateDnsNAPTRAnswerRegexpCase(t *testing.T) {
	naptr := new(dns.NAPTR)
	path := []string{utils.Regexp}
	value := "value"
	err := updateDnsNAPTRAnswer(naptr, path, value)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if naptr.Regexp != value {
		t.Errorf("Expected v.Regexp to be %q, got %q", value, naptr.Regexp)
	}
}

func TestLibDnsUpdateDnsNAPTRAnswerServiceCase(t *testing.T) {
	naptr := new(dns.NAPTR)
	path := []string{utils.Service}
	value := "value"
	err := updateDnsNAPTRAnswer(naptr, path, value)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if naptr.Service != value {
		t.Errorf("Expected v.Service to be %q, got %q", value, naptr.Service)
	}
}

func TestLibDnsUpdateDnsNAPTRAnswerFlagsCase(t *testing.T) {
	var naptr dns.NAPTR
	path := []string{utils.Flags}
	value := "example_flags_value"
	err := updateDnsNAPTRAnswer(&naptr, path, value)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if naptr.Flags != value {
		t.Errorf("Expected v.Flags to be %q, got %q", value, naptr.Flags)
	}
}

func TestLibDnsUpdateDnsNAPTRAnswerPreferenceCase(t *testing.T) {
	var naptr dns.NAPTR
	path := []string{utils.Preference}
	value := int64(100)
	err := updateDnsNAPTRAnswer(&naptr, path, value)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if naptr.Preference != uint16(value) {
		t.Errorf("Expected v.Preference to be %d, got %d", value, naptr.Preference)
	}
}

func TestLibDnsUpdateDnsNAPTRAnswerOrderCase(t *testing.T) {
	var naptr dns.NAPTR
	path := []string{utils.Order}
	value := int64(50)
	err := updateDnsNAPTRAnswer(&naptr, path, value)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if naptr.Order != uint16(value) {
		t.Errorf("Expected v.Order to be %d, got %d", value, naptr.Order)
	}
}

func TestLibDnsUpdateDnsNAPTRAnswerHdrCase(t *testing.T) {
	var naptr dns.NAPTR
	path := []string{utils.DNSHdr, "path"}
	value := "hdr_value"
	err := updateDnsNAPTRAnswer(&naptr, path, value)
	if err == nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestLibDnsUpdateDnsNAPTRAnswerWrongPath(t *testing.T) {
	var naptr dns.NAPTR
	testCases := []struct {
		path  []string
		value any
	}{
		{[]string{}, "value"},
		{[]string{"invalid_path"}, "value"},
		{[]string{utils.DNSHdr}, "value"},
		{[]string{utils.DNSHdr, "subpath1", "subpath2"}, "value"},
	}
	for _, tc := range testCases {
		err := updateDnsNAPTRAnswer(&naptr, tc.path, tc.value)
		if err != utils.ErrWrongPath {
			t.Errorf("Expected error %v for path %v, got %v", utils.ErrWrongPath, tc.path, err)
		}
	}
}

func TestLibDnsNewDNSAnswerSRV(t *testing.T) {
	qType := dns.TypeSRV
	qName := "cgrates.com"
	a, err := newDNSAnswer(qType, qName)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	srv, ok := a.(*dns.SRV)
	if !ok {
		t.Errorf("Expected a DNS SRV record, got %T", a)
	}
	if srv.Hdr.Name != qName || srv.Hdr.Rrtype != qType {
		t.Errorf("Unexpected header values. Expected Name: %s, Rrtype: %d, got Name: %s, Rrtype: %d",
			qName, qType, srv.Hdr.Name, srv.Hdr.Rrtype)
	}
}

func TestLibDnsUpdateDnsAnswerSRVCase(t *testing.T) {
	q := make([]dns.RR, 0)
	qType := dns.TypeSRV
	qName := "cgrates.com"
	path := []string{"home_path"}
	value := "value"
	srv := &dns.SRV{
		Hdr:    dns.RR_Header{Name: qName, Rrtype: qType, Class: dns.ClassINET, Ttl: 60},
		Target: "cgrates.com",
		Port:   8080,
	}
	q = append(q, srv)
	_, err := updateDnsAnswer(q, qType, qName, path, value, true)
	if err == nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if srv.Target == value {
		t.Errorf("cgrates.com")
	}
}

func TestLibDnsUpdateDnsAnswerACase(t *testing.T) {
	q := make([]dns.RR, 0)
	qType := dns.TypeA
	qName := "cgrates.com"
	path := []string{"home_path"}
	value := "192.168.1.1"
	a := &dns.A{
		Hdr: dns.RR_Header{Name: qName, Rrtype: qType, Class: dns.ClassINET, Ttl: 60},
		A:   net.ParseIP(value),
	}
	q = append(q, a)
	_, err := updateDnsAnswer(q, qType, qName, path, value, true)
	if err == nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if a.A.String() != value {
		t.Errorf("Expected a.A to be %q, got %q", value, a.A.String())
	}
}
