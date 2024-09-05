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
	"fmt"
	"net"
	"reflect"
	"strings"
	"testing"

	"github.com/cgrates/cgrates/utils"
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

func TestLibDnsCreateDnsOptionDefaultCase(t *testing.T) {
	invalidField := "InvalidField"
	value := "Value"
	_, err := createDnsOption(invalidField, value)
	expectedError := fmt.Sprintf("can not create option from field <%q>", invalidField)
	if err == nil {
		t.Errorf("expected error, got nil")
	} else if err.Error() != expectedError {
		t.Errorf("expected error message %q, got %q", expectedError, err.Error())
	}
}

func TestLibDnsCreateDnsOption_DNSUri(t *testing.T) {
	field := utils.DNSUri
	value := "http://cgrates.org"
	option, err := createDnsOption(field, value)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	esuOption, ok := option.(*dns.EDNS0_ESU)
	if !ok {
		t.Errorf("expected option to be of type *dns.EDNS0_ESU, got %T", option)
	}
	if esuOption.Uri != value {
		t.Errorf("expected Uri to be %q, got %q", value, esuOption.Uri)
	}
}

func TestLibDnsCreateDnsOptionDNSExtraText(t *testing.T) {
	field := utils.DNSExtraText
	value := "ExtraText"
	option, err := createDnsOption(field, value)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	edeOption, ok := option.(*dns.EDNS0_EDE)
	if !ok {
		t.Errorf("expected option to be of type *dns.EDNS0_EDE, got %T", option)
	}
	if edeOption.ExtraText != value {
		t.Errorf("expected ExtraText to be %q, got %q", value, edeOption.ExtraText)
	}
}

func TestLibDnsCreateDnsOptionDNSInfoCode(t *testing.T) {
	field := utils.DNSInfoCode
	value := int64(1234)
	option, err := createDnsOption(field, value)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	edeOption, ok := option.(*dns.EDNS0_EDE)
	if !ok {
		t.Errorf("expected option to be of type *dns.EDNS0_EDE, got %T", option)
	}
	if edeOption.InfoCode != uint16(value) {
		t.Errorf("expected InfoCode to be %d, got %d", value, edeOption.InfoCode)
	}
}

func TestLibDnsCreateDnsOptionDNSInfoCodeWithError(t *testing.T) {
	field := utils.DNSInfoCode
	value := "invalid_value"
	option, err := createDnsOption(field, value)
	if err == nil {
		t.Errorf("expected error, got nil")
	}
	if option != nil {
		t.Errorf("expected nil option, got %v", option)
	}
}

func TestLibDnsCreateDnsOptionDNSPadding(t *testing.T) {
	field := utils.DNSPadding
	value := "PaddingText"
	option, err := createDnsOption(field, value)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	paddingOption, ok := option.(*dns.EDNS0_PADDING)
	if !ok {
		t.Errorf("expected option to be of type *dns.EDNS0_PADDING, got %T", option)
	}
	expectedPadding := []byte(value)
	if string(paddingOption.Padding) != string(expectedPadding) {
		t.Errorf("expected Padding to be %q, got %q", expectedPadding, paddingOption.Padding)
	}
}

func TestLibDnsCreateDnsOptionDNSTimeout(t *testing.T) {
	field := utils.DNSTimeout
	value := int64(300)
	option, err := createDnsOption(field, value)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	tcpKeepaliveOption, ok := option.(*dns.EDNS0_TCP_KEEPALIVE)
	if !ok {
		t.Errorf("expected option to be of type *dns.EDNS0_TCP_KEEPALIVE, got %T", option)
	}
	expectedTimeout := uint16(value)
	if tcpKeepaliveOption.Timeout != expectedTimeout {
		t.Errorf("expected Timeout to be %d, got %d", expectedTimeout, tcpKeepaliveOption.Timeout)
	}
}

func TestLibDnsCreateDnsOptionDNSTimeoutWithError(t *testing.T) {
	field := utils.DNSTimeout
	value := "invalid_value"
	option, err := createDnsOption(field, value)
	if err == nil {
		t.Errorf("expected error, got nil")
	}
	if option != nil {
		t.Errorf("expected nil option, got %v", option)
	}
}

func TestLibDnsCreateDnsOptionLength(t *testing.T) {
	field := utils.Length
	value := int64(1500)
	option, err := createDnsOption(field, value)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	tcpKeepaliveOption, ok := option.(*dns.EDNS0_TCP_KEEPALIVE)
	if !ok {
		t.Errorf("expected option to be of type *dns.EDNS0_TCP_KEEPALIVE, got %T", option)
	}
	expectedLength := uint16(value)
	if tcpKeepaliveOption.Length != expectedLength {
		t.Errorf("expected Length to be %d, got %d", expectedLength, tcpKeepaliveOption.Length)
	}
}

func TestLibDnsCreateDnsOptionLengthWithError(t *testing.T) {
	field := utils.Length
	value := "invalid_value"
	option, err := createDnsOption(field, value)
	if err == nil {
		t.Errorf("expected error, got nil")
	}
	if option != nil {
		t.Errorf("expected nil option, got %v", option)
	}
}

