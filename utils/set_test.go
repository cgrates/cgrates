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
	"reflect"
	"sort"
	"testing"
)

func TestNewStringSet(t *testing.T) {
	input := []string{}
	exp := make(StringSet)
	if rcv := NewStringSet(input); !reflect.DeepEqual(rcv, exp) {
		t.Errorf("Expected: %+v, received: %+v", exp, rcv)
	}
	input = []string{"test"}
	exp.AddSlice(input)
	if rcv := NewStringSet(input); !reflect.DeepEqual(rcv, exp) {
		t.Errorf("Expected: %+v, received: %+v", exp, rcv)
	}
	input = []string{"test1", "test2", "test3"}
	exp = make(StringSet)
	exp.AddSlice(input)
	if rcv := NewStringSet(input); !reflect.DeepEqual(rcv, exp) {
		t.Errorf("Expected: %+v, received: %+v", exp, rcv)
	}
}

func TestAdd(t *testing.T) {
	s := make(StringSet)
	eOut := StringSet{
		"test": struct{}{},
	}
	if reflect.DeepEqual(eOut, s) {
		t.Errorf("Expecting: %+v, received: %+v", eOut, s)
	}
	s.Add("test")
	if !reflect.DeepEqual(eOut, s) {
		t.Errorf("Expecting: %+v, received: %+v", eOut, s)
	}
}

func TestRemove(t *testing.T) {
	eOut := make(StringSet)
	s := StringSet{
		"test": struct{}{},
	}
	if reflect.DeepEqual(eOut, s) {
		t.Errorf("Expecting: %+v, received: %+v", eOut, s)
	}
	s.Remove("test")
	if !reflect.DeepEqual(eOut, s) {
		t.Errorf("Expecting: %+v, received: %+v", eOut, s)
	}
}

func TestHas(t *testing.T) {
	s := StringSet{}
	if s.Has("test") {
		t.Error("Expecting: false, received: true")
	}
	s = StringSet{
		"test": struct{}{},
	}
	if !s.Has("test") {
		t.Error("Expecting: true, received: false")
	}
}

func TestAddSlice(t *testing.T) {
	s := StringSet{
		"test": struct{}{}}
	eOut := StringSet{
		"test":  struct{}{},
		"test1": struct{}{},
		"test2": struct{}{}}
	s.AddSlice([]string{"test1", "test2"})
	if !reflect.DeepEqual(eOut, s) {
		t.Errorf("Expecting: %+v, received: %+v", eOut, s)
	}
}

func TestAsSlice(t *testing.T) {
	s := StringSet{}
	eOut := make([]string, 0)
	if rcv := s.AsSlice(); !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", eOut, rcv)
	}
	s = StringSet{
		"test":  struct{}{},
		"test1": struct{}{},
		"test2": struct{}{}}
	eOut = []string{"test", "test1", "test2"}
	rcv := s.AsSlice()
	sort.Strings(rcv)
	if !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", eOut, rcv)
	}
}

func TestAsOrderedSlice(t *testing.T) {
	s := StringSet{}
	eOut := make([]string, 0)
	if rcv := s.AsOrderedSlice(); !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", eOut, rcv)
	}
	s = StringSet{
		"test3":  struct{}{},
		"test12": struct{}{},
		"test2":  struct{}{}}
	eOut = []string{"test12", "test2", "test3"}
	rcv := s.AsOrderedSlice()
	if !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", eOut, rcv)
	}
}

func TestSetSha1(t *testing.T) {
	s := StringSet{
		"test3":  struct{}{},
		"test12": struct{}{},
		"test2":  struct{}{}}
	eOut := "8fbb49ecf2ee4116bc492505865d2125a78f2161"
	if rcv := s.Sha1(); rcv != eOut {
		t.Errorf("Expecting: %+v, received: %+v", eOut, rcv)
	}

	s2 := StringSet{
		"test2":  struct{}{},
		"test3":  struct{}{},
		"test12": struct{}{},
	}
	if rcv := s2.Sha1(); rcv != eOut {
		t.Errorf("Expecting: %+v, received: %+v", eOut, rcv)
	}
}

func TestSize(t *testing.T) {
	s := StringSet{}
	if rcv := s.Size(); rcv != 0 {
		t.Errorf("Expecting: 0, received %+v", rcv)
	}
	s = StringSet{
		"test0": struct{}{},
		"test1": struct{}{},
		"test2": struct{}{}}
	if rcv := s.Size(); rcv != 3 {
		t.Errorf("Expecting: 3, received %+v", rcv)
	}
}

func TestIntersect(t *testing.T) {
	s1 := StringSet{
		"test0": struct{}{},
		"test1": struct{}{},
		"test2": struct{}{}}
	s2 := StringSet{
		"test0": struct{}{},
		"test2": struct{}{},
		"test3": struct{}{}}
	eOut := StringSet{
		"test0": struct{}{},
		"test2": struct{}{}}
	s1.Intersect(s2)
	if !reflect.DeepEqual(eOut, s1) {
		t.Errorf("Expecting: %+v, received: %+v", eOut, s1)
	}
	s1 = StringSet{
		"test0": struct{}{},
		"test1": struct{}{},
		"test2": struct{}{}}
	s2 = StringSet{
		"test3": struct{}{},
		"test4": struct{}{},
		"test5": struct{}{}}
	s1.Intersect(s2)
	eOut = make(StringSet)
	if !reflect.DeepEqual(eOut, s1) {
		t.Errorf("Expecting: %+v, received: %+v", eOut, s1)
	}
}
