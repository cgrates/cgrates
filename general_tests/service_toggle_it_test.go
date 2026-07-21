//go:build integration

// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package general_tests

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

func TestServiceToggle(t *testing.T) {
	var dbCfg engine.DBCfg
	switch *utils.DBType {
	case utils.MetaInternal:
		dbCfg = engine.InternalDBCfg
	case utils.MetaMySQL, utils.MetaRedis, utils.MetaMongo, utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("unsupported dbtype value")
	}

	anzDBPath := t.TempDir()
	cfgJSON := `{
"logger": {
	"type": "*stdout"
},
"db": {
	"dbConns": {
		"*default": {
			"dbType": "*internal",
			"opts":{
		"internalDBRewriteInterval": "0s",
		"internalDBDumpInterval": "0s"
	}
    	}
	},
},
"accounts": {
	"enabled": %[1]v
},
"actions": {
	"enabled": %[1]v
},
"admins": {
	"enabled": %[1]v
},
"analyzers": {
	"enabled": %[1]v,
 	"dbPath": "%[2]s",
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
"configs": {
	"enabled": %[1]v
},
"ees": {
	"enabled": %[1]v
},
"ers": {
	"enabled": %[1]v
},
"efs": {
	"enabled": %[1]v
},
// "rankings": {
// 	"enabled": %[1]v,
// 	"storeInterval": "-1"
// },
"rates": {
	"enabled": %[1]v
},
"resources": {
	"enabled": %[1]v,
	"storeInterval": "-1"
},
"ips": {
	"enabled": %[1]v,
	"storeInterval": "-1"
},
"routes": {
	"enabled": %[1]v
},
"sessions": {
	"enabled": %[1]v
},
"stats": {
	"enabled": %[1]v,
	"storeInterval": "-1"
},
"thresholds": {
	"enabled": %[1]v,
	"storeInterval": "-1"
},
"tpes": {
	"enabled": %[1]v
},
// "trends": {
// 	"enabled": %[1]v,
// 	"storeInterval": "-1"
// }
}`

	// Start a cgr-engine instance that has all the services
	// from the slice above enabled.
	ng := engine.TestEngine{
		ConfigJSON: fmt.Sprintf(cfgJSON, "true", anzDBPath),
		DBCfg:      dbCfg,
		Encoding:   *utils.Encoding,
		// LogBuffer:  &bytes.Buffer{},
	}
	// defer fmt.Println(ng.LogBuffer)
	client, cfg := ng.Run(t)
	checkServiceStates(t, client, utils.StateServiceUP)

	// Toggle the state of all services via config reload.
	fullCfgPath := filepath.Join(cfg.ConfigPath, "zzz_dynamic_cgrates.json") // path to the original json config file
	if err := os.WriteFile(fullCfgPath, fmt.Appendf(nil, cfgJSON, "false", anzDBPath), 0644); err != nil {
		t.Fatal(err)
	}
	var reply string
	if err := client.Call(context.Background(), utils.ConfigSv1ReloadConfig, &config.ReloadArgs{
		Section: utils.MetaAll,
	}, &reply); err != nil {
		t.Errorf("ConfigSv1.ReloadConfig unexpected err: %v", err)
	}
	checkServiceStates(t, client, utils.StateServiceDOWN)

	// Toggle the state once again to make sure the actions are repeatable.
	if err := os.WriteFile(fullCfgPath, fmt.Appendf(nil, cfgJSON, "true", anzDBPath), 0644); err != nil {
		t.Fatal(err)
	}
	if err := client.Call(context.Background(), utils.ConfigSv1ReloadConfig, &config.ReloadArgs{
		Section: utils.MetaAll,
	}, &reply); err != nil {
		t.Errorf("ConfigSv1.ReloadConfig unexpected err: %v", err)
	}
	checkServiceStates(t, client, utils.StateServiceUP)
}

func checkServiceStates(t *testing.T, client *birpc.Client, want string) {
	t.Helper()

	// The following services' ShouldRun method always returns true, therefore
	// they cannot be stopped.
	// NOTE: Some services are not needed for cgr-engine to properly function
	// and could have their ShouldRun methods revised. For example, while
	// CGRConfig is definitely needed, ConfigService just registers the
	// CGRConfig service methods. We could call the Shutdown function on it to
	// unregister them without affecting other services at all.
	alwaysUp := []string{
		utils.CacheS,
		utils.CapS,
		utils.CommonListenerS,
		utils.ConfigS,
		utils.ConnManager,
		utils.CoreS,
		utils.DB,
		utils.FilterS,
		utils.GlobalVarS,
		utils.GuardianS,
		utils.LoggerS,
	}

	services := []string{
		utils.AccountS,
		utils.ActionS,
		utils.AdminS,
		utils.AnalyzerS,
		utils.AttributeS,
		utils.CDRServer,
		utils.ChargerS,
		utils.EEs,
		utils.EFs,
		utils.ERs,
		utils.RateS,
		utils.ResourceS,
		utils.IPs,
		utils.RouteS,
		utils.SessionS,
		utils.StatS,
		utils.TPeS,
		utils.ThresholdS,
		// utils.RegistrarC,
		// utils.LoaderS,
		// utils.TrendS,
		// utils.RankingS,
	}

	// Ensure the service manager finished processing before checking states.
	// Applies only for services that have ping methods defined.
	timeout := 100 * time.Millisecond
	afterShutdown := want == utils.StateServiceDOWN
	if !afterShutdown {
		for _, id := range alwaysUp {
			engine.WaitForServiceStart(t, client, id, timeout)
		}
	}
	waitForService := engine.WaitForServiceStart
	if afterShutdown {
		waitForService = engine.WaitForServiceShutdown
	}
	for _, id := range services {
		waitForService(t, client, id, timeout)
	}

	var status map[string]string
	if err := client.Call(context.Background(), utils.ServiceManagerV1ServiceStatus,
		&servmanager.ArgsServiceID{
			ServiceID: utils.MetaAll,
		}, &status); err != nil {
		t.Error(err)
	}

	for _, id := range alwaysUp {
		if got := status[id]; got != utils.StateServiceUP {
			t.Errorf("service %q state=%q, should always be %q", id, got, utils.StateServiceUP)
		}
	}
	for _, id := range services {
		if got := status[id]; got != want {
			t.Errorf("service %q state=%q, want %q", id, got, want)
		}
	}
}
