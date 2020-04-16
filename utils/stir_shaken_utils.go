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
	"io/ioutil"
	"net/http"
	"os"
	"time"
	"unicode"

	"github.com/dgrijalva/jwt-go"
)

// NewECDSAPrvKey creates a private key from the path
func NewECDSAPrvKey(prvKeyPath string, timeout time.Duration) (prvKey *ecdsa.PrivateKey, err error) {
	var prvkeyBuf []byte
	if prvkeyBuf, err = GetDataAtPath(prvKeyPath, timeout); err != nil {
		return
	}
	return jwt.ParseECPrivateKeyFromPEM(prvkeyBuf)
}

// NewECDSAPubKey returns a public key from the path
func NewECDSAPubKey(pubKeyPath string, timeout time.Duration) (pubKey *ecdsa.PublicKey, err error) {
	var pubkeyBuf []byte
	if pubkeyBuf, err = GetDataAtPath(pubKeyPath, timeout); err != nil {
		return
	}
	return jwt.ParseECPublicKeyFromPEM(pubkeyBuf)
}

// getURLFile returns the file from URL
func getURLFile(urlVal string, timeout time.Duration) (body []byte, err error) {
	httpClient := http.Client{
		Timeout: timeout,
	}
	var resp *http.Response
	if resp, err = httpClient.Get(urlVal); err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("http status error: %v", resp.StatusCode)
		return
	}
	return ioutil.ReadAll(resp.Body)
}

func GetDataAtPath(path string, timeout time.Duration) (body []byte, err error) {
	if IsURL(path) {
		return getURLFile(path, timeout)
	}
	var file *os.File
	if file, err = os.Open(path); err != nil {
		return
	}
	body, err = ioutil.ReadAll(file)
	file.Close()
	return
}

func EncodeBase64JSON(val interface{}) (enc string, err error) {
	var b []byte
	if b, err = json.Marshal(val); err != nil {
		return
	}
	enc = jwt.EncodeSegment(b)
	return
}

func DecodeBase64JSON(data string, val interface{}) (err error) {
	var b []byte
	if b, err = jwt.DecodeSegment(data); err != nil {
		return
	}
	return json.Unmarshal(b, val)
}

func RemoveWhiteSpaces(str string) string {
	rout := make([]rune, 0, len(str))
	for _, r := range str {
		if !unicode.IsSpace(r) {
			rout = append(rout, r)
		}
	}
	return string(rout)
}
