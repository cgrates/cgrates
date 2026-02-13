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

package utils

import (
	"encoding/hex"
	"reflect"
	"testing"
)

func TestDecodePLMN(t *testing.T) {
	tests := []struct {
		name        string
		hex         string
		expectedMCC string
		expectedMNC string
	}{
		{
			name:        "2-digit MNC (547/05)",
			hex:         "45f750",
			expectedMCC: "547",
			expectedMNC: "05",
		},
		{
			name:        "3-digit MNC (310/260)",
			hex:         "130062",
			expectedMCC: "310",
			expectedMNC: "260",
		},
		{
			name:        "2-digit MNC (262/01)",
			hex:         "62f210",
			expectedMCC: "262",
			expectedMNC: "01",
		},
		{
			name:        "3GPP test PLMN 2-digit MNC (001/01)",
			hex:         "00f110",
			expectedMCC: "001",
			expectedMNC: "01",
		},
		{
			name:        "3GPP test PLMN 3-digit MNC (001/001)",
			hex:         "001100",
			expectedMCC: "001",
			expectedMNC: "001",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := hex.DecodeString(tt.hex)
			if err != nil {
				t.Fatalf("invalid test hex: %v", err)
			}
			mcc, mnc := decodePLMN(data)
			if mcc != tt.expectedMCC {
				t.Errorf("MCC: got %q, want %q", mcc, tt.expectedMCC)
			}
			if mnc != tt.expectedMNC {
				t.Errorf("MNC: got %q, want %q", mnc, tt.expectedMNC)
			}
		})
	}
}

func TestDecodeULI(t *testing.T) {
	tests := []struct {
		name     string
		hex      string
		expected *ULI
	}{
		{
			name: "CGI",
			hex:  "0062f21012345678",
			expected: &ULI{
				CGI: &CGI{MCC: "262", MNC: "01", LAC: 0x1234, CI: 0x5678},
			},
		},
		{
			name: "SAI",
			hex:  "0162f2101234abcd",
			expected: &ULI{
				SAI: &SAI{MCC: "262", MNC: "01", LAC: 0x1234, SAC: 0xABCD},
			},
		},
		{
			name: "RAI",
			hex:  "0262f210123456",
			expected: &ULI{
				RAI: &RAI{MCC: "262", MNC: "01", LAC: 0x1234, RAC: 0x56},
			},
		},
		{
			name: "TAI",
			hex:  "8013006204d2",
			expected: &ULI{
				TAI: &TAI{MCC: "310", MNC: "260", TAC: 1234},
			},
		},
		{
			name: "ECGI",
			hex:  "8145f75000000101",
			expected: &ULI{
				ECGI: &ECGI{MCC: "547", MNC: "05", ECI: 257},
			},
		},
		{
			name: "TAI+ECGI",
			hex:  "8245f750000145f75000000101",
			expected: &ULI{
				TAI:  &TAI{MCC: "547", MNC: "05", TAC: 1},
				ECGI: &ECGI{MCC: "547", MNC: "05", ECI: 257},
			},
		},
		{
			name: "NCGI",
			hex:  "871300620123456789",
			expected: &ULI{
				NCGI: &NCGI{MCC: "310", MNC: "260", NCI: 0x123456789},
			},
		},
		{
			name: "5GS TAI",
			hex:  "88130062123456",
			expected: &ULI{
				TAI5GS: &TAI5GS{MCC: "310", MNC: "260", TAC: 0x123456},
			},
		},
		{
			name: "5GS TAI+NCGI",
			hex:  "891300620000011300620000000101",
			expected: &ULI{
				TAI5GS: &TAI5GS{MCC: "310", MNC: "260", TAC: 1},
				NCGI:   &NCGI{MCC: "310", MNC: "260", NCI: 257},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := hex.DecodeString(tt.hex)
			if err != nil {
				t.Fatalf("invalid test hex: %v", err)
			}
			got, err := DecodeULI(data)
			if err != nil {
				t.Fatalf("DecodeULI failed: %v", err)
			}
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("got %s, want %s", ToJSON(got), ToJSON(tt.expected))
			}
		})
	}
}

