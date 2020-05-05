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

// NewNMData returns the interface wraped in NMInterface struture
func NewNMData(val interface{}) *NMData { return &NMData{data: val} }

// NMData most basic NM structure
type NMData struct{ data interface{} }

func (nmi *NMData) String() string {
	return IfaceAsString(nmi.data)
}

// Interface returns the wraped interface
func (nmi *NMData) Interface() interface{} {
	return nmi.data
}

// Field is not implemented only used in order to implement the NM interface
func (nmi *NMData) Field(path PathItems) (val NMInterface, err error) {
	return nil, ErrNotImplemented
}

// Set sets the wraped interface when the path is empty
// This behaivior is in order to modify the wraped interface
// witout aserting the type of the NMInterface
func (nmi *NMData) Set(path PathItems, val NMInterface) (addedNew bool, err error) {
	if len(path) != 0 {
		return false, ErrWrongPath
	}
	nmi.data = val.Interface()
	return
}

// Remove is not implemented only used in order to implement the NM interface
func (nmi *NMData) Remove(path PathItems) (err error) {
	return ErrNotImplemented
}

// Type returns the type of the NM interface
func (nmi *NMData) Type() NMType {
	return NMDataType
}

// Empty returns true if the NM is empty(no data)
func (nmi *NMData) Empty() bool {
	return nmi == nil || nmi.data == nil
}

// Len is not implemented only used in order to implement the NM interface
func (nmi *NMData) Len() int {
	return 0
}
