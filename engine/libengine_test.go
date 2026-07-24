// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package engine

import (
	"testing"
)

func TestNewRPCClientSet(t *testing.T) {
	tests := []struct {
		name      string
		clientSet *RPCClientSet
	}{
		{
			name:      "default case",
			clientSet: NewRPCClientSet(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if tt.clientSet == nil {
				t.Errorf("Expected RPCClientSet to be non-nil")
			}
			if tt.clientSet.set == nil {
				t.Errorf("Expected 'set' map to be initialized, got nil")
			}
			if len(tt.clientSet.set) != 0 {
				t.Errorf("Expected 'set' map to be empty, got %d items", len(tt.clientSet.set))
			}
		})
	}
}
