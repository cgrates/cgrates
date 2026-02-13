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

//go:generate go run ../data/scripts/gen_mccmnc.go

import (
	"encoding/binary"
	"errors"
	"fmt"
	"strings"
)

// Geographic location types per 3GPP TS 29.061 section 16.4.7.2
const (
	ULITypeCGI        = 0   // Cell Global Identity (2G)
	ULITypeSAI        = 1   // Service Area Identity (3G)
	ULITypeRAI        = 2   // Routing Area Identity (2G/3G)
	ULITypeTAI        = 128 // Tracking Area Identity (4G)
	ULITypeECGI       = 129 // E-UTRAN Cell Global Identifier (4G)
	ULITypeTAIECGI    = 130 // TAI and ECGI (4G)
	ULITypeNCGI       = 135 // NR Cell Global Identifier (5G)
	ULIType5GSTAI     = 136 // 5G Tracking Area Identity
	ULIType5GSTAINCGI = 137 // 5GS TAI and NCGI (5G)
)

// ULI holds decoded 3GPP-User-Location-Info.
type ULI struct {
	CGI    *CGI    `json:"CGI,omitempty"`
	SAI    *SAI    `json:"SAI,omitempty"`
	RAI    *RAI    `json:"RAI,omitempty"`
	TAI    *TAI    `json:"TAI,omitempty"`
	ECGI   *ECGI   `json:"ECGI,omitempty"`
	TAI5GS *TAI5GS `json:"TAI5GS,omitempty"`
	NCGI   *NCGI   `json:"NCGI,omitempty"`
}

// CGI is Cell Global Identity (2G GSM).
type CGI struct {
	MCC string `json:"MCC"`
	MNC string `json:"MNC"`
	LAC uint16 `json:"LAC"`
	CI  uint16 `json:"CI"`
}

// SAI is Service Area Identity (3G UMTS).
type SAI struct {
	MCC string `json:"MCC"`
	MNC string `json:"MNC"`
	LAC uint16 `json:"LAC"`
	SAC uint16 `json:"SAC"`
}

// RAI is Routing Area Identity (2G/3G).
type RAI struct {
	MCC string `json:"MCC"`
	MNC string `json:"MNC"`
	LAC uint16 `json:"LAC"`
	RAC uint8  `json:"RAC"`
}

// TAI is Tracking Area Identity (4G LTE).
type TAI struct {
	MCC string `json:"MCC"`
	MNC string `json:"MNC"`
	TAC uint16 `json:"TAC"`
}

// ECGI is E-UTRAN Cell Global Identifier (4G LTE).
type ECGI struct {
	MCC string `json:"MCC"`
	MNC string `json:"MNC"`
	ECI uint32 `json:"ECI"`
}

// TAI5GS is 5G Tracking Area Identity.
type TAI5GS struct {
	MCC string `json:"MCC"`
	MNC string `json:"MNC"`
	TAC uint32 `json:"TAC"` // 24-bit (vs 16-bit in 4G TAI)
}

// NCGI is NR Cell Global Identifier (5G).
type NCGI struct {
	MCC string `json:"MCC"`
	MNC string `json:"MNC"`
	NCI uint64 `json:"NCI"` // 36-bit NR Cell Identity
}

// DecodeULI parses 3GPP-User-Location-Info from bytes per 3GPP TS 29.061 section 16.4.7.2.
func DecodeULI(data []byte) (*ULI, error) {
	if len(data) < 2 {
		return nil, errors.New("ULI data too short")
	}

	uli := &ULI{}
	locType := data[0]
	pos := 1

	switch locType {
	case ULITypeCGI:
		if len(data) < 8 { // 1 type + 3 PLMN + 2 LAC + 2 CI
			return nil, errors.New("insufficient data for CGI")
		}
		uli.CGI = decodeCGI(data[pos:])

	case ULITypeSAI:
		if len(data) < 8 { // 1 type + 3 PLMN + 2 LAC + 2 SAC
			return nil, errors.New("insufficient data for SAI")
		}
		uli.SAI = decodeSAI(data[pos:])

	case ULITypeRAI:
		if len(data) < 7 { // 1 type + 3 PLMN + 2 LAC + 1 RAC
			return nil, errors.New("insufficient data for RAI")
		}
		uli.RAI = decodeRAI(data[pos:])

	case ULITypeTAI:
		if len(data) < 6 { // 1 type + 3 PLMN + 2 TAC
			return nil, errors.New("insufficient data for TAI")
		}
		uli.TAI = decodeTAI(data[pos:])

	case ULITypeECGI:
		if len(data) < 8 { // 1 type + 3 PLMN + 4 ECI
			return nil, errors.New("insufficient data for ECGI")
		}
		uli.ECGI = decodeECGI(data[pos:])

	case ULITypeTAIECGI:
		if len(data) < 13 { // 1 type + 5 TAI + 7 ECGI
			return nil, errors.New("insufficient data for TAI+ECGI")
		}
		uli.TAI = decodeTAI(data[pos:])
		uli.ECGI = decodeECGI(data[pos+5:])

	case ULITypeNCGI:
		if len(data) < 9 { // 1 type + 8 NCGI
			return nil, errors.New("insufficient data for NCGI")
		}
		uli.NCGI = decodeNCGI(data[pos:])

	case ULIType5GSTAI:
		if len(data) < 7 { // 1 type + 6 5GS TAI
			return nil, errors.New("insufficient data for 5GS TAI")
		}
		uli.TAI5GS = decodeTAI5GS(data[pos:])

	case ULIType5GSTAINCGI:
		if len(data) < 15 { // 1 type + 6 5GS TAI + 8 NCGI
			return nil, errors.New("insufficient data for 5GS TAI+NCGI")
		}
		uli.TAI5GS = decodeTAI5GS(data[pos:])
		uli.NCGI = decodeNCGI(data[pos+6:])

	default:
		return nil, fmt.Errorf("unsupported ULI location type: %d", locType)
	}

	return uli, nil
}

