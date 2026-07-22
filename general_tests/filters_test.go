// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package general_tests

import (
	"testing"
	"time"

	"github.com/cgrates/birpc"
	v1 "github.com/cgrates/cgrates/apier/v1"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestFilterPassDestinations(t *testing.T) {
	if err := engine.Cache.Set(utils.CacheDestinations, "DE",
		&engine.Destination{Id: "DE", Prefixes: []string{"+49"}}, nil, true, ""); err != nil {
		t.Errorf("Expecting: nil, received: %s", err)
	}
	if err := engine.Cache.Set(utils.CacheDestinations, "EU_LANDLINE",
		&engine.Destination{Id: "EU_LANDLINE", Prefixes: []string{"+49"}}, nil, true, ""); err != nil {
		t.Errorf("Expecting: nil, received: %s", err)
	}
	config.CgrConfig().FilterSCfg().ApierSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaApier)}
	internalAPIerSv1Chan := make(chan birpc.ClientConnector, 1)
	connMgr := engine.NewConnManager(config.CgrConfig(), map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaApier): internalAPIerSv1Chan,
	})
	data, err := engine.NewInternalDB(nil, nil, true, nil, config.CgrConfig().DataDbCfg().Items)
	if err != nil {
		t.Error(err)
	}
	dm := engine.NewDataManager(data, config.CgrConfig().CacheCfg(), connMgr)
	srv, err := engine.NewService(&v1.APIerSv1{DataManager: dm})
	if err != nil {
		t.Error(err)
	}
	internalAPIerSv1Chan <- srv
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
	cfg := config.NewDefaultCGRConfig()
	cfg.FilterSCfg().ApierSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaApier)}
	internalAPIerSv1Chan := make(chan birpc.ClientConnector, 1)
	connMgr := engine.NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaApier): internalAPIerSv1Chan,
	})
	data, err := engine.NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	if err != nil {
		t.Error(err)
	}
	dmFilterPass := engine.NewDataManager(data, cfg.CacheCfg(), connMgr)
	filterS := engine.NewFilterS(cfg, connMgr, dmFilterPass)
	if err := engine.Cache.Set(utils.CacheReverseDestinations, "+49",
		[]string{"DE", "EU_LANDLINE"}, nil, true, ""); err != nil {
		t.Errorf("Expecting: nil, received: %s", err)
	}

	apiSrv, err := engine.NewService(&v1.APIerSv1{DataManager: dmFilterPass})
	if err != nil {
		t.Fatal(err)
	}
	internalAPIerSv1Chan <- apiSrv
	engine.SetConnManager(connMgr)
	failEvent := map[string]any{
		utils.Destination: "+5086517174963",
	}
	passEvent := map[string]any{
		utils.Destination: "+4986517174963",
	}
	fEv := utils.MapStorage{utils.MetaReq: failEvent}
	pEv := utils.MapStorage{utils.MetaReq: passEvent}
	if pass, err := filterS.Pass("cgrates.org",
		[]string{"*destinations:~*req.Destination:EU"}, fEv); err != nil {
		t.Error(err)
	} else if pass {
		t.Errorf("Expecting: %+v, received: %+v", false, pass)
	}
	if pass, err := filterS.Pass("cgrates.org",
		[]string{"*destinations:~*req.Destination:EU_LANDLINE"}, pEv); err != nil {
		t.Error(err)
	} else if !pass {
		t.Errorf("Expecting: %+v, received: %+v", true, pass)
	}
}