func TestULI_GetField(t *testing.T) {
	uli := &ULI{
		TAI: &TAI{
			MCC: "547",
			MNC: "05",
			TAC: 1,
		},
		ECGI: &ECGI{
			MCC: "547",
			MNC: "05",
			ECI: 257,
		},
		TAI5GS: &TAI5GS{
			MCC: "310",
			MNC: "260",
			TAC: 0x123456,
		},
		NCGI: &NCGI{
			MCC: "310",
			MNC: "260",
			NCI: 0x123456789,
		},
	}

	tests := []struct {
		path     string
		expected any
	}{
		{"TAI", uli.TAI},
		{"TAI.MCC", "547"},
		{"TAI.MNC", "05"},
		{"TAI.TAC", uint16(1)},
		{"ECGI", uli.ECGI},
		{"ECGI.MCC", "547"},
		{"ECGI.MNC", "05"},
		{"ECGI.ECI", uint32(257)},
		{"TAI5GS", uli.TAI5GS},
		{"TAI5GS.MCC", "310"},
		{"TAI5GS.MNC", "260"},
		{"TAI5GS.TAC", uint32(0x123456)},
		{"NCGI", uli.NCGI},
		{"NCGI.MCC", "310"},
		{"NCGI.MNC", "260"},
		{"NCGI.NCI", uint64(0x123456789)},
		{"TAI.MCC.Name", "French Polynesia"},
		{"NCGI.MCC.Name", "United States"},
		{"NCGI.MNC.Name", "T-Mobile USA"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got, err := uli.GetField(tt.path)
			if err != nil {
				t.Fatalf("GetField(%q) error: %v", tt.path, err)
			}
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("GetField(%q): got %v (%T), want %v (%T)",
					tt.path, got, got, tt.expected, tt.expected)
			}
		})
	}
}

func TestULI_GetField_Errors(t *testing.T) {
	uli := &ULI{
		TAI: &TAI{MCC: "547", MNC: "05", TAC: 1},
	}

	tests := []struct {
		name string
		path string
	}{
		{"missing ECGI", "ECGI"},
		{"missing CGI", "CGI"},
		{"missing TAI5GS", "TAI5GS"},
		{"missing NCGI", "NCGI"},
		{"invalid field", "TAI.INVALID"},
		{"invalid component", "INVALID"},
		{"empty path", ""},
		{"invalid MCC subfield", "TAI.MCC.Invalid"},
		{"invalid MNC subfield", "TAI.MNC.Invalid"},
		{"subfield on TAC", "TAI.TAC.Name"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := uli.GetField(tt.path)
			if err == nil {
				t.Errorf("GetField(%q) should have returned error", tt.path)
			}
		})
	}
}

