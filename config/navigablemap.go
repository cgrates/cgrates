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
	"errors"
	"fmt"
	"net"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/cgrates/cgrates/utils"
)

// CGRReplier is the interface supported by replies convertible to CGRReply
type NavigableMapper interface {
	AsNavigableMap([]*FCTemplate) (*NavigableMap, error)
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
	Path   []string    // path in map
	Data   interface{} // value of the element
	Config *FCTemplate // so we can store additional configuration
}

// NavigableMap is a map who's values can be navigated via path
// data can be retrieved as ordered
// NavigableMap is not thread safe due to performance demands, could come back if needed
type NavigableMap struct {
	data  map[string]interface{} // layered map
	order [][]string             // order of field paths
}

// Set will set items into NavigableMap populating also order
// apnd parameter allows appending the data if both sides are []*NMItem
func (nM *NavigableMap) Set(path []string, data interface{}, apnd, ordered bool) {
	if ordered {
		nM.order = append(nM.order, path)
	}
	mp := nM.data
	for i, spath := range path {
		if i == len(path)-1 { // last path
			oData, has := mp[spath]
			if !has || !apnd { // no need to append
				mp[spath] = data
				return
			}
			dataItms, isNMItems := data.([]*NMItem)
			if !isNMItems { // new data is not items
				mp[spath] = data
				return
			}
			oItms, isNMItems := oData.([]*NMItem)
			if !isNMItems { // previous data is not items, simply overwrite
				mp[spath] = data
				return
			}
			mp[spath] = append(oItms, dataItms...)
			return
		}
		if _, has := mp[spath]; !has {
			mp[spath] = make(map[string]interface{})
		}
		mp = mp[spath].(map[string]interface{}) // so we can check further down
	}
}

// FieldAsInterface returns the field value as interface{} for the path specified
// implements DataProvider
// supports spath with selective elements in case of []*NMItem
func (nM *NavigableMap) FieldAsInterface(fldPath []string) (fldVal interface{}, err error) {
	lenPath := len(fldPath)
	if lenPath == 0 {
		return nil, errors.New("empty field path")
	}
	lastMp := nM.data // last map when layered
	for i, spath := range fldPath {
		if i == lenPath-1 { // lastElement
			return nM.getLastItem(lastMp, spath)
		}
		var dp interface{}
		if dp, err = nM.getNextMap(lastMp, spath); err != nil {
			return
		}
		switch mv := dp.(type) { // used for cdr when populating eventCost whitin
		case map[string]interface{}:
			lastMp = mv
		case DataProvider:
			return mv.FieldAsInterface(fldPath[i+1:])
		default:
			return nil, fmt.Errorf("cannot cast field: <%+v> type: %T with path: <%s> to map[string]interface{}",
				dp, dp, spath)
		}
	}
	err = errors.New("end of function")
	return
}

// getLastItem returns the item from the map
// checking if it needs to return the item or an element of him if the item is a slice
func (nM *NavigableMap) getLastItem(mp map[string]interface{}, spath string) (val interface{}, err error) {
	var idx *int
	spath, idx = nM.getIndex(spath)
	var has bool
	val, has = mp[spath]
	if !has {
		return nil, utils.ErrNotFound
	}
	if idx == nil {
		return val, nil
	}
	switch vt := val.(type) {
	case []string:
		if *idx > len(vt) {
			return nil, utils.ErrNotFound
			// return nil, fmt.Errorf("selector index %d out of range", *idx)
		}
		return vt[*idx], nil
	case []*NMItem:
		if *idx > len(vt) {
			return nil, utils.ErrNotFound
			// return nil, fmt.Errorf("selector index %d out of range", *idx)
		}
		return vt[*idx].Data, nil
	default:
	}
	// only if all above fails use reflect:
	vr := reflect.ValueOf(val)
	if vr.Kind() == reflect.Ptr {
		vr = vr.Elem()
	}
	if vr.Kind() != reflect.Slice && vr.Kind() != reflect.Array {
		return nil, utils.ErrNotFound
		// return nil, fmt.Errorf("selector index used on non slice type(%T)", val)
	}
	if *idx > vr.Len() {
		return nil, utils.ErrNotFound
		// return nil, fmt.Errorf("selector index %d out of range", *idx)
	}
	return vr.Index(*idx).Interface(), nil
}

