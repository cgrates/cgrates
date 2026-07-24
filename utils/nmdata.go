// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package utils

// NewNMData returns the interface wraped in NMInterface struture
func NewNMData(val any) *NMData { return &NMData{data: val} }

// NMData most basic NM structure
type NMData struct{ data any }

func (nmi *NMData) String() string {
	return IfaceAsString(nmi.data)
}

// Interface returns the wraped interface
func (nmi *NMData) Interface() any {
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
