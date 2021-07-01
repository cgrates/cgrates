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

package ers

import (
	"reflect"
	"testing"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

func TestERsProcessPartialEvent(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	erS := NewERService(cfg, nil, nil)
	event := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "EventERsProcessPartial",
		Event: map[string]interface{}{
			utils.OriginID: "originID",
		},
	}
	rdrCfg := &config.EventReaderCfg{
		ID:             utils.MetaDefault,
		Type:           utils.MetaNone,
		RunDelay:       0,
		ConcurrentReqs: 0,
		SourcePath:     "/var/spool/cgrates/ers/in",
		ProcessedPath:  "/var/spool/cgrates/ers/out",
		Filters:        []string{},
		Opts:           make(map[string]interface{}),
	}

	args := &erEvent{
		cgrEvent: event,
		rdrCfg:   rdrCfg,
	}
	if err := erS.processPartialEvent(args.cgrEvent, args.rdrCfg); err != nil {
		t.Error(err)
	} else {
		rcv := <-erS.rdrEvents
		if !reflect.DeepEqual(rcv, args) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>", args, rcv)
		}
	}
}
