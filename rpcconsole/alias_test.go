/*
Real-time Online/Offline Charging System (OCS) for Telecom/ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/

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
