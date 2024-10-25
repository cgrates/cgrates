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
	"bytes"
	"io"
	"math"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"
	"time"
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

func TestGetReaderFromPath(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		timeout    time.Duration
		expectErr  bool
		expectData string
	}{
		{
			name:       "Valid file path",
			path:       "testfile.txt",
			timeout:    1 * time.Second,
			expectErr:  false,
			expectData: "This is a test file.",
		},
		{
			name:      "Invalid file path",
			path:      "invalidfile.txt",
			timeout:   1 * time.Second,
			expectErr: true,
		},
		{
			name:      "Valid HTTP URL",
			path:      "http://cgrates.com",
			timeout:   1 * time.Second,
			expectErr: false,
		},
		{
			name:      "HTTP URL returns non-200 status",
			path:      "http://cgrates.com/non200",
			timeout:   1 * time.Second,
			expectErr: true,
		},
	}

	testFileName := "testfile.txt"
	err := os.WriteFile(testFileName, []byte("This is a test file."), 0644)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}
	defer os.Remove(testFileName)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/non200" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Write([]byte("test"))
	}))
	defer ts.Close()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.path == "http://cgrates.com" {
				tt.path = ts.URL
			} else if tt.path == "http://cgrates.com/non200" {
				tt.path = ts.URL + "/non200"
			}

			reader, err := GetReaderFromPath(tt.path, tt.timeout)

			if (err != nil) != tt.expectErr {
				t.Errorf("expected error: %v, got: %v", tt.expectErr, err)
				return
			}

			if !tt.expectErr && reader != nil {
				defer reader.Close()

				data, err := io.ReadAll(reader)
				if err != nil {
					t.Errorf("failed to read from reader: %v", err)
					return
				}

				if string(data) != tt.expectData && tt.expectData != "" {
					t.Errorf("expected data: %s, got: %s", tt.expectData, string(data))
				}
			}
		})
	}
}
