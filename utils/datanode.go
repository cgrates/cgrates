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

package utils

import (
	"strconv"
	"strings"
)

// CompilePathSlice returns the path as a slice accepted by DataNode structure
// field1[field2][index], field3
// will become:
// field1.field2.index.field3
func CompilePathSlice(spath []string) (path []string) {
	path = make([]string, 0, len(spath))
	for _, p := range spath {
		idxStart := strings.Index(p, IdxStart)
		if idxStart == -1 || !strings.HasSuffix(p, IdxEnd) {
			path = append(path, p)
			continue
		}
		path = append(path, p[:idxStart])
		path = append(path, strings.Split(p[idxStart+1:len(p)-1], IdxCombination)...)
	}
	return
}
func CompilePath(spath string) (path []string) {
	return CompilePathSlice(strings.Split(spath, NestingSep))
}

func NewDataNode(path []string) (n *DataNode) {
	n = new(DataNode)
	if len(path) == 0 { // is most probably a leaf
		return
	}
	obj := NewDataNode(path[1:])
	if path[0] == "0" { // only support the 0 index when creating new array
		n.Type = NMSliceType
		n.Slice = []*DataNode{obj}
		return
	}
	n.Type = NMMapType // consider it a map otherwise
	n.Map = map[string]*DataNode{path[0]: obj}
	return
}

func NewLeafNode(val interface{}) *DataNode {
	return &DataNode{
		Type: NMDataType,
		Value: &DataLeaf{
			Data: val,
		},
	}
}

// DataNode a structure for storing data
type DataNode struct {
	Type  NMType               `json:"-"`
	Map   map[string]*DataNode `json:",omitempty"`
	Slice []*DataNode          `json:",omitempty"`
	Value *DataLeaf            `json:",omitempty"`
}

// DataLeaf is an item in the DataNode
type DataLeaf struct {
	Data        interface{} // value of the element
	Path        []string    // path in map
	NewBranch   bool
	AttributeID string
	// Config *FCTemplate // so we can store additional configuration
}

func (dl *DataLeaf) String() string {
	return IfaceAsString(dl.Data)
}

// Field returns the value found at path
// the path equivalent for:
// field1[field2][index].field3
// should be of following form:
// field1.field2.index.field3
func (n *DataNode) Field(path []string) (*DataLeaf, error) {
	switch n.Type { // based on current type return the value
	case NMDataType:
		if len(path) != 0 { // only return if the path is empty
			return nil, ErrNotFound
		}
		return n.Value, nil
	case NMMapType:
		if len(path) == 0 {
			return nil, ErrWrongPath
		}
		node, has := n.Map[path[0]]
		if !has {
			return nil, ErrNotFound
		}
		return node.Field(path[1:]) // let the next node handle the value
	case NMSliceType:
		if len(path) == 0 {
			return nil, ErrWrongPath
		}
		idx, err := strconv.Atoi(path[0]) // convert the path to index
		if err != nil {
			return nil, err
		}
		if idx < 0 { // in case the index is negative add the slice lenght
			idx += len(n.Slice)
		}
		if idx < 0 || idx >= len(n.Slice) { // check if the index is in range [0,len(slice))
			return nil, ErrNotFound
		}
		return n.Slice[idx].Field(path[1:])
	}
	// this is possible if the node was created but no value was assigned to it
	return nil, ErrWrongPath
}

func (n *DataNode) FieldAsInterface(path []string) (interface{}, error) {
	return n.fieldAsInterface(CompilePathSlice(path))
}
func (n *DataNode) fieldAsInterface(path []string) (interface{}, error) {
	switch n.Type { // based on current type return the value
	case NMDataType:
		if len(path) != 0 { // only return if the path is empty
			return nil, ErrNotFound
		}
		return n.Value.Data, nil
	case NMMapType:
		if len(path) == 0 {
			return n.Map, nil
		}
		node, has := n.Map[path[0]]
		if !has {
			return nil, ErrNotFound
		}
		return node.fieldAsInterface(path[1:]) // let the next node handle the value
	case NMSliceType:
		if len(path) == 0 {
			return n.Slice, nil
		}
		idx, err := strconv.Atoi(path[0]) // convert the path to index
		if err != nil {
			return nil, err
		}
		if idx < 0 { // in case the index is negative add the slice lenght
			idx += len(n.Slice)
		}
		if idx < 0 || idx >= len(n.Slice) { // check if the index is in range [0,len(slice))
			return nil, ErrNotFound
		}
		return n.Slice[idx].fieldAsInterface(path[1:])
	}
	// this is possible if the node was created but no value was assigned to it
	return nil, ErrWrongPath
}

