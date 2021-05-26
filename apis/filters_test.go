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

package apis

import (
	"reflect"
	"testing"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestApisSetGetGetIDsCountFilters(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg, nil)
	dataDB := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(dataDB, nil, connMgr)
	admS := NewAdminSv1(cfg, dm, connMgr)
	fltr := &engine.FilterWithAPIOpts{
		Filter: &engine.Filter{
			Tenant: utils.CGRateSorg,
			ID:     "fltr_for_attr",
			Rules: []*engine.FilterRule{
				{
					Type:    utils.MetaString,
					Element: "~*req.Subject",
					Values:  []string{"1004", "6774", "22312"},
				},
			},
		},
	}
	var reply string
	err := admS.SetFilter(context.Background(), fltr, &reply)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	if !reflect.DeepEqual(`"OK"`, utils.ToJSON(&reply)) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", `"OK"`, utils.ToJSON(&reply))
	}

	var replyGet engine.Filter
	argsGet := &utils.TenantID{
		Tenant: utils.CGRateSorg,
		ID:     "fltr_for_attr",
	}

	err = admS.GetFilter(context.Background(), argsGet, &replyGet)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	if !reflect.DeepEqual(&replyGet, fltr.Filter) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", fltr.Filter, &replyGet)
	}

	var replyCnt int
	argsCnt := &utils.TenantWithAPIOpts{
		Tenant: utils.CGRateSorg,
	}
	err = admS.GetFilterIDsCount(context.Background(), argsCnt, &replyCnt)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	if !reflect.DeepEqual(utils.ToJSON(&replyCnt), utils.ToJSON(1)) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", utils.ToJSON(1), utils.ToJSON(&replyCnt))
	}
	fltr2 := &engine.FilterWithAPIOpts{
		Filter: &engine.Filter{
			Tenant: utils.CGRateSorg,
			ID:     "fltr_for_attr2",
			Rules: []*engine.FilterRule{
				{
					Type:    utils.MetaString,
					Element: "~*req.Subject",
					Values:  []string{"1004", "6774", "22312"},
				},
			},
		},
	}
	var reply2 string
	err = admS.SetFilter(context.Background(), fltr2, &reply2)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	var replyCnt2 int
	argsCnt2 := &utils.TenantWithAPIOpts{
		Tenant: utils.CGRateSorg,
	}
	err = admS.GetFilterIDsCount(context.Background(), argsCnt2, &replyCnt2)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	if !reflect.DeepEqual(utils.ToJSON(&replyCnt2), utils.ToJSON(2)) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", utils.ToJSON(2), utils.ToJSON(&replyCnt2))
	}

}
