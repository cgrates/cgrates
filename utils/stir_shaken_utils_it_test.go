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
	"testing"
	"time"
)

var (
	stirShakenTests = []func(t *testing.T){
		testGetReaderFromPathGetError,
		testGetReaderFromPathStatusCode,
		testNewECDSAPrvKey,
		testNewECDSAPublicKey,
	}
)

func TestStirShakenUtils(t *testing.T) {
	for _, test := range stirShakenTests {
		t.Run("StirShakenUtils", test)
	}
}

func testGetReaderFromPathGetError(t *testing.T) {
	urlPath := "https://www.example.org/cert.cer"
	expErr := "Get \"https://www.example.org/cert.cer\": context deadline exceeded (Client.Timeout exceeded while awaiting headers)"
	if _, err := GetReaderFromPath(urlPath, time.Duration(10)); err == nil || err.Error() != expErr {
		t.Errorf("Expected %+v, received %+v", expErr, err)
	}
}

func testGetReaderFromPathStatusCode(t *testing.T) {
	urlPath := "https://www.example.org/cert.cer"
	expErr := "http status error: 404"
	if _, err := GetReaderFromPath(urlPath, time.Duration(0)); err == nil || err.Error() != expErr {
		t.Errorf("Expected %+v, received %+v", expErr, err)
	}
}

func testNewECDSAPrvKey(t *testing.T) {
	urlPath := "https://en.wikipedia.org/wiki/Elliptic_Curve_Digital_Signature_Algorithm"
	expPrvKey := "Invalid Key: Key must be PEM encoded PKCS1 or PKCS8 private key"
	if _, err := NewECDSAPrvKey(urlPath, time.Duration(0)); err == nil || err.Error() != expPrvKey {
		t.Errorf("Expected %+v, received %+v", expPrvKey, err)
	}
}

func testNewECDSAPublicKey(t *testing.T) {
	urlPath := "https://en.wikipedia.org/wiki/Wiki"
	expPublKey := "Invalid Key: Key must be PEM encoded PKCS1 or PKCS8 private key"
	if _, err := NewECDSAPubKey(urlPath, 0); err == nil || err.Error() != expPublKey {
		t.Errorf("Expected %+v, received %+v", expPublKey, err)
	}
}
