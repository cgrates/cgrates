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
	"strconv"
	"strings"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"github.com/miekg/dns"
)

const (
	QueryType = "QueryType"
	QueryName = "QueryName"
	dnsOption = "Option"
)

// dnsWriteErr writes the error with code back to the client
func dnsWriteMsg(w dns.ResponseWriter, msg *dns.Msg) (err error) {
	if err = w.WriteMsg(msg); err != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<%s> error: <%s> when writing on connection",
				utils.DNSAgent, err.Error()))
	}
	return
}

// appendDNSAnswer will append the right answer payload to the message
func appendDNSAnswer(msg *dns.Msg) (err error) {
	switch msg.Question[0].Qtype {
	case dns.TypeA:
		msg.Answer = append(msg.Answer,
			&dns.A{
				Hdr: dns.RR_Header{
					Name:   msg.Question[0].Name,
					Rrtype: dns.TypeA,
					Class:  dns.ClassINET,
					Ttl:    60},
			},
		)
	case dns.TypeNAPTR:
		msg.Answer = append(msg.Answer,
			&dns.NAPTR{
				Hdr: dns.RR_Header{
					Name:   msg.Question[0].Name,
					Rrtype: msg.Question[0].Qtype,
					Class:  dns.ClassINET,
					Ttl:    60},
			},
		)
	default:
		return fmt.Errorf("unsupported DNS type: <%v>", msg.Question[0].Qtype)
	}
	return
}

func newDnsDP(req *dns.Msg) utils.DataProvider {
	var opts interface{}
	if o := req.IsEdns0(); o != nil {
		opts = o
	}
	return &dnsDP{
		req:  config.NewObjectDP(req),
		opts: config.NewObjectDP(opts),
	}
}

type dnsDP struct {
	req  utils.DataProvider
	opts utils.DataProvider
}

func (dp dnsDP) String() string { return dp.req.String() }
func (dp dnsDP) FieldAsInterface(fldPath []string) (interface{}, error) {
	if len(fldPath) != 0 && fldPath[0] == dnsOption {
		return dp.opts.FieldAsInterface(fldPath[1:])
	}
	return dp.req.FieldAsInterface(fldPath)
}
func (dp dnsDP) FieldAsString(fldPath []string) (string, error) {
	valIface, err := dp.FieldAsInterface(fldPath)
	if err != nil {
		return "", err
	}
	return utils.IfaceAsString(valIface), nil
}

