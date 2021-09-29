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
	"strings"
	"testing"

	"github.com/cgrates/cgrates/utils"
	"github.com/miekg/dns"
)

func TestAppendDNSAnswerTypeNAPTR(t *testing.T) {
	m := new(dns.Msg)
	m.SetQuestion("3.6.9.4.7.1.7.1.5.6.8.9.4.e164.arpa.", dns.TypeNAPTR)
	if err := appendDNSAnswer(m); err != nil {
		t.Error(err)
	}
	if len(m.Answer) != 1 {
		t.Fatalf("Unexpected number of Answers : %+v", len(m.Answer))
	} else if m.Answer[0].Header().Name != "3.6.9.4.7.1.7.1.5.6.8.9.4.e164.arpa." {
		t.Errorf("expecting: <3.6.9.4.7.1.7.1.5.6.8.9.4.e164.arpa.>, received: <%+v>", m.Answer[0].Header().Name)
	} else if m.Answer[0].Header().Rrtype != 35 {
		t.Errorf("expecting: <35>, received: <%+v>", m.Answer[0].Header().Rrtype)
	} else if m.Answer[0].Header().Class != dns.ClassINET {
		t.Errorf("expecting: <%+v>, received: <%+v>", dns.ClassINET, m.Answer[0].Header().Rrtype)
	} else if m.Answer[0].Header().Ttl != 60 {
		t.Errorf("expecting: <60>, received: <%+v>", m.Answer[0].Header().Rrtype)
	}
}

func TestAppendDNSAnswerTypeA(t *testing.T) {
	m := new(dns.Msg)
	m.SetQuestion("3.6.9.4.7.1.7.1.5.6.8.9.4.e164.arpa.", dns.TypeA)
	if err := appendDNSAnswer(m); err != nil {
		t.Error(err)
	}
	if len(m.Answer) != 1 {
		t.Fatalf("Unexpected number of Answers : %+v", len(m.Answer))
	} else if m.Answer[0].Header().Name != "3.6.9.4.7.1.7.1.5.6.8.9.4.e164.arpa." {
		t.Errorf("expecting: <3.6.9.4.7.1.7.1.5.6.8.9.4.e164.arpa.>, received: <%+v>", m.Answer[0].Header().Name)
	} else if m.Answer[0].Header().Rrtype != 1 {
		t.Errorf("expecting: <1>, received: <%+v>", m.Answer[0].Header().Rrtype)
	} else if m.Answer[0].Header().Class != dns.ClassINET {
		t.Errorf("expecting: <%+v>, received: <%+v>", dns.ClassINET, m.Answer[0].Header().Rrtype)
	} else if m.Answer[0].Header().Ttl != 60 {
		t.Errorf("expecting: <60>, received: <%+v>", m.Answer[0].Header().Rrtype)
	}
}

func TestAppendDNSAnswerUnexpectedType(t *testing.T) {
	m := new(dns.Msg)
	m.SetQuestion("3.6.9.4.7.1.7.1.5.6.8.9.4.e164.arpa.", dns.TypeAFSDB)
	if err := appendDNSAnswer(m); err == nil || err.Error() != "unsupported DNS type: <18>" {
		t.Error(err)
	}
}