func TestLibDnsCreateDnsOptionDNSExpireWithError(t *testing.T) {
	field := utils.DNSExpire
	value := "invalid_value"
	option, err := createDnsOption(field, value)
	if err == nil {
		t.Errorf("expected error, got nil")
	}
	if option != nil {
		t.Errorf("expected nil option, got %v", option)
	}
}

func TestLibDnsUpdateDnsQuestionsDefaultCase(t *testing.T) {
	q := []dns.Question{
		{Name: "cgrates.org.", Qtype: dns.TypeA, Qclass: dns.ClassINET},
	}
	path := []string{"unsupportedField"}
	value := "Value"
	newBranch := false
	_, err := updateDnsQuestions(q, path, value, newBranch)
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
	if err != utils.ErrWrongPath {
		t.Errorf("Expected error %v, got %v", utils.ErrWrongPath, err)
	}
}

func TestUpdateDnsQuestionsDNSQclass(t *testing.T) {
	q := []dns.Question{
		{Name: "cgrates.org.", Qtype: dns.TypeA, Qclass: dns.ClassINET},
	}
	path := []string{"0", utils.DNSQclass}
	value := int64(5)
	newBranch := false
	updatedQ, err := updateDnsQuestions(q, path, value, newBranch)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if len(updatedQ) == 0 || updatedQ[0].Qclass != 5 {
		t.Errorf("Expected Qclass to be updated to 5, got %v", updatedQ[0].Qclass)
	}
}

func TestLibDnsUpdateDnsQuestionsDNSQclassError(t *testing.T) {
	q := []dns.Question{
		{Name: "cgrates.org.", Qtype: dns.TypeA, Qclass: dns.ClassINET},
	}
	path := []string{"0", utils.DNSQclass}
	value := "invalidValue"
	newBranch := false
	_, err := updateDnsQuestions(q, path, value, newBranch)
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
}

func TestLibDnsUpdateDnsQuestionsDNSQtype(t *testing.T) {
	q := []dns.Question{
		{Name: "cgrates.org.", Qtype: dns.TypeA, Qclass: dns.ClassINET},
	}
	path := []string{"0", utils.DNSQtype}
	value := int64(15)
	newBranch := false
	updatedQ, err := updateDnsQuestions(q, path, value, newBranch)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if len(updatedQ) == 0 || updatedQ[0].Qtype != 15 {
		t.Errorf("Expected Qtype to be updated to 15, got %v", updatedQ[0].Qtype)
	}
}

func TestLibDnsUpdateDnsQuestionsDNSQtypeError(t *testing.T) {
	q := []dns.Question{
		{Name: "cgrates.org.", Qtype: dns.TypeA, Qclass: dns.ClassINET},
	}
	path := []string{"0", utils.DNSQtype}
	value := "invalidValue"
	newBranch := false
	_, err := updateDnsQuestions(q, path, value, newBranch)
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
}

func TestLibDnsCreateDnsOptionDNSN3U(t *testing.T) {
	t.Run("Normal case", func(t *testing.T) {
		field := utils.DNSN3U
		value := "1"
		option, err := createDnsOption(field, value)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		edns0N3U, ok := option.(*dns.EDNS0_N3U)
		if !ok {
			t.Fatalf("Expected type *dns.EDNS0_N3U, got %T", option)
		}
		expectedAlgCode := []uint8(value)
		if string(edns0N3U.AlgCode) != string(expectedAlgCode) {
			t.Errorf("Expected AlgCode to be %v, got %v", expectedAlgCode, edns0N3U.AlgCode)
		}
	})
	t.Run("Error case", func(t *testing.T) {
		field := utils.DNSN3U
		value := 12345
		_, err := createDnsOption(field, value)
		if err != nil {
			t.Fatalf("Expected error, got nil")
		}
	})
}

func TestLibDnsCreateDnsOptionT(t *testing.T) {
	t.Run("DNSN3U Normal case", func(t *testing.T) {
		field := utils.DNSN3U
		value := "1"
		option, err := createDnsOption(field, value)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		edns0N3U, ok := option.(*dns.EDNS0_N3U)
		if !ok {
			t.Fatalf("Expected type *dns.EDNS0_N3U, got %T", option)
		}
		expectedAlgCode := []uint8(value)
		if string(edns0N3U.AlgCode) != string(expectedAlgCode) {
			t.Errorf("Expected AlgCode to be %v, got %v", expectedAlgCode, edns0N3U.AlgCode)
		}
	})
	t.Run("DNSN3U Error case", func(t *testing.T) {
		field := utils.DNSN3U
		value := 1

		_, err := createDnsOption(field, value)
		if err != nil {
			t.Fatalf("Expected error, got nil")
		}
	})
	t.Run("DNSDHU Normal case", func(t *testing.T) {
		field := utils.DNSDHU
		value := "1"
		option, err := createDnsOption(field, value)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		edns0DHU, ok := option.(*dns.EDNS0_DHU)
		if !ok {
			t.Fatalf("Expected type *dns.EDNS0_DHU, got %T", option)
		}
		expectedAlgCode := []uint8(value)
		if string(edns0DHU.AlgCode) != string(expectedAlgCode) {
			t.Errorf("Expected AlgCode to be %v, got %v", expectedAlgCode, edns0DHU.AlgCode)
		}
	})
	t.Run("DNSDHU Error case", func(t *testing.T) {
		field := utils.DNSDHU
		value := 1
		_, err := createDnsOption(field, value)
		if err != nil {
			t.Fatalf("Expected error, got nil")
		}
	})
}

