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

// NewNMInterface returns the interface wraped in NMInterface struture
func NewNMInterface(val interface{}) *NMInterface { return &NMInterface{data: val} }

// NMInterface most basic NM structure
type NMInterface struct{ data interface{} }

func (nmi *NMInterface) String() string {
	return IfaceAsString(nmi.data)
}

// Interface returns the wraped interface
func (nmi *NMInterface) Interface() interface{} {
	return nmi.data
}

// Field not implemented only used in order to implement the NM interface
func (nmi *NMInterface) Field(path PathItems) (val NM, err error) {
	return nil, ErrNotImplemented
}

// Set not implemented only used in order to implement the NM interface
// special case when the path is empty the interface should be seted
// this is in order to modify the wraped interface
func (nmi *NMInterface) Set(path PathItems, val NM) (err error) {
	if len(path) != 0 {
		return ErrWrongPath
	}
	nmi.data = val.Interface()
	return
}

// Remove not implemented only used in order to implement the NM interface
func (nmi *NMInterface) Remove(path PathItems) (err error) {
	return ErrNotImplemented
}

// Type returns the type of the NM interface
func (nmi *NMInterface) Type() NMType {
	return NMInterfaceType
}

// Empty returns true if the NM is empty(no data)
func (nmi *NMInterface) Empty() bool {
	return nmi == nil || nmi.data == nil
}

// GetField not implemented only used in order to implement the NM interface
func (nmi *NMInterface) GetField(path PathItem) (val NM, err error) {
	return nil, ErrNotImplemented
}

// SetField not implemented only used in order to implement the NM interface
// special case when the path is empty the interface should be seted
// this is in order to modify the wraped interface
func (nmi *NMInterface) SetField(path PathItem, val NM) (err error) {
	// if path != nil {
	// 	return ErrWrongPath
	// }
	nmi.data = val.Interface()
	return
}

// Len not implemented only used in order to implement the NM interface
func (nmi *NMInterface) Len() int {
	return 0
}
