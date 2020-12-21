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

/*
import (
	"reflect"
	"sync"
	"testing"

	v2 "github.com/cgrates/cgrates/apier/v2"

	v1 "github.com/cgrates/cgrates/apier/v1"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/cores"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

//TestApiersCoverage for cover testing
func TestApiersCoverage(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	filterSChan := make(chan *engine.FilterS, 1)
	filterSChan <- nil
	shdChan := utils.NewSyncedChan()
	chS := engine.NewCacheS(cfg, nil, nil)
	cfg.ThresholdSCfg().Enabled = true
	cfg.SchedulerCfg().Enabled = true
	server := cores.NewServer(nil)
	srvDep := map[string]*sync.WaitGroup{utils.DataDB: new(sync.WaitGroup)}
	db := NewDataDBService(cfg, nil, srvDep)
	cfg.StorDbCfg().Type = utils.INTERNAL
	stordb := NewStorDBService(cfg, srvDep)
	anz := NewAnalyzerService(cfg, server, filterSChan, shdChan, make(chan rpcclient.ClientConnector, 1), srvDep)
	schS := NewSchedulerService(cfg, db, chS, filterSChan, server, make(chan rpcclient.ClientConnector, 1), nil, anz, srvDep)
	apiSv1 := NewAPIerSv1Service(cfg, db, stordb, filterSChan, server, schS, new(ResponderService),
		make(chan rpcclient.ClientConnector, 1), nil, anz, srvDep)
	apiSv2 := NewAPIerSv2Service(apiSv1, cfg, server, make(chan rpcclient.ClientConnector, 1), anz, srvDep)
	if apiSv1.IsRunning() {
		t.Errorf("Expected service to be down")
	}
	if apiSv2.IsRunning() {
		t.Errorf("Expected service to be down")
	}

	apiSv1.api = &v1.APIerSv1{}
	apiSv2.api = &v2.APIerSv2{}
	if !apiSv1.IsRunning() {
		t.Errorf("Expected service to be running")
	}
	if !apiSv2.IsRunning() {
		t.Errorf("Expected service to be running")
	}

	err := apiSv1.Start()
	if err == nil || err != utils.ErrServiceAlreadyRunning {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", utils.ErrServiceAlreadyRunning, err)
	}
	err2 := apiSv2.Start()
	if err2 == nil || err2 != utils.ErrServiceAlreadyRunning {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", utils.ErrServiceAlreadyRunning, err2)
	}
	err = apiSv1.Reload()
	if err != nil {
		t.Errorf("\nExpecting <nil>,\n Received <%+v>", err)
	}
	err2 = apiSv2.Reload()
	if err2 != nil {
		t.Errorf("\nExpecting <nil>,\n Received <%+v>", err2)
	}
	serviceName := apiSv1.ServiceName()
	if serviceName != utils.APIerSv1 {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", utils.APIerSv1, serviceName)
	}
	serviceName2 := apiSv2.ServiceName()
	if serviceName2 != utils.APIerSv2 {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", utils.APIerSv2, serviceName2)
	}
	getApi1 := apiSv1.GetAPIerSv1()
	if !reflect.DeepEqual(getApi1, apiSv1.api) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", apiSv1.api, getApi1)
	}
	getApiChan1 := apiSv1.GetAPIerSv1Chan()
	if !reflect.DeepEqual(getApiChan1, apiSv1.APIerSv1Chan) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", apiSv1.APIerSv1Chan, getApiChan1)
	}
	shouldRun := apiSv1.ShouldRun()
	if shouldRun != false {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", false, shouldRun)
	}
	shouldRun2 := apiSv2.ShouldRun()
	if shouldRun2 != false {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", false, shouldRun2)
	}
	//populates apiSv1 and apiSv2 with something in order to call the close function
	apiSv1.stopChan = make(chan struct{}, 1)
	apiSv1.stopChan <- struct{}{}
	apiSv1.connChan = make(chan rpcclient.ClientConnector, 1)
	apiSv1.connChan <- chS
	shutdownApi1 := apiSv1.Shutdown()
	if shutdownApi1 != nil {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", nil, shutdownApi1)
	}
	apiSv2.connChan = make(chan rpcclient.ClientConnector, 1)
	apiSv2.connChan <- chS
	shutdownApi2 := apiSv2.Shutdown()
	if shutdownApi2 != nil {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", nil, shutdownApi2)
	}
}
*/