// getNextMap returns the next map from the given map
// used only for path parsing
func (nM *NavigableMap) getNextMap(mp map[string]interface{}, spath string) (interface{}, error) {
	var idx *int
	spath, idx = nM.getIndex(spath)
	mi, has := mp[spath]
	if !has {
		return nil, utils.ErrNotFound
	}
	if idx == nil {
		switch mv := mi.(type) {
		case map[string]interface{}:
			return mv, nil
		case *map[string]interface{}:
			return *mv, nil
		case NavigableMap:
			return mv.data, nil
		case *NavigableMap:
			return mv.data, nil
		case DataProvider: // used for cdr when populating eventCost whitin
			return mv, nil
		default:
		}
	} else {
		switch mv := mi.(type) {
		case []interface{}:
			// in case we create the map using json and we marshall the value into a map[string]interface{}
			// we can have slice of interfaces that is masking a slice of map[string]interface{}
			// this is for CostDetails BalanceSummaries
			if *idx < len(mv) {
				mm := mv[*idx]
				switch mmv := mm.(type) {
				case map[string]interface{}:
					return mmv, nil
				case *map[string]interface{}:
					return *mmv, nil
				case NavigableMap:
					return mmv.data, nil
				case *NavigableMap:
					return mmv.data, nil
				default:
				}
			}
		case []map[string]interface{}:
			if *idx < len(mv) {
				return mv[*idx], nil
			}
		case []NavigableMap:
			if *idx < len(mv) {
				return mv[*idx].data, nil
			}
		case []*NavigableMap:
			if *idx < len(mv) {
				return mv[*idx].data, nil
			}
		case []DataProvider: // used for cdr when populating eventCost whitin
			if *idx < len(mv) {
				return mv[*idx], nil
			}
		default:
		}
		return nil, utils.ErrNotFound // xml compatible
	}
	return nil, fmt.Errorf("cannot cast field: <%+v> type: %T with path: <%s> to map[string]interface{}",
		mi, mi, spath)
}

// getIndex returns the path and index if index present
// path[index]=>path,index
// path=>path,nil
func (nM *NavigableMap) getIndex(spath string) (opath string, idx *int) {
	idxStart := strings.Index(spath, utils.IdxStart)
	if idxStart == -1 || !strings.HasSuffix(spath, utils.IdxEnd) {
		return spath, nil
	}
	slctr := spath[idxStart+1 : len(spath)-1]
	opath = spath[:idxStart]
	if strings.HasPrefix(slctr, utils.DynamicDataPrefix) {
		return
	}
	idxVal, err := strconv.Atoi(slctr)
	if err != nil {
		return spath, nil
	}
	return opath, &idxVal
}

// FieldAsString returns the field value as string for the path specified
// implements DataProvider
func (nM *NavigableMap) FieldAsString(fldPath []string) (fldVal string, err error) {
	var valIface interface{}
	valIface, err = nM.FieldAsInterface(fldPath)
	if err != nil {
		return
	}
	return utils.IfaceAsString(valIface), nil
}

// String is part of engine.DataProvider interface
func (nM *NavigableMap) String() string {
	return utils.ToJSON(nM.data)
}

// RemoteHost is part of engine.DataProvider interface
func (nM *NavigableMap) RemoteHost() net.Addr {
	return utils.LocalAddr()
}

// indexMapElements will recursively go through map and index the element paths into elmns
func indexMapElements(mp map[string]interface{}, path []string, vals *[]interface{}) {
	for k, v := range mp {
		vPath := append(path, k)
		if mpIface, isMap := v.(map[string]interface{}); isMap {
			indexMapElements(mpIface, vPath, vals)
			continue
		}
		valsOut := append(*vals, v)
		*vals = valsOut
	}
}

// Values returns the values in map, ordered by order information
func (nM *NavigableMap) Values() (vals []interface{}) {
	if len(nM.data) == 0 {
		return
	}
	if len(nM.order) == 0 {
		indexMapElements(nM.data, []string{}, &vals)
		return
	}
	vals = make([]interface{}, len(nM.order))
	for i, path := range nM.order {
		val, _ := nM.FieldAsInterface(path)
		vals[i] = val
	}
	return
}

// AsNavigableMap implements both NavigableMapper as well as DataProvider interfaces
func (nM *NavigableMap) AsNavigableMap(
	tpl []*FCTemplate) (oNM *NavigableMap, err error) {
	return nil, utils.ErrNotImplemented
}

