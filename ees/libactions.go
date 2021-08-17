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
	"encoding/gob"
	"encoding/json"
	"net/http"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func init() {
	gob.Register(new(HTTPPosterRequest))
	gob.Register(new(sqlPosterRequest))

	engine.RegisterActionFunc(utils.MetaHTTPPost, callURL)
	engine.RegisterActionFunc(utils.HttpPostAsync, callURLAsync)
	engine.RegisterActionFunc(utils.MetaPostEvent, postEvent)
}

func getOneData(ub *engine.Account, extraData interface{}) ([]byte, error) {
	switch {
	case ub != nil:
		return json.Marshal(ub)
	case extraData != nil:
		return json.Marshal(extraData)
	}
	return nil, nil
}

func callURL(ub *engine.Account, a *engine.Action, _ engine.Actions, extraData interface{}) error {
	body, err := getOneData(ub, extraData)
	if err != nil {
		return err
	}
	pstr, err := NewHTTPjsonMapEE(&config.EventExporterCfg{
		ID:             a.Id,
		ExportPath:     a.ExtraParameters,
		Attempts:       config.CgrConfig().GeneralCfg().PosterAttempts,
		FailedPostsDir: config.CgrConfig().GeneralCfg().FailedPostsDir,
	}, config.CgrConfig(), nil, nil)
	if err != nil {
		return err
	}
	return ExportWithAttempts(pstr, &HTTPPosterRequest{Body: body, Header: make(http.Header)}, "")
}

// Does not block for posts, no error reports
func callURLAsync(ub *engine.Account, a *engine.Action, _ engine.Actions, extraData interface{}) error {
	body, err := getOneData(ub, extraData)
	if err != nil {
		return err
	}
	pstr, err := NewHTTPjsonMapEE(&config.EventExporterCfg{
		ID:             a.Id,
		ExportPath:     a.ExtraParameters,
		Attempts:       config.CgrConfig().GeneralCfg().PosterAttempts,
		FailedPostsDir: config.CgrConfig().GeneralCfg().FailedPostsDir,
	}, config.CgrConfig(), nil, nil)
	if err != nil {
		return err
	}
	go ExportWithAttempts(pstr, &HTTPPosterRequest{Body: body, Header: make(http.Header)}, "")
	return nil
}

func postEvent(_ *engine.Account, a *engine.Action, _ engine.Actions, extraData interface{}) error {
	body, err := json.Marshal(extraData)
	if err != nil {
		return err
	}
	pstr, err := NewHTTPjsonMapEE(&config.EventExporterCfg{
		ID:             a.Id,
		ExportPath:     a.ExtraParameters,
		Attempts:       config.CgrConfig().GeneralCfg().PosterAttempts,
		FailedPostsDir: config.CgrConfig().GeneralCfg().FailedPostsDir,
	}, config.CgrConfig(), nil, nil)
	if err != nil {
		return err
	}
	return ExportWithAttempts(pstr, &HTTPPosterRequest{Body: body, Header: make(http.Header)}, "")
}
