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

package apis

import (
	"reflect"
	"testing"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
)

func TestConfigNewConfigSv1(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	expected := &ConfigSv1{
		cfg: cfg,
	}
	result := NewConfigSv1(cfg)
	if !reflect.DeepEqual(expected, result) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestConfigReloadConfigError(t *testing.T) {
	cfgDflt := config.NewDefaultCGRConfig()
	cfg := NewConfigSv1(cfgDflt)
	var reply *string
	args := &config.ReloadArgs{}
	err := cfg.ReloadConfig(context.Background(), args, reply)
	expected := "MANDATORY_IE_MISSING: [Path]"
	if err == nil || err.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, err)
	}
}
