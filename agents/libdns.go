/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
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

func newDnsReply(req *dns.Msg) (rply *dns.Msg) {
	rply = new(dns.Msg)
	rply.SetReply(req)
	if len(req.Question) > 0 {
		rply.Question = make([]dns.Question, len(req.Question))
		copy(rply.Question, req.Question)
	}
	if opts := rply.IsEdns0(); opts != nil {
		rply.SetEdns0(4096, false).IsEdns0().Option = opts.Option
	}
	return
}

// dnsWriteErr writes the error with code back to the client
func dnsWriteMsg(w dns.ResponseWriter, msg *dns.Msg) (err error) {
	if err = w.WriteMsg(msg); err != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<%s> error: <%s> when writing on connection",
				utils.DNSAgent, err.Error()))
	}
	return
}

func newDnsDP(req *dns.Msg) utils.DataProvider {
	return &dnsDP{
		req:  config.NewObjectDP(req),
		opts: config.NewObjectDP(req.IsEdns0()),
	}
}

type dnsDP struct {
	req  utils.DataProvider
	opts utils.DataProvider
}

func (dp dnsDP) String() string { return dp.req.String() }
func (dp dnsDP) FieldAsInterface(fldPath []string) (o any, e error) {
	if len(fldPath) != 0 && strings.HasPrefix(fldPath[0], utils.DNSOption) {
		return dp.opts.FieldAsInterface(fldPath)
	}
	return dp.req.FieldAsInterface(fldPath)
}
func (dp dnsDP) FieldAsString(fldPath []string) (string, error) {
	valIface, err := dp.FieldAsInterface(fldPath)
	if err != nil {
		return utils.EmptyString, err
	}
	return utils.IfaceAsString(valIface), nil
}

