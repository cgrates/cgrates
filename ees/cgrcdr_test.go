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
package ees

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
)

func TestCgrCDRInitDialectorUnsupported(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	cgr := &CgrCDR{
		cfg:  cgrCfg.EEsCfg().Exporters[0],
		reqs: newConcReq(0),
	}
	errExpect := fmt.Sprintf("db type <%s> not supported", cgr.cfg.ExportPath)
	if err := cgr.initDialector(); err == nil || err.Error() == errExpect {
		t.Errorf("Expected %v but received %v", errExpect, err)
	}
}

func TestCgrCDRInitDialectorMySQL(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	cgrCfg.EEsCfg().Exporters[0].Opts.SQLDBName = utils.StringPointer("cgrates")
	cgrCfg.EEsCfg().Exporters[0].ExportPath = `mysql://cgrates:CGRateS.org@127.0.0.1:3306`
	cgr := &CgrCDR{
		cfg:  cgrCfg.EEsCfg().Exporters[0],
		reqs: newConcReq(0),
	}
	dialectExpect := mysql.Open(fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&loc=Local&parseTime=true&sql_mode='ALLOW_INVALID_DATES'",
		"cgrates", "CGRateS.org", "127.0.0.1", "3306", "cgrates"))
	if err := cgr.initDialector(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(cgr.dialect, dialectExpect) {
		t.Errorf("Expected %v but received %v", utils.ToJSON(dialectExpect), utils.ToJSON(cgr.dialect))
	}
	if cgr.tableName != utils.CDRsTBL {
		t.Errorf("Expected tableName %q but received %q", utils.CDRsTBL, cgr.tableName)
	}
}

func TestCgrCDRInitDialectorPostgres(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	cgrCfg.EEsCfg().Exporters[0].Opts.SQLDBName = utils.StringPointer("cgrates")
	cgrCfg.EEsCfg().Exporters[0].ExportPath = `postgres://cgrates:CGRateS.org@127.0.0.1:5432`
	cgr := &CgrCDR{
		cfg:  cgrCfg.EEsCfg().Exporters[0],
		reqs: newConcReq(0),
	}
	dialectExpect := postgres.Open(fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=%s",
		"127.0.0.1", "5432", "cgrates", "cgrates", "CGRateS.org", utils.SQLDefaultPgSSLMode))
	if err := cgr.initDialector(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(cgr.dialect, dialectExpect) {
		t.Errorf("Expected %v but received %v", utils.ToJSON(dialectExpect), utils.ToJSON(cgr.dialect))
	}
}

func TestCgrCDRInitDialectorURLError(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	cgrCfg.EEsCfg().Exporters[0].ExportPath = ":exportpath"
	cgr := &CgrCDR{
		cfg:  cgrCfg.EEsCfg().Exporters[0],
		reqs: newConcReq(0),
	}
	errExpect := `parse ":exportpath": missing protocol scheme`
	if err := cgr.initDialector(); err == nil || err.Error() != errExpect {
		t.Errorf("Expected %v but received %v", errExpect, err)
	}
}

func TestCgrCDRExportEventDisconnected(t *testing.T) {
	cgr := &CgrCDR{reqs: newConcReq(0)}
	cgrEv := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "evt1",
		Event:  map[string]any{utils.OriginID: "origin1"},
	}
	if err := cgr.ExportEvent(context.Background(), nil, cgrEv); err != utils.ErrDisconnected {
		t.Errorf("Expected %v but received %v", utils.ErrDisconnected, err)
	}
}
