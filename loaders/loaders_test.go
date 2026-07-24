// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package loaders

import (
	"fmt"
	"testing"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestLoaderServiceV1Load_UnknownLoader(t *testing.T) {
	loaderService := &LoaderService{}
	args := &ArgsProcessFolder{
		LoaderID:  "unknown_loader",
		ForceLock: false,
	}
	var reply string
	err := loaderService.V1Load(args, &reply)
	if err == nil || err.Error() != fmt.Sprintf("UNKNOWN_LOADER: %s", args.LoaderID) {
		t.Errorf("V1Load() error = %v, wantErr %v", err, fmt.Errorf("UNKNOWN_LOADER: %s", args.LoaderID))
	}
	if reply != "" {
		t.Errorf("V1Load() reply = %v, want %v", reply, "")
	}
}

func TestLoaderServiceReload(t *testing.T) {
	loaderService := &LoaderService{}
	dataManager := &engine.DataManager{}
	timezone := "UTC"
	exitChan := make(chan bool)
	filterS := &engine.FilterS{}
	connMgr := &engine.ConnManager{}
	loaderConfigs := []*config.LoaderSCfg{
		{Id: "loader1", Enabled: true, CacheSConns: []string{"conn1"}},
		{Id: "loader2", Enabled: false, CacheSConns: []string{"conn2"}},
		{Id: "loader3", Enabled: true, CacheSConns: []string{"conn3"}},
	}

	loaderService.Reload(dataManager, loaderConfigs, timezone, exitChan, filterS, connMgr)

	if len(loaderService.ldrs) != 2 {
		t.Errorf("expected 2 loaders, got %d", len(loaderService.ldrs))
	}

	if _, exists := loaderService.ldrs["loader1"]; !exists {
		t.Errorf("expected loader1 to be present in ldrs map")
	}

	if _, exists := loaderService.ldrs["loader3"]; !exists {
		t.Errorf("expected loader3 to be present in ldrs map")
	}

	if _, exists := loaderService.ldrs["loader2"]; exists {
		t.Errorf("did not expect loader2 to be present in ldrs map")
	}
}

func TestLoaderServiceV1RemoveValidLoader(t *testing.T) {
	loaderService := &LoaderService{
		ldrs: map[string]*Loader{},
	}
	args := &ArgsProcessFolder{
		LoaderID:  "existing_loader",
		ForceLock: false,
	}
	var reply string
	err := loaderService.V1Remove(args, &reply)
	if err == nil {
		t.Errorf("V1Remove() error = %v, wantErr %v", err, nil)
	}
	if reply == utils.OK {
		t.Errorf("V1Remove() reply = %v, want %v", reply, utils.OK)
	}
}

func TestLoaderServiceV1RemoveUnknownLoader(t *testing.T) {
	loaderService := &LoaderService{
		ldrs: map[string]*Loader{},
	}
	args := &ArgsProcessFolder{
		LoaderID:  "unknown_loader",
		ForceLock: false,
	}
	var reply string
	err := loaderService.V1Remove(args, &reply)
	if err == nil || err.Error() != fmt.Sprintf("UNKNOWN_LOADER: %s", args.LoaderID) {
		t.Errorf("V1Remove() error = %v, wantErr %v", err, fmt.Errorf("UNKNOWN_LOADER: %s", args.LoaderID))
	}
	if reply != "" {
		t.Errorf("V1Remove() reply = %v, want %v", reply, "")
	}
}

func TestLoaderServiceV1RemoveForceUnlockLoader(t *testing.T) {
	loaderService := &LoaderService{
		ldrs: map[string]*Loader{},
	}
	args := &ArgsProcessFolder{
		LoaderID:  "locked_loader",
		ForceLock: true,
	}
	var reply string
	err := loaderService.V1Remove(args, &reply)
	if err == nil {
		t.Errorf("V1Remove() error = %v, wantErr %v", err, nil)
	}
	if reply == utils.OK {
		t.Errorf("V1Remove() reply = %v, want %v", reply, utils.OK)
	}
}

func TestLoaderServiceEnabledNoLoaders(t *testing.T) {
	loaderService := &LoaderService{
		ldrs: make(map[string]*Loader),
	}
	if got := loaderService.Enabled(); got != false {
		t.Errorf("Enabled() = %v, want false", got)
	}
}

func TestLoaderServiceEnabledAllDisabled(t *testing.T) {
	loaderService := &LoaderService{
		ldrs: map[string]*Loader{
			"loader1": {enabled: false},
			"loader2": {enabled: false},
		},
	}
	if got := loaderService.Enabled(); got != false {
		t.Errorf("Enabled() = %v, want false", got)
	}
}

func TestLoaderServiceEnabledAtLeastOneEnabled(t *testing.T) {
	loaderService := &LoaderService{
		ldrs: map[string]*Loader{
			"loader1": {enabled: false},
			"loader2": {enabled: true},
		},
	}
	if got := loaderService.Enabled(); got != true {
		t.Errorf("Enabled() = %v, want true", got)
	}
}

func TestNewLoaderService(t *testing.T) {
	dm := &engine.DataManager{}
	filterS := &engine.FilterS{}
	connMgr := &engine.ConnManager{}
	exitChan := make(chan bool)

	ldrsCfg := []*config.LoaderSCfg{
		{Id: "loader1", Enabled: true, CacheSConns: []string{"conn1"}},
		{Id: "loader2", Enabled: false, CacheSConns: []string{"conn2"}},
		{Id: "loader3", Enabled: true, CacheSConns: []string{"conn3"}},
	}

	ldrS := NewLoaderService(dm, ldrsCfg, "UTC", exitChan, filterS, connMgr)

	if len(ldrS.ldrs) != 2 {
		t.Errorf("NewLoaderService() created %d loaders, want %d", len(ldrS.ldrs), 2)
	}

	if _, exists := ldrS.ldrs["loader1"]; !exists {
		t.Errorf("NewLoaderService() missing loader with ID 'loader1'")
	}

	if _, exists := ldrS.ldrs["loader3"]; !exists {
		t.Errorf("NewLoaderService() missing loader with ID 'loader3'")
	}

	if _, exists := ldrS.ldrs["loader2"]; exists {
		t.Errorf("NewLoaderService() should not include loader with ID 'loader2'")
	}
}
