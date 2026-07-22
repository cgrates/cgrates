// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package ers

import (
	"sync"
	"testing"
)

func TestAmqpClientIsAvailable(t *testing.T) {
	tests := []struct {
		name      string
		available bool
	}{
		{
			name:      "Available",
			available: true,
		},
		{
			name:      "Not Available",
			available: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &amqpClient{
				available: tt.available,
				mu:        sync.RWMutex{},
			}

			got := client.isAvailable()
			if got != tt.available {
				t.Errorf("isAvailable() = %v; want %v", got, tt.available)
			}
		})
	}
}
