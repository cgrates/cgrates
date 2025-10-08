//go:build integration
// +build integration

/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
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

	expected := "Invalid section: <config_db>"
	if err := cfg.V1ReloadConfig(context.Background(), args, &reply); err == nil || err.Error() != expected {
		t.Errorf("%T and %T", expected, err.Error())
		t.Errorf("Expected %q \n but received \n %q", expected, err)
	}
}