func updateDNSMsgFromNM(msg *dns.Msg, nm *utils.OrderedNavigableMap) (err error) {
	msgFields := make(utils.StringSet) // work around to NMap issue
	for el := nm.GetFirstElement(); el != nil; el = el.Next() {
		path := el.Value
		itm, _ := nm.Field(path)
		switch path[0] { // go for each posible field
		case utils.Id:
			var vItm int64
			if vItm, err = utils.IfaceAsTInt64(itm.Data); err != nil {
				return fmt.Errorf("item: <%s>, err: %s", path[0], err.Error())
			}
			msg.Id = uint16(vItm)
		case utils.Response:
			var vItm bool
			if vItm, err = utils.IfaceAsBool(itm.Data); err != nil {
				return fmt.Errorf("item: <%s>, err: %s", path[0], err.Error())
			}
			msg.Response = vItm
		case utils.Opcode:
			var vItm int64
			if vItm, err = utils.IfaceAsTInt64(itm.Data); err != nil {
				return fmt.Errorf("item: <%s>, err: %s", path[0], err.Error())
			}
			msg.Opcode = int(vItm)
		case utils.Authoritative:
			var vItm bool
			if vItm, err = utils.IfaceAsBool(itm.Data); err != nil {
				return fmt.Errorf("item: <%s>, err: %s", path[0], err.Error())
			}
			msg.Authoritative = vItm
		case utils.Truncated:
			var vItm bool
			if vItm, err = utils.IfaceAsBool(itm.Data); err != nil {
				return fmt.Errorf("item: <%s>, err: %s", path[0], err.Error())
			}
			msg.Truncated = vItm
		case utils.RecursionDesired:
			var vItm bool
			if vItm, err = utils.IfaceAsBool(itm.Data); err != nil {
				return fmt.Errorf("item: <%s>, err: %s", path[0], err.Error())
			}
			msg.RecursionDesired = vItm
		case utils.RecursionAvailable:
			var vItm bool
			if vItm, err = utils.IfaceAsBool(itm.Data); err != nil {
				return fmt.Errorf("item: <%s>, err: %s", path[0], err.Error())
			}
			msg.RecursionAvailable = vItm
		case utils.Zero:
			var vItm bool
			if vItm, err = utils.IfaceAsBool(itm.Data); err != nil {
				return fmt.Errorf("item: <%s>, err: %s", path[0], err.Error())
			}
			msg.Zero = vItm
		case utils.AuthenticatedData:
			var vItm bool
			if vItm, err = utils.IfaceAsBool(itm.Data); err != nil {
				return fmt.Errorf("item: <%s>, err: %s", path[0], err.Error())
			}
			msg.AuthenticatedData = vItm
		case utils.CheckingDisabled:
			var vItm bool
			if vItm, err = utils.IfaceAsBool(itm.Data); err != nil {
				return fmt.Errorf("item: <%s>, err: %s", path[0], err.Error())
			}
			msg.CheckingDisabled = vItm
		case utils.Rcode:
			var vItm int64
			if vItm, err = utils.IfaceAsTInt64(itm.Data); err != nil {
				return fmt.Errorf("item: <%s>, err: %s", path[0], err.Error())
			}
			msg.Rcode = int(vItm)
		case utils.Question:
			if msg.Question, err = updateDnsQuestions(msg.Question, path[1:len(path)-1], itm.Data, itm.NewBranch); err != nil {
				return fmt.Errorf("item: <%s>, err: %s", path[:len(path)-1], err.Error())
			}
		case utils.Answer:
			newBranch := itm.NewBranch ||
				len(msg.Answer) == 0 ||
				msgFields.Has(path[0])
			if newBranch { // force append if the same path was already used
				msgFields = make(utils.StringSet)      // reset the fields inside since we have a new message
				msgFields.Add(strings.Join(path, ".")) // detect new branch
			}
			if msg.Answer, err = updateDnsAnswer(msg.Answer, msg.Question[0].Qtype, msg.Question[0].Name, path[1:len(path)-1], itm.Data, newBranch); err != nil {
				return fmt.Errorf("item: <%s>, err: %s", path[:len(path)-1], err.Error())
			}
		case utils.Ns: //ToDO
		case utils.Extra: //ToDO
		case dnsOption:
			opts := msg.IsEdns0()
			if opts == nil {
				opts = msg.SetEdns0(4096, false).IsEdns0()
			}
			if opts.Option, err = updateDnsOption(opts.Option, path[1:len(path)-1], itm.Data, itm.NewBranch); err != nil {
				return fmt.Errorf("item: <%s>, err: %s", path[:len(path)-1], err.Error())
			}
		default:
		}

	}
	return
}

// updateDnsQuestion
func updateDnsQuestions(q []dns.Question, path []string, value interface{}, newBranch bool) (_ []dns.Question, err error) {
	var idx int
	var field string
	switch len(path) {
	case 1: // only the field so update the last one
		if newBranch || len(q) == 0 {
			q = append(q, dns.Question{})
		}
		idx = len(q) - 1
		field = path[0]
	case 2: // the index is specified
		if idx, err = strconv.Atoi(path[0]); err != nil {
			return
		}
		if lq := len(q); idx > lq {
			err = utils.ErrWrongPath
			return
		} else if lq == idx {
			q = append(q, dns.Question{})
		}
		field = path[1]
	default:
		err = utils.ErrWrongPath
		return
	}
	switch field {
	case utils.Name:
		q[idx].Name = utils.IfaceAsString(value)
	case utils.Qtype:
		var vItm int64
		if vItm, err = utils.IfaceAsTInt64(value); err != nil {
			return
		}
		q[idx].Qtype = uint16(vItm)
	case utils.Qclass:
		var vItm int64
		if vItm, err = utils.IfaceAsTInt64(value); err != nil {
			return
		}
		q[idx].Qclass = uint16(vItm)
	default:
		err = utils.ErrWrongPath
		return
	}

	return q, nil
}