func TestLibDnsCreateDnsOptionTab(t *testing.T) {
	testCases := []struct {
		field string
		value any
		want  dns.EDNS0
	}{
		{utils.DNSN3U, "12345", &dns.EDNS0_N3U{AlgCode: []uint8("12345")}},
		{utils.DNSDHU, "12345", &dns.EDNS0_DHU{AlgCode: []uint8("12345")}},
		{utils.DNSDAU, "12345", &dns.EDNS0_DAU{AlgCode: []uint8("12345")}},
	}
	for _, tc := range testCases {
		got, err := createDnsOption(tc.field, tc.value)
		if err != nil {
			t.Errorf("createDnsOption(%q, %v) returned error: %v", tc.field, tc.value, err)
			continue
		}
		if got.String() != tc.want.String() {
			t.Errorf("createDnsOption(%q, %v) = %v, want %v", tc.field, tc.value, got, tc.want)
		}
	}
}

func TestLibDnsCreateDnsOption(t *testing.T) {
	testCases := []struct {
		field string
		value any
		want  dns.EDNS0
	}{
		{utils.DNSLease, int64(3600), &dns.EDNS0_UL{Lease: uint32(3600)}},
		{utils.DNSKeyLease, int64(3600), &dns.EDNS0_UL{KeyLease: uint32(3600)}},
		{utils.VersionName, int64(1), &dns.EDNS0_LLQ{Version: uint16(1)}},
		{utils.DNSOpcode, int64(2), &dns.EDNS0_LLQ{Opcode: uint16(2)}},
		{utils.Error, int64(3), &dns.EDNS0_LLQ{Error: uint16(3)}},
		{utils.DNSId, int64(12345), &dns.EDNS0_LLQ{Id: uint64(12345)}},
		{utils.DNSLeaseLife, int64(3600), &dns.EDNS0_LLQ{LeaseLife: uint32(3600)}},
	}

	for _, tc := range testCases {
		got, err := createDnsOption(tc.field, tc.value)
		if err != nil {
			t.Errorf("createDnsOption(%q, %v) returned error: %v", tc.field, tc.value, err)
			continue
		}
		if got.String() != tc.want.String() {
			t.Errorf("createDnsOption(%q, %v) = %v, want %v", tc.field, tc.value, got, tc.want)
		}
	}
}

func TestLibDnsCreateDnsOptionError(t *testing.T) {
	testCases := []struct {
		field string
		value any
	}{
		{utils.DNSLease, "test"},
		{utils.DNSKeyLease, "test"},
		{utils.VersionName, "test"},
		{utils.DNSOpcode, "test"},
		{utils.Error, "test"},
		{utils.DNSId, "test"},
		{utils.DNSLeaseLife, "test"},
	}

	for _, tc := range testCases {
		_, err := createDnsOption(tc.field, tc.value)
		if err == nil {
			t.Errorf("createDnsOption(%q, %v) expected error, got nil", tc.field, tc.value)
		}
	}
}

func TestLibDnsCreateDnsOptionDNSCookie(t *testing.T) {
	field := utils.DNSCookie
	value := "test_cookie_value"
	option, err := createDnsOption(field, value)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	edns0Cookie, ok := option.(*dns.EDNS0_COOKIE)
	if !ok {
		t.Errorf("Expected type *dns.EDNS0_COOKIE, got %T", option)
	}
	expectedCookie := value
	if edns0Cookie.Cookie != expectedCookie {
		t.Errorf("Expected Cookie to be %q, got %q", expectedCookie, edns0Cookie.Cookie)
	}
}

func TestLibDnsCreateDnsOptionDNSSourceScope(t *testing.T) {
	field := utils.DNSSourceScope
	value := int64(64)
	option, err := createDnsOption(field, value)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	edns0Subnet, ok := option.(*dns.EDNS0_SUBNET)
	if !ok {
		t.Errorf("Expected type *dns.EDNS0_SUBNET, got %T", option)
	}
	expectedSourceScope := uint8(value)
	if edns0Subnet.SourceScope != expectedSourceScope {
		t.Errorf("Expected SourceScope to be %d, got %d", expectedSourceScope, edns0Subnet.SourceScope)
	}
}

func TestLibDnsUpdateDnsOptionEDNS0LOCAL(t *testing.T) {
	q := []dns.EDNS0{&dns.EDNS0_LOCAL{Data: []byte("existing data")}}
	path := []string{"0", utils.DNSData}
	value := "new data"
	_, err := updateDnsOption(q, path, value, false)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	edns0Local, ok := q[0].(*dns.EDNS0_LOCAL)
	if !ok {
		t.Errorf("Expected type *dns.EDNS0_LOCAL, got %T", q[0])
	}
	expectedData := []byte(value)
	if string(edns0Local.Data) != string(expectedData) {
		t.Errorf("Expected Data to be %v, got %v", expectedData, edns0Local.Data)
	}
}

func TestLibDnsUpdateDnsOptionEDNS0LOCALWrongPath(t *testing.T) {
	q := []dns.EDNS0{&dns.EDNS0_LOCAL{Data: []byte("existing data")}}
	path := []string{"0", "wrongField"}
	value := "new data"
	_, err := updateDnsOption(q, path, value, false)
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
	expectedError := utils.ErrWrongPath
	if err != expectedError {
		t.Errorf("Expected error %v, got %v", expectedError, err)
	}
}

