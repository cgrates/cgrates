// +build integration

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
package config

import (
	"net"
	"path"
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/utils"
)

func TestNewCgrJsonCfgFromHttp(t *testing.T) {
	addr := "https://raw.githubusercontent.com/cgrates/cgrates/master/data/conf/samples/tutmongo/cgrates.json"
	expVal, err := NewCgrJsonCfgFromFile(path.Join("/usr", "share", "cgrates", "conf", "samples", "tutmongo", "cgrates.json"))
	if err != nil {
		t.Fatal(err)
	}

	if _, err = net.DialTimeout("tcp", addr, time.Second); err != nil { // check if site is up
		return
	}

	if rply, err := NewCgrJsonCfgFromHttp(addr); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expVal, rply) {
		t.Errorf("Expected: %s ,received: %s", utils.ToJSON(expVal), utils.ToJSON(rply))
	}

}

func TestNewCGRConfigFromPath(t *testing.T) {
	addr := "https://raw.githubusercontent.com/cgrates/cgrates/master/data/conf/samples/tutmongo/cgrates.json"
	expVal, err := NewCGRConfigFromPath(path.Join("/usr", "share", "cgrates", "conf", "samples", "tutmongo"))
	if err != nil {
		t.Fatal(err)
	}

	if _, err = net.DialTimeout("tcp", addr, time.Second); err != nil { // check if site is up
		return
	}

	if rply, err := NewCGRConfigFromPath(addr); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expVal, rply) {
		t.Errorf("Expected: %s ,received: %s", utils.ToJSON(expVal), utils.ToJSON(rply))
	}

}
