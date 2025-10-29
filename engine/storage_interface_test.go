/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/

package engine

import (
	"reflect"
	"testing"

	"github.com/ugorji/go/codec"
)

func TestStorageInterfaceNewMarshaler(t *testing.T) {
	str := "json"

	rcv, err := NewMarshaler(str)

	exp := new(JSONMarshaler)
	if err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("expected %v, received %v", exp, rcv)
	}
}

func TestSorageInterfaceMarshal(t *testing.T) {
	type Test struct {
		Field string
	}
	arg := Test{
		Field: "test",
	}
	jm := BSONMarshaler{}

	rcv, err := jm.Marshal(&arg)
	if err != nil {
		t.Error(err)
	}

	exp := []byte{21, 0, 0, 0, 2, 102, 105, 101, 108, 100, 0, 5, 0, 0, 0, 116, 101, 115, 116, 0, 0}

	if !reflect.DeepEqual(exp, rcv) {
		t.Errorf("expected %v, received %v", exp, rcv)
	}
}

func TestStorageInterfaceUnmarshal(t *testing.T) {
	type Test struct {
		Field string
	}
	arg := Test{
		Field: "test",
	}
	jm := BSONMarshaler{}

	rcv, err := jm.Marshal(&arg)
	if err != nil {
		t.Error(err)
	}

	var um Test

	err = jm.Unmarshal(rcv, &um)
	if err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(um, arg) {
		t.Errorf("expected %v, received %v", arg, um)
	}
}

func TestStorageInterfaceMarshalBufMarshaler(t *testing.T) {
	type Test struct {
		Field string
	}
	arg := Test{
		Field: "test",
	}
	jbm := JSONBufMarshaler{}

	rcv, err := jbm.Marshal(arg)
	if err != nil {
		t.Error(err)
	}

	exp := []byte{123, 34, 70, 105, 101, 108, 100, 34, 58, 34, 116, 101, 115, 116, 34, 125, 10}

	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("expeted %v, received %v", exp, rcv)
	}
}

func TestStorageInterfaceUnmarshalBufMarshaler(t *testing.T) {
	type Test struct {
		Field string
	}
	arg := Test{
		Field: "test",
	}
	jbm := JSONBufMarshaler{}

	rcv, err := jbm.Marshal(arg)
	if err != nil {
		t.Error(err)
	}

	var um Test

	err = jbm.Unmarshal(rcv, &um)
	if err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(arg, um) {
		t.Errorf("expeted %v, received %v", arg, um)
	}
}

func TestStorageInterfaceNewBincMarshaler(t *testing.T) {
	rcv := NewBincMarshaler()

	exp := &BincMarshaler{new(codec.BincHandle)}

	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("expeted %v, received %v", exp, rcv)
	}
}

func TestStorageInterfaceMarshalBinc(t *testing.T) {
	type Test struct {
		Field string
	}
	arg := Test{
		Field: "test",
	}
	bm := NewBincMarshaler()

	rcv, err := bm.Marshal(arg)
	if err != nil {
		t.Error(err)
	}

	exp := []byte{117, 73, 70, 105, 101, 108, 100, 72, 116, 101, 115, 116}

	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("expeted %v, received %v", exp, rcv)
	}
}

func TestStorageInterfaceUnmarshalBinc(t *testing.T) {
	type Test struct {
		Field string
	}
	arg := Test{
		Field: "test",
	}
	bm := NewBincMarshaler()

	rcv, err := bm.Marshal(arg)
	if err != nil {
		t.Error(err)
	}

	var um Test

	err = bm.Unmarshal(rcv, &um)
	if err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(arg, um) {
		t.Errorf("expeted %v, received %v", arg, um)
	}
}

func TestStorageInterfaceUnmarshalGOB(t *testing.T) {
	type Test struct {
		Field string
	}
	arg := Test{
		Field: "test",
	}
	bm := GOBMarshaler{}

	rcv, err := bm.Marshal(arg)
	if err != nil {
		t.Error(err)
	}

	var um Test

	err = bm.Unmarshal(rcv, &um)
	if err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(arg, um) {
		t.Errorf("expeted %v, received %v", arg, um)
	}
}