// Set will set the value at de specified path
// the path should be in the same format as the path given to Field
func (n *DataNode) Set(path []string, val interface{}) (addedNew bool, err error) {
	if len(path) == 0 { // the path is empty so overwrite curent node data
		switch v := val.(type) { // cast the value to see if is a supported type for node
		case map[string]*DataNode:
			n.Type = NMMapType
			n.Map = v
		case []*DataNode:
			n.Type = NMSliceType
			n.Slice = v
		case *DataLeaf:
			n.Type = NMDataType
			n.Value = v
		case *DataNode:
			n.Type = v.Type
			n.Map = v.Map
			n.Slice = v.Slice
			n.Value = v.Value
		default:
			n.Type = NMDataType
			n.Value = &DataLeaf{
				Data: val,
			}
		}
		return
	}
	switch n.Type {
	case NMDataType:
		return false, ErrWrongPath
	case NMMapType:
		node, has := n.Map[path[0]]
		if !has { // create the node if not exists
			node = NewDataNode(path[1:])
			n.Map[path[0]] = node
		}
		addedNew, err = node.Set(path[1:], val) // set the value in the node
		return addedNew || !has, err
	case NMSliceType:
		idx, err := strconv.Atoi(path[0])
		if err != nil {
			return false, err
		}
		if idx == len(n.Slice) { // special case when the index is the length so append
			node := NewDataNode(path[1:])
			n.Slice = append(n.Slice, node)
			_, err = node.Set(path[1:], val)
			return true, err
		}
		if idx < 0 { // in case the index is negative add the slice lenght
			idx += len(n.Slice)
			path[0] = strconv.Itoa(idx) // update the slice to reflect on orderNavMap
		}
		if idx < 0 || idx >= len(n.Slice) { // check if the index is in range [0,len(slice))
			return false, ErrNotFound
		}
		return n.Slice[idx].Set(path[1:], val)
	}
	// this is possible if the node was created but no value was assigned to it
	return false, ErrWrongPath
}

// IsEmpty return if the node is empty/ has no data
func (n DataNode) IsEmpty() bool {
	return n.Value == nil ||
		len(n.Map) == 0 ||
		len(n.Slice) == 0
}

// Remove will remove the value at given path
// the path should be in the same format as the path given to Field
// also removes the empty nodes found in path
func (n *DataNode) Remove(path []string) error {
	if len(path) == 0 { // the path is empty so make the node empty
		n.Map = nil
		n.Slice = nil
		n.Value = nil
		return nil
	}
	switch n.Type {
	case NMDataType: // no remove for data type
		return ErrWrongPath
	case NMMapType:
		node, has := n.Map[path[0]]
		if !has { // the element doesn't exist so ignore
			return nil
		}
		err := node.Remove(path[1:])
		if node.IsEmpty() { // remove the element if empty
			delete(n.Map, path[0])
		}
		return err
	case NMSliceType:
		idx, err := strconv.Atoi(path[0])
		if err != nil {
			return err // the only error is when we expect an index but is not int
		}
		if idx < 0 {
			idx += len(n.Slice)
			path[0] = strconv.Itoa(idx) // update the path for OrdNavMap
		}
		if idx < 0 || idx >= len(n.Slice) { // the index is not in range so ignore
			return nil
		}
		err = n.Slice[idx].Remove(path[1:])
		if n.Slice[idx].IsEmpty() { // remove the element if empty
			n.Slice = append(n.Slice[:idx], n.Slice[idx+1:]...)
		}
		return err
	}
	// this is possible if the node was created but no value was assigned to it
	return ErrWrongPath
}

