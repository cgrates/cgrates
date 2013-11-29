/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2013 ITsysCOM

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

package cdrexporter

import (
	"testing"
)

func TestMaxLen(t *testing.T) {
	result, err := filterField("test", 4, false, false, false, false)
	expected := "test"
	if err != nil || result != expected {
		t.Errorf("Expected \"test\" was \"%s\"", result)
	}
}

func TestRPadding(t *testing.T) {
	result, err := filterField("test", 8, false, false, false, false)
	expected := "test    "
	if err != nil || result != expected {
		t.Errorf("Expected \"%s \" was \"%s\"", expected, result)
	}
}

func TestLPadding(t *testing.T) {
	result, err := filterField("test", 8, false, false, true, false)
	expected := "    test"
	if err != nil || result != expected {
		t.Errorf("Expected \"%s \" was \"%s\"", expected, result)
	}
}

func TestRStrip(t *testing.T) {
	result, err := filterField("test", 2, true, false, false, false)
	expected := "te"
	if err != nil || result != expected {
		t.Errorf("Expected \"%s \" was \"%s\"", expected, result)
	}
}

func TestLStrip(t *testing.T) {
	result, err := filterField("test", 2, true, true, false, false)
	expected := "st"
	if err != nil || result != expected {
		t.Errorf("Expected \"%s \" was \"%s\"", expected, result)
	}
}

func TestStripNotAllowed(t *testing.T) {
	_, err := filterField("test", 2, false, false, false, false)
	if err == nil {
		t.Error("Expected error")
	}
}

func TestLZeroPadding(t *testing.T) {
	result, err := filterField("12", 8, false, false, true, true)
	expected := "00000012"
	if err != nil || result != expected {
		t.Errorf("Expected \"%s \" was \"%s\"", expected, result)
	}
}

func TestRZeroPadding(t *testing.T) {
	result, err := filterField("12", 8, false, false, false, true)
	expected := "12      "
	if err != nil || result != expected {
		t.Errorf("Expected \"%s \" was \"%s\"", expected, result)
	}
}