// decodePLMN extracts MCC and MNC from 3 bytes (TS 24.008 section 10.5.1.13).
// Each digit is one nibble: [MCC2|MCC1] [MNC3|MCC3] [MNC2|MNC1].
// MNC3=0xF means the MNC is only 2 digits.
func decodePLMN(data []byte) (mcc, mnc string) {
	mcc1 := data[0] & 0x0F
	mcc2 := data[0] >> 4
	mcc3 := data[1] & 0x0F
	mnc3 := data[1] >> 4
	mnc1 := data[2] & 0x0F
	mnc2 := data[2] >> 4

	mcc = fmt.Sprintf("%d%d%d", mcc1, mcc2, mcc3)

	if mnc3 == 0x0F {
		mnc = fmt.Sprintf("%d%d", mnc1, mnc2)
	} else {
		mnc = fmt.Sprintf("%d%d%d", mnc1, mnc2, mnc3)
	}
	return
}

func decodeCGI(data []byte) *CGI {
	// PLMN + LAC + CI (TS 29.274 section 8.21.1)
	mcc, mnc := decodePLMN(data)
	return &CGI{
		MCC: mcc,
		MNC: mnc,
		LAC: binary.BigEndian.Uint16(data[3:5]),
		CI:  binary.BigEndian.Uint16(data[5:7]),
	}
}

func decodeSAI(data []byte) *SAI {
	// PLMN + LAC + SAC (TS 29.274 section 8.21.2)
	mcc, mnc := decodePLMN(data)
	return &SAI{
		MCC: mcc,
		MNC: mnc,
		LAC: binary.BigEndian.Uint16(data[3:5]),
		SAC: binary.BigEndian.Uint16(data[5:7]),
	}
}

func decodeRAI(data []byte) *RAI {
	// PLMN + LAC + RAC (TS 29.274 section 8.21.3)
	mcc, mnc := decodePLMN(data)
	return &RAI{
		MCC: mcc,
		MNC: mnc,
		LAC: binary.BigEndian.Uint16(data[3:5]),
		RAC: data[5],
	}
}

func decodeTAI(data []byte) *TAI {
	// PLMN + TAC (TS 29.274 section 8.21.4)
	mcc, mnc := decodePLMN(data)
	return &TAI{
		MCC: mcc,
		MNC: mnc,
		TAC: binary.BigEndian.Uint16(data[3:5]),
	}
}

func decodeECGI(data []byte) *ECGI {
	mcc, mnc := decodePLMN(data)
	// the leading 4 bits are spare (TS 29.274 section 8.21.5)
	eci := binary.BigEndian.Uint32(data[3:7]) & 0x0FFFFFFF
	return &ECGI{
		MCC: mcc,
		MNC: mnc,
		ECI: eci,
	}
}

func decodeTAI5GS(data []byte) *TAI5GS {
	mcc, mnc := decodePLMN(data)
	// TAC is 24 bits in 5GS (TS 38.413 section 9.3.3.11), unlike 16 bits in 4G TAI
	tac := uint32(data[3])<<16 | uint32(data[4])<<8 | uint32(data[5])
	return &TAI5GS{
		MCC: mcc,
		MNC: mnc,
		TAC: tac,
	}
}

