//go:build flaky

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

package general_tests

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// TODO: also test how toggling dispatchers behaves
func TestServiceToggle(t *testing.T) {
	cfgJSON := `{
"data_db": {
	"db_type": "*internal"
},
"stor_db": {
	"db_type": "*internal"
},
"analyzers": {
	"enabled": %[1]v,
 	"db_path": "/tmp",
},
"apiers": {
	"enabled": %[1]v
},
"attributes": {
	"enabled": %[1]v
},
"cdrs": {
	"enabled": %[1]v
},
"chargers": {
	"enabled": %[1]v
},
"ees": {
	"enabled": %[1]v
},
"ers": {
	"enabled": %[1]v
},
"rals": {
	"enabled": %[1]v
},
"resources": {
	"enabled": %[1]v,
	"store_interval": "-1"
},
"routes": {
	"enabled": %[1]v
},
"schedulers": {
	"enabled": %[1]v
},
"sessions": {
	"enabled": %[1]v
},
"stats": {
	"enabled": %[1]v,
	"store_interval": "-1"
},
"thresholds": {
	"enabled": %[1]v,
	"store_interval": "-1"
}
}`

	// List of services that can be toggled (by setting "enabled" true/false).
	services := []string{
		utils.AnalyzerSv1,
		utils.APIerSv1,
		utils.APIerSv2,
		utils.AttributeSv1,
		utils.CDRsV1,
		utils.ChargerSv1,
		utils.EeSv1,
		utils.ErSv1,
		utils.RALsV1,
		utils.ResourceSv1,
		utils.RouteSv1,
		utils.SchedulerSv1,
		utils.SessionSv1,
		utils.StatSv1,
		utils.ThresholdSv1,
		// utils.TrendSv1,
	}

	// Start a cgr-engine instance that has all the services
	// from the slice above enabled.
	ng := engine.TestEngine{
		ConfigJSON: fmt.Sprintf(cfgJSON, "true"),
	}
	client, cfg := ng.Run(t)

	// Ensure the services are up by calling waitForService helper for each service.
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	for _, service := range services {
		engine.WaitForService(t, ctx, client, service)
	}

	// Toggle the state of all services via config reload.
	fullCfgPath := filepath.Join(cfg.ConfigPath, "cgrates.json") // path to the original json config file
	if err := os.WriteFile(fullCfgPath, []byte(fmt.Sprintf(cfgJSON, "false")), 0644); err != nil {
		t.Fatal(err)
	}
	var reply string
	if err := client.Call(context.Background(), utils.ConfigSv1ReloadConfig, &config.ReloadArgs{
		Path:    cfg.ConfigPath,
		Section: utils.MetaAll,
	}, &reply); err != nil {
		t.Errorf("ConfigSv1.ReloadConfig unexpected err: %v", err)
	}

	// Ping the services again to ensure they aren't reachable anymore.
	for _, service := range services {
		method := service + ".Ping"
		if err := client.Call(context.Background(), method, nil, &reply); err == nil ||
			!strings.HasPrefix(err.Error(), "rpc: can't find service") {
			t.Errorf("could still ping %s when disabled", service)
		}
	}
}
