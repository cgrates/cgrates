/*
Real-time Online/Offline Charging System (OerS) for Telecom & ISP environments
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
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

type actHTTPPost struct {
	config *config.CGRConfig
	aCfg   *engine.APAction
}

func (aL *actHTTPPost) id() string {
	return aL.aCfg.ID
}

func (aL *actHTTPPost) cfg() *engine.APAction {
	return aL.aCfg
}

// execute implements actioner interface
func (aL *actHTTPPost) execute(_ context.Context, data utils.MapStorage, _ string) (err error) {
	var body []byte
	if body, err = json.Marshal(data); err != nil {
		return
	}
	var partExec bool
	for _, actD := range aL.cfg().Diktats {
		var pstr *engine.HTTPPoster
		if pstr, err = engine.NewHTTPPoster(config.CgrConfig().GeneralCfg().ReplyTimeout, actD.Path,
			utils.ContentJSON, aL.config.GeneralCfg().PosterAttempts); err != nil {
			return
		}
		if async, has := aL.cfg().Opts[utils.MetaAsync]; has && utils.IfaceAsString(async) == utils.TrueStr {
			go aL.post(pstr, body, actD.Path)
		} else if err = aL.post(pstr, body, actD.Path); err != nil {
			partExec = true
		}
	}
	if partExec {
		err = utils.ErrPartiallyExecuted
	}
	return
}

func (aL *actHTTPPost) post(pstr *engine.HTTPPoster, body []byte, path string) (err error) {
	if err = pstr.PostValues(body, make(http.Header)); err != nil {
		utils.Logger.Warning(fmt.Sprintf("<%s> Failed posting event to: <%s> because: %s", utils.ActionS, path, err.Error()))
		if aL.config.GeneralCfg().FailedPostsDir != utils.MetaNone {
			engine.AddFailedPost(path, utils.MetaHTTPjson, utils.ActionsPoster+utils.HierarchySep+aL.cfg().Type, body, make(map[string]interface{}))
			err = nil
		}
	}
	return
}

type actExport struct {
	tnt     string
	config  *config.CGRConfig
	connMgr *engine.ConnManager
	aCfg    *engine.APAction
}

func (aL *actExport) id() string {
	return aL.aCfg.ID
}

func (aL *actExport) cfg() *engine.APAction {
	return aL.aCfg
}

// execute implements actioner interface
func (aL *actExport) execute(_ context.Context, data utils.MapStorage, _ string) (err error) {
	var exporterIDs []string
	if expIDs, has := aL.cfg().Opts[utils.MetaExporterIDs]; has {
		exporterIDs = strings.Split(utils.IfaceAsString(expIDs), utils.InfieldSep)
	}
	var rply map[string]map[string]interface{}
	return aL.connMgr.Call(aL.config.ActionSCfg().EEsConns, nil,
		utils.EeSv1ProcessEvent, &utils.CGREventWithEeIDs{
			EeIDs: exporterIDs,
			CGREvent: &utils.CGREvent{
				Tenant: aL.tnt,
				Time:   utils.TimePointer(time.Now()),
				ID:     utils.GenUUID(),
				Event:  data[utils.MetaReq].(map[string]interface{}),
				Opts:   data[utils.MetaOpts].(map[string]interface{}),
			},
		}, &rply)
}
