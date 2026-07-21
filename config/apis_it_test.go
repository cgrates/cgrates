//go:build integration
// +build integration

// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package config

import (
	"path"
	"testing"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"
)

func TestV1ReloadConfig(t *testing.T) {
	cfg := NewDefaultCGRConfig()
	cfg.db = &CgrJsonCfg{}
	cfg.ConfigPath = path.Join("/usr", "share", "cgrates", "conf", "samples", "tutmongo2")
	args := &ReloadArgs{
		Section: utils.MetaAll,
	}

	cfg.rldCh = make(chan string, 100)

	var reply string
	if err := cfg.V1ReloadConfig(context.Background(), args, &reply); err != nil {
		t.Error(err)
	} else if reply != "OK" {
		t.Errorf("Expected %v \n but received \n %v", "OK", reply)
	}

	args = &ReloadArgs{
		Section: ConfigDBJSON,
	}

	expected := "Invalid section: <configDB>"
	if err := cfg.V1ReloadConfig(context.Background(), args, &reply); err == nil || err.Error() != expected {
		t.Errorf("%T and %T", expected, err.Error())
		t.Errorf("Expected %q \n but received \n %q", expected, err)
	}
}
