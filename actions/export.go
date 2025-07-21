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

package actions

import (
	"cmp"
	"encoding/json"
	"net/http"
	"slices"
	"strings"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/ees"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func newActHTTPPost(ctx *context.Context, tnt string, cgrEv *utils.CGREvent,
	fltrS *engine.FilterS, cfg *config.CGRConfig, aCfg *utils.APAction) (aL *actHTTPPost, err error) {
	weights := make(map[string]float64)   // stores sorting weights by Diktat ID
	diktats := make([]*utils.APDiktat, 0) // list of diktats which have *balancePath in opts, will be weight sorted later
	data := cgrEv.AsDataProvider()
	for _, diktat := range aCfg.Diktats {
		if pass, err := fltrS.Pass(ctx, cfg.GeneralCfg().DefaultTenant, diktat.FilterIDs, data); err != nil {
			return nil, err
		} else if !pass {
			continue
		}
		weight, err := engine.WeightFromDynamics(ctx, diktat.Weights, fltrS, cfg.GeneralCfg().DefaultTenant, data)
		if err != nil {
			return nil, err
		}
		weights[diktat.ID] = weight
		diktats = append(diktats, diktat)
	}
	// Sort by weight (higher values first).
	slices.SortFunc(diktats, func(a, b *utils.APDiktat) int {
		return cmp.Compare(weights[b.ID], weights[a.ID])
	})
	aL = &actHTTPPost{
		config: cfg,
		aCfg:   aCfg,
		pstrs:  make([]*ees.HTTPjsonMapEE, len(diktats)),
	}
	for i, actD := range diktats {
		attempts, err := engine.GetIntOpts(ctx, tnt, cgrEv.AsDataProvider(), nil, fltrS, cfg.ActionSCfg().Opts.PosterAttempts,
			utils.MetaPosterAttempts)
		if err != nil {
			return nil, err
		}
		eeCfg := config.NewEventExporterCfg(aL.id(), utils.EmptyString,
			utils.IfaceAsString(actD.Opts[utils.MetaURL]),
			cfg.EEsCfg().ExporterCfg(utils.MetaDefault).FailedPostsDir,
			attempts, nil)
		aL.pstrs[i], _ = ees.NewHTTPjsonMapEE(eeCfg, cfg, nil, nil)
		if blocker, err := engine.BlockerFromDynamics(ctx, actD.Blockers, aL.fltrS, aL.config.GeneralCfg().DefaultTenant, data); err != nil {
			return nil, err
		} else if blocker {
			break
		}
	}
	return
}

type actHTTPPost struct {
	config *config.CGRConfig
	aCfg   *utils.APAction
	fltrS  *engine.FilterS

	pstrs []*ees.HTTPjsonMapEE
}

func (aL *actHTTPPost) id() string {
	return aL.aCfg.ID
}

func (aL *actHTTPPost) cfg() *utils.APAction {
	return aL.aCfg
}

// execute implements actioner interface
func (aL *actHTTPPost) execute(ctx *context.Context, data utils.MapStorage, _ string) (err error) {
	var body []byte
	if body, err = json.Marshal(data); err != nil {
		return
	}
	var partExec bool
	for _, pstr := range aL.pstrs {
		if async, has := aL.cfg().Opts[utils.MetaAsync]; has && utils.IfaceAsString(async) == utils.TrueStr {
			go ees.ExportWithAttempts(context.Background(), pstr, &ees.HTTPPosterRequest{Body: body, Header: make(http.Header)}, utils.EmptyString,
				nil, aL.config.GeneralCfg().DefaultTenant)
		} else if err = ees.ExportWithAttempts(ctx, pstr, &ees.HTTPPosterRequest{Body: body, Header: make(http.Header)}, utils.EmptyString,
			nil, aL.config.GeneralCfg().DefaultTenant); err != nil {
			if pstr.Cfg().FailedPostsDir != utils.MetaNone {
				err = nil
			} else {
				partExec = true
			}
		}
	}
	if partExec {
		err = utils.ErrPartiallyExecuted
	}
	return
}

type actExport struct {
	tnt     string
	config  *config.CGRConfig
	connMgr *engine.ConnManager
	aCfg    *utils.APAction
}

func (aL *actExport) id() string {
	return aL.aCfg.ID
}

func (aL *actExport) cfg() *utils.APAction {
	return aL.aCfg
}

// execute implements actioner interface
func (aL *actExport) execute(ctx *context.Context, data utils.MapStorage, _ string) (err error) {
	var exporterIDs []string
	if expIDs, has := aL.cfg().Opts[utils.MetaExporterIDs]; has {
		exporterIDs = strings.Split(utils.IfaceAsString(expIDs), utils.InfieldSep)
	}
	var rply map[string]map[string]any
	return aL.connMgr.Call(ctx, aL.config.ActionSCfg().EEsConns,
		utils.EeSv1ProcessEvent, &utils.CGREventWithEeIDs{
			EeIDs: exporterIDs,
			CGREvent: &utils.CGREvent{
				Tenant:  aL.tnt,
				ID:      utils.GenUUID(),
				Event:   data[utils.MetaReq].(map[string]any),
				APIOpts: data[utils.MetaOpts].(map[string]any),
			},
		}, &rply)
}
