package utils

import (
	"errors"
	"fmt"
	"net"
	"reflect"
	"strconv"
	"strings"
)

// NewNavigableMap constructs a NavigableMapOld1
func NewNavigableMapOld1(data map[string]interface{}) *NavigableMapOld1 {
	if data == nil {
		data = make(map[string]interface{})
	}
	return &NavigableMapOld1{data: data}
}

// NMItem is an item in the NavigableMapOld1
type NMItem struct {
	Path []string    // path in map
	Data interface{} // value of the element
}

// NavigableMapOld1 is the current implementation of navigableMap
// this is used only by benchmarks
type NavigableMapOld1 struct {
	data  map[string]interface{} // layered map
	order [][]string             // order of field paths
}

// Set will set items into NavigableMapOld1 populating also order
// apnd parameter allows appending the data if both sides are []*NMItem
func (nM *NavigableMapOld1) Set(path []string, data interface{}, apnd bool) {
	nM.order = append(nM.order, path)
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

// GetField returns a field in it's original format, without converting from ie. *NMItem
func (nM *NavigableMapOld1) GetField(path []string) (fldVal interface{}, err error) {
	lenPath := len(path)
	if lenPath == 0 {
		return nil, errors.New("empty field path")
	}
	lastMp := nM.data // last map when layered
	for i, spath := range path {
		if i == lenPath-1 { // lastElement
			return nM.getLastRealItem(lastMp, spath)
		}
		var dp interface{}
		if dp, err = nM.getNextMap(lastMp, spath); err != nil {
			return
		}
		switch mv := dp.(type) { // used for cdr when populating eventCost whitin
		case map[string]interface{}:
			lastMp = mv
		case DataProvider:
			return mv.FieldAsInterface(path[i+1:])
		default:
			return nil, fmt.Errorf("cannot cast field: <%+v> type: %T with path: <%s> to map[string]interface{}",
				dp, dp, spath)
		}
	}
	panic("BUG") // should never make it here
}

// getLastItem returns the item from the map
// checking if it needs to return the item or an element of him if the item is a slice
func (nM *NavigableMapOld1) getLastRealItem(mp map[string]interface{}, spath string) (val interface{}, err error) {
	var idx *int
	spath, idx = nM.getIndex(spath)
	var has bool
	val, has = mp[spath]
	if !has {
		return nil, ErrNotFound
	}
	if idx == nil {
		return val, nil
	}
	switch vt := val.(type) {
	case []string:
		if *idx > len(vt) {
			return nil, fmt.Errorf("selector index %d out of range", *idx)
		}
		return vt[*idx], nil
	default:
	}
	// only if all above fails use reflect:
	vr := reflect.ValueOf(val)
	if vr.Kind() == reflect.Ptr {
		vr = vr.Elem()
	}
	if vr.Kind() != reflect.Slice && vr.Kind() != reflect.Array {
		return nil, fmt.Errorf("selector index used on non slice type(%T)", val)
	}
	if *idx > vr.Len() {
		return nil, fmt.Errorf("selector index %d out of range", *idx)
	}
	return vr.Index(*idx).Interface(), nil
}

// FieldAsInterface returns the field value as interface{} for the path specified
// implements DataProvider
// supports spath with selective elements in case of []*NMItem
func (nM *NavigableMapOld1) FieldAsInterface(fldPath []string) (fldVal interface{}, err error) {
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
func (nM *NavigableMapOld1) getLastItem(mp map[string]interface{}, spath string) (val interface{}, err error) {
	var idx *int
	spath, idx = nM.getIndex(spath)
	var has bool
	val, has = mp[spath]
	if !has {
		return nil, ErrNotFound
	}
	if idx == nil {
		return val, nil
	}
	switch vt := val.(type) {
	case []string:
		if *idx > len(vt) {
			return nil, ErrNotFound
			// return nil, fmt.Errorf("selector index %d out of range", *idx)
		}
		return vt[*idx], nil
	case []*NMItem:
		if *idx > len(vt) {
			return nil, ErrNotFound
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
		return nil, ErrNotFound
		// return nil, fmt.Errorf("selector index used on non slice type(%T)", val)
	}
	if *idx > vr.Len() {
		return nil, ErrNotFound
		// return nil, fmt.Errorf("selector index %d out of range", *idx)
	}
	return vr.Index(*idx).Interface(), nil
}

// getNextMap returns the next map from the given map
// used only for path parsing
func (nM *NavigableMapOld1) getNextMap(mp map[string]interface{}, spath string) (interface{}, error) {
	var idx *int
	spath, idx = nM.getIndex(spath)
	mi, has := mp[spath]
	if !has {
		return nil, ErrNotFound
	}
	if idx == nil {
		switch mv := mi.(type) {
		case map[string]interface{}:
			return mv, nil
		case *map[string]interface{}:
			return *mv, nil
		case NavigableMapOld1:
			return mv.data, nil
		case *NavigableMapOld1:
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
				case NavigableMapOld1:
					return mmv.data, nil
				case *NavigableMapOld1:
					return mmv.data, nil
				default:
				}
			}
		case []map[string]interface{}:
			if *idx < len(mv) {
				return mv[*idx], nil
			}
		case []NavigableMapOld1:
			if *idx < len(mv) {
				return mv[*idx].data, nil
			}
		case []*NavigableMapOld1:
			if *idx < len(mv) {
				return mv[*idx].data, nil
			}
		case []DataProvider: // used for cdr when populating eventCost whitin
			if *idx < len(mv) {
				return mv[*idx], nil
			}
		default:
		}
		return nil, ErrNotFound // xml compatible
	}
	return nil, fmt.Errorf("cannot cast field: <%+v> type: %T with path: <%s> to map[string]interface{}",
		mi, mi, spath)
}

// getIndex returns the path and index if index present
// path[index]=>path,index
// path=>path,nil
func (nM *NavigableMapOld1) getIndex(spath string) (opath string, idx *int) {
	idxStart := strings.Index(spath, IdxStart)
	if idxStart == -1 || !strings.HasSuffix(spath, IdxEnd) {
		return spath, nil
	}
	slctr := spath[idxStart+1 : len(spath)-1]
	opath = spath[:idxStart]
	if strings.HasPrefix(slctr, DynamicDataPrefix) {
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
func (nM *NavigableMapOld1) FieldAsString(fldPath []string) (fldVal string, err error) {
	var valIface interface{}
	valIface, err = nM.FieldAsInterface(fldPath)
	if err != nil {
		return
	}
	return IfaceAsString(valIface), nil
}

// String is part of engine.DataProvider interface
func (nM *NavigableMapOld1) String() string {
	return ToJSON(nM.data)
}

// RemoteHost is part of engine.DataProvider interface
func (nM *NavigableMapOld1) RemoteHost() net.Addr {
	return LocalAddr()
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
func (nM *NavigableMapOld1) Values() (vals []interface{}) {
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

// Merge will update nM with values from a second one
func (nM *NavigableMapOld1) Merge(nM2 *NavigableMapOld1) {
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
			pathStr := strings.Join(path, NestingSep)
			pathIdx[pathStr]++
			if pathIdx[pathStr] > len(valItms) {
				val = valItms[len(valItms)-1:] // slice with only last element in, so we can set it unlimited
			} else {
				val = []*NMItem{valItms[pathIdx[pathStr]-1]} // set only one item per path
			}
		}
		nM.Set(path, val, true)
	}
	return
}

// indexMapPaths parses map returning the parsed branchPath, useful when not having order for NavigableMapOld1
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
