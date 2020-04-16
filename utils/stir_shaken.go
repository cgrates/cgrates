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
	"time"
)

// NewPASSporTHeader returns a new PASSporT headder with:
// extension shaken, ES256 algorithm and the given x5u
func NewPASSporTHeader(x5uVal string) *PASSporTHeader {
	return &PASSporTHeader{
		Alg: STIRAlg,
		Ppt: STIRPpt,
		Typ: STIRTyp,
		X5u: x5uVal,
	}
}

// PASSporTHeader is the JOSE header for PASSporT claim
type PASSporTHeader struct {
	Typ string `json:"typ"` // the media type MUST be the string "passport"
	Alg string `json:"alg"` // the cryptographic algorithm for the signature part
	Ppt string `json:"ppt"` // the extension that is used this is optional
	X5u string `json:"x5u"` // the URI referring to the resource for the X.509 public key certificate or certificate chain corresponding to the key used to digitally sign the JWS
}

// NewPASSporTDestinationsIdentity returns a new PASSporTDestinationsIdentity with the given ids
func NewPASSporTDestinationsIdentity(tn, uri []string) *PASSporTDestinationsIdentity {
	return &PASSporTDestinationsIdentity{
		Tn:  tn,
		URI: uri,
	}
}

// PASSporTDestinationsIdentity is the destination identities of any type
type PASSporTDestinationsIdentity struct {
	Tn  []string `json:"tn,omitempty"`  // the telephone numbers
	URI []string `json:"uri,omitempty"` // the identities in URI form
}

// NewPASSporTOriginsIdentity returns a new PASSporTOriginsIdentity with the given id
func NewPASSporTOriginsIdentity(tn, uri string) *PASSporTOriginsIdentity {
	return &PASSporTOriginsIdentity{
		Tn:  tn,
		URI: uri,
	}
}

// PASSporTOriginsIdentity is the origin identities of any type
type PASSporTOriginsIdentity struct {
	Tn  string `json:"tn,omitempty"`  // the telephone number
	URI string `json:"uri,omitempty"` // the identity in URI form
}

// NewPASSporTPayload returns an new PASSporTPayload with the given origin and destination
func NewPASSporTPayload(attest, originID string, dest PASSporTDestinationsIdentity, orig PASSporTOriginsIdentity) PASSporTPayload {
	return PASSporTPayload{
		ATTest: attest,
		Dest:   dest,
		IAT:    time.Now().Unix(),
		Orig:   orig,
		OrigID: originID,
	}
}

// PASSporTPayload is the JOSE claim for PASSporT
type PASSporTPayload struct {
	ATTest string                       `json:"attest"` // the atestation value: 'A', 'B', or 'C'.These values correspond to 'Full Attestation', 'Partial Attestation', and 'Gateway Attestation', respectively. Not used for verification
	Dest   PASSporTDestinationsIdentity `json:"dest"`   // the destinations identity
	IAT    int64                        `json:"iat"`    // is the date and time of issuance of the JWT
	Orig   PASSporTOriginsIdentity      `json:"orig"`   // the originator identity
	OrigID string                       `json:"origid"` // is an opaque unique identifier representing an element on the path of a given SIP request. Not used for verification
}