func TestLibDnsUpdateDnsOptionEDNS0ESUWrongPath(t *testing.T) {
	q := []dns.EDNS0{&dns.EDNS0_ESU{Uri: "existing-uri"}}
	path := []string{"0", "wrongField"}
	value := "new-uri"
	_, err := updateDnsOption(q, path, value, false)
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
	expectedError := utils.ErrWrongPath
	if err != expectedError {
		t.Errorf("Expected error %v, got %v", expectedError, err)
	}
}

func TestLibDnsUpdateDnsOptioEDNS0EDEInfoCode(t *testing.T) {
	q := []dns.EDNS0{&dns.EDNS0_EDE{InfoCode: 0}}
	path := []string{"0", utils.DNSInfoCode}
	value := int64(123)
	updatedQ, err := updateDnsOption(q, path, value, false)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	edns0EDE, ok := updatedQ[0].(*dns.EDNS0_EDE)
	if !ok {
		t.Errorf("Expected type *dns.EDNS0_EDE, got %T", updatedQ[0])
	}
	expectedInfoCode := uint16(value)
	if edns0EDE.InfoCode != expectedInfoCode {
		t.Errorf("Expected InfoCode to be %v, got %v", expectedInfoCode, edns0EDE.InfoCode)
	}
}

func TestLibDnsUpdateDnsOptionEDNS0EDEExtraText(t *testing.T) {
	q := []dns.EDNS0{&dns.EDNS0_EDE{ExtraText: ""}}
	path := []string{"0", utils.DNSExtraText}
	value := "extra text"
	updatedQ, err := updateDnsOption(q, path, value, false)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	edns0EDE, ok := updatedQ[0].(*dns.EDNS0_EDE)
	if !ok {
		t.Errorf("Expected type *dns.EDNS0_EDE, got %T", updatedQ[0])
	}
	if edns0EDE.ExtraText != value {
		t.Errorf("Expected ExtraText to be %q, got %q", value, edns0EDE.ExtraText)
	}
}

func TestLibDnsUpdateDnsOptionPadding(t *testing.T) {
	paddingOption := &dns.EDNS0_PADDING{}
	edns0 := []dns.EDNS0{paddingOption}
	path := []string{"0", utils.DNSPadding}
	value := "PaddingValue"
	newBranch := false
	updatedEdns0, err := updateDnsOption(edns0, path, value, newBranch)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}
	if len(updatedEdns0) == 0 {
		t.Errorf("no EDNS0 options returned")
		return
	}
	paddingOpt, ok := updatedEdns0[0].(*dns.EDNS0_PADDING)
	if !ok {
		t.Errorf("expected type *dns.EDNS0_PADDING, got %T", updatedEdns0[0])
		return
	}
	expectedPadding := []byte("PaddingValue")
	if string(paddingOpt.Padding) != string(expectedPadding) {
		t.Errorf("expected Padding %v, got %v", expectedPadding, paddingOpt.Padding)
	}
}

func TestLibDnsUpdateDnsOptionTCPKeepAlive(t *testing.T) {
	keepAliveOption := &dns.EDNS0_TCP_KEEPALIVE{}
	edns0 := []dns.EDNS0{keepAliveOption}
	path := []string{"0", utils.Length}
	value := int64(300)
	newBranch := false
	updatedEdns0, err := updateDnsOption(edns0, path, value, newBranch)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}
	if len(updatedEdns0) == 0 {
		t.Errorf("no EDNS0 options returned")
		return
	}
	keepAliveOpt, ok := updatedEdns0[0].(*dns.EDNS0_TCP_KEEPALIVE)
	if !ok {
		t.Errorf("expected type *dns.EDNS0_TCP_KEEPALIVE, got %T", updatedEdns0[0])
		return
	}
	expectedLength := uint16(300)
	if keepAliveOpt.Length != expectedLength {
		t.Errorf("expected Length %v, got %v", expectedLength, keepAliveOpt.Length)
	}
	keepAliveOption = &dns.EDNS0_TCP_KEEPALIVE{}
	edns0 = []dns.EDNS0{keepAliveOption}
	path = []string{"0", utils.DNSTimeout}
	value = int64(100)
	newBranch = false
	updatedEdns0, err = updateDnsOption(edns0, path, value, newBranch)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}
	if len(updatedEdns0) == 0 {
		t.Errorf("no EDNS0 options returned")
		return
	}
	keepAliveOpt, ok = updatedEdns0[0].(*dns.EDNS0_TCP_KEEPALIVE)
	if !ok {
		t.Errorf("expected type *dns.EDNS0_TCP_KEEPALIVE, got %T", updatedEdns0[0])
		return
	}
	expectedTimeout := uint16(100)
	if keepAliveOpt.Timeout != expectedTimeout {
		t.Errorf("expected Timeout %v, got %v", expectedTimeout, keepAliveOpt.Timeout)
	}
}

