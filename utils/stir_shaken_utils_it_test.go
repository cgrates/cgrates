//go:build flaky

// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

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
