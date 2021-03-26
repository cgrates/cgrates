/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOev.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>
*/

package utils

import (
	"encoding/xml"
	"strings"
)

// XMLElement is specially crafted to be automatically marshalled by encoding/xml
type XMLElement struct {
	XMLName    xml.Name
	Value      string        `xml:",chardata"`
	Attributes []*xml.Attr   `xml:",attr"`
	Elements   []*XMLElement `xml:"omitempty"`
}

// NMAsXMLElements returns the values as []*XMLElement which can be later marshaled
// considers each value returned by .Values() in the form of []*NMItem, otherwise errors
func NMAsXMLElements(nm *OrderedNavigableMap) (ents []*XMLElement, err error) {
	pathIdx := make(map[string]*XMLElement) // Keep the index of elements based on path
	for el := nm.GetFirstElement(); el != nil; el = el.Next() {
		path := el.Value
		nmItm, _ := nm.Field(path) // this should never return error cause we get the path from the order
		if nmItm.NewBranch {
			pathIdx = make(map[string]*XMLElement) // reset cache so we can start having other elements with same path
		}
		path = path[:len(path)-1] // remove the last index
		val := nmItm.String()
		var pathCached bool
		for i := len(path); i > 0; i-- {
			var cachedElm *XMLElement
			if cachedElm, pathCached = pathIdx[strings.Join(path[:i], "")]; !pathCached {
				continue
			}
			if i == len(path) { // lastElmnt, overwrite value or add attribute
				if nmItm.AttributeID != "" {
					cachedElm.Attributes = append(cachedElm.Attributes,
						&xml.Attr{
							Name:  xml.Name{Local: nmItm.AttributeID},
							Value: val,
						})
				} else {
					cachedElm.Value = val
				}
				break
			}
			// create elements in reverse order so we can append already created
			var newElm *XMLElement
			for j := len(path); j > i; j-- {
				elm := &XMLElement{XMLName: xml.Name{Local: path[j-1]}}
				pathIdx[strings.Join(path[:j], "")] = elm
				if newElm == nil {
					if nmItm.AttributeID != "" {
						elm.Attributes = append(elm.Attributes,
							&xml.Attr{
								Name:  xml.Name{Local: nmItm.AttributeID},
								Value: val,
							})
					} else {
						elm.Value = val
					}
					newElm = elm // last element
				} else {
					elm.Elements = append(elm.Elements, newElm)
					newElm = elm
				}
			}
			cachedElm.Elements = append(cachedElm.Elements, newElm)
		}
		if !pathCached { // not an update but new element to be created
			var newElm *XMLElement
			for i := len(path); i > 0; i-- {
				elm := &XMLElement{XMLName: xml.Name{Local: path[i-1]}}
				pathIdx[strings.Join(path[:i], "")] = elm
				if newElm == nil { // last element, create data inside
					if nmItm.AttributeID != "" {
						elm.Attributes = append(elm.Attributes,
							&xml.Attr{
								Name:  xml.Name{Local: nmItm.AttributeID},
								Value: val,
							})
					} else {
						elm.Value = val
					}
					newElm = elm // last element
				} else {
					elm.Elements = append(elm.Elements, newElm)
					newElm = elm
				}
			}
			ents = append(ents, newElm)
		}
	}
	return
}