func TestUpdateDNSMsgFromNM(t *testing.T) {
	m := new(dns.Msg)
	m.SetQuestion("3.6.9.4.7.1.7.1.5.6.8.9.4.e164.arpa.", dns.TypeNAPTR)

	nM := utils.NewOrderedNavigableMap()
	path := []string{utils.Rcode}
	itm := &utils.DataNode{Type: utils.NMDataType, Value: &utils.DataLeaf{
		Data: 10,
	}}
	nM.SetAsSlice(&utils.FullPath{
		Path:      strings.Join(path, utils.NestingSep),
		PathSlice: path,
	}, []*utils.DataNode{itm})
	if err := updateDNSMsgFromNM(m, nM); err != nil {
		t.Fatal(err)
	}
	if m.Rcode != 10 {
		t.Errorf("expecting: <10>, received: <%+v>", m.Rcode)
	}

	nM = utils.NewOrderedNavigableMap()
	path = []string{utils.Rcode}
	itm = &utils.DataNode{Type: utils.NMDataType, Value: &utils.DataLeaf{
		Data: "RandomValue",
	}}
	nM.SetAsSlice(&utils.FullPath{
		Path:      strings.Join(path, utils.NestingSep),
		PathSlice: path,
	}, []*utils.DataNode{itm})
	if err := updateDNSMsgFromNM(m, nM); err == nil ||
		err.Error() != `item: <Rcode>, err: strconv.ParseInt: parsing "RandomValue": invalid syntax` {
		t.Error(err)
	}

	nM = utils.NewOrderedNavigableMap()
	path = []string{utils.Answer, utils.Order}
	itm = &utils.DataNode{Type: utils.NMDataType, Value: &utils.DataLeaf{
		Data: "RandomValue",
	}}
	nM.SetAsSlice(&utils.FullPath{
		Path:      strings.Join(path, utils.NestingSep),
		PathSlice: path,
	}, []*utils.DataNode{itm})
	if err := updateDNSMsgFromNM(m, nM); err == nil ||
		err.Error() != `item: <[Answer Order]>, err: strconv.ParseInt: parsing "RandomValue": invalid syntax` {
		t.Error(err)
	}

	nM = utils.NewOrderedNavigableMap()
	path = []string{utils.Answer, utils.Preference}
	itm = &utils.DataNode{Type: utils.NMDataType, Value: &utils.DataLeaf{
		Data: "RandomValue",
	}}
	nM.SetAsSlice(&utils.FullPath{
		Path:      strings.Join(path, utils.NestingSep),
		PathSlice: path,
	}, []*utils.DataNode{itm})
	if err := updateDNSMsgFromNM(m, nM); err == nil ||
		err.Error() != `item: <[Answer Preference]>, err: strconv.ParseInt: parsing "RandomValue": invalid syntax` {
		t.Error(err)
	}

	m = new(dns.Msg)
	m.SetQuestion("3.6.9.4.7.1.7.1.5.6.8.9.4.e164.arpa.", dns.TypeA)
	nM = utils.NewOrderedNavigableMap()
	path = []string{utils.Answer, utils.Order}
	itm = &utils.DataNode{Type: utils.NMDataType, Value: &utils.DataLeaf{
		Data: 10,
	}}
	nM.SetAsSlice(&utils.FullPath{
		Path:      strings.Join(path, utils.NestingSep),
		PathSlice: path,
	}, []*utils.DataNode{itm})
	if err := updateDNSMsgFromNM(m, nM); err == nil ||
		err.Error() != `item: <[Answer Order]>, err: unsuported dns option type <*dns.A>` {
		t.Error(err)
	}

	nM = utils.NewOrderedNavigableMap()
	path = []string{utils.Answer, utils.Preference}
	itm = &utils.DataNode{Type: utils.NMDataType, Value: &utils.DataLeaf{
		Data: 10,
	}}
	nM.SetAsSlice(&utils.FullPath{
		Path:      strings.Join(path, utils.NestingSep),
		PathSlice: path,
	}, []*utils.DataNode{itm})
	if err := updateDNSMsgFromNM(m, nM); err == nil ||
		err.Error() != `item: <[Answer Preference]>, err: unsuported dns option type <*dns.A>` {
		t.Error(err)
	}

	nM = utils.NewOrderedNavigableMap()
	path = []string{utils.Answer, utils.Flags}
	itm = &utils.DataNode{Type: utils.NMDataType, Value: &utils.DataLeaf{
		Data: 10,
	}}
	nM.SetAsSlice(&utils.FullPath{
		Path:      strings.Join(path, utils.NestingSep),
		PathSlice: path,
	}, []*utils.DataNode{itm})
	if err := updateDNSMsgFromNM(m, nM); err == nil ||
		err.Error() != `item: <[Answer Flags]>, err: unsuported dns option type <*dns.A>` {
		t.Error(err)
	}

	nM = utils.NewOrderedNavigableMap()
	path = []string{utils.Answer, utils.Service}
	itm = &utils.DataNode{Type: utils.NMDataType, Value: &utils.DataLeaf{
		Data: 10,
	}}
	nM.SetAsSlice(&utils.FullPath{
		Path:      strings.Join(path, utils.NestingSep),
		PathSlice: path,
	}, []*utils.DataNode{itm})
	if err := updateDNSMsgFromNM(m, nM); err == nil ||
		err.Error() != `item: <[Answer Service]>, err: unsuported dns option type <*dns.A>` {
		t.Error(err)
	}

	nM = utils.NewOrderedNavigableMap()
	path = []string{utils.Answer, utils.Regexp}
	itm = &utils.DataNode{Type: utils.NMDataType, Value: &utils.DataLeaf{
		Data: 10,
	}}
	nM.SetAsSlice(&utils.FullPath{
		Path:      strings.Join(path, utils.NestingSep),
		PathSlice: path,
	}, []*utils.DataNode{itm})
	if err := updateDNSMsgFromNM(m, nM); err == nil ||
		err.Error() != `item: <[Answer Regexp]>, err: unsuported dns option type <*dns.A>` {
		t.Error(err)
	}

	nM = utils.NewOrderedNavigableMap()
	path = []string{utils.Answer, utils.Replacement}
	itm = &utils.DataNode{Type: utils.NMDataType, Value: &utils.DataLeaf{
		Data: 10,
	}}
	nM.SetAsSlice(&utils.FullPath{
		Path:      strings.Join(path, utils.NestingSep),
		PathSlice: path,
	}, []*utils.DataNode{itm})
	if err := updateDNSMsgFromNM(m, nM); err == nil ||
		err.Error() != `item: <[Answer Replacement]>, err: unsuported dns option type <*dns.A>` {
		t.Error(err)
	}

}
