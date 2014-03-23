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

package cdre

import (
	"testing"
)

func TestMaxLen(t *testing.T) {
	result, err := FmtFieldWidth("test", 4, "", "")
	expected := "test"
	if err != nil || result != expected {
		t.Errorf("Expected \"test\" was \"%s\"", result)
	}
}

func TestRPadding(t *testing.T) {
	result, err := FmtFieldWidth("test", 8, "", "right")
	expected := "test    "
	if err != nil || result != expected {
		t.Errorf("Expected \"%s \" was \"%s\"", expected, result)
	}
}

func TestPaddingFiller(t *testing.T) {
	result, err := FmtFieldWidth("", 8, "", "right")
	expected := "        "
	if err != nil || result != expected {
		t.Errorf("Expected \"%s \" was \"%s\"", expected, result)
	}
}

func TestLPadding(t *testing.T) {
	result, err := FmtFieldWidth("test", 8, "", "left")
	expected := "    test"
	if err != nil || result != expected {
		t.Errorf("Expected \"%s \" was \"%s\"", expected, result)
	}
}

func TestZeroLPadding(t *testing.T) {
	result, err := FmtFieldWidth("test", 8, "", "zeroleft")
	expected := "0000test"
	if err != nil || result != expected {
		t.Errorf("Expected \"%s \" was \"%s\"", expected, result)
	}
}

func TestRStrip(t *testing.T) {
	result, err := FmtFieldWidth("test", 2, "right", "")
	expected := "te"
	if err != nil || result != expected {
		t.Errorf("Expected \"%s \" was \"%s\"", expected, result)
	}
}

func TestXRStrip(t *testing.T) {
	result, err := FmtFieldWidth("test", 3, "xright", "")
	expected := "tex"
	if err != nil || result != expected {
		t.Errorf("Expected \"%s \" was \"%s\"", expected, result)
	}
}

func TestLStrip(t *testing.T) {
	result, err := FmtFieldWidth("test", 2, "left", "")
	expected := "st"
	if err != nil || result != expected {
		t.Errorf("Expected \"%s \" was \"%s\"", expected, result)
	}
}

func TestXLStrip(t *testing.T) {
	result, err := FmtFieldWidth("test", 3, "xleft", "")
	expected := "xst"
	if err != nil || result != expected {
		t.Errorf("Expected \"%s \" was \"%s\"", expected, result)
	}
}

func TestStripNotAllowed(t *testing.T) {
	_, err := FmtFieldWidth("test", 3, "", "")
	if err == nil {
		t.Error("Expected error")
	}
}

func TestPaddingNotAllowed(t *testing.T) {
	_, err := FmtFieldWidth("test", 5, "", "")
	if err == nil {
		t.Error("Expected error")
	}
}
