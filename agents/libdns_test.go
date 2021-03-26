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
	"strings"
	"testing"

	"github.com/cgrates/cgrates/utils"
	"github.com/miekg/dns"
)

func TestE164FromNAPTR(t *testing.T) {
	if e164, err := e164FromNAPTR("8.7.6.5.4.3.2.1.0.1.6.e164.arpa."); err != nil {
		t.Error(err)
	} else if e164 != "61012345678" {
		t.Errorf("received: <%s>", e164)
	}
}

func TestDomainNameFromNAPTR(t *testing.T) {
	if dName := domainNameFromNAPTR("8.7.6.5.4.3.2.1.0.1.6.e164.arpa."); dName != "e164.arpa" {
		t.Errorf("received: <%s>", dName)
	}
	if dName := domainNameFromNAPTR("8.7.6.5.4.3.2.1.0.1.6.e164.itsyscom.com."); dName != "e164.itsyscom.com" {
		t.Errorf("received: <%s>", dName)
	}
	if dName := domainNameFromNAPTR("8.7.6.5.4.3.2.1.0.1.6.itsyscom.com."); dName != "8.7.6.5.4.3.2.1.0.1.6.itsyscom.com" {
		t.Errorf("received: <%s>", dName)
	}
}

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

func TestDNSDPFieldAsInterface(t *testing.T) {
	m := new(dns.Msg)
	m.SetQuestion("3.6.9.4.7.1.7.1.5.6.8.9.4.e164.arpa.", dns.TypeNAPTR)
	dp := newDNSDataProvider(m, nil)
	expected := m.Question[0]
	if data, err := dp.FieldAsInterface([]string{"test"}); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(data, expected) {
		t.Errorf("expecting: <%+v>, received: <%+v>", expected, data)
	}
}

func TestDNSDPFieldAsInterfaceEmptyPath(t *testing.T) {
	m := new(dns.Msg)
	m.SetQuestion("3.6.9.4.7.1.7.1.5.6.8.9.4.e164.arpa.", dns.TypeNAPTR)
	dp := newDNSDataProvider(m, nil)
	if _, err := dp.FieldAsInterface([]string{}); err == nil ||
		err.Error() != "empty field path" {
		t.Error(err)
	}
}

func TestDNSDPFieldAsInterfaceFromCache(t *testing.T) {
	m := new(dns.Msg)
	m.SetQuestion("3.6.9.4.7.1.7.1.5.6.8.9.4.e164.arpa.", dns.TypeNAPTR)
	dp := newDNSDataProvider(m, nil)
	expected := m.Question[0]
	if data, err := dp.FieldAsInterface([]string{"test"}); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(data, expected) {
		t.Errorf("expecting: <%+v>, received: <%+v>", expected, data)
	}
	if data, err := dp.FieldAsInterface([]string{"test"}); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(data, expected) {
		t.Errorf("expecting: <%+v>, received: <%+v>", expected, data)
	}
}

func TestDNSDPFieldAsString(t *testing.T) {
	m := new(dns.Msg)
	m.SetQuestion("3.6.9.4.7.1.7.1.5.6.8.9.4.e164.arpa.", dns.TypeNAPTR)
	dp := newDNSDataProvider(m, nil)
	expected := utils.ToJSON(m.Question[0])
	if data, err := dp.FieldAsString([]string{"test"}); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(data, expected) {
		t.Errorf("expecting: <%+v>, received: <%+v>", expected, data)
	}
}

func TestDNSDPFieldAsStringEmptyPath(t *testing.T) {
	m := new(dns.Msg)
	m.SetQuestion("3.6.9.4.7.1.7.1.5.6.8.9.4.e164.arpa.", dns.TypeNAPTR)
	dp := newDNSDataProvider(m, nil)
	if _, err := dp.FieldAsString([]string{}); err == nil ||
		err.Error() != "empty field path" {
		t.Error(err)
	}
}

