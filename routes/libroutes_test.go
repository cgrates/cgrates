// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package routes

import (
	"testing"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestRouteSortDispatcherSortRoutes(t *testing.T) {
	ssd := RouteSortDispatcher{}
	suppls := map[string]*RouteWithWeight{}
	suplEv := new(utils.CGREvent)
	extraOpts := new(optsGetRoutes)
	expErr := "unsupported sorting strategy: "
	if _, err := ssd.SortRoutes(context.Background(), "", "", suppls, suplEv, extraOpts); err.Error() != expErr {
		t.Errorf("Expected error <%v>, Received <%v>", expErr, err)
	}
}

func TestRouteLazyPassErr(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()

	filters := []*engine.FilterRule{
		{
			Type:    "nr1",
			Element: "nr1",
			Values:  []string{"nr1"},
		},
		{
			Type:    "nr2",
			Element: "nr2",
			Values:  []string{"nr2"},
		},
	}

	ev := new(utils.CGREvent)
	data := utils.MapStorage{
		"FirstLevel":        map[string]any{},
		"AnotherFirstLevel": "ValAnotherFirstLevel",
	}

	expErr := "NOT_IMPLEMENTED:nr1"
	if _, err := routeLazyPass(context.Background(), filters, ev,
		data, cfg, nil); err == nil || err.Error() != expErr {
		t.Errorf("Expected error <%v>, received <%v>", expErr, err)
	}

}

func TestRouteLazyPassTrue(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	filter, err := engine.NewFilterFromInline(
		"cgrates.org",
		"*string:~*opts.<~*opts.*originID;~*req.RunID;-Cost>:~*opts.<~*opts.*originID;~*req.RunID;-Cost>",
	)
	if err != nil {
		t.Fatal(err)
	}

	rules := filter.Rules

	ev := &utils.CGREvent{
		ID:     "cgrId",
		Tenant: "cgrates.org",
		Event: utils.MapStorage{

			utils.RunID: utils.MetaDefault,
		},
		APIOpts: utils.MapStorage{
			utils.MetaOriginID:  "Uniq",
			"Uniq*default-Cost": 10,
		},
	}
	data := utils.MapStorage{
		utils.MetaReq: utils.MapStorage{

			utils.RunID: utils.MetaDefault,
		},
		utils.MetaOpts: utils.MapStorage{
			utils.MetaOriginID:  "Uniq",
			"Uniq*default-Cost": 10,
		},
	}

	if ok, err := routeLazyPass(context.Background(), rules, ev,
		data, cfg, nil); err != nil {
		t.Error(err)
	} else if !ok {
		t.Error("Returned false, expecting true")
	}

}