func updateDnsOption(q []dns.EDNS0, path []string, value interface{}, newBranch bool) (_ []dns.EDNS0, err error) {
	var idx int
	var field string
	switch len(path) {
	case 1: // only the field so update the last one
		field = path[0]
		if newBranch ||
			len(q) == 0 {
			var o dns.EDNS0
			if o, err = createDnsOption(field, value); err != nil {
				return
			}
			return append(q, o), nil
		}
		idx = len(q) - 1
	case 2: // the index is specified
		field = path[1]
		if idx, err = strconv.Atoi(path[0]); err != nil {
			return
		}
		if lq := len(q); idx > lq {
			err = utils.ErrWrongPath
			return
		} else if lq == idx {
			var o dns.EDNS0
			if o, err = createDnsOption(field, value); err != nil {
				return
			}
			return append(q, o), nil
		}
	default:
		err = utils.ErrWrongPath
		return
	}
	switch v := q[idx].(type) {
	case *dns.EDNS0_NSID:
		if field != "Nsid" {
			err = utils.ErrWrongPath
			return
		}
		v.Nsid = utils.IfaceAsString(value)
	case *dns.EDNS0_SUBNET:
		switch field {
		case "Family":
			var vItm int64
			if vItm, err = utils.IfaceAsTInt64(value); err != nil {
				return
			}
			v.Family = uint16(vItm)
		case "SourceNetmask":
			var vItm int64
			if vItm, err = utils.IfaceAsTInt64(value); err != nil {
				return
			}
			v.SourceNetmask = uint8(vItm)
		case "SourceScope":
			var vItm int64
			if vItm, err = utils.IfaceAsTInt64(value); err != nil {
				return
			}
			v.SourceScope = uint8(vItm)
		case "Address":
			v.Address = net.ParseIP(utils.IfaceAsString(value))
		default:
			err = utils.ErrWrongPath
			return
		}
	case *dns.EDNS0_COOKIE:
		if field != "Cookie" {
			err = utils.ErrWrongPath
			return
		}
		v.Cookie = utils.IfaceAsString(value)
	case *dns.EDNS0_UL:
		switch field {
		case "Lease":
			var vItm int64
			if vItm, err = utils.IfaceAsTInt64(value); err != nil {
				return
			}
			v.Lease = uint32(vItm)
		case "KeyLease":
			var vItm int64
			if vItm, err = utils.IfaceAsTInt64(value); err != nil {
				return
			}
			v.KeyLease = uint32(vItm)
		default:
			err = utils.ErrWrongPath
			return
		}
	case *dns.EDNS0_LLQ:
		switch field {
		case "Version":
			var vItm int64
			if vItm, err = utils.IfaceAsTInt64(value); err != nil {
				return
			}
			v.Version = uint16(vItm)
		case "Opcode":
			var vItm int64
			if vItm, err = utils.IfaceAsTInt64(value); err != nil {
				return
			}
			v.Opcode = uint16(vItm)
		case "Error":
			var vItm int64
			if vItm, err = utils.IfaceAsTInt64(value); err != nil {
				return
			}
			v.Error = uint16(vItm)
		case "Id":
			var vItm int64
			if vItm, err = utils.IfaceAsTInt64(value); err != nil {
				return
			}
			v.Id = uint64(vItm)
		case "LeaseLife":
			var vItm int64
			if vItm, err = utils.IfaceAsTInt64(value); err != nil {
				return
			}
			v.LeaseLife = uint32(vItm)
		default:
			err = utils.ErrWrongPath
			return
		}
	case *dns.EDNS0_DAU:
		if field != "DAU" {
			err = utils.ErrWrongPath
			return
		}
		v.AlgCode = []uint8(utils.IfaceAsString(value))
	case *dns.EDNS0_DHU:
		if field != "DHU" {
			err = utils.ErrWrongPath
			return
		}
		v.AlgCode = []uint8(utils.IfaceAsString(value))
	case *dns.EDNS0_N3U:
		if field != "N3U" {
			err = utils.ErrWrongPath
			return
		}
		v.AlgCode = []uint8(utils.IfaceAsString(value))
	case *dns.EDNS0_EXPIRE:
		if field != "Expire" {
			err = utils.ErrWrongPath
			return
		}
		var vItm int64
		if vItm, err = utils.IfaceAsTInt64(value); err != nil {
			return
		}
		v.Expire = uint32(vItm)
	case *dns.EDNS0_TCP_KEEPALIVE:
		switch field {
		case "Length": //
			var vItm int64
			if vItm, err = utils.IfaceAsTInt64(value); err != nil {
				return
			}
			v.Length = uint16(vItm)
		case "Timeout": //
			var vItm int64
			if vItm, err = utils.IfaceAsTInt64(value); err != nil {
				return
			}
			v.Timeout = uint16(vItm)
		default:
			err = utils.ErrWrongPath
			return
		}
	case *dns.EDNS0_PADDING:
		if field != "Padding" {
			err = utils.ErrWrongPath
			return
		}
		v.Padding = []byte(utils.IfaceAsString(value))
	case *dns.EDNS0_EDE:
		switch field {
		case "InfoCode":
			var vItm int64
			if vItm, err = utils.IfaceAsTInt64(value); err != nil {
				return
			}
			v.InfoCode = uint16(vItm)
		case "ExtraText":
			v.ExtraText = utils.IfaceAsString(value)
		default:
			err = utils.ErrWrongPath
			return
		}
	case *dns.EDNS0_ESU:
		if field != "Uri" {
			err = utils.ErrWrongPath
			return
		}
		v.Uri = utils.IfaceAsString(value)
	case *dns.EDNS0_LOCAL: // if already there you can change data
		if field != "Data" {
			err = utils.ErrWrongPath
			return
		}
		v.Data = []byte(utils.IfaceAsString(value))
	case nil:
		err = fmt.Errorf("unsuported dns option type <%T>", v)
	default:
		err = fmt.Errorf("unsuported dns option type <%T>", v)
	}
	return q, err
}