// Merge will update nM with values from a second one
func (nM *NavigableMap) Merge(nM2 *NavigableMap) {
	if nM2 == nil {
		return
	}
	if len(nM2.order) == 0 {
		indexMapPaths(nM2.data, nil, &nM.order)
	}
	pathIdx := make(map[string]int) // will hold references for last index exported in case of []*NMItem
	for _, path := range nM2.order {
		val, _ := nM2.FieldAsInterface(path)
		if valItms, isItms := val.([]*NMItem); isItms {
			pathStr := strings.Join(path, utils.NestingSep)
			pathIdx[pathStr]++
			if pathIdx[pathStr] > len(valItms) {
				val = valItms[len(valItms)-1:] // slice with only last element in, so we can set it unlimited
			} else {
				val = []*NMItem{valItms[pathIdx[pathStr]-1]} // set only one item per path
			}
		}
		nM.Set(path, val, true,
			(len(nM.order) != 0 || len(nM.data) == 0))
	}
	return
}

// indexMapPaths parses map returning the parsed branchPath, useful when not having order for NavigableMap
func indexMapPaths(mp map[string]interface{}, branchPath []string, parsedPaths *[][]string) {
	for k, v := range mp {
		if mpIface, isMap := v.(map[string]interface{}); isMap {
			indexMapPaths(mpIface, append(branchPath, k), parsedPaths)
			continue
		}
		tmpPaths := append(*parsedPaths, append(branchPath, k))
		*parsedPaths = tmpPaths
	}
}

// AsCGREvent builds a CGREvent considering Time as time.Now()
// and Event as linear map[string]interface{} with joined paths
// treats particular case when the value of map is []*NMItem - used in agents/AgentRequest
func (nM *NavigableMap) AsCGREvent(tnt string, pathSep string) (cgrEv *utils.CGREvent) {
	if nM == nil || len(nM.data) == 0 {
		return
	}
	cgrEv = &utils.CGREvent{
		Tenant: tnt,
		ID:     utils.UUIDSha1Prefix(),
		Time:   utils.TimePointer(time.Now()),
		Event:  make(map[string]interface{})}
	if len(nM.order) == 0 {
		indexMapPaths(nM.data, nil, &nM.order)
	}
	for _, branchPath := range nM.order {
		val, _ := nM.FieldAsInterface(branchPath)
		if nmItms, isNMItems := val.([]*NMItem); isNMItems { // special case when we have added multiple items inside a key, used in agents
			for _, nmItm := range nmItms {
				if nmItm.Config == nil ||
					nmItm.Config.AttributeID == "" {
					val = nmItm.Data // first item which is not an attribute will become the value
					break
				}
			}
		}
		cgrEv.Event[strings.Join(branchPath, pathSep)] = val
	}
	return
}

// AsMapStringIface builds a linear map[string]interface{} with joined paths
// treats particular case when the value of map is []*NMItem - used in agents/AgentRequest
func (nM *NavigableMap) AsMapStringIface(pathSep string) (mp map[string]interface{}) {
	if nM == nil || len(nM.data) == 0 {
		return
	}
	mp = make(map[string]interface{})
	if len(nM.order) == 0 {
		indexMapPaths(nM.data, nil, &nM.order)
	}
	for _, branchPath := range nM.order {
		val, _ := nM.FieldAsInterface(branchPath)
		if nmItms, isNMItems := val.([]*NMItem); isNMItems { // special case when we have added multiple items inside a key, used in agents
			for _, nmItm := range nmItms {
				if nmItm.Config == nil ||
					nmItm.Config.AttributeID == "" {
					val = nmItm.Data // first item which is not an attribute will become the value
					break
				}
			}
		}
		mp[strings.Join(branchPath, pathSep)] = val
	}
	return
}

// XMLElement is specially crafted to be automatically marshalled by encoding/xml
type XMLElement struct {
	XMLName    xml.Name
	Value      string        `xml:",chardata"`
	Attributes []*xml.Attr   `xml:",attr"`
	Elements   []*XMLElement `xml:"omitempty"`
}

