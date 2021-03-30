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
	"testing"
	"time"

	v1 "github.com/cgrates/cgrates/apier/v1"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

func TestFilterPassDestinations(t *testing.T) {
	if err := engine.Cache.Set(utils.CacheReverseDestinations, "+49",
		[]string{"DE", "EU_LANDLINE"}, nil, true, ""); err != nil {
		t.Errorf("Expecting: nil, received: %s", err)
	}
	config.CgrConfig().FilterSCfg().ApierSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaApier)}
	internalAPIerSv1Chan := make(chan rpcclient.ClientConnector, 1)
	connMgr := engine.NewConnManager(config.CgrConfig(), map[string]chan rpcclient.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaApier): internalAPIerSv1Chan,
	})
	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, config.CgrConfig().CacheCfg(), connMgr)
	internalAPIerSv1Chan <- &v1.APIerSv1{DataManager: dm}
	engine.SetConnManager(connMgr)
	cd := &utils.MapStorage{
		utils.Category:      "call",
		utils.Tenant:        "cgrates.org",
		utils.Subject:       "dan",
		utils.Destination:   "+4986517174963",
		utils.TimeStart:     time.Date(2013, time.October, 7, 14, 50, 0, 0, time.UTC),
		utils.TimeEnd:       time.Date(2013, time.October, 7, 14, 52, 12, 0, time.UTC),
		utils.DurationIndex: 132 * time.Second,
		utils.ExtraFields:   map[string]string{"navigation": "off"},
	}
	rf, err := engine.NewFilterRule(utils.MetaDestinations, "~Destination", []string{"DE"})
	if err != nil {
		t.Error(err)
	}
	if passes, err := rf.Pass(cd); err != nil {
		t.Error(err)
	} else if !passes {
		t.Error("Not passing")
	}
	rf, err = engine.NewFilterRule(utils.MetaDestinations, "~Destination", []string{"RO"})
	if err != nil {
		t.Error(err)
	}
	if passes, err := rf.Pass(cd); err != nil {
		t.Error(err)
	} else if passes {
		t.Error("Passing")
	}
	//not
	rf, err = engine.NewFilterRule(utils.MetaNotDestinations, "~Destination", []string{"DE"})
	if err != nil {
		t.Error(err)
	}
	if passes, err := rf.Pass(cd); err != nil {
		t.Error(err)
	} else if passes {
		t.Error("Passing")
	}
}

func TestInlineFilterPassFiltersForEvent(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.FilterSCfg().ApierSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaApier)}
	internalAPIerSv1Chan := make(chan rpcclient.ClientConnector, 1)
	connMgr := engine.NewConnManager(cfg, map[string]chan rpcclient.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaApier): internalAPIerSv1Chan,
	})
	data := engine.NewInternalDB(nil, nil, true)
	dmFilterPass := engine.NewDataManager(data, cfg.CacheCfg(), connMgr)
	filterS := engine.NewFilterS(cfg, connMgr, dmFilterPass)
	if err := engine.Cache.Set(utils.CacheReverseDestinations, "+49",
		[]string{"DE", "EU_LANDLINE"}, nil, true, ""); err != nil {
		t.Errorf("Expecting: nil, received: %s", err)
	}
	internalAPIerSv1Chan <- &v1.APIerSv1{DataManager: dmFilterPass}
	engine.SetConnManager(connMgr)
	failEvent := map[string]interface{}{
		utils.Destination: "+5086517174963",
	}
	passEvent := map[string]interface{}{
		utils.Destination: "+4986517174963",
	}
	fEv := utils.MapStorage{utils.MetaReq: failEvent}
	pEv := utils.MapStorage{utils.MetaReq: passEvent}
	if pass, err := filterS.Pass("cgrates.org",
		[]string{"*destinations:~*req.Destination:EU"}, fEv); err != nil {
		t.Errorf(err.Error())
	} else if pass {
		t.Errorf("Expecting: %+v, received: %+v", false, pass)
	}
	if pass, err := filterS.Pass("cgrates.org",
		[]string{"*destinations:~*req.Destination:EU_LANDLINE"}, pEv); err != nil {
		t.Errorf(err.Error())
	} else if !pass {
		t.Errorf("Expecting: %+v, received: %+v", true, pass)
	}
}
