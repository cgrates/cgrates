// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package utils

import (
	"testing"
	"time"
)

func TestSyncedChan(t *testing.T) {
	defer func() {
		if v := recover(); v != nil {
			t.Error("Expected to not panic")
		}
	}()
	sc := NewSyncedChan()
	sc.CloseOnce()
	sc.CloseOnce()
	sc.CloseOnce()
	select {
	case <-sc.Done():
	case <-time.After(10 * time.Millisecond):
		t.Error("Timeout")
	}
}
