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

package tpes

import (
	"reflect"
	"testing"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestNewTPeS(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	connMng := engine.NewConnManager(cfg)
	dataDB, err := engine.NewDataDBConn(cfg.DataDbCfg().Type,
		cfg.DataDbCfg().Host, cfg.DataDbCfg().Port,
		cfg.DataDbCfg().Name, cfg.DataDbCfg().User,
		cfg.DataDbCfg().Password, cfg.GeneralCfg().DBDataEncoding,
		cfg.DataDbCfg().Opts, cfg.DataDbCfg().Items)
	if err != nil {
		t.Error(err)
	}
	defer dataDB.Close()
	dm := engine.NewDataManager(dataDB, config.CgrConfig().CacheCfg(), connMng)
	tpExporterTypes.Add("not_valid")
	// utils.Logger, err = utils.NewLogger(utils.MetaStdLog, utils.EmptyString, 6)
	// if err != nil {
	// 	t.Error(err)
	// }
	// // utils.Logger.SetLogLevel(7)
	// buff := new(bytes.Buffer)
	// log.SetOutput(buff)
	_ = NewTPeS(cfg, dm, connMng)
	tpExporterTypes.Remove("not_valid")
	// expected := "<not_valid>"
	// if rcv := buff.String(); !strings.Contains(rcv, expected) {
	// 	t.Errorf("Expected %v, received %v", expected, rcv)
	// }
}

func TestGetTariffPlansKeys(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	connMng := engine.NewConnManager(cfg)
	dataDB, err := engine.NewDataDBConn(cfg.DataDbCfg().Type,
		cfg.DataDbCfg().Host, cfg.DataDbCfg().Port,
		cfg.DataDbCfg().Name, cfg.DataDbCfg().User,
		cfg.DataDbCfg().Password, cfg.GeneralCfg().DBDataEncoding,
		cfg.DataDbCfg().Opts, cfg.DataDbCfg().Items)
	if err != nil {
		t.Error(err)
	}
	defer dataDB.Close()
	dm := engine.NewDataManager(dataDB, config.CgrConfig().CacheCfg(), connMng)

	//Attributes
	rcv, _ := getTariffPlansKeys(context.Background(), dm, "cgrates.org", utils.MetaAttributes)
	exp := []string{"TEST_ATTRIBUTES_TEST"}
	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("Expected %v\n but received %v", exp, rcv)
	}

	//Actions
	rcv, _ = getTariffPlansKeys(context.Background(), dm, "cgrates.org", utils.MetaActions)
	exp = []string{"SET_BAL"}
	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("Expected %v\n but received %v", exp, rcv)
	}

	//Accounts
	rcv, _ = getTariffPlansKeys(context.Background(), dm, "cgrates.org", utils.MetaAccounts)
	exp = []string{"Account_simple", "Account_complicated"}
	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("Expected %v\n but received %v", exp, rcv)
	}

	//Chargers
	rcv, _ = getTariffPlansKeys(context.Background(), dm, "cgrates.org", utils.MetaChargers)
	exp = []string{"Chargers1"}
	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("Expected %v\n but received %v", exp, rcv)
	}

	//Filters
	rcv, _ = getTariffPlansKeys(context.Background(), dm, "cgrates.org", utils.MetaFilters)
	exp = []string{"fltr_for_prf"}
	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("Expected %v\n but received %v", exp, rcv)
	}

	//Rates
	rcv, _ = getTariffPlansKeys(context.Background(), dm, "cgrates.org", utils.MetaRateS)
	exp = []string{"TEST_RATE_TEST"}
	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("Expected %v\n but received %v", exp, rcv)
	}

	//Resources
	rcv, _ = getTariffPlansKeys(context.Background(), dm, "cgrates.org", utils.MetaResources)
	exp = []string{"ResGroup1"}
	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("Expected %v\n but received %v", exp, rcv)
	}

	//Routes
	rcv, _ = getTariffPlansKeys(context.Background(), dm, "cgrates.org", utils.MetaRoutes)
	exp = []string{"ROUTE_2003"}
	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("Expected %v\n but received %v", exp, rcv)
	}

	//Stats
	rcv, _ = getTariffPlansKeys(context.Background(), dm, "cgrates.org", utils.MetaStats)
	exp = []string{"SQ_2"}
	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("Expected %v\n but received %v", exp, rcv)
	}

	//Thresholds
	rcv, _ = getTariffPlansKeys(context.Background(), dm, "cgrates.org", utils.MetaThresholds)
	exp = []string{"THD_2"}
	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("Expected %v\n but received %v", exp, rcv)
	}

	//Dispatchers
	rcv, _ = getTariffPlansKeys(context.Background(), dm, "cgrates.org", utils.MetaDispatchers)
	exp = []string{"Dsp1"}
	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("Expected %v\n but received %v", exp, rcv)
	}

	//Dispatchers
	rcv, _ = getTariffPlansKeys(context.Background(), dm, "cgrates.org", utils.MetaDispatcherHosts)
	exp = []string{"DSH1"}
	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("Expected %v\n but received %v", exp, rcv)
	}

	//Unsupported
	_, err = getTariffPlansKeys(context.Background(), dm, "cgrates.org", "not_valid")
	errExpect := "Unsuported exporter type"
	if err.Error() != errExpect {
		t.Errorf("Expected %v\n but received %v", errExpect, err.Error())
	}
}

