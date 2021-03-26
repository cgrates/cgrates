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
	"errors"
	"fmt"
	"net"
	"strings"

	"github.com/cgrates/cgrates/utils"
	"github.com/miekg/dns"
)

const (
	QueryType   = "QueryType"
	E164Address = "E164Address"
	QueryName   = "QueryName"
	DomainName  = "DomainName"
)

// e164FromNAPTR extracts the E164 address out of a NAPTR name record
func e164FromNAPTR(name string) (e164 string, err error) {
	i := strings.Index(name, ".e164.")
	if i == -1 {
		return "", errors.New("unknown format")
	}
	e164 = utils.ReverseString(
		strings.Replace(name[:i], ".", "", -1))
	return
}

// domainNameFromNAPTR extracts the domain part out of a NAPTR name record
func domainNameFromNAPTR(name string) (dName string) {
	i := strings.Index(name, ".e164.")
	if i == -1 {
		dName = name
	} else {
		dName = name[i:]
	}
	return strings.Trim(dName, ".")
}

// newDADataProvider constructs a DataProvider for a diameter message
func newDNSDataProvider(req *dns.Msg,
	w dns.ResponseWriter) utils.DataProvider {
	return &dnsDP{req: req, w: w,
		cache: utils.MapStorage{}}
}

// dnsDP implements engien.DataProvider, serving as dns.Msg decoder
// cache is used to cache queries within the message
type dnsDP struct {
	req   *dns.Msg
	w     dns.ResponseWriter
	cache utils.MapStorage
}

// String is part of utils.DataProvider interface
// when called, it will display the already parsed values out of cache
func (dP *dnsDP) String() string {
	return utils.ToJSON(dP.req)
}

// FieldAsString is part of utils.DataProvider interface
func (dP *dnsDP) FieldAsString(fldPath []string) (data string, err error) {
	var valIface interface{}
	valIface, err = dP.FieldAsInterface(fldPath)
	if err != nil {
		return
	}
	return utils.IfaceAsString(valIface), nil
}

// RemoteHost is part of utils.DataProvider interface
func (dP *dnsDP) RemoteHost() net.Addr {
	return utils.NewNetAddr(dP.w.RemoteAddr().Network(), dP.w.RemoteAddr().String())
}

// FieldAsInterface is part of utils.DataProvider interface
func (dP *dnsDP) FieldAsInterface(fldPath []string) (data interface{}, err error) {
	if data, err = dP.cache.FieldAsInterface(fldPath); err != nil {
		if err != utils.ErrNotFound { // item found in cache
			return nil, err
		}
		err = nil // cancel previous err
	} else {
		return // data was found in cache
	}
	data = ""
	// Return Question[0] by default
	if len(dP.req.Question) != 0 {
		data = dP.req.Question[0]
	}
	dP.cache.Set(fldPath, data)
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

// updateDNSMsgFromNM will update DNS message with values from NavigableMap
func updateDNSMsgFromNM(msg *dns.Msg, nm *utils.OrderedNavigableMap) (err error) {
	msgFields := make(utils.StringSet) // work around to NMap issue
	for el := nm.GetFirstElement(); el != nil; el = el.Next() {
		path := el.Value
		cfgItm, _ := nm.Field(path)
		// path = path[:len(path)-1] // no need to remove the last index here as this uses only the first level
		apnd := len(msg.Answer) == 0
		if msgFields.Has(path[0]) { // force append if the same path was already used
			apnd = true
		}
		if apnd {
			if err = appendDNSAnswer(msg); err != nil {
				return
			}
			msgFields = make(utils.StringSet) // reset the fields inside since we have a new message
		}
		itmData := cfgItm.Data
		switch path[0] {
		case utils.Rcode:
			var itm int64
			if itm, err = utils.IfaceAsInt64(itmData); err != nil {
				return fmt.Errorf("item: <%s>, err: %s", path[0], err.Error())
			}
			msg.Rcode = int(itm)
		case utils.Order:
			if msg.Question[0].Qtype != dns.TypeNAPTR {
				return fmt.Errorf("field <%s> only works with NAPTR", utils.Order)
			}
			var itm int64
			if itm, err = utils.IfaceAsInt64(itmData); err != nil {
				return fmt.Errorf("item: <%s>, err: %s", path[0], err.Error())
			}
			msg.Answer[len(msg.Answer)-1].(*dns.NAPTR).Order = uint16(itm)
		case utils.Preference:
			if msg.Question[0].Qtype != dns.TypeNAPTR {
				return fmt.Errorf("field <%s> only works with NAPTR", utils.Preference)
			}
			var itm int64
			if itm, err = utils.IfaceAsInt64(itmData); err != nil {
				return fmt.Errorf("item: <%s>, err: %s", path[0], err.Error())
			}
			msg.Answer[len(msg.Answer)-1].(*dns.NAPTR).Preference = uint16(itm)
		case utils.Flags:
			if msg.Question[0].Qtype != dns.TypeNAPTR {
				return fmt.Errorf("field <%s> only works with NAPTR", utils.Flags)
			}
			msg.Answer[len(msg.Answer)-1].(*dns.NAPTR).Flags = utils.IfaceAsString(itmData)
		case utils.Service:
			if msg.Question[0].Qtype != dns.TypeNAPTR {
				return fmt.Errorf("field <%s> only works with NAPTR", utils.Service)
			}
			msg.Answer[len(msg.Answer)-1].(*dns.NAPTR).Service = utils.IfaceAsString(itmData)
		case utils.Regexp:
			if msg.Question[0].Qtype != dns.TypeNAPTR {
				return fmt.Errorf("field <%s> only works with NAPTR", utils.Regexp)
			}
			msg.Answer[len(msg.Answer)-1].(*dns.NAPTR).Regexp = utils.IfaceAsString(itmData)
		case utils.Replacement:
			if msg.Question[0].Qtype != dns.TypeNAPTR {
				return fmt.Errorf("field <%s> only works with NAPTR", utils.Replacement)
			}
			msg.Answer[len(msg.Answer)-1].(*dns.NAPTR).Replacement = utils.IfaceAsString(itmData)
		}

		msgFields.Add(path[0]) // detect new branch

	}
	return
}
