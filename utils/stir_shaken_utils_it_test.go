//go:build flaky

// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package utils

import (
	"testing"
	"time"
)

var (
	stirShakenTests = []func(t *testing.T){
		testGetReaderFromPathGetError,
		testGetReaderFromPathStatusCode,
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