func TestULIConverter(t *testing.T) {
	tests := []struct {
		name     string
		params   string
		hex      string // raw ULI bytes in hex, decoded to binary before Convert
		expected any
	}{
		{
			name:     "Extract TAI.MCC",
			params:   "*3gpp_uli:TAI.MCC",
			hex:      "8245f750000145f75000000101",
			expected: "547",
		},
		{
			name:     "Extract TAI.MNC",
			params:   "*3gpp_uli:TAI.MNC",
			hex:      "8245f750000145f75000000101",
			expected: "05",
		},
		{
			name:     "Extract TAI.TAC",
			params:   "*3gpp_uli:TAI.TAC",
			hex:      "8245f750000145f75000000101",
			expected: uint16(1),
		},
		{
			name:     "Extract ECGI.ECI",
			params:   "*3gpp_uli:ECGI.ECI",
			hex:      "8245f750000145f75000000101",
			expected: uint32(257),
		},
		{
			name:     "Extract TAI5GS.MCC from 5GS TAI",
			params:   "*3gpp_uli:TAI5GS.MCC",
			hex:      "88130062123456",
			expected: "310",
		},
		{
			name:     "Extract TAI5GS.TAC from 5GS TAI",
			params:   "*3gpp_uli:TAI5GS.TAC",
			hex:      "88130062123456",
			expected: uint32(0x123456),
		},
		{
			name:     "Extract NCGI.NCI from NCGI",
			params:   "*3gpp_uli:NCGI.NCI",
			hex:      "871300620123456789",
			expected: uint64(0x123456789),
		},
		{
			name:     "Extract TAI5GS.MCC.Name",
			params:   "*3gpp_uli:TAI5GS.MCC.Name",
			hex:      "88130062123456",
			expected: "United States",
		},
		{
			name:     "Extract NCGI.MNC.Name",
			params:   "*3gpp_uli:NCGI.MNC.Name",
			hex:      "871300620123456789",
			expected: "T-Mobile USA",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conv, err := NewULIConverter(tt.params)
			if err != nil {
				t.Fatalf("NewULIConverter failed: %v", err)
			}
			raw, err := hex.DecodeString(tt.hex)
			if err != nil {
				t.Fatalf("invalid test hex: %v", err)
			}
			got, err := conv.Convert(string(raw))
			if err != nil {
				t.Fatalf("Convert failed: %v", err)
			}
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("got %v (%T), want %v (%T)",
					got, got, tt.expected, tt.expected)
			}
		})
	}
}

func TestDecodeULI_InsufficientData(t *testing.T) {
	tests := []struct {
		name string
		hex  string
	}{
		{"CGI too short", "0062f210123456"},
		{"SAI too short", "0162f210123456"},
		{"RAI too short", "0262f2101234"},
		{"TAI too short", "8062f21012"},
		{"ECGI too short", "8162f210123456"},
		{"TAI+ECGI too short", "8262f2101234567890"},
		{"5GS TAI too short", "881300621234"},
		{"NCGI too short", "8713006201234567"},
		{"5GS TAI+NCGI too short", "8913006200000113006200000001"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := hex.DecodeString(tt.hex)
			if err != nil {
				t.Fatalf("invalid test hex: %v", err)
			}
			_, err = DecodeULI(data)
			if err == nil {
				t.Error("DecodeULI should have returned error for insufficient data")
			}
		})
	}
}

func TestDecodeULI_UnsupportedType(t *testing.T) {
	data, err := hex.DecodeString("0362f21012345678")
	if err != nil {
		t.Fatalf("invalid test hex: %v", err)
	}

	_, err = DecodeULI(data)
	if err == nil {
		t.Error("DecodeULI should have returned error for unsupported type")
	}
}

func TestULIConverter_EmptyPath(t *testing.T) {
	conv, err := NewULIConverter("*3gpp_uli")
	if err != nil {
		t.Fatalf("NewULIConverter failed: %v", err)
	}
	raw, err := hex.DecodeString("8013006204d2")
	if err != nil {
		t.Fatalf("invalid test hex: %v", err)
	}
	got, err := conv.Convert(string(raw))
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}
	uli, ok := got.(*ULI)
	if !ok {
		t.Fatalf("expected *ULI, got %T", got)
	}
	if uli.TAI == nil {
		t.Fatal("TAI should not be nil")
	}
	if uli.TAI.MCC != "310" {
		t.Errorf("TAI.MCC: got %q, want %q", uli.TAI.MCC, "310")
	}
}

func TestULIConverter_Errors(t *testing.T) {
	conv, err := NewULIConverter("*3gpp_uli:TAI.MCC")
	if err != nil {
		t.Fatalf("NewULIConverter failed: %v", err)
	}

	tests := []struct {
		name  string
		input any
	}{
		{"empty string", ""},
		{"unsupported ULI type", string([]byte{0x03, 0x62, 0xF2, 0x10, 0x12, 0x34, 0x56, 0x78})},
		{"truncated data", string([]byte{0x82, 0x45})},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := conv.Convert(tt.input)
			if err == nil {
				t.Error("Convert should have returned error")
			}
		})
	}
}
