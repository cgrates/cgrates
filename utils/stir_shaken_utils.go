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
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
	"unicode"

	"github.com/dgrijalva/jwt-go"
)

// NewECDSAPrvKeyFromReader creates a private key from io.Reader
func NewECDSAPrvKeyFromReader(reader io.Reader) (prvKey *ecdsa.PrivateKey, err error) {
	var prvkeyBuf []byte
	if prvkeyBuf, err = io.ReadAll(reader); err != nil {
		return
	}
	return jwt.ParseECPrivateKeyFromPEM(prvkeyBuf)
}

// NewECDSAPubKeyFromReader returns a public key from io.Reader
func NewECDSAPubKeyFromReader(reader io.Reader) (pubKey *ecdsa.PublicKey, err error) {
	var pubkeyBuf []byte
	if pubkeyBuf, err = io.ReadAll(reader); err != nil {
		return
	}
	return jwt.ParseECPublicKeyFromPEM(pubkeyBuf)
}

// NewECDSAPrvKey creates a private key from the path
func NewECDSAPrvKey(prvKeyPath string, timeout time.Duration) (prvKey *ecdsa.PrivateKey, err error) {
	var prvKeyBuf io.ReadCloser
	if prvKeyBuf, err = GetReaderFromPath(prvKeyPath, timeout); err != nil {
		return
	}
	prvKey, err = NewECDSAPrvKeyFromReader(prvKeyBuf)
	prvKeyBuf.Close()
	return
}

// NewECDSAPubKey returns a public key from the path
func NewECDSAPubKey(pubKeyPath string, timeout time.Duration) (pubKey *ecdsa.PublicKey, err error) {
	var pubKeyBuf io.ReadCloser
	if pubKeyBuf, err = GetReaderFromPath(pubKeyPath, timeout); err != nil {
		return
	}
	pubKey, err = NewECDSAPubKeyFromReader(pubKeyBuf)
	pubKeyBuf.Close()
	return
}

// GetReaderFromPath returns the reader at the given path
func GetReaderFromPath(path string, timeout time.Duration) (r io.ReadCloser, err error) {
	if !IsURL(path) {
		return os.Open(path)
	}
	httpClient := http.Client{
		Timeout: timeout,
	}
	var resp *http.Response
	if resp, err = httpClient.Get(path); err != nil {
		return
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		err = fmt.Errorf("http status error: %v", resp.StatusCode)
		return
	}
	return resp.Body, nil
}

// EncodeBase64JSON encodes the structure in json and then the string in base64
func EncodeBase64JSON(val interface{}) (enc string, err error) {
	var b []byte
	if b, err = json.Marshal(val); err != nil {
		return
	}
	enc = jwt.EncodeSegment(b)
	return
}

// DecodeBase64JSON decodes the base64 json string in the given interface
func DecodeBase64JSON(data string, val interface{}) (err error) {
	var b []byte
	if b, err = jwt.DecodeSegment(data); err != nil {
		return
	}
	return json.Unmarshal(b, val)
}

// RemoveWhiteSpaces removes white spaces from string
func RemoveWhiteSpaces(str string) string {
	rout := make([]rune, 0, len(str))
	for _, r := range str {
		if !unicode.IsSpace(r) {
			rout = append(rout, r)
		}
	}
	return string(rout)
}
