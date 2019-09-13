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
package services

import (
	"path"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

func TestAttributeSReload(t *testing.T) {
	cfg, err := config.NewDefaultCGRConfig()
	if err != nil {
		t.Fatal(err)
	}
	filterSChan := make(chan *engine.FilterS, 1)
	filterSChan <- nil
	engineShutdown := make(chan bool, 1)
	chS := engine.NewCacheS(cfg, nil)
	close(chS.GetPrecacheChannel(utils.CacheAttributeProfiles))
	close(chS.GetPrecacheChannel(utils.CacheAttributeFilterIndexes))
	server := utils.NewServer()
	srvMngr := servmanager.NewServiceManager(cfg /*dm*/, nil,
		chS /*cdrStorage*/, nil,
		/*loadStorage*/ nil, filterSChan,
		server, engineShutdown)
	attrS := NewAttributeService()
	srvMngr.AddService(attrS)
	if err = srvMngr.StartServices(); err != nil {
		t.Error(err)
	}
	if attrS.IsRunning() {
		t.Errorf("Expected service to be down")
	}
	var reply string
	cfg.V1ReloadConfig(&config.ConfigReloadWithArgDispatcher{
		Path:    path.Join("/usr", "share", "cgrates", "conf", "samples", "tutmongo"),
		Section: config.ATTRIBUTE_JSN,
	}, &reply)
	time.Sleep(1) //need to switch to gorutine
	if !attrS.IsRunning() {
		t.Errorf("Expected service to be running")
	}
	cfg.AttributeSCfg().Enabled = false
	cfg.GetReloadChan(config.ATTRIBUTE_JSN) <- struct{}{}
	time.Sleep(1)
	if attrS.IsRunning() {
		t.Errorf("Expected service to be down")
	}
	engineShutdown <- true
}
