// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package engine

import (
	"net/http"
	"testing"
)

func TestSetHTTPPstrTransport(t *testing.T) {
	tmp := httpPstrTransport
	SetHTTPPstrTransport(nil)
	if httpPstrTransport != nil {
		t.Error("Expected the transport to be nil", httpPstrTransport)
	}
	httpPstrTransport = tmp
}

func TestSetCdrStorage(t *testing.T) {
	tmp := cdrStorage
	SetCdrStorage(nil)
	if cdrStorage != nil {
		t.Error("Expected the cdrStorage to be nil", cdrStorage)
	}
	cdrStorage = tmp
}

func TestSetDataStorage(t *testing.T) {
	tmp := dm
	SetDataStorage(nil)
	if dm != nil {
		t.Error("Expected the dm to be nil", dm)
	}
	dm = tmp
}

func TestGlobalvarsGetHTTPPstrTransport(t *testing.T) {
	httpPstrTransport = &http.Transport{}
	transport := GetHTTPPstrTransport()
	if transport == nil {
		t.Error("Expected transport to be initialized, but got nil")
	}
}