func TestV1ExportTariffPlan(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	connMng := engine.NewConnManager(cfg)
	dataDB, err := engine.NewDataDBConn(cfg.DataDbCfg().Type,
		cfg.DataDbCfg().Host, cfg.DataDbCfg().Port,
		cfg.DataDbCfg().Name, cfg.DataDbCfg().User,
		cfg.DataDbCfg().Password, cfg.GeneralCfg().DBDataEncoding,
		cfg.DataDbCfg().Opts, cfg.DataDbCfg().Items)
	if err != nil {
		t.Error(err)
	}
	defer dataDB.Close()
	dm := engine.NewDataManager(dataDB, config.CgrConfig().CacheCfg(), connMng)
	tpE := NewTPeS(cfg, dm, connMng)
	var reply []byte
	args := &ArgsExportTP{
		Tenant: utils.EmptyString,
		ExportItems: map[string][]string{
			utils.MetaAttributes: {"TEST_ATTRIBUTES_TEST"},
		},
	}
	err = tpE.V1ExportTariffPlan(context.Background(), args, &reply)
	if err != nil {
		t.Error(err)
	}
}

func TestV1ExportTariffPlanZeroExp(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	connMng := engine.NewConnManager(cfg)
	dataDB, err := engine.NewDataDBConn(cfg.DataDbCfg().Type,
		cfg.DataDbCfg().Host, cfg.DataDbCfg().Port,
		cfg.DataDbCfg().Name, cfg.DataDbCfg().User,
		cfg.DataDbCfg().Password, cfg.GeneralCfg().DBDataEncoding,
		cfg.DataDbCfg().Opts, cfg.DataDbCfg().Items)
	if err != nil {
		t.Error(err)
	}
	defer dataDB.Close()
	dm := engine.NewDataManager(dataDB, config.CgrConfig().CacheCfg(), connMng)
	tpE := NewTPeS(cfg, dm, connMng)
	var reply []byte
	args := &ArgsExportTP{
		Tenant:      utils.EmptyString,
		ExportItems: map[string][]string{},
	}
	err = tpE.V1ExportTariffPlan(context.Background(), args, &reply)
	if err != nil {
		t.Error(err)
	}
}

func TestV1ExportTariffPlanZeroIDs(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	connMng := engine.NewConnManager(cfg)
	dataDB, err := engine.NewDataDBConn(cfg.DataDbCfg().Type,
		cfg.DataDbCfg().Host, cfg.DataDbCfg().Port,
		cfg.DataDbCfg().Name, cfg.DataDbCfg().User,
		cfg.DataDbCfg().Password, cfg.GeneralCfg().DBDataEncoding,
		cfg.DataDbCfg().Opts, cfg.DataDbCfg().Items)
	if err != nil {
		t.Error(err)
	}
	defer dataDB.Close()
	dm := engine.NewDataManager(dataDB, config.CgrConfig().CacheCfg(), connMng)
	tpE := NewTPeS(cfg, dm, connMng)
	var reply []byte
	args := &ArgsExportTP{
		Tenant: utils.EmptyString,
		ExportItems: map[string][]string{
			utils.MetaAttributes: {},
		},
	}
	err = tpE.V1ExportTariffPlan(context.Background(), args, &reply)
	if err != nil {
		t.Error(err)
	}
}

func TestV1ExportTariffPlanInvalidExpType(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	connMng := engine.NewConnManager(cfg)
	dataDB, err := engine.NewDataDBConn(cfg.DataDbCfg().Type,
		cfg.DataDbCfg().Host, cfg.DataDbCfg().Port,
		cfg.DataDbCfg().Name, cfg.DataDbCfg().User,
		cfg.DataDbCfg().Password, cfg.GeneralCfg().DBDataEncoding,
		cfg.DataDbCfg().Opts, cfg.DataDbCfg().Items)
	if err != nil {
		t.Error(err)
	}
	defer dataDB.Close()
	dm := engine.NewDataManager(dataDB, config.CgrConfig().CacheCfg(), connMng)
	tpE := NewTPeS(cfg, dm, connMng)
	var reply []byte
	args := &ArgsExportTP{
		Tenant: utils.EmptyString,
		ExportItems: map[string][]string{
			"not_valid": {},
		},
	}
	err = tpE.V1ExportTariffPlan(context.Background(), args, &reply)
	errExp := "UNSUPPORTED_TPEXPORTER_TYPE:not_valid"
	if err.Error() != errExp {
		t.Errorf("Expected %v\n but received %v", errExp, err.Error())
	}

}