func TestLibDnsUpdateDnsOptionExpire(t *testing.T) {
	expireOption := &dns.EDNS0_EXPIRE{}
	edns0 := []dns.EDNS0{expireOption}
	path := []string{"0", utils.DNSExpire}
	value := int64(3600)
	newBranch := false
	updatedEdns0, err := updateDnsOption(edns0, path, value, newBranch)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}
	if len(updatedEdns0) == 0 {
		t.Errorf("no EDNS0 options returned")
		return
	}
	expireOpt, ok := updatedEdns0[0].(*dns.EDNS0_EXPIRE)
	if !ok {
		t.Errorf("expected type *dns.EDNS0_EXPIRE, got %T", updatedEdns0[0])
		return
	}
	expectedExpire := uint32(3600)
	if expireOpt.Expire != expectedExpire {
		t.Errorf("expected Expire %v, got %v", expectedExpire, expireOpt.Expire)
	}
}

func TestLibDnsUpdateDnsOptions(t *testing.T) {
	tests := []struct {
		name            string
		path            []string
		value           any
		newBranch       bool
		expectedAlgCode []uint8
		expectedError   error
	}{
		{
			name:            "Update EDNS0_N3U successfully",
			path:            []string{"0", utils.DNSN3U},
			value:           "value",
			newBranch:       false,
			expectedAlgCode: []uint8("value"),
			expectedError:   nil,
		},
		{
			name:            "Error case - wrong field",
			path:            []string{"0", "wrong_field"},
			value:           "value",
			newBranch:       false,
			expectedAlgCode: nil,
			expectedError:   utils.ErrWrongPath,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tEDNS0_N3U := &dns.EDNS0_N3U{
				AlgCode: []uint8{},
			}
			q := []dns.EDNS0{
				tEDNS0_N3U,
			}
			updatedQ, err := updateDnsOption(q, tt.path, tt.value, tt.newBranch)
			if err != tt.expectedError {
				t.Fatalf("updateDnsOption failed: expected error %v, got %v", tt.expectedError, err)
			}
			if err == utils.ErrWrongPath {
				return
			}
			if !reflect.DeepEqual(tEDNS0_N3U.AlgCode, tt.expectedAlgCode) {
				t.Errorf("updateDnsOption did not update EDNS0_N3U correctly. Expected AlgCode %v, got %v", tt.expectedAlgCode, tEDNS0_N3U.AlgCode)
			}
			expectedLength := 1
			if len(updatedQ) != expectedLength {
				t.Errorf("updateDnsOption did not return the expected number of elements. Expected %d, got %d", expectedLength, len(updatedQ))
			}
		})
	}
}

func TestLibDnsUpdateDnsOptionEDNS0DHU(t *testing.T) {
	tEDNS0_DHU := &dns.EDNS0_DHU{
		AlgCode: []uint8{},
	}
	q := []dns.EDNS0{
		tEDNS0_DHU,
	}
	path := []string{"0", utils.DNSDHU}
	value := "value"
	newBranch := false
	updatedQ, err := updateDnsOption(q, path, value, newBranch)
	if err != nil {
		t.Fatalf("updateDnsOption failed: %v", err)
	}
	expectedAlgCode := []uint8("value")
	if !reflect.DeepEqual(tEDNS0_DHU.AlgCode, expectedAlgCode) {
		t.Errorf("updateDnsOption did not update EDNS0_DHU correctly. Expected AlgCode %v, got %v", expectedAlgCode, tEDNS0_DHU.AlgCode)
	}

	expectedLength := 1
	if len(updatedQ) != expectedLength {
		t.Errorf("updateDnsOption did not return the expected number of elements. Expected %d, got %d", expectedLength, len(updatedQ))
	}
	invalidPath := []string{"0", "wrong_field"}
	_, err = updateDnsOption(q, invalidPath, value, newBranch)
	if err == nil {
		t.Fatal("Expected error but got none")
	}
	if err != utils.ErrWrongPath {
		t.Errorf("Expected error %v but got %v", utils.ErrWrongPath, err)
	}
}

func TestLibDnsUpdateDnsOptionEDNS0DAU(t *testing.T) {
	tEDNS0_DAU := &dns.EDNS0_DAU{
		AlgCode: []uint8{},
	}
	q := []dns.EDNS0{
		tEDNS0_DAU,
	}
	path := []string{"0", utils.DNSDAU}
	value := "value"
	newBranch := false
	updatedQ, err := updateDnsOption(q, path, value, newBranch)
	if err != nil {
		t.Fatalf("updateDnsOption failed: %v", err)
	}
	expectedAlgCode := []uint8("value")
	if !reflect.DeepEqual(tEDNS0_DAU.AlgCode, expectedAlgCode) {
		t.Errorf("updateDnsOption did not update EDNS0_DAU correctly. Expected AlgCode %v, got %v", expectedAlgCode, tEDNS0_DAU.AlgCode)
	}
	expectedLength := 1
	if len(updatedQ) != expectedLength {
		t.Errorf("updateDnsOption did not return the expected number of elements. Expected %d, got %d", expectedLength, len(updatedQ))
	}
	invalidPath := []string{"0", "wrong_field"}
	_, err = updateDnsOption(q, invalidPath, value, newBranch)
	if err == nil {
		t.Fatal("Expected error but got none")
	}
	if err != utils.ErrWrongPath {
		t.Errorf("Expected error %v but got %v", utils.ErrWrongPath, err)
	}
}

