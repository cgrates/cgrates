// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

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
		path = StripTrailingIndex(path)
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
