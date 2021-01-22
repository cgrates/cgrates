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

package config

import (
	"encoding/xml"
	"fmt"
	"strings"
	"time"

	"github.com/cgrates/cgrates/utils"
)

// NMItem is an item in the NavigableMap
type NMItem struct {
	Path   []string    // path in map
	Data   interface{} // value of the element
	Config *FCTemplate // so we can store additional configuration
}

func (nmi *NMItem) String() string {
	return utils.ToJSON(nmi)
}

// Interface returns the wraped interface
func (nmi *NMItem) Interface() interface{} {
	return nmi.Data
}

// Field not implemented only used in order to implement the NM interface
func (nmi *NMItem) Field(path utils.PathItems) (val utils.NMInterface, err error) {
	return nil, utils.ErrNotImplemented
}

// Set not implemented only used in order to implement the NM interface
// special case when the path is empty the interface should be seted
// this is in order to modify the wraped interface
func (nmi *NMItem) Set(path utils.PathItems, val utils.NMInterface) (added bool, err error) {
	if len(path) != 0 {
		return false, utils.ErrWrongPath
	}
	nmi.Data = val.Interface()
	return
}

// Remove not implemented only used in order to implement the NM interface
func (nmi *NMItem) Remove(path utils.PathItems) (err error) {
	return utils.ErrNotImplemented
}

// Type returns the type of the NM interface
func (nmi *NMItem) Type() utils.NMType {
	return utils.NMDataType
}

// Empty returns true if the NM is empty(no data)
func (nmi *NMItem) Empty() bool {
	return nmi == nil || nmi.Data == nil
}

// Len not implemented only used in order to implement the NM interface
func (nmi *NMItem) Len() int {
	return 0
}

// XMLElement is specially crafted to be automatically marshalled by encoding/xml
type XMLElement struct {
	XMLName    xml.Name
	Value      string        `xml:",chardata"`
	Attributes []*xml.Attr   `xml:",attr"`
	Elements   []*XMLElement `xml:"omitempty"`
}

// NMAsXMLElements returns the values as []*XMLElement which can be later marshaled
// considers each value returned by .Values() in the form of []*NMItem, otherwise errors
func NMAsXMLElements(nm *utils.OrderedNavigableMap) (ents []*XMLElement, err error) {
	pathIdx := make(map[string]*XMLElement) // Keep the index of elements based on path
	for el := nm.GetFirstElement(); el != nil; el = el.Next() {
		path := el.Value
		var nmIt utils.NMInterface
		nmIt, _ = nm.Field(path) // this should never return error cause we get the path from the order
		nmItm, isNMItem := nmIt.(*NMItem)
		if !isNMItem {
			return nil, fmt.Errorf("value: %+v is not []*NMItem", path)
		}
		if nmItm.Config != nil && nmItm.Config.NewBranch {
			pathIdx = make(map[string]*XMLElement) // reset cache so we can start having other elements with same path
		}
		val := utils.IfaceAsString(nmItm.Data)
		var pathCached bool
		for i := len(nmItm.Path); i > 0; i-- {
			var cachedElm *XMLElement
			if cachedElm, pathCached = pathIdx[strings.Join(nmItm.Path[:i], "")]; !pathCached {
				continue
			}
			if i == len(nmItm.Path) { // lastElmnt, overwrite value or add attribute
				if nmItm.Config != nil && nmItm.Config.AttributeID != "" {
					cachedElm.Attributes = append(cachedElm.Attributes,
						&xml.Attr{
							Name:  xml.Name{Local: nmItm.Config.AttributeID},
							Value: val,
						})
				} else {
					cachedElm.Value = val
				}
				break
			}
			// create elements in reverse order so we can append already created
			var newElm *XMLElement
			for j := len(nmItm.Path); j > i; j-- {
				elm := &XMLElement{XMLName: xml.Name{Local: nmItm.Path[j-1]}}
				pathIdx[strings.Join(nmItm.Path[:j], "")] = elm
				if newElm == nil {
					if nmItm.Config != nil &&
						nmItm.Config.AttributeID != "" {
						elm.Attributes = append(elm.Attributes,
							&xml.Attr{
								Name:  xml.Name{Local: nmItm.Config.AttributeID},
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
			for i := len(nmItm.Path); i > 0; i-- {
				elm := &XMLElement{XMLName: xml.Name{Local: nmItm.Path[i-1]}}
				pathIdx[strings.Join(nmItm.Path[:i], "")] = elm
				if newElm == nil { // last element, create data inside
					if nmItm.Config != nil &&
						nmItm.Config.AttributeID != "" {
						elm.Attributes = append(elm.Attributes,
							&xml.Attr{
								Name:  xml.Name{Local: nmItm.Config.AttributeID},
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

// NMAsCGREvent builds a CGREvent considering Time as time.Now()
// and Event as linear map[string]interface{} with joined paths
// treats particular case when the value of map is []*NMItem - used in agents/AgentRequest
func NMAsCGREvent(nM *utils.OrderedNavigableMap, tnt string, pathSep string, opts *utils.OrderedNavigableMap) (cgrEv *utils.CGREvent) {
	if nM == nil {
		return
	}
	el := nM.GetFirstElement()
	if el == nil {
		return
	}
	cgrEv = &utils.CGREvent{
		Tenant: tnt,
		ID:     utils.UUIDSha1Prefix(),
		Time:   utils.TimePointer(time.Now()),
		Event:  make(map[string]interface{}),
		Opts:   NMAsMapInterface(opts, pathSep),
	}
	for ; el != nil; el = el.Next() {
		branchPath := el.Value
		val, _ := nM.Field(branchPath) // this should never return error cause we get the path from the order
		opath := utils.GetPathWithoutIndex(branchPath.String())
		if nmItm, isNMItem := val.(*NMItem); isNMItem { // special case when we have added multiple items inside a key, used in agents
			if nmItm.Config != nil &&
				nmItm.Config.AttributeID != "" {
				continue
			}
		}
		if _, has := cgrEv.Event[opath]; !has {
			cgrEv.Event[opath] = val.Interface() // first item which is not an attribute will become the value
		}
	}
	return
}

// NMAsMapInterface builds a linear map[string]interface{} with joined paths
// treats particular case when the value of map is []*NMItem - used in agents/AgentRequest
func NMAsMapInterface(nM *utils.OrderedNavigableMap, pathSep string) (mp map[string]interface{}) {
	mp = make(map[string]interface{})
	el := nM.GetFirstElement()
	if el == nil {
		return
	}
	for ; el != nil; el = el.Next() {
		branchPath := el.Value
		val, _ := nM.Field(branchPath) // this should never return error cause we get the path from the order
		opath := utils.GetPathWithoutIndex(branchPath.String())
		if nmItm, isNMItem := val.(*NMItem); isNMItem { // special case when we have added multiple items inside a key, used in agents
			if nmItm.Config != nil &&
				nmItm.Config.AttributeID != "" {
				continue
			}
		}
		if _, has := mp[opath]; !has {
			mp[opath] = val.Interface() // first item which is not an attribute will become the value
		}
	}
	return
}