func createDnsOption(field string, value interface{}) (o dns.EDNS0, err error) {
	switch field {
	case "Nsid": // EDNS0_NSID
		o = &dns.EDNS0_NSID{Nsid: utils.IfaceAsString(value)}
	case "Family": // EDNS0_SUBNET
		var vItm int64
		if vItm, err = utils.IfaceAsTInt64(value); err != nil {
			return
		}
		o = &dns.EDNS0_SUBNET{Family: uint16(vItm)}
	case "SourceNetmask": // EDNS0_SUBNET
		var vItm int64
		if vItm, err = utils.IfaceAsTInt64(value); err != nil {
			return
		}
		o = &dns.EDNS0_SUBNET{SourceNetmask: uint8(vItm)}
	case "SourceScope": // EDNS0_SUBNET
		var vItm int64
		if vItm, err = utils.IfaceAsTInt64(value); err != nil {
			return
		}
		o = &dns.EDNS0_SUBNET{SourceScope: uint8(vItm)}
	case "Address": // EDNS0_SUBNET
		o = &dns.EDNS0_SUBNET{Address: net.ParseIP(utils.IfaceAsString(value))}
	case "Cookie": // EDNS0_COOKIE
		o = &dns.EDNS0_COOKIE{Cookie: utils.IfaceAsString(value)}
	case "Lease": // EDNS0_UL
		var vItm int64
		if vItm, err = utils.IfaceAsTInt64(value); err != nil {
			return
		}
		o = &dns.EDNS0_UL{Lease: uint32(vItm)}
	case "KeyLease": // EDNS0_UL
		var vItm int64
		if vItm, err = utils.IfaceAsTInt64(value); err != nil {
			return
		}
		o = &dns.EDNS0_UL{KeyLease: uint32(vItm)}
	case "Version": // EDNS0_LLQ
		var vItm int64
		if vItm, err = utils.IfaceAsTInt64(value); err != nil {
			return
		}
		o = &dns.EDNS0_LLQ{Version: uint16(vItm)}
	case "Opcode": // EDNS0_LLQ
		var vItm int64
		if vItm, err = utils.IfaceAsTInt64(value); err != nil {
			return
		}
		o = &dns.EDNS0_LLQ{Opcode: uint16(vItm)}
	case "Error": // EDNS0_LLQ
		var vItm int64
		if vItm, err = utils.IfaceAsTInt64(value); err != nil {
			return
		}
		o = &dns.EDNS0_LLQ{Error: uint16(vItm)}
	case "Id": // EDNS0_LLQ
		var vItm int64
		if vItm, err = utils.IfaceAsTInt64(value); err != nil {
			return
		}
		o = &dns.EDNS0_LLQ{Id: uint64(vItm)}
	case "LeaseLife": // EDNS0_LLQ
		var vItm int64
		if vItm, err = utils.IfaceAsTInt64(value); err != nil {
			return
		}
		o = &dns.EDNS0_LLQ{LeaseLife: uint32(vItm)}
	case "DAU": // EDNS0_DAU
		o = &dns.EDNS0_DAU{AlgCode: []uint8(utils.IfaceAsString(value))}
	case "DHU": // EDNS0_DHU
		o = &dns.EDNS0_DHU{AlgCode: []uint8(utils.IfaceAsString(value))}
	case "N3U": // EDNS0_N3U
		o = &dns.EDNS0_N3U{AlgCode: []uint8(utils.IfaceAsString(value))}
	case "Expire": // EDNS0_EXPIRE
		var vItm int64
		if vItm, err = utils.IfaceAsTInt64(value); err != nil {
			return
		}
		o = &dns.EDNS0_EXPIRE{Expire: uint32(vItm)}
	case "Length": // EDNS0_TCP_KEEPALIVE
		var vItm int64
		if vItm, err = utils.IfaceAsTInt64(value); err != nil {
			return
		}
		o = &dns.EDNS0_TCP_KEEPALIVE{Length: uint16(vItm)}
	case "Timeout": // EDNS0_TCP_KEEPALIVE
		var vItm int64
		if vItm, err = utils.IfaceAsTInt64(value); err != nil {
			return
		}
		o = &dns.EDNS0_TCP_KEEPALIVE{Timeout: uint16(vItm)}
	case "Padding": // EDNS0_PADDING
		o = &dns.EDNS0_PADDING{Padding: []byte(utils.IfaceAsString(value))}
	case "InfoCode": // EDNS0_EDE
		var vItm int64
		if vItm, err = utils.IfaceAsTInt64(value); err != nil {
			return
		}
		o = &dns.EDNS0_EDE{InfoCode: uint16(vItm)}
	case "ExtraText": // EDNS0_EDE
		o = &dns.EDNS0_EDE{ExtraText: utils.IfaceAsString(value)}
	case "Uri": // EDNS0_ESU
		o = &dns.EDNS0_ESU{Uri: utils.IfaceAsString(value)}
	default:
		err = fmt.Errorf("can not create option from field <%q>", field)
	}
	return
}

