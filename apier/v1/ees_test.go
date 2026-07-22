// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package v1

import (
	"testing"

	"github.com/cgrates/cgrates/ees"
)

func TestNewEeSv1(t *testing.T) {
	eeS := &ees.EventExporterS{}
	eeSv1 := NewEeSv1(eeS)
	if eeSv1 == nil {
		t.Fatalf("Expected non-nil EeSv1, got nil")
	}
	if eeSv1.eeS != eeS {
		t.Errorf("Expected eeS field to be set correctly")
	}
}