// AsXMLElements returns the values as []*XMLElement which can be later marshaled
// considers each value returned by .Values() in the form of []*NMItem, otherwise errors
func (nM *NavigableMap) AsXMLElements() (ents []*XMLElement, err error) {
	pathIdx := make(map[string]*XMLElement) // Keep the index of elements based on path
	for _, val := range nM.Values() {
		nmItms, isNMItems := val.([]*NMItem)
		if !isNMItems {
			return nil, fmt.Errorf("value: %+v is not []*NMItem", val)
		}
		for _, nmItm := range nmItms {
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
					if nmItm.Config != nil &&
						nmItm.Config.AttributeID != "" {
						cachedElm.Attributes = append(cachedElm.Attributes,
							&xml.Attr{xml.Name{Local: nmItm.Config.AttributeID}, val})
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
								&xml.Attr{xml.Name{Local: nmItm.Config.AttributeID}, val})
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
								&xml.Attr{xml.Name{Local: nmItm.Config.AttributeID}, val})
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
	}
	return
}

func getPathFromValue(in reflect.Value, prefix string) (out []string) {
	switch in.Kind() {
	case reflect.Ptr:
		return getPathFromValue(in.Elem(), prefix)
	case reflect.Array, reflect.Slice:
		prefix = strings.TrimSuffix(prefix, utils.NestingSep)
		for i := 0; i < in.Len(); i++ {
			pref := fmt.Sprintf("%s[%v]", prefix, i)
			out = append(out, pref)
			out = append(out, getPathFromValue(in.Index(i), pref+utils.NestingSep)...)
		}
	case reflect.Map:
		iter := reflect.ValueOf(in).MapRange()
		for iter.Next() {
			pref := prefix + iter.Key().String()
			out = append(out, pref)
			out = append(out, getPathFromValue(iter.Value(), pref+utils.NestingSep)...)
		}
	case reflect.Struct:
		inType := in.Type()
		for i := 0; i < in.NumField(); i++ {
			pref := prefix + inType.Field(i).Name
			out = append(out, pref)
			out = append(out, getPathFromValue(in.Field(i), pref+utils.NestingSep)...)
		}
	case reflect.Invalid, reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr, reflect.Float32, reflect.Float64, reflect.Complex64, reflect.Complex128, reflect.String, reflect.Chan, reflect.Func, reflect.UnsafePointer, reflect.Interface:
	default:
	}
	return
}

func getPathFromInterface(in interface{}, prefix string) (out []string) {
	switch vin := in.(type) {
	case map[string]interface{}:
		for k, val := range vin {
			pref := prefix + k
			out = append(out, pref)
			out = append(out, getPathFromInterface(val, pref+utils.NestingSep)...)
		}
	case []map[string]interface{}:
		prefix = strings.TrimSuffix(prefix, utils.NestingSep)
		for i, val := range vin {
			pref := fmt.Sprintf("%s[%v]", prefix, i)
			out = append(out, pref)
			out = append(out, getPathFromInterface(val, pref+utils.NestingSep)...)
		}
	case []interface{}:
		prefix = strings.TrimSuffix(prefix, utils.NestingSep)
		for i, val := range vin {
			pref := fmt.Sprintf("%s[%v]", prefix, i)
			out = append(out, pref)
			out = append(out, getPathFromInterface(val, pref+utils.NestingSep)...)
		}
	case []string:
		prefix = strings.TrimSuffix(prefix, utils.NestingSep)
		for i := range vin {
			pref := fmt.Sprintf("%s[%v]", prefix, i)
			out = append(out, pref)
		}
	case nil, int, int32, int64, uint32, uint64, bool, float32, float64, []uint8, time.Duration, time.Time, string: //no path
	default: //reflect based
		out = getPathFromValue(reflect.ValueOf(vin), prefix)
	}
	return
}

// GetKeys returns all posibble keys
func (nM *NavigableMap) GetKeys() (keys []string) {
	for k, val := range nM.data {
		keys = append(keys, k)
		keys = append(keys, getPathFromInterface(val, k+utils.NestingSep)...)
	}
	return
}

// Remove will remove the items from the given path
func (nM *NavigableMap) Remove(path []string) {
	// if ordered {
	// 	nM.order = append(nM.order, path)
	// }
	mp := nM.data
	for i, spath := range path {
		oData, has := mp[spath]
		if !has { // no need to remove
			return
		}
		if i == len(path)-1 { // last path
			delete(mp, spath)
			return
		}
		defer func(np map[string]interface{}, p string) {
			o, has := np[p]
			if !has {
				return
			}
			if o == nil {
				delete(np, p)
				return
			}
			v, ok := o.(map[string]interface{})
			if !ok {
				return
			}
			if len(v) == 0 {
				delete(np, p)
			}
		}(mp, spath)
		mp = oData.(map[string]interface{}) // so we can check further down
	}
}