func TestDNSDPFieldAsStringFromCache(t *testing.T) {
	m := new(dns.Msg)
	m.SetQuestion("3.6.9.4.7.1.7.1.5.6.8.9.4.e164.arpa.", dns.TypeNAPTR)
	dp := newDNSDataProvider(m, nil)
	expected := utils.ToJSON(m.Question[0])
	if data, err := dp.FieldAsString([]string{"test"}); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(data, expected) {
		t.Errorf("expecting: <%+v>, received: <%+v>", expected, data)
	}
	if data, err := dp.FieldAsString([]string{"test"}); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(data, expected) {
		t.Errorf("expecting: <%+v>, received: <%+v>", expected, data)
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
	path = []string{utils.Order}
	itm = &utils.DataNode{Type: utils.NMDataType, Value: &utils.DataLeaf{
		Data: "RandomValue",
	}}
	nM.SetAsSlice(&utils.FullPath{
		Path:      strings.Join(path, utils.NestingSep),
		PathSlice: path,
	}, []*utils.DataNode{itm})
	if err := updateDNSMsgFromNM(m, nM); err == nil ||
		err.Error() != `item: <Order>, err: strconv.ParseInt: parsing "RandomValue": invalid syntax` {
		t.Error(err)
	}

	nM = utils.NewOrderedNavigableMap()
	path = []string{utils.Preference}
	itm = &utils.DataNode{Type: utils.NMDataType, Value: &utils.DataLeaf{
		Data: "RandomValue",
	}}
	nM.SetAsSlice(&utils.FullPath{
		Path:      strings.Join(path, utils.NestingSep),
		PathSlice: path,
	}, []*utils.DataNode{itm})
	if err := updateDNSMsgFromNM(m, nM); err == nil ||
		err.Error() != `item: <Preference>, err: strconv.ParseInt: parsing "RandomValue": invalid syntax` {
		t.Error(err)
	}

	m = new(dns.Msg)
	m.SetQuestion("3.6.9.4.7.1.7.1.5.6.8.9.4.e164.arpa.", dns.TypeA)
	nM = utils.NewOrderedNavigableMap()
	path = []string{utils.Order}
	itm = &utils.DataNode{Type: utils.NMDataType, Value: &utils.DataLeaf{
		Data: 10,
	}}
	nM.SetAsSlice(&utils.FullPath{
		Path:      strings.Join(path, utils.NestingSep),
		PathSlice: path,
	}, []*utils.DataNode{itm})
	if err := updateDNSMsgFromNM(m, nM); err == nil ||
		err.Error() != `field <Order> only works with NAPTR` {
		t.Error(err)
	}

	nM = utils.NewOrderedNavigableMap()
	path = []string{utils.Preference}
	itm = &utils.DataNode{Type: utils.NMDataType, Value: &utils.DataLeaf{
		Data: 10,
	}}
	nM.SetAsSlice(&utils.FullPath{
		Path:      strings.Join(path, utils.NestingSep),
		PathSlice: path,
	}, []*utils.DataNode{itm})
	if err := updateDNSMsgFromNM(m, nM); err == nil ||
		err.Error() != `field <Preference> only works with NAPTR` {
		t.Error(err)
	}

	nM = utils.NewOrderedNavigableMap()
	path = []string{utils.Flags}
	itm = &utils.DataNode{Type: utils.NMDataType, Value: &utils.DataLeaf{
		Data: 10,
	}}
	nM.SetAsSlice(&utils.FullPath{
		Path:      strings.Join(path, utils.NestingSep),
		PathSlice: path,
	}, []*utils.DataNode{itm})
	if err := updateDNSMsgFromNM(m, nM); err == nil ||
		err.Error() != `field <Flags> only works with NAPTR` {
		t.Error(err)
	}

	nM = utils.NewOrderedNavigableMap()
	path = []string{utils.Service}
	itm = &utils.DataNode{Type: utils.NMDataType, Value: &utils.DataLeaf{
		Data: 10,
	}}
	nM.SetAsSlice(&utils.FullPath{
		Path:      strings.Join(path, utils.NestingSep),
		PathSlice: path,
	}, []*utils.DataNode{itm})
	if err := updateDNSMsgFromNM(m, nM); err == nil ||
		err.Error() != `field <Service> only works with NAPTR` {
		t.Error(err)
	}

	nM = utils.NewOrderedNavigableMap()
	path = []string{utils.Regexp}
	itm = &utils.DataNode{Type: utils.NMDataType, Value: &utils.DataLeaf{
		Data: 10,
	}}
	nM.SetAsSlice(&utils.FullPath{
		Path:      strings.Join(path, utils.NestingSep),
		PathSlice: path,
	}, []*utils.DataNode{itm})
	if err := updateDNSMsgFromNM(m, nM); err == nil ||
		err.Error() != `field <Regexp> only works with NAPTR` {
		t.Error(err)
	}

	nM = utils.NewOrderedNavigableMap()
	path = []string{utils.Replacement}
	itm = &utils.DataNode{Type: utils.NMDataType, Value: &utils.DataLeaf{
		Data: 10,
	}}
	nM.SetAsSlice(&utils.FullPath{
		Path:      strings.Join(path, utils.NestingSep),
		PathSlice: path,
	}, []*utils.DataNode{itm})
	if err := updateDNSMsgFromNM(m, nM); err == nil ||
		err.Error() != `field <Replacement> only works with NAPTR` {
		t.Error(err)
	}

}
