// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

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
