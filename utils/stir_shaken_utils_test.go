// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package utils

import (
	"bytes"
	"math"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
)

func TestRemoveWhiteSpaces(t *testing.T) {
	strWithWS := `   A	String
	With	White Spaces`
	expected := `AStringWithWhiteSpaces`
	if rply := RemoveWhiteSpaces(strWithWS); rply != expected {
		t.Errorf("Expected: %q, received: %q", expected, rply)
	}
}

func TestEncodeBase64JSON(t *testing.T) {
	var args any
	args = math.NaN()
	if _, err := EncodeBase64JSON(args); err == nil {
		t.Errorf("Expected error")
	}
	args = map[string]any{"Q": 1}
	expected := `eyJRIjoxfQ`
	if rply, err := EncodeBase64JSON(args); err != nil {
		t.Error(err)
	} else if rply != expected {
		t.Errorf("Expected: %q,received: %q", expected, rply)
	}
}

func TestDecodeBase64JSON(t *testing.T) {
	args := `eyJRIjoxfQ`
	var rply1 string
	if err := DecodeBase64JSON(args, &rply1); err == nil {
		t.Errorf("Expected error")
	}
	var rply2 map[string]any
	expected := map[string]any{"Q": 1.}
	if err := DecodeBase64JSON(args, &rply2); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rply2) {
		t.Errorf("Expected: %s,received: %s", ToJSON(expected), ToJSON(rply2))
	}
	args = `eyJRIjoxfQ,`
	if err := DecodeBase64JSON(args, &rply2); err == nil {
		t.Errorf("Expected error")
	}
}

type testErrReader struct{}

func (testErrReader) Read([]byte) (int, error) { return 0, ErrNotFound }

func TestNewECDSAPrvKeyFromReader(t *testing.T) {
	if _, err := NewECDSAPrvKeyFromReader(new(testErrReader)); err == nil {
		t.Errorf("Expected error")
	}
	r := bytes.NewBuffer([]byte("invalid certificate"))
	if _, err := NewECDSAPrvKeyFromReader(r); err == nil {
		t.Errorf("Expected error")
	}
}

func TestNewECDSAPubKeyFromReader(t *testing.T) {
	if _, err := NewECDSAPubKeyFromReader(new(testErrReader)); err == nil {
		t.Errorf("Expected error")
	}
	r := bytes.NewBuffer([]byte("invalid certificate"))
	if _, err := NewECDSAPubKeyFromReader(r); err == nil {
		t.Errorf("Expected error")
	}
}

func TestNewECDSAPrvKey(t *testing.T) {
	server := httptest.NewServer(http.FileServer(http.Dir("../data/stir")))
	defer server.Close()
	expected, err := jwt.ParseECPrivateKeyFromPEM([]byte(`
-----BEGIN EC PRIVATE KEY-----
MHcCAQEEICcL1+2nj9ylMlTKjSpIGx03gALK0cISciviwudQuvb9oAoGCCqGSM49
AwEHoUQDQgAEjS4zmWotYqKWB2/sn+4v1uUoPAQ2N2ZtrUsmewkl3ErAbIokXSZS
rucJPPszlBtYbbhcmbXC7DKP9u9Pq/GnVg==
-----END EC PRIVATE KEY-----`))
	if err != nil {
		t.Fatal(err)
	}
	prvKey, err := NewECDSAPrvKey(server.URL+"/stir_privatekey.pem", time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(expected, prvKey) {
		t.Errorf("Expected %+v, received %+v", expected, prvKey)
	}
}

func TestNewECDSAPubKey(t *testing.T) {
	server := httptest.NewServer(http.FileServer(http.Dir("../data/stir")))
	defer server.Close()
	expected, err := jwt.ParseECPublicKeyFromPEM([]byte(`
-----BEGIN PUBLIC KEY-----
MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEjS4zmWotYqKWB2/sn+4v1uUoPAQ2
N2ZtrUsmewkl3ErAbIokXSZSrucJPPszlBtYbbhcmbXC7DKP9u9Pq/GnVg==
-----END PUBLIC KEY-----`))
	if err != nil {
		t.Fatal(err)
	}
	pubKey, err := NewECDSAPubKey(server.URL+"/stir_pubkey.pem", time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(expected, pubKey) {
		t.Errorf("Expected %+v, received %+v", expected, pubKey)
	}
}

func TestNewECDSAPrvKeyError(t *testing.T) {
	_, err := NewECDSAPrvKey("string", time.Duration(10))
	if err == nil || err.Error() != "open string: no such file or directory" {
		t.Errorf("Expected <open string: no such file or directory>, received <%v>", err)
	}
}

func TestNewECDSAPubKeyError(t *testing.T) {
	_, err := NewECDSAPubKey("string", time.Duration(10))
	if err == nil || err.Error() != "open string: no such file or directory" {
		t.Errorf("Expected <open string: no such file or directory>, received <%v>", err)
	}
}

func TestGetReaderFromPathError(t *testing.T) {
	_, err := GetReaderFromPath("string", time.Duration(10))
	if err == nil || err.Error() != "open string: no such file or directory" {
		t.Errorf("Expected <open string: no such file or directory>, received <%v>", err)
	}
}
