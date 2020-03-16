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
	engine.Cache.Set(utils.CacheReverseDestinations, "+49",
		[]string{"DE", "EU_LANDLINE"}, nil, true, "")
	config.CgrConfig().FilterSCfg().ApierSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaApier)}
	internalAPIerSv1Chan := make(chan rpcclient.ClientConnector, 1)
	connMgr := engine.NewConnManager(config.CgrConfig(), map[string]chan rpcclient.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaApier): internalAPIerSv1Chan,
	})
	data := engine.NewInternalDB(nil, nil, true, config.CgrConfig().DataDbCfg().Items)
	dm := engine.NewDataManager(data, config.CgrConfig().CacheCfg(), connMgr)
	internalAPIerSv1Chan <- &v1.APIerSv1{DataManager: dm}
	engine.SetConnManager(connMgr)
	cd := &engine.CallDescriptor{
		Category:      "call",
		Tenant:        "cgrates.org",
		Subject:       "dan",
		Destination:   "+4986517174963",
		TimeStart:     time.Date(2013, time.October, 7, 14, 50, 0, 0, time.UTC),
		TimeEnd:       time.Date(2013, time.October, 7, 14, 52, 12, 0, time.UTC),
		DurationIndex: 132 * time.Second,
		ExtraFields:   map[string]string{"navigation": "off"},
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
	cfg, _ := config.NewDefaultCGRConfig()
	cfg.FilterSCfg().ApierSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaApier)}
	internalAPIerSv1Chan := make(chan rpcclient.ClientConnector, 1)
	connMgr := engine.NewConnManager(cfg, map[string]chan rpcclient.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaApier): internalAPIerSv1Chan,
	})
	data := engine.NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dmFilterPass := engine.NewDataManager(data, cfg.CacheCfg(), connMgr)
	filterS := engine.NewFilterS(cfg, connMgr, dmFilterPass)
	engine.Cache.Set(utils.CacheReverseDestinations, "+49",
		[]string{"DE", "EU_LANDLINE"}, nil, true, "")
	internalAPIerSv1Chan <- &v1.APIerSv1{DataManager: dmFilterPass}
	engine.SetConnManager(connMgr)
	failEvent := map[string]interface{}{
		utils.Destination: "+5086517174963",
	}
	passEvent := map[string]interface{}{
		utils.Destination: "+4986517174963",
	}
	fEv := utils.NavigableMap{utils.MetaReq: failEvent}
	pEv := utils.NavigableMap{utils.MetaReq: passEvent}
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
