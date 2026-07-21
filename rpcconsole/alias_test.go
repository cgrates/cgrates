// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package rpcconsole_test

import (
	"testing"

	"github.com/cgrates/cgrates/rpcconsole"
)

func TestAlias(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"AdminSv1.SetActionProfile", "admins.setActionProfile"},
		{"SessionSv1.ProcessEvent", "sessions.processEvent"},
		{"IPsV1.AllocateIP", "ips.allocateIP"},
		{"IPsV1.STIRAuthenticate", "ips.stirAuthenticate"},
		{"CDRsV1.GetCDRs", "cdrs.getCDRs"},
		{"ConfigSv1.GetConfigAsJSON", "configs.getConfigAsJSON"},
		{"AccountSv1.GetAccountIDs", "accounts.getAccountIDs"},
		{"AgentV1.STIRIdentity", "agent.stirIdentity"},
		{"CoreSv1.Status", "cores.status"},
		{"ServiceManagerV1.StartEngine", "servicemanager.startEngine"},
		{"ErSv1.ExportEvent", "ers.exportEvent"},
		{"notadottedname", "notadottedname"},
	}
	for _, tc := range tests {
		t.Run(tc.in, func(t *testing.T) {
			if got := rpcconsole.Alias(tc.in); got != tc.want {
				t.Fatalf("Alias(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}
