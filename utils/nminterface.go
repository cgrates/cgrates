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

func NewNMInterface(val interface{}) *NMInterface { return &NMInterface{data: val} }

type NMInterface struct{ data interface{} }

func (nmi *NMInterface) String() string         { return IfaceAsString(nmi.data) }
func (nmi *NMInterface) Interface() interface{} { return nmi.data }
func (nmi *NMInterface) Field(path PathItems) (val NM, err error) {
	return nil, ErrNotImplemented
}
func (nmi *NMInterface) Set(path PathItems, val NM) (err error) {
	return ErrNotImplemented
}
func (nmi *NMInterface) Remove(path PathItems) (err error) {
	return ErrNotImplemented
}
func (nmi *NMInterface) Type() NMType { return NMInterfaceType }
func (nmi *NMInterface) Empty() bool  { return nmi == nil || nmi.data == nil }

func (nmi *NMInterface) GetField(path *PathItem) (val NM, err error) { return nil, ErrNotImplemented }

func (nmi *NMInterface) SetField(path *PathItem, val NM) (err error) { return ErrNotImplemented }

func (nmi *NMInterface) Len() int { return 0 }