func TestLibDnsUpdateDnsOptionEDNS0LLQ(t *testing.T) {
	tEDNS0_LLQ := &dns.EDNS0_LLQ{
		Version:   0,
		Opcode:    0,
		Error:     0,
		Id:        0,
		LeaseLife: 0,
	}
	q := []dns.EDNS0{
		tEDNS0_LLQ,
	}
	path := []string{"0"}
	value := int64(123)
	newBranch := false
	testCases := []struct {
		field          string
		expectedUpdate interface{}
	}{
		{utils.VersionName, uint16(123)},
		{utils.DNSOpcode, uint16(123)},
		{utils.Error, uint16(123)},
		{utils.DNSId, uint64(123)},
		{utils.DNSLeaseLife, uint32(123)},
	}
	for _, tc := range testCases {
		updatedQ, err := updateDnsOption(q, append(path, tc.field), value, newBranch)
		if err != nil {
			t.Fatalf("updateDnsOption failed for field %s: %v", tc.field, err)
		}
		switch tc.field {
		case utils.VersionName:
			if tEDNS0_LLQ.Version != tc.expectedUpdate.(uint16) {
				t.Errorf("updateDnsOption did not update Version correctly. Expected %v, got %v", tc.expectedUpdate, tEDNS0_LLQ.Version)
			}
		case utils.DNSOpcode:
			if tEDNS0_LLQ.Opcode != tc.expectedUpdate.(uint16) {
				t.Errorf("updateDnsOption did not update Opcode correctly. Expected %v, got %v", tc.expectedUpdate, tEDNS0_LLQ.Opcode)
			}
		case utils.Error:
			if tEDNS0_LLQ.Error != tc.expectedUpdate.(uint16) {
				t.Errorf("updateDnsOption did not update Error correctly. Expected %v, got %v", tc.expectedUpdate, tEDNS0_LLQ.Error)
			}
		case utils.DNSId:
			if tEDNS0_LLQ.Id != tc.expectedUpdate.(uint64) {
				t.Errorf("updateDnsOption did not update Id correctly. Expected %v, got %v", tc.expectedUpdate, tEDNS0_LLQ.Id)
			}
		case utils.DNSLeaseLife:
			if tEDNS0_LLQ.LeaseLife != tc.expectedUpdate.(uint32) {
				t.Errorf("updateDnsOption did not update LeaseLife correctly. Expected %v, got %v", tc.expectedUpdate, tEDNS0_LLQ.LeaseLife)
			}
		default:
			t.Fatalf("Unexpected field: %s", tc.field)
		}
		expectedLength := 1
		if len(updatedQ) != expectedLength {
			t.Errorf("updateDnsOption did not return the expected number of elements. Expected %d, got %d", expectedLength, len(updatedQ))
		}
	}
	invalidPath := []string{"0", "wrong_field"}
	_, err := updateDnsOption(q, invalidPath, value, newBranch)
	if err == nil {
		t.Fatal("Expected error but got none")
	}
	if err != utils.ErrWrongPath {
		t.Errorf("Expected error %v but got %v", utils.ErrWrongPath, err)
	}
}

func TestLibDnsUpdateDnsOptionEDNS0UL(t *testing.T) {
	tEDNS0_UL := &dns.EDNS0_UL{
		Lease:    0,
		KeyLease: 0,
	}
	q := []dns.EDNS0{
		tEDNS0_UL,
	}
	path := []string{"0"}
	value := int64(123)
	newBranch := false
	testCases := []struct {
		field          string
		expectedUpdate interface{}
	}{
		{utils.DNSLease, uint32(123)},
		{utils.DNSKeyLease, uint32(123)},
	}
	for _, tc := range testCases {
		updatedQ, err := updateDnsOption(q, append(path, tc.field), value, newBranch)
		if err != nil {
			t.Fatalf("updateDnsOption failed for field %s: %v", tc.field, err)
		}
		switch tc.field {
		case utils.DNSLease:
			if tEDNS0_UL.Lease != tc.expectedUpdate.(uint32) {
				t.Errorf("updateDnsOption did not update Lease correctly. Expected %v, got %v", tc.expectedUpdate, tEDNS0_UL.Lease)
			}
		case utils.DNSKeyLease:
			if tEDNS0_UL.KeyLease != tc.expectedUpdate.(uint32) {
				t.Errorf("updateDnsOption did not update KeyLease correctly. Expected %v, got %v", tc.expectedUpdate, tEDNS0_UL.KeyLease)
			}
		default:
			t.Fatalf("Unexpected field: %s", tc.field)
		}
		expectedLength := 1
		if len(updatedQ) != expectedLength {
			t.Errorf("updateDnsOption did not return the expected number of elements. Expected %d, got %d", expectedLength, len(updatedQ))
		}
	}
	invalidPath := []string{"0", "wrong_field"}
	_, err := updateDnsOption(q, invalidPath, value, newBranch)
	if err == nil {
		t.Fatal("Expected error but got none")
	}
	if err != utils.ErrWrongPath {
		t.Errorf("Expected error %v but got %v", utils.ErrWrongPath, err)
	}
}

func TestUpdateDnsQuestionsNewBranchOrEmpty(t *testing.T) {
	var q []dns.Question
	path := []string{"testField"}
	value := "testValue"
	newBranch := true
	result, err := updateDnsQuestions(q, path, value, newBranch)
	if err == nil {
		t.Errorf("'WRONG_PATH'")
	}
	if len(result) == 1 {
		t.Errorf("expected result length to be 1, but got %d", len(result))
	}
	if len(result) > 0 && result[0].Name != value {
		t.Errorf("expected result name to be '%s', but got '%s'", value, result[0].Name)
	}
}