func updateDnsAnswer(q []dns.RR, qType uint16, qName string, path []string, value interface{}, newBranch bool) (_ []dns.RR, err error) {
	var idx int
	switch len(path) {
	case 1: // only the field so update the last one
		if newBranch || len(q) == 0 {
			var a dns.RR
			if a, err = newDNSAnswer(qType, qName); err != nil {
				return
			}
			q = append(q, a)
		}
		idx = len(q) - 1
	case 2: // the index is specified
		if idx, err = strconv.Atoi(path[0]); err != nil {
			return
		}
		if lq := len(q); idx > lq {
			err = utils.ErrWrongPath
			return
		} else if lq == idx {
			var a dns.RR
			if a, err = newDNSAnswer(qType, qName); err != nil {
				return
			}
			q = append(q, a)
		}
		path = path[1:]
	default:
		err = utils.ErrWrongPath
		return
	}
	switch v := q[idx].(type) {
	case *dns.NAPTR:
		err = updateDnsNAPTRAnswer(v, path, value)
	case nil:
		err = fmt.Errorf("unsuported dns option type <%T>", v)
	default:
		err = fmt.Errorf("unsuported dns option type <%T>", v)
	}

	return q, err
}

// appendDNSAnswer will append the right answer payload to the message
func newDNSAnswer(qType uint16, qName string) (a dns.RR, err error) {
	hdr := dns.RR_Header{
		Name:   qName,
		Rrtype: qType,
		Class:  dns.ClassINET,
		Ttl:    60,
	}
	switch qType {
	case dns.TypeA:
		a = &dns.A{Hdr: hdr}
	case dns.TypeNAPTR:
		a = &dns.NAPTR{Hdr: hdr}
	default:
		err = fmt.Errorf("unsupported DNS type: <%v>", qType)
	}
	return
}

