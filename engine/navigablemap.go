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

package engine

import (
	"encoding/xml"
	"errors"
	"fmt"
	"strings"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

// CGRReplier is the interface supported by replies convertible to CGRReply
type NavigableMapper interface {
	AsNavigableMap([]*config.CfgCdrField) (*NavigableMap, error)
}

// NewNavigableMap constructs a NavigableMap
func NewNavigableMap(data map[string]interface{}) *NavigableMap {
	if data == nil {
		data = make(map[string]interface{})
	}
	return &NavigableMap{data: data}
}

// NMItem is an item in the NavigableMap
type NMItem struct {
	Path   []string            // path in map
	Data   interface{}         // value of the element
	Config *config.CfgCdrField // so we can store additional configuration
}

// NavigableMap is a map who's values can be navigated via path
// data can be retrieved as ordered
// NavigableMap is not thread safe due to performance demands, could come back if needed
type NavigableMap struct {
	data  map[string]interface{} // layered map
	order [][]string             // order of field paths
}

// Add will add items into NavigableMap populating also order
func (nM *NavigableMap) Set(itm *NMItem, ordered bool) {
	mp := nM.data
	for i, spath := range itm.Path {
		if i == len(itm.Path)-1 { // last path
			mp[spath] = itm
			return
		}
		if _, has := mp[spath]; !has {
			mp[spath] = make(map[string]interface{})
		}
		mp = mp[spath].(map[string]interface{}) // so we can check further down
	}
	if ordered {
		nM.order = append(nM.order, itm.Path)
	}
}

// FieldAsInterface returns the field value as interface{} for the path specified
// implements DataProvider
func (nM *NavigableMap) FieldAsInterface(fldPath []string) (fldVal interface{}, err error) {
	lenPath := len(fldPath)
	if lenPath == 0 {
		return nil, errors.New("empty field path")
	}
	lastMp := nM.data // last map when layered
	var canCast bool
	for i, spath := range fldPath {
		if i == lenPath-1 { // lastElement
			itmIface, has := lastMp[spath]
			if !has {
				return nil, utils.ErrNotFound
			}
			if itm, cast := itmIface.(*NMItem); cast {
				return itm.Data, nil
			}
			return itmIface, nil
		} else {
			elmnt, has := lastMp[spath]
			if !has {
				err = fmt.Errorf("no map at path: <%s>", spath)
				return
			}
			lastMp, canCast = elmnt.(map[string]interface{})
			if !canCast {
				lastMpNM, canCast := elmnt.(*NavigableMap) // attempt to cast into NavigableMap
				if !canCast {
					err = fmt.Errorf("cannot cast field: <%+v> type: %T with path: <%s> to map[string]interface{}",
						elmnt, elmnt, spath)
					return
				}
				lastMp = lastMpNM.data
			}
		}
	}
	err = errors.New("end of function")
	return
}

// FieldAsString returns the field value as string for the path specified
// implements DataProvider
func (nM *NavigableMap) FieldAsString(fldPath []string) (fldVal string, err error) {
	var valIface interface{}
	valIface, err = nM.FieldAsInterface(fldPath)
	if err != nil {
		return
	}
	var canCast bool
	if fldVal, canCast = utils.CastFieldIfToString(valIface); !canCast {
		return "", fmt.Errorf("cannot cast field: %s to string", utils.ToJSON(valIface))
	}
	return
}

func (nM *NavigableMap) String() string {
	return utils.ToJSON(nM.data)
}

// AsMapStringInterface returns the data out of map, ignoring the order part
func (nM *NavigableMap) AsMapStringInterface() map[string]interface{} {
	return nM.data
}

// indexMapElements will recursively go through map and index the element paths into elmns
func indexMapElements(mp map[string]interface{}, path []string, elms *[]*NMItem) {
	for k, v := range mp {
		vPath := append(path, k)
		if mpIface, isMap := v.(map[string]interface{}); isMap {
			indexMapElements(mpIface, vPath, elms)
			continue
		}
		var elmsOut []*NMItem
		if nMItem, isNMItem := v.(*NMItem); isNMItem {
			elmsOut = append(*elms, nMItem)
		} else {
			elmsOut = append(*elms, &NMItem{Path: vPath, Data: v})
		}

		*elms = elmsOut
	}
}

// Items returns the items in map, ordered by order information
func (nM *NavigableMap) Items() (itms []*NMItem) {
	if len(nM.data) == 0 {
		return
	}
	if len(nM.order) == 0 {
		indexMapElements(nM.data, []string{}, &itms)
		return
	}
	itms = make([]*NMItem, len(nM.order))
	for i, path := range nM.order {
		val, _ := nM.FieldAsInterface(path)
		itms[i] = &NMItem{Data: val, Path: path}
	}
	return
}

// AsNavigableMap implements both NavigableMapper as well as DataProvider interfaces
func (nM *NavigableMap) AsNavigableMap(tpl []*config.CfgCdrField) (oNM *NavigableMap, err error) {
	return nil, utils.ErrNotImplemented
}

// Merge will update nM with values from a second one
func (nM *NavigableMap) Merge(nM2 *NavigableMap) {
	if nM2 == nil {
		return
	}
	for k, v := range nM2.data {
		nM.data[k] = v
	}
	if len(nM2.order) != 0 {
		nM.order = append(nM.order, nM2.order...)
	}
}

// MarshalXML implements xml.Marshaler
func (nM *NavigableMap) MarshalXML(e *xml.Encoder, start xml.StartElement) (err error) {
	tokens := []xml.Token{start}
	for _, itm := range nM.Items() {
		strVal, canCast := utils.CastFieldIfToString(itm.Data)
		if !canCast {
			return fmt.Errorf("cannot cast field: %s to string", utils.ToJSON(itm.Data))
		}
		t := xml.StartElement{Name: xml.Name{"", strings.Join(itm.Path, ">")}}
		tokens = append(tokens, t, xml.CharData([]byte(strVal)), xml.EndElement{t.Name})
	}

	tokens = append(tokens, xml.EndElement{start.Name})

	for _, t := range tokens {
		err := e.EncodeToken(t)
		if err != nil {
			return err
		}
	}

	// flush to ensure tokens are written
	return e.Flush()
	if err != nil {
		return err
	}

	return nil
}