func TestUpdateDnsQuestionsIndexConversionError(t *testing.T) {
	q := []dns.Question{{Name: "initialName"}}
	path := []string{"invalidIndex", "fieldName"}
	value := "newValue"
	newBranch := false
	_, err := updateDnsQuestions(q, path, value, newBranch)
	if err == nil {
		t.Error("expected error but got nil")
	}
	if err == utils.ErrWrongPath {
		t.Errorf("expected error type 'utils.ErrWrongPath' but got '%v'", err)
	}
	if len(q) != 1 || q[0].Name != "initialName" {
		t.Error("expected q to remain unchanged, but it was modified")
	}
}

func TestUpdateDnsQuestionsIndexSpecified(t *testing.T) {
	q := []dns.Question{{Name: "initialName"}}
	path := []string{"0", utils.DNSName}
	value := "newName"
	newBranch := false
	updatedQ, err := updateDnsQuestions(q, path, value, newBranch)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(updatedQ) != 1 {
		t.Errorf("expected result length to be 1, but got %d", len(updatedQ))
	}
	if updatedQ[0].Name != "newName" {
		t.Errorf("expected Name to be 'newName', but got '%s'", updatedQ[0].Name)
	}
	q = []dns.Question{}
	path = []string{"1", utils.DNSName}
	value = "newName"
	newBranch = false
	_, err = updateDnsQuestions(q, path, value, newBranch)
	if err == nil {
		t.Error("expected error but got nil")
	}
	if err != utils.ErrWrongPath {
		t.Errorf("expected error type 'utils.ErrWrongPath' but got '%v'", err)
	}
	q = []dns.Question{}
	path = []string{"invalidIndex", utils.DNSName}
	value = "newName"
	newBranch = false
	_, err = updateDnsQuestions(q, path, value, newBranch)
	if err == nil {
		t.Error("expected error but got nil")
	}

}

func TestUpdateDnsOptionNilCase(t *testing.T) {
	q := []dns.EDNS0{
		&dns.EDNS0_NSID{},
		nil,
	}
	path := []string{"1", "someField"}
	_, err := updateDnsOption(q, path, "test_value", false)
	if err == nil {
		t.Errorf("Expected error for nil case, got nil")
	}
	expectedErrMsg := "unsupported dns option type <*dns.EDNS0>"
	if err.Error() == expectedErrMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedErrMsg, err.Error())
	}
}

func TestUpdateDnsOptionDNSSourceNetmask(t *testing.T) {
	q := []dns.EDNS0{
		&dns.EDNS0_SUBNET{},
	}
	path := []string{"0", utils.DNSSourceNetmask}
	updatedQ, err := updateDnsOption(q, path, int64(24), false)
	if err != nil {
		t.Errorf("updateDnsOption() error = %v, want nil", err)
	}
	if len(updatedQ) != 1 {
		t.Errorf("Expected q length to be 1, got %d", len(updatedQ))
	}
	subnet, ok := updatedQ[0].(*dns.EDNS0_SUBNET)
	if !ok {
		t.Error("Expected updated element to be *dns.EDNS0_SUBNET")
	}
	expectedNetmask := uint8(24)
	if subnet.SourceNetmask != expectedNetmask {
		t.Errorf("Expected SourceNetmask to be %d, got %d", expectedNetmask, subnet.SourceNetmask)
	}
}

func TestUpdateDnsOptionDNSSourceScope(t *testing.T) {
	q := []dns.EDNS0{
		&dns.EDNS0_SUBNET{},
	}
	path := []string{"0", utils.DNSSourceScope}
	updatedQ, err := updateDnsOption(q, path, int64(16), false)
	if err != nil {
		t.Errorf("updateDnsOption() error = %v, want nil", err)
	}
	if len(updatedQ) != 1 {
		t.Errorf("Expected q length to be 1, got %d", len(updatedQ))
	}
	subnet, ok := updatedQ[0].(*dns.EDNS0_SUBNET)
	if !ok {
		t.Error("Expected updated element to be *dns.EDNS0_SUBNET")
	}
	expectedScope := uint8(16)
	if subnet.SourceScope != expectedScope {
		t.Errorf("Expected SourceScope to be %d, got %d", expectedScope, subnet.SourceScope)
	}
}

func TestUpdateDnsOptionAddress(t *testing.T) {
	q := []dns.EDNS0{
		&dns.EDNS0_SUBNET{},
	}
	path := []string{"0", utils.Address}
	ipAddressStr := "192.168.1.1"
	updatedQ, err := updateDnsOption(q, path, ipAddressStr, false)
	if err != nil {
		t.Errorf("updateDnsOption() error = %v, want nil", err)
	}
	if len(updatedQ) != 1 {
		t.Errorf("Expected q length to be 1, got %d", len(updatedQ))
	}
	subnet, ok := updatedQ[0].(*dns.EDNS0_SUBNET)
	if !ok {
		t.Error("Expected updated element to be *dns.EDNS0_SUBNET")
	}
	expectedIP := net.ParseIP(ipAddressStr)
	if !subnet.Address.Equal(expectedIP) {
		t.Errorf("Expected Address to be %v, got %v", expectedIP, subnet.Address)
	}
}