func updateDnsNAPTRAnswer(v *dns.NAPTR, path []string, value interface{}) (err error) {
	if len(path) < 1 ||
		(path[0] != "Hdr" && len(path) != 1) ||
		(path[0] == "Hdr" && len(path) != 2) {
		return utils.ErrWrongPath
	}
	switch path[0] {
	case "Hdr":
		return updateDnsRRHeader(&v.Hdr, path[1:], value)
	case "Order":
		var vItm int64
		if vItm, err = utils.IfaceAsTInt64(value); err != nil {
			return
		}
		v.Order = uint16(vItm)
	case "Preference":
		var vItm int64
		if vItm, err = utils.IfaceAsTInt64(value); err != nil {
			return
		}
		v.Preference = uint16(vItm)
	case "Flags":
		v.Flags = utils.IfaceAsString(value)
	case "Service":
		v.Service = utils.IfaceAsString(value)
	case "Regexp":
		v.Regexp = utils.IfaceAsString(value)
	case "Replacement":
		v.Replacement = utils.IfaceAsString(value)
	default:
		return utils.ErrWrongPath
	}
	return
}

func updateDnsRRHeader(v *dns.RR_Header, path []string, value interface{}) (err error) {
	if len(path) != 1 {
		return utils.ErrWrongPath
	}
	switch path[0] {
	case "Name":
		v.Name = utils.IfaceAsString(value)
	case "Rrtype":
		var vItm int64
		if vItm, err = utils.IfaceAsTInt64(value); err != nil {
			return
		}
		v.Rrtype = uint16(vItm)
	case "Class":
		var vItm int64
		if vItm, err = utils.IfaceAsTInt64(value); err != nil {
			return
		}
		v.Class = uint16(vItm)
	case "Ttl":
		var vItm int64
		if vItm, err = utils.IfaceAsTInt64(value); err != nil {
			return
		}
		v.Ttl = uint32(vItm)
	case "Rdlength":
		var vItm int64
		if vItm, err = utils.IfaceAsTInt64(value); err != nil {
			return
		}
		v.Rdlength = uint16(vItm)
	default:
		return utils.ErrWrongPath
	}
	return
}
