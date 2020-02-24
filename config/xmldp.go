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

package config

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/antchfx/xmlquery"
	"github.com/cgrates/cgrates/utils"
)

// NewXmlProvider constructs a utils.DataProvider
func NewXmlProvider(req *xmlquery.Node, cdrPath utils.HierarchyPath) (dP utils.DataProvider) {
	dP = &XmlProvider{req: req, cdrPath: cdrPath, cache: utils.NavigableMap{}}
	return
}

// XmlProvider implements utils.DataProvider so we can pass it to filters
type XmlProvider struct {
	req     *xmlquery.Node
	cdrPath utils.HierarchyPath //used to compute relative path
	cache   utils.NavigableMap
}

// String is part of utils.DataProvider interface
// when called, it will display the already parsed values out of cache
func (xP *XmlProvider) String() string {
	return utils.ToJSON(xP)
}

// FieldAsInterface is part of utils.DataProvider interface
func (xP *XmlProvider) FieldAsInterface(fldPath []string) (data interface{}, err error) {
	if len(fldPath) == 0 {
		return nil, utils.ErrNotFound
	}
	if data, err = xP.cache.FieldAsInterface(fldPath); err == nil ||
		err != utils.ErrNotFound { // item found in cache
		return
	}
	err = nil // cancel previous err
	relPath := utils.HierarchyPath(fldPath[len(xP.cdrPath):])
	var slctrStr string
	for i := range relPath {
		if sIdx := strings.Index(relPath[i], "["); sIdx != -1 {
			slctrStr = relPath[i][sIdx:]
			if slctrStr[len(slctrStr)-1:] != "]" {
				return nil, fmt.Errorf("filter rule <%s> needs to end in ]", slctrStr)
			}
			relPath[i] = relPath[i][:sIdx]
			if slctrStr[1:2] != "@" {
				i, err := strconv.Atoi(slctrStr[1 : len(slctrStr)-1])
				if err != nil {
					return nil, err
				}
				slctrStr = "[" + strconv.Itoa(i+1) + "]"
			}
			relPath[i] = relPath[i] + slctrStr
		}
	}
	data, err = ElementText(xP.req, relPath.AsString("/", false))
	xP.cache.Set(fldPath, data)
	return
}

// FieldAsString is part of utils.DataProvider interface
func (xP *XmlProvider) FieldAsString(fldPath []string) (data string, err error) {
	var valIface interface{}
	valIface, err = xP.FieldAsInterface(fldPath)
	if err != nil {
		return
	}
	return utils.IfaceAsString(valIface), nil
}

// RemoteHost is part of utils.DataProvider interface
func (xP *XmlProvider) RemoteHost() net.Addr {
	return utils.LocalAddr()
}

// ElementText will process the node to extract the elementName's text out of it (only first one found)
// returns utils.ErrNotFound if the element is not found in the node
// Make the method exportable until we remove the cdrc
func ElementText(xmlElement *xmlquery.Node, elmntPath string) (string, error) {
	elmnt := xmlquery.FindOne(xmlElement, elmntPath)
	if elmnt == nil {
		return "", utils.ErrNotFound
	}
	return elmnt.InnerText(), nil
}
