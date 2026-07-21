// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package agents

import (
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
