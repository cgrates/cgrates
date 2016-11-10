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

import "testing"

func TestVerifyCredential(t *testing.T) {
	var hashedPasswords = map[string]string{
		"1234": "MTIzNA==",
		"bar":  "YmFy",
	}

	var verifyCredentialTests = []struct {
		username string
		password string
		userList map[string]string
		result   bool
	}{
		{"test", "1234", map[string]string{"test": hashedPasswords["1234"]}, true},
		{"test", "0000", map[string]string{"test": hashedPasswords["1234"]}, false},
		{"foo", "bar", map[string]string{"test": "1234", "foo": hashedPasswords["bar"]}, true},
		{"foo", "1234", map[string]string{"test": "1234", "foo": hashedPasswords["bar"]}, false},
		{"none", "1234", map[string]string{"test": "1234", "foo": hashedPasswords["bar"]}, false},
	}

	for _, tt := range verifyCredentialTests {
		r := verifyCredential(tt.username, tt.password, tt.userList)
		if r != tt.result {
			t.Errorf("verifyCredential(%s, %s, %v) => %t, want %t", tt.username, tt.password, tt.userList, r, tt.result)
		}
	}
}
