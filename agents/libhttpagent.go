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
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/antchfx/xmlquery"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

// newHADataProvider constructs a DataProvider
func newHADataProvider(reqPayload string,
	req *http.Request) (dP config.DataProvider, err error) {
	switch reqPayload {
	default:
		return nil, fmt.Errorf("unsupported decoder type <%s>", reqPayload)
	case utils.MetaUrl:
		return newHTTPUrlDP(req)
	case utils.MetaXml:
		return newHTTPXmlDP(req)

	}
}

func newHTTPUrlDP(req *http.Request) (dP config.DataProvider, err error) {
	dP = &httpUrlDP{req: req, cache: config.NewNavigableMap(nil)}
	return
}

// httpUrlDP implements engine.DataProvider, serving as url data decoder
// decoded data is only searched once and cached
type httpUrlDP struct {
	req   *http.Request
	cache *config.NavigableMap
}

// String is part of engine.DataProvider interface
// when called, it will display the already parsed values out of cache
func (hU *httpUrlDP) String() string {
	//return utils.ToJSON(hU.cache.AsMapStringInterface())
	return "" // ToDo: fixme
}

// FieldAsInterface is part of engine.DataProvider interface
func (hU *httpUrlDP) FieldAsInterface(fldPath []string) (data interface{}, err error) {
	if len(fldPath) != 1 {
		return nil, utils.ErrNotFound
	}
	if data, err = hU.cache.FieldAsInterface(fldPath); err != nil {
		if err != utils.ErrNotFound { // item found in cache
			return
		}
		err = nil // cancel previous err
	} else {
		return // data found in cache
	}
	data = hU.req.FormValue(fldPath[0])
	hU.cache.Set(fldPath, data, false)
	return
}

// FieldAsString is part of engine.DataProvider interface
func (hU *httpUrlDP) FieldAsString(fldPath []string) (data string, err error) {
	var valIface interface{}
	valIface, err = hU.FieldAsInterface(fldPath)
	if err != nil {
		return
	}
	data, err = utils.IfaceAsString(valIface)
	return
}

// AsNavigableMap is part of engine.DataProvider interface
func (hU *httpUrlDP) AsNavigableMap([]*config.FCTemplate) (
	nm *config.NavigableMap, err error) {
	return nil, utils.ErrNotImplemented
}

func newHTTPXmlDP(req *http.Request) (dP config.DataProvider, err error) {
	byteData, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}
	//convert the byteData into a xmlquery Node
	doc, err := xmlquery.Parse(strings.NewReader(string(byteData)))
	if err != nil {
		return nil, err
	}
	dP = &httpXmlDP{xmlDoc: doc, cache: config.NewNavigableMap(nil)}
	return
}

// httpXmlDP implements engine.DataProvider, serving as xml data decoder
// decoded data is only searched once and cached
type httpXmlDP struct {
	cache  *config.NavigableMap
	xmlDoc *xmlquery.Node
}

// String is part of engine.DataProvider interface
// when called, it will display the already parsed values out of cache
func (hU *httpXmlDP) String() string {
	//return utils.ToJSON(hU.cache.AsMapStringInterface())
	return "" // ToDo: fixme
}

// FieldAsInterface is part of engine.DataProvider interface
func (hU *httpXmlDP) FieldAsInterface(fldPath []string) (data interface{}, err error) {
	//if path is missing return here error because if it arrived in xmlquery library will panic
	if len(fldPath) == 0 {
		return nil, fmt.Errorf("Empty path")
	}
	if data, err = hU.cache.FieldAsInterface(fldPath); err == nil ||
		err != utils.ErrNotFound { // item found in cache
		return
	}
	err = nil // cancel previous err
	var slctrStr string
	for i := range fldPath {
		if sIdx := strings.Index(fldPath[i], "["); sIdx != -1 {
			slctrStr = fldPath[i][sIdx:]
			if slctrStr[len(slctrStr)-1:] != "]" {
				return nil, fmt.Errorf("filter rule <%s> needs to end in ]", slctrStr)
			}
			fldPath[i] = fldPath[i][:sIdx]
			if slctrStr[1:2] != "@" {
				i, err := strconv.Atoi(slctrStr[1 : len(slctrStr)-1])
				if err != nil {
					return nil, err
				}
				slctrStr = "[" + strconv.Itoa(i+1) + "]"
			}
			fldPath[i] = fldPath[i] + slctrStr
		}
	}
	//convert fldPath to HierarchyPath
	path := utils.HierarchyPath(fldPath)
	elmnt := xmlquery.FindOne(hU.xmlDoc, path.AsString("/", false))
	if elmnt == nil {
		return
	}
	//add the content in data and cache it
	data = elmnt.InnerText()
	hU.cache.Set(fldPath, data, false)
	return
}

// FieldAsString is part of engine.DataProvider interface
func (hU *httpXmlDP) FieldAsString(fldPath []string) (data string, err error) {
	var valIface interface{}
	valIface, err = hU.FieldAsInterface(fldPath)
	if err != nil {
		return
	}
	data, err = utils.IfaceAsString(valIface)
	return
}

// AsNavigableMap is part of engine.DataProvider interface
func (hU *httpXmlDP) AsNavigableMap([]*config.FCTemplate) (
	nm *config.NavigableMap, err error) {
	return nil, utils.ErrNotImplemented
}

// httpAgentReplyEncoder will encode  []*engine.NMElement
// and write content to http writer
type httpAgentReplyEncoder interface {
	Encode(*config.NavigableMap) error
}

// newHAReplyEncoder constructs a httpAgentReqDecoder based on encoder type
func newHAReplyEncoder(encType string,
	w http.ResponseWriter) (rE httpAgentReplyEncoder, err error) {
	switch encType {
	default:
		return nil, fmt.Errorf("unsupported encoder type <%s>", encType)
	case utils.MetaXml:
		return newHAXMLEncoder(w)
	}
}

func newHAXMLEncoder(w http.ResponseWriter) (xE httpAgentReplyEncoder, err error) {
	return &haXMLEncoder{w: w}, nil
}

type haXMLEncoder struct {
	w http.ResponseWriter
}

// Encode implements httpAgentReplyEncoder
func (xE *haXMLEncoder) Encode(nM *config.NavigableMap) (err error) {
	var xmlElmnts []*config.XMLElement
	if xmlElmnts, err = nM.AsXMLElements(); err != nil {
		return
	}
	if len(xmlElmnts) == 0 {
		return
	}
	var xmlOut []byte
	if xmlOut, err = xml.MarshalIndent(xmlElmnts, "", "  "); err != nil {
		return
	}
	if _, err = xE.w.Write([]byte(xml.Header)); err != nil {
		return
	}
	_, err = xE.w.Write(xmlOut)
	return
}