// Append will append the value at de specified path
// the path should be in the same format as the path given to Field
func (n *DataNode) Append(path []string, val *DataLeaf) (idx int, err error) {
	if len(path) == 0 { // the path is empty so overwrite curent node data
		switch n.Type {
		case NMMapType:
			return -1, ErrWrongPath
		case NMDataType:
			if n.Value != nil && n.Value.Data != nil {
				return -1, ErrWrongPath
			}
			// is empty so make a slice to be compatible with append
			n.Type = NMSliceType
			n.Value = nil
			n.Slice = []*DataNode{{Type: NMDataType, Value: val}}
			return 0, nil
		default:
			n.Type = NMSliceType
			n.Slice = append(n.Slice, &DataNode{Type: NMDataType, Value: val})
			return len(n.Slice) - 1, nil
		}
	}
	switch n.Type {
	case NMDataType:
		return -1, ErrWrongPath
	case NMMapType:
		node, has := n.Map[path[0]]
		if !has { // create the node if not exists
			node = NewDataNode(path[1:])
			n.Map[path[0]] = node
		}
		return node.Append(path[1:], val) // set the value in the node
	case NMSliceType:
		idx, err := strconv.Atoi(path[0])
		if err != nil {
			return -1, err
		}
		if idx == len(n.Slice) { // special case when the index is the length so append
			node := NewDataNode(path[1:])
			n.Slice = append(n.Slice, node)
			return node.Append(path[1:], val)
		}
		if idx < 0 { // in case the index is negative add the slice lenght
			idx += len(n.Slice)
			path[0] = strconv.Itoa(idx) // update the slice to reflect on orderNavMap
		}
		if idx < 0 || idx >= len(n.Slice) { // check if the index is in range [0,len(slice))
			return -1, ErrNotFound
		}
		return n.Slice[idx].Append(path[1:], val)
	}
	// this is possible if the node was created but no value was assigned to it
	return -1, ErrWrongPath
}

// Compose will set the value at de specified path
// the path should be in the same format as the path given to Field
func (n *DataNode) Compose(path []string, val *DataLeaf) (err error) {
	if len(path) == 0 { // the path is empty so overwrite curent node data
		switch n.Type {
		case NMMapType:
			return ErrWrongPath
		case NMDataType:
			if n.Value == nil || n.Value.Data == nil {
				// is empty so make a slice to be compatible with append
				n.Type = NMSliceType
				n.Value = nil
				n.Slice = []*DataNode{{Type: NMDataType, Value: val}}
				return
			}
			n.Value.Data = n.Value.String() + val.String()
		default:
			if len(n.Slice) == 0 {
				n.Type = NMSliceType
				n.Slice = []*DataNode{{Type: NMDataType, Value: val}}
				return
			}
			n.Slice[len(n.Slice)-1].Value.Data = n.Slice[len(n.Slice)-1].Value.String() + val.String()
		}
		return
	}
	switch n.Type {
	case NMDataType:
		return ErrWrongPath
	case NMMapType:
		node, has := n.Map[path[0]]
		if !has { // create the node if not exists
			node = NewDataNode(path[1:])
			n.Map[path[0]] = node
		}
		return node.Compose(path[1:], val) // set the value in the node
	case NMSliceType:
		idx, err := strconv.Atoi(path[0])
		if err != nil {
			return err
		}
		if idx == len(n.Slice) { // special case when the index is the length so append
			node := NewDataNode(path[1:])
			n.Slice = append(n.Slice, node)
			return node.Compose(path[1:], val)
		}
		if idx < 0 { // in case the index is negative add the slice lenght
			idx += len(n.Slice)
			path[0] = strconv.Itoa(idx) // update the slice to reflect on orderNavMap
		}
		if idx < 0 || idx >= len(n.Slice) { // check if the index is in range [0,len(slice))
			return ErrNotFound
		}
		return n.Slice[idx].Compose(path[1:], val)
	}
	// this is possible if the node was created but no value was assigned to it
	return ErrWrongPath
}
