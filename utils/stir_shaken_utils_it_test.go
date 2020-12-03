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
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
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
	urlPath := "https://raw.githubusercontent.com/cgrates/cgrates/master/data/stir/stir_privatekey.pem"
	expected, err := jwt.ParseECPrivateKeyFromPEM([]byte(`
-----BEGIN EC PRIVATE KEY-----
MHcCAQEEICcL1+2nj9ylMlTKjSpIGx03gALK0cISciviwudQuvb9oAoGCCqGSM49
AwEHoUQDQgAEjS4zmWotYqKWB2/sn+4v1uUoPAQ2N2ZtrUsmewkl3ErAbIokXSZS
rucJPPszlBtYbbhcmbXC7DKP9u9Pq/GnVg==
-----END EC PRIVATE KEY-----`))
	if err != nil {
		t.Error(err)
	}
	if prvKey, err := NewECDSAPrvKey(urlPath, time.Duration(0)); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, prvKey) {
		t.Errorf("Expected %+v, received %+v", expected, prvKey)
	}
}

func testNewECDSAPublicKey(t *testing.T) {
	urlPath := "https://raw.githubusercontent.com/cgrates/cgrates/master/data/stir/stir_pubkey.pem"
	expPublKey, err := jwt.ParseECPublicKeyFromPEM([]byte(` 
-----BEGIN PUBLIC KEY-----
MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEjS4zmWotYqKWB2/sn+4v1uUoPAQ2
N2ZtrUsmewkl3ErAbIokXSZSrucJPPszlBtYbbhcmbXC7DKP9u9Pq/GnVg==
-----END PUBLIC KEY-----`))
	if err != nil {
		t.Error(err)
	}
	if publKey, err := NewECDSAPubKey(urlPath, 0); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expPublKey, publKey) {
		t.Errorf("Expected %+v, received %+v", expPublKey, publKey)
	}
}