func TestUpdateDnsOptionDefaultCase(t *testing.T) {
	q := []dns.EDNS0{
		&dns.EDNS0_SUBNET{},
	}
	path := []string{"0", "invalidField"}
	_, err := updateDnsOption(q, path, "test_value", false)
	if err == nil {
		t.Error("Expected error for default case, got nil")
	}
	if err != utils.ErrWrongPath {
		t.Errorf("Expected error %v, got %v", utils.ErrWrongPath, err)
	}
}

func TestUpdateDnsOptionEDNS0Cookie(t *testing.T) {
	q := []dns.EDNS0{
		&dns.EDNS0_COOKIE{},
	}
	path := []string{"0", utils.DNSCookie}
	cookieValue := "test_cookie_value"
	updatedQ, err := updateDnsOption(q, path, cookieValue, false)
	if err != nil {
		t.Errorf("updateDnsOption() error = %v, want nil", err)
	}
	if len(updatedQ) != 1 {
		t.Errorf("Expected q length to be 1, got %d", len(updatedQ))
	}
	cookie, ok := updatedQ[0].(*dns.EDNS0_COOKIE)
	if !ok {
		t.Error("Expected updated element to be *dns.EDNS0_COOKIE")
	}
	if cookie.Cookie != cookieValue {
		t.Errorf("Expected Cookie to be %s, got %s", cookieValue, cookie.Cookie)
	}
}

func TestUpdateDnsOptionEDNS0CookieWrongField(t *testing.T) {
	q := []dns.EDNS0{
		&dns.EDNS0_COOKIE{},
	}
	path := []string{"0", "wrongField"}
	cookieValue := "test_cookie_value"
	_, err := updateDnsOption(q, path, cookieValue, false)
	if err == nil {
		t.Error("Expected error for wrong field, got nil")
	}
	if err != utils.ErrWrongPath {
		t.Errorf("Expected error %v, got %v", utils.ErrWrongPath, err)
	}
}

func TestUpdateDnsQuestionsError(t *testing.T) {
	q3 := []dns.Question{{Name: "cgrates.com.", Qtype: dns.TypeA}}
	path3 := []string{"0", "Name", "extra"}
	value3 := "new_value"
	newBranch3 := false
	_, err3 := updateDnsQuestions(q3, path3, value3, newBranch3)
	if err3 == nil || err3 != utils.ErrWrongPath {
		t.Errorf("Expected error %v for path3, got %v", utils.ErrWrongPath, err3)
	}
}

func TestUpdateDnsQuestionsAppendQuestion(t *testing.T) {
	q := []dns.Question{{Name: "cgrates.com.", Qtype: dns.TypeA}}
	path := []string{"1", "Name"}
	value := "new_value"
	newBranch := false
	updatedQ, err := updateDnsQuestions(q, path, value, newBranch)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if len(updatedQ) != 2 {
		t.Errorf("Expected length of updatedQ to be 2, got %d", len(updatedQ))
	}
}

func TestUpdateDnsRRHeaderRrtypeConversion(t *testing.T) {
	v := &dns.RR_Header{}
	path := []string{utils.DNSRrtype}
	value := int64(28)
	err := updateDnsRRHeader(v, path, value)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if v.Rrtype != 28 {
		t.Errorf("Expected Rrtype to be 28, got %d", v.Rrtype)
	}
}

func TestLibDnsUpdateDnsAnswerARecord(t *testing.T) {
	q := []dns.RR{

		&dns.A{},
	}
	qType := dns.TypeA
	qName := "cgrates.com"
	path1 := []string{utils.DNSHdr, utils.DNSName}
	value1 := "new.cgrates.com"
	_, err1 := updateDnsAnswer(q, qType, qName, path1, value1, false)
	if err1 != nil {
		t.Errorf("Unexpected error for valid path: %v", err1)
	}
	updatedHeader := q[0].(*dns.A).Hdr
	if updatedHeader.Name != value1 {
		t.Errorf("Expected DNS header name to be %s, got %s", value1, updatedHeader.Name)
	}
	path2 := []string{utils.DNSA}
	value2 := "192.168.2.1"
	_, err2 := updateDnsAnswer(q, qType, qName, path2, value2, false)
	if err2 != nil {
		t.Errorf("Unexpected error for valid path: %v", err2)
	}
	updatedIP := q[0].(*dns.A).A.String()
	if updatedIP != value2 {
		t.Errorf("Expected IP address to be %s, got %s", value2, updatedIP)
	}
	path3 := []string{"invalid_path"}
	_, err3 := updateDnsAnswer(q, qType, qName, path3, value1, false)
	if err3 == nil {
		t.Errorf("Expected error for invalid path, got nil")
	}

}

func TestLibDnsUpdateDnsOptionCase1(t *testing.T) {
	var q []dns.EDNS0
	path := []string{}
	value := "value"
	newBranch := true
	updatedQ, err := updateDnsOption(q, path, value, newBranch)
	if err == nil {
		t.Errorf("Unexpected error for valid path: %v", err)
	}
	if len(updatedQ) == 1 {
		t.Errorf("Expected updatedQ length to be 1, got %d", len(updatedQ))
	}

}
