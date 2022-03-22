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

package ees

import (
	"reflect"
	"testing"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

func TestNewAMQPee(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dc := &utils.SafeMapStorage{
		MapStorage: utils.MapStorage{
			utils.NumberOfEvents:  int64(0),
			utils.PositiveExports: utils.StringSet{},
			utils.NegativeExports: 5,
		}}
	cfg.EEsCfg().GetDefaultExporter().ConcurrentRequests = 2
	rcv := NewAMQPee(cfg.EEsCfg().GetDefaultExporter(), dc)
	exp := &AMQPee{
		cfg:  cfg.EEsCfg().GetDefaultExporter(),
		dc:   dc,
		reqs: newConcReq(cfg.EEsCfg().GetDefaultExporter().ConcurrentRequests),
	}
	rcv.reqs = nil
	exp.reqs = nil
	exp.parseOpts(cfg.EEsCfg().GetDefaultExporter().Opts)
	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("Expected %v\n but received %v", exp, rcv)
	}
}

// func TestAMQPExportEvent(t *testing.T) {
// 	cfg := config.NewDefaultCGRConfig()
// 	dc := &utils.SafeMapStorage{
// 		MapStorage: utils.MapStorage{
// 			utils.NumberOfEvents:  int64(0),
// 			utils.PositiveExports: utils.StringSet{},
// 			utils.NegativeExports: 5,
// 		}}
// 	// cfg.EEsCfg().GetDefaultExporter().ConcurrentRequests = 2
// 	// cfg.EEsCfg().GetDefaultExporter().Opts = &config.EventExporterOpts{

// 	// }
// 	pstr := NewAMQPee(cfg.EEsCfg().GetDefaultExporter(), dc)
// 	content := "some_content"
// 	pstr.postChan =
// 	if err := pstr.ExportEvent(context.Background(), content, ""); err != nil {
// 		t.Error(err)
// 	}

// }