func decodeNCGI(data []byte) *NCGI {
	mcc, mnc := decodePLMN(data)
	// the leading 4 bits are spare (TS 38.413 section 9.3.1.7)
	// TODO: check why Wireshark's ULI dissector uses trail-spare for NCGI
	nci := uint64(data[3]&0x0F)<<32 |
		uint64(data[4])<<24 |
		uint64(data[5])<<16 |
		uint64(data[6])<<8 |
		uint64(data[7])
	return &NCGI{
		MCC: mcc,
		MNC: mnc,
		NCI: nci,
	}
}

// GetField retrieves a value found at the specified path (e.g. "TAI.MCC", "ECGI.ECI", "TAI.MCC.Name").
func (uli *ULI) GetField(path string) (any, error) {
	parts := strings.SplitN(path, ".", 3)
	if len(parts) == 0 || parts[0] == "" {
		return nil, errors.New("empty path")
	}

	var loc any
	var mcc, mnc string

	switch parts[0] {
	case "CGI":
		if uli.CGI == nil {
			return nil, errors.New("CGI not present in ULI")
		}
		loc, mcc, mnc = uli.CGI, uli.CGI.MCC, uli.CGI.MNC
	case "SAI":
		if uli.SAI == nil {
			return nil, errors.New("SAI not present in ULI")
		}
		loc, mcc, mnc = uli.SAI, uli.SAI.MCC, uli.SAI.MNC
	case "RAI":
		if uli.RAI == nil {
			return nil, errors.New("RAI not present in ULI")
		}
		loc, mcc, mnc = uli.RAI, uli.RAI.MCC, uli.RAI.MNC
	case "TAI":
		if uli.TAI == nil {
			return nil, errors.New("TAI not present in ULI")
		}
		loc, mcc, mnc = uli.TAI, uli.TAI.MCC, uli.TAI.MNC
	case "ECGI":
		if uli.ECGI == nil {
			return nil, errors.New("ECGI not present in ULI")
		}
		loc, mcc, mnc = uli.ECGI, uli.ECGI.MCC, uli.ECGI.MNC
	case "TAI5GS":
		if uli.TAI5GS == nil {
			return nil, errors.New("TAI5GS not present in ULI")
		}
		loc, mcc, mnc = uli.TAI5GS, uli.TAI5GS.MCC, uli.TAI5GS.MNC
	case "NCGI":
		if uli.NCGI == nil {
			return nil, errors.New("NCGI not present in ULI")
		}
		loc, mcc, mnc = uli.NCGI, uli.NCGI.MCC, uli.NCGI.MNC
	default:
		return nil, fmt.Errorf("unknown ULI component: %s", parts[0])
	}

	if len(parts) == 1 {
		return loc, nil
	}

	switch parts[1] {
	case "MCC":
		if len(parts) == 3 {
			if parts[2] == "Name" {
				return countryName(mcc)
			}
			return nil, fmt.Errorf("unknown MCC subfield: %s", parts[2])
		}
		return mcc, nil
	case "MNC":
		if len(parts) == 3 {
			if parts[2] == "Name" {
				return networkName(mcc, mnc)
			}
			return nil, fmt.Errorf("unknown MNC subfield: %s", parts[2])
		}
		return mnc, nil
	default:
		if len(parts) == 3 {
			return nil, fmt.Errorf("unknown subfield: %s.%s", parts[1], parts[2])
		}
		return uliFieldValue(loc, parts[1])
	}
}

func uliFieldValue(loc any, field string) (any, error) {
	switch l := loc.(type) {
	case *CGI:
		switch field {
		case "LAC":
			return l.LAC, nil
		case "CI":
			return l.CI, nil
		}
	case *SAI:
		switch field {
		case "LAC":
			return l.LAC, nil
		case "SAC":
			return l.SAC, nil
		}
	case *RAI:
		switch field {
		case "LAC":
			return l.LAC, nil
		case "RAC":
			return l.RAC, nil
		}
	case *TAI:
		if field == "TAC" {
			return l.TAC, nil
		}
	case *ECGI:
		if field == "ECI" {
			return l.ECI, nil
		}
	case *TAI5GS:
		if field == "TAC" {
			return l.TAC, nil
		}
	case *NCGI:
		if field == "NCI" {
			return l.NCI, nil
		}
	}

	return nil, fmt.Errorf("unknown field: %s", field)
}

func countryName(mcc string) (string, error) {
	if name, ok := mccCountry[mcc]; ok {
		return name, nil
	}
	return "", fmt.Errorf("unknown MCC: %s", mcc)
}

func networkName(mcc, mnc string) (string, error) {
	if name, ok := mccmncNetwork[mcc+"-"+mnc]; ok {
		return name, nil
	}
	return "", fmt.Errorf("unknown MCC-MNC: %s-%s", mcc, mnc)
}