func updateDNSMsgFromNM(msg *dns.Msg, nm *utils.OrderedNavigableMap, qType uint16, qName string) (err error) {
	msgFields := make(utils.StringSet) // work around to NMap issue
	for el := nm.GetFirstElement(); el != nil; el = el.Next() {
		path := el.Value
		itm, _ := nm.Field(path)
		switch path[0] { // go for each posible field
		case utils.DNSId:
			var vItm int64
			if vItm, err = utils.IfaceAsTInt64(itm.Data); err != nil {
				return fmt.Errorf("item: <%s>, err: %s", path[0], err.Error())
			}
			msg.Id = uint16(vItm)
		case utils.DNSResponse:
			var vItm bool
			if vItm, err = utils.IfaceAsBool(itm.Data); err != nil {
				return fmt.Errorf("item: <%s>, err: %s", path[0], err.Error())
			}
			msg.Response = vItm
		case utils.DNSOpcode:
			var vItm int64
			if vItm, err = utils.IfaceAsTInt64(itm.Data); err != nil {
				return fmt.Errorf("item: <%s>, err: %s", path[0], err.Error())
			}
			msg.Opcode = int(vItm)
		case utils.DNSAuthoritative:
			var vItm bool
			if vItm, err = utils.IfaceAsBool(itm.Data); err != nil {
				return fmt.Errorf("item: <%s>, err: %s", path[0], err.Error())
			}
			msg.Authoritative = vItm
		case utils.DNSTruncated:
			var vItm bool
			if vItm, err = utils.IfaceAsBool(itm.Data); err != nil {
				return fmt.Errorf("item: <%s>, err: %s", path[0], err.Error())
			}
			msg.Truncated = vItm
		case utils.DNSRecursionDesired:
			var vItm bool
			if vItm, err = utils.IfaceAsBool(itm.Data); err != nil {
				return fmt.Errorf("item: <%s>, err: %s", path[0], err.Error())
			}
			msg.RecursionDesired = vItm
		case utils.DNSRecursionAvailable:
			var vItm bool
			if vItm, err = utils.IfaceAsBool(itm.Data); err != nil {
				return fmt.Errorf("item: <%s>, err: %s", path[0], err.Error())
			}
			msg.RecursionAvailable = vItm
		case utils.DNSZero:
			var vItm bool
			if vItm, err = utils.IfaceAsBool(itm.Data); err != nil {
				return fmt.Errorf("item: <%s>, err: %s", path[0], err.Error())
			}
			msg.Zero = vItm
		case utils.DNSAuthenticatedData:
			var vItm bool
			if vItm, err = utils.IfaceAsBool(itm.Data); err != nil {
				return fmt.Errorf("item: <%s>, err: %s", path[0], err.Error())
			}
			msg.AuthenticatedData = vItm
		case utils.DNSCheckingDisabled:
			var vItm bool
			if vItm, err = utils.IfaceAsBool(itm.Data); err != nil {
				return fmt.Errorf("item: <%s>, err: %s", path[0], err.Error())
			}
			msg.CheckingDisabled = vItm
		case utils.DNSRcode:
			var vItm int64
			if vItm, err = utils.IfaceAsTInt64(itm.Data); err != nil {
				return fmt.Errorf("item: <%s>, err: %s", path[0], err.Error())
			}
			msg.Rcode = int(vItm)
		case utils.DNSQuestion:
			if msg.Question, err = updateDnsQuestions(msg.Question, path[1:len(path)-1], itm.Data, itm.NewBranch); err != nil {
				return fmt.Errorf("item: <%s>, err: %s", path[:len(path)-1], err.Error())
			}
		case utils.DNSAnswer:
			newBranch := itm.NewBranch ||
				len(msg.Answer) == 0 ||
				msgFields.Has(path[0])
			if newBranch { // force append if the same path was already used
				msgFields = make(utils.StringSet)      // reset the fields inside since we have a new message
				msgFields.Add(strings.Join(path, ".")) // detect new branch
			}
			if msg.Answer, err = updateDnsAnswer(msg.Answer, qType, qName, path[1:len(path)-1], itm.Data, newBranch); err != nil {
				return fmt.Errorf("item: <%s>, err: %s", path[:len(path)-1], err.Error())
			}
		case utils.DNSNs: //ToDO
		case utils.DNSExtra: //ToDO
		case utils.DNSOption:
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
func updateDnsQuestions(q []dns.Question, path []string, value any, newBranch bool) (_ []dns.Question, err error) {
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
	case utils.DNSName:
		q[idx].Name = utils.IfaceAsString(value)
	case utils.DNSQtype:
		var vItm int64
		if vItm, err = utils.IfaceAsTInt64(value); err != nil {
			return
		}
		q[idx].Qtype = uint16(vItm)
	case utils.DNSQclass:
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

func updateDnsOption(q []dns.EDNS0, path []string, value any, newBranch bool) (_ []dns.EDNS0, err error) {
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
		if field != utils.DNSNsid {
			err = utils.ErrWrongPath
			return
		}
		v.Nsid = utils.IfaceAsString(value)
	case *dns.EDNS0_SUBNET:
		switch field {
		case utils.DNSFamily:
			var vItm int64
			if vItm, err = utils.IfaceAsTInt64(value); err != nil {
				return
			}
			v.Family = uint16(vItm)
		case utils.DNSSourceNetmask:
			var vItm int64
			if vItm, err = utils.IfaceAsTInt64(value); err != nil {
				return
			}
			v.SourceNetmask = uint8(vItm)
		case utils.DNSSourceScope:
			var vItm int64
			if vItm, err = utils.IfaceAsTInt64(value); err != nil {
				return
			}
			v.SourceScope = uint8(vItm)
		case utils.Address:
			v.Address = net.ParseIP(utils.IfaceAsString(value))
		default:
			err = utils.ErrWrongPath
			return
		}
	case *dns.EDNS0_COOKIE:
		if field != utils.DNSCookie {
			err = utils.ErrWrongPath
			return
		}
		v.Cookie = utils.IfaceAsString(value)
	case *dns.EDNS0_UL:
		switch field {
		case utils.DNSLease:
			var vItm int64
			if vItm, err = utils.IfaceAsTInt64(value); err != nil {
				return
			}
			v.Lease = uint32(vItm)
		case utils.DNSKeyLease:
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
		case utils.VersionName:
			var vItm int64
			if vItm, err = utils.IfaceAsTInt64(value); err != nil {
				return
			}
			v.Version = uint16(vItm)
		case utils.DNSOpcode:
			var vItm int64
			if vItm, err = utils.IfaceAsTInt64(value); err != nil {
				return
			}
			v.Opcode = uint16(vItm)
		case utils.Error:
			var vItm int64
			if vItm, err = utils.IfaceAsTInt64(value); err != nil {
				return
			}
			v.Error = uint16(vItm)
		case utils.DNSId:
			var vItm int64
			if vItm, err = utils.IfaceAsTInt64(value); err != nil {
				return
			}
			v.Id = uint64(vItm)
		case utils.DNSLeaseLife:
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
		if field != utils.DNSDAU {
			err = utils.ErrWrongPath
			return
		}
		v.AlgCode = []uint8(utils.IfaceAsString(value))
	case *dns.EDNS0_DHU:
		if field != utils.DNSDHU {
			err = utils.ErrWrongPath
			return
		}
		v.AlgCode = []uint8(utils.IfaceAsString(value))
	case *dns.EDNS0_N3U:
		if field != utils.DNSN3U {
			err = utils.ErrWrongPath
			return
		}
		v.AlgCode = []uint8(utils.IfaceAsString(value))
	case *dns.EDNS0_EXPIRE:
		if field != utils.DNSExpire {
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
		case utils.Length: //
			var vItm int64
			if vItm, err = utils.IfaceAsTInt64(value); err != nil {
				return
			}
			v.Length = uint16(vItm)
		case utils.DNSTimeout: //
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
		if field != utils.DNSPadding {
			err = utils.ErrWrongPath
			return
		}
		v.Padding = []byte(utils.IfaceAsString(value))
	case *dns.EDNS0_EDE:
		switch field {
		case utils.DNSInfoCode:
			var vItm int64
			if vItm, err = utils.IfaceAsTInt64(value); err != nil {
				return
			}
			v.InfoCode = uint16(vItm)
		case utils.DNSExtraText:
			v.ExtraText = utils.IfaceAsString(value)
		default:
			err = utils.ErrWrongPath
			return
		}
	case *dns.EDNS0_ESU:
		if field != utils.DNSUri {
			err = utils.ErrWrongPath
			return
		}
		v.Uri = utils.IfaceAsString(value)
	case *dns.EDNS0_LOCAL: // if already there you can change data
		if field != utils.DNSData {
			err = utils.ErrWrongPath
			return
		}
		v.Data = []byte(utils.IfaceAsString(value))
	case nil:
		err = fmt.Errorf("unsupported dns option type <%T>", v)
	default:
		err = fmt.Errorf("unsupported dns option type <%T>", v)
	}
	return q, err
}

func createDnsOption(field string, value any) (o dns.EDNS0, err error) {
	switch field {
	case utils.DNSNsid: // EDNS0_NSID
		o = &dns.EDNS0_NSID{Nsid: utils.IfaceAsString(value)}
	case utils.DNSFamily: // EDNS0_SUBNET
		var vItm int64
		if vItm, err = utils.IfaceAsTInt64(value); err != nil {
			return
		}
		o = &dns.EDNS0_SUBNET{Family: uint16(vItm)}
	case utils.DNSSourceNetmask: // EDNS0_SUBNET
		var vItm int64
		if vItm, err = utils.IfaceAsTInt64(value); err != nil {
			return
		}
		o = &dns.EDNS0_SUBNET{SourceNetmask: uint8(vItm)}
	case utils.DNSSourceScope: // EDNS0_SUBNET
		var vItm int64
		if vItm, err = utils.IfaceAsTInt64(value); err != nil {
			return
		}
		o = &dns.EDNS0_SUBNET{SourceScope: uint8(vItm)}
	case utils.Address: // EDNS0_SUBNET
		o = &dns.EDNS0_SUBNET{Address: net.ParseIP(utils.IfaceAsString(value))}
	case utils.DNSCookie: // EDNS0_COOKIE
		o = &dns.EDNS0_COOKIE{Cookie: utils.IfaceAsString(value)}
	case utils.DNSLease: // EDNS0_UL
		var vItm int64
		if vItm, err = utils.IfaceAsTInt64(value); err != nil {
			return
		}
		o = &dns.EDNS0_UL{Lease: uint32(vItm)}
	case utils.DNSKeyLease: // EDNS0_UL
		var vItm int64
		if vItm, err = utils.IfaceAsTInt64(value); err != nil {
			return
		}
		o = &dns.EDNS0_UL{KeyLease: uint32(vItm)}
	case utils.VersionName: // EDNS0_LLQ
		var vItm int64
		if vItm, err = utils.IfaceAsTInt64(value); err != nil {
			return
		}
		o = &dns.EDNS0_LLQ{Version: uint16(vItm)}
	case utils.DNSOpcode: // EDNS0_LLQ
		var vItm int64
		if vItm, err = utils.IfaceAsTInt64(value); err != nil {
			return
		}
		o = &dns.EDNS0_LLQ{Opcode: uint16(vItm)}
	case utils.Error: // EDNS0_LLQ
		var vItm int64
		if vItm, err = utils.IfaceAsTInt64(value); err != nil {
			return
		}
		o = &dns.EDNS0_LLQ{Error: uint16(vItm)}
	case utils.DNSId: // EDNS0_LLQ
		var vItm int64
		if vItm, err = utils.IfaceAsTInt64(value); err != nil {
			return
		}
		o = &dns.EDNS0_LLQ{Id: uint64(vItm)}
	case utils.DNSLeaseLife: // EDNS0_LLQ
		var vItm int64
		if vItm, err = utils.IfaceAsTInt64(value); err != nil {
			return
		}
		o = &dns.EDNS0_LLQ{LeaseLife: uint32(vItm)}
	case utils.DNSDAU: // EDNS0_DAU
		o = &dns.EDNS0_DAU{AlgCode: []uint8(utils.IfaceAsString(value))}
	case utils.DNSDHU: // EDNS0_DHU
		o = &dns.EDNS0_DHU{AlgCode: []uint8(utils.IfaceAsString(value))}
	case utils.DNSN3U: // EDNS0_N3U
		o = &dns.EDNS0_N3U{AlgCode: []uint8(utils.IfaceAsString(value))}
	case utils.DNSExpire: // EDNS0_EXPIRE
		var vItm int64
		if vItm, err = utils.IfaceAsTInt64(value); err != nil {
			return
		}
		o = &dns.EDNS0_EXPIRE{Expire: uint32(vItm)}
	case utils.Length: // EDNS0_TCP_KEEPALIVE
		var vItm int64
		if vItm, err = utils.IfaceAsTInt64(value); err != nil {
			return
		}
		o = &dns.EDNS0_TCP_KEEPALIVE{Length: uint16(vItm)}
	case utils.DNSTimeout: // EDNS0_TCP_KEEPALIVE
		var vItm int64
		if vItm, err = utils.IfaceAsTInt64(value); err != nil {
			return
		}
		o = &dns.EDNS0_TCP_KEEPALIVE{Timeout: uint16(vItm)}
	case utils.DNSPadding: // EDNS0_PADDING
		o = &dns.EDNS0_PADDING{Padding: []byte(utils.IfaceAsString(value))}
	case utils.DNSInfoCode: // EDNS0_EDE
		var vItm int64
		if vItm, err = utils.IfaceAsTInt64(value); err != nil {
			return
		}
		o = &dns.EDNS0_EDE{InfoCode: uint16(vItm)}
	case utils.DNSExtraText: // EDNS0_EDE
		o = &dns.EDNS0_EDE{ExtraText: utils.IfaceAsString(value)}
	case utils.DNSUri: // EDNS0_ESU
		o = &dns.EDNS0_ESU{Uri: utils.IfaceAsString(value)}
	default:
		err = fmt.Errorf("can not create option from field <%q>", field)
	}
	return
}

func updateDnsAnswer(q []dns.RR, qType uint16, qName string, path []string, value any, newBranch bool) (_ []dns.RR, err error) {
	var idx int
	if lPath := len(path); lPath == 0 {
		err = utils.ErrWrongPath
		return
	} else {
		var hasIdx bool
		if idx, err = strconv.Atoi(path[0]); err == nil {
			hasIdx = true
		}
		err = nil
		if !hasIdx || lPath == 1 { // only the field so update the last one
			if newBranch || len(q) == 0 {
				var a dns.RR
				if a, err = newDNSAnswer(qType, qName); err != nil {
					return
				}
				q = append(q, a)
			}
			idx = len(q) - 1
		} else { // the index is specified
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
		}
	}

	switch v := q[idx].(type) {
	case *dns.NAPTR:
		err = updateDnsNAPTRAnswer(v, path, value)
	case *dns.SRV:
		err = updateDnsSRVAnswer(v, path, value)
	case *dns.A:
		if len(path) < 1 ||
			(path[0] != utils.DNSHdr && len(path) != 1) ||
			(path[0] == utils.DNSHdr && len(path) != 2) {
			err = utils.ErrWrongPath
			return
		}
		switch path[0] {
		case utils.DNSHdr:
			err = updateDnsRRHeader(&v.Hdr, path[1:], value)
		case utils.DNSA:
			v.A = net.ParseIP(utils.IfaceAsString(value))
			if v.A == nil {
				err = fmt.Errorf("invalid IP address <%v>",
					utils.IfaceAsString(value))
				return
			}
		default:
			err = utils.ErrWrongPath
		}
	case nil:
		err = fmt.Errorf("unsupported dns option type <%T>", v)
	default:
		err = fmt.Errorf("unsupported dns option type <%T>", v)
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
	case dns.TypeSRV:
		a = &dns.SRV{Hdr: hdr}
	default:
		err = fmt.Errorf("unsupported DNS type: <%v>", dns.TypeToString[qType])
	}
	return
}

func updateDnsSRVAnswer(v *dns.SRV, path []string, value any) (err error) {
	if len(path) < 1 ||
		(path[0] != utils.DNSHdr && len(path) != 1) ||
		(path[0] == utils.DNSHdr && len(path) != 2) {
		err = utils.ErrWrongPath
		return
	}
	switch path[0] {
	case utils.DNSHdr:
		err = updateDnsRRHeader(&v.Hdr, path[1:], value)
	case utils.DNSPriority:
		var vItm int64
		if vItm, err = utils.IfaceAsTInt64(value); err != nil {
			return
		}
		v.Priority = uint16(vItm)
	case utils.Weight:
		var vItm int64
		if vItm, err = utils.IfaceAsTInt64(value); err != nil {
			return
		}
		v.Weight = uint16(vItm)
	case utils.DNSPort:
		var vItm int64
		if vItm, err = utils.IfaceAsTInt64(value); err != nil {
			return
		}
		v.Port = uint16(vItm)
	case utils.DNSTarget:
		v.Target = utils.IfaceAsString(value)
	default:
		err = utils.ErrWrongPath
	}
	return
}

func updateDnsNAPTRAnswer(v *dns.NAPTR, path []string, value any) (err error) {
	if len(path) < 1 ||
		(path[0] != utils.DNSHdr && len(path) != 1) ||
		(path[0] == utils.DNSHdr && len(path) != 2) {
		return utils.ErrWrongPath
	}
	switch path[0] {
	case utils.DNSHdr:
		return updateDnsRRHeader(&v.Hdr, path[1:], value)
	case utils.Order:
		var vItm int64
		if vItm, err = utils.IfaceAsTInt64(value); err != nil {
			return
		}
		v.Order = uint16(vItm)
	case utils.Preference:
		var vItm int64
		if vItm, err = utils.IfaceAsTInt64(value); err != nil {
			return
		}
		v.Preference = uint16(vItm)
	case utils.Flags:
		v.Flags = utils.IfaceAsString(value)
	case utils.Service:
		v.Service = utils.IfaceAsString(value)
	case utils.Regexp:
		v.Regexp = utils.IfaceAsString(value)
	case utils.Replacement:
		v.Replacement = utils.IfaceAsString(value)
	default:
		return utils.ErrWrongPath
	}
	return
}

func updateDnsRRHeader(v *dns.RR_Header, path []string, value any) (err error) {
	if len(path) != 1 {
		return utils.ErrWrongPath
	}
	switch path[0] {
	case utils.DNSName:
		v.Name = utils.IfaceAsString(value)
	case utils.DNSRrtype:
		var vItm int64
		if vItm, err = utils.IfaceAsTInt64(value); err != nil {
			return
		}
		v.Rrtype = uint16(vItm)
	case utils.DNSClass:
		var vItm int64
		if vItm, err = utils.IfaceAsTInt64(value); err != nil {
			return
		}
		v.Class = uint16(vItm)
	case utils.DNSTtl:
		var vItm int64
		if vItm, err = utils.IfaceAsTInt64(value); err != nil {
			return
		}
		v.Ttl = uint32(vItm)
	case utils.DNSRdlength:
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
