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

package ees

import (
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func NewHTTPPostEe(cgrCfg *config.CGRConfig, cfgIdx int, filterS *engine.FilterS,
	dc utils.MapStorage) (httpPost *HTTPPost, err error) {
	httpPost = &HTTPPost{id: cgrCfg.EEsCfg().Exporters[cfgIdx].ID,
		cgrCfg: cgrCfg, cfgIdx: cfgIdx, filterS: filterS, dc: dc}
	httpPost.httpPoster, err = engine.NewHTTPPoster(cgrCfg.GeneralCfg().ReplyTimeout,
		cgrCfg.EEsCfg().Exporters[cfgIdx].ExportPath,
		utils.PosterTransportContentTypes[cgrCfg.EEsCfg().Exporters[cfgIdx].Type],
		cgrCfg.EEsCfg().Exporters[cfgIdx].Attempts)
	return
}

// FileCSVee implements EventExporter interface for .csv files
type HTTPPost struct {
	id         string
	cgrCfg     *config.CGRConfig
	cfgIdx     int // index of config instance within ERsCfg.Readers
	filterS    *engine.FilterS
	httpPoster *engine.HTTPPoster
	sync.RWMutex
	dc utils.MapStorage
}

// ID returns the identificator of this exporter
func (httpPost *HTTPPost) ID() string {
	return httpPost.id
}

// OnEvicted implements EventExporter, doing the cleanup before exit
func (httpPost *HTTPPost) OnEvicted(_ string, _ interface{}) {
	return
}

// ExportEvent implements EventExporter
func (httpPost *HTTPPost) ExportEvent(cgrEv *utils.CGREvent) (err error) {
	httpPost.Lock()
	defer func() {
		if err != nil {
			httpPost.dc[utils.NegativeExports].(utils.StringSet).Add(cgrEv.ID)
		} else {
			httpPost.dc[utils.PositiveExports].(utils.StringSet).Add(cgrEv.ID)
		}
		httpPost.Unlock()
	}()
	httpPost.dc[utils.NumberOfEvents] = httpPost.dc[utils.NumberOfEvents].(int64) + 1

	urlVals := url.Values{}
	hdr := http.Header{}
	if len(httpPost.cgrCfg.EEsCfg().Exporters[httpPost.cfgIdx].ContentFields()) == 0 {
		for k, v := range cgrEv.Event {
			urlVals.Set(k, utils.IfaceAsString(v))
		}
	} else {
		oNm := map[string]*utils.OrderedNavigableMap{
			utils.MetaExp: utils.NewOrderedNavigableMap(),
		}
		req := utils.MapStorage(cgrEv.Event)
		eeReq := engine.NewEventRequest(req, httpPost.dc, cgrEv.APIOpts,
			httpPost.cgrCfg.EEsCfg().Exporters[httpPost.cfgIdx].Tenant,
			httpPost.cgrCfg.GeneralCfg().DefaultTenant,
			utils.FirstNonEmpty(httpPost.cgrCfg.EEsCfg().Exporters[httpPost.cfgIdx].Timezone,
				httpPost.cgrCfg.GeneralCfg().DefaultTimezone),
			httpPost.filterS, oNm)
		if err = eeReq.SetFields(httpPost.cgrCfg.EEsCfg().Exporters[httpPost.cfgIdx].ContentFields()); err != nil {
			return
		}
		for el := eeReq.OrdNavMP[utils.MetaExp].GetFirstElement(); el != nil; el = el.Next() {
			path := el.Value
			nmIt, _ := eeReq.OrdNavMP[utils.MetaExp].Field(path)
			path = path[:len(path)-1] // remove the last index
			urlVals.Set(strings.Join(path, utils.NestingSep), nmIt.String())
		}
		if hdr, err = httpPost.composeHeader(); err != nil {
			return
		}
	}
	updateEEMetrics(httpPost.dc, cgrEv.Event, utils.FirstNonEmpty(httpPost.cgrCfg.EEsCfg().Exporters[httpPost.cfgIdx].Timezone,
		httpPost.cgrCfg.GeneralCfg().DefaultTimezone))
	if err = httpPost.httpPoster.PostValues(urlVals, hdr); err != nil &&
		httpPost.cgrCfg.GeneralCfg().FailedPostsDir != utils.MetaNone {
		engine.AddFailedPost(httpPost.cgrCfg.EEsCfg().Exporters[httpPost.cfgIdx].ExportPath,
			httpPost.cgrCfg.EEsCfg().Exporters[httpPost.cfgIdx].Type, utils.EventExporterS,
			&engine.HTTPPosterRequest{
				Header: hdr,
				Body:   urlVals,
			}, httpPost.cgrCfg.EEsCfg().Exporters[httpPost.cfgIdx].Opts)
	}
	return
}

func (httpPost *HTTPPost) GetMetrics() utils.MapStorage {
	return httpPost.dc.Clone()
}

// Compose and cache the header
func (httpPost *HTTPPost) composeHeader() (hdr http.Header, err error) {
	hdr = make(http.Header)
	if len(httpPost.cgrCfg.EEsCfg().Exporters[httpPost.cfgIdx].HeaderFields()) == 0 {
		return
	}
	oNm := map[string]*utils.OrderedNavigableMap{
		utils.MetaHdr: utils.NewOrderedNavigableMap(),
	}
	eeReq := engine.NewEventRequest(nil, httpPost.dc, nil,
		httpPost.cgrCfg.EEsCfg().Exporters[httpPost.cfgIdx].Tenant,
		httpPost.cgrCfg.GeneralCfg().DefaultTenant,
		utils.FirstNonEmpty(httpPost.cgrCfg.EEsCfg().Exporters[httpPost.cfgIdx].Timezone,
			httpPost.cgrCfg.GeneralCfg().DefaultTimezone),
		httpPost.filterS, oNm)
	if err = eeReq.SetFields(httpPost.cgrCfg.EEsCfg().Exporters[httpPost.cfgIdx].HeaderFields()); err != nil {
		return
	}
	for el := eeReq.OrdNavMP[utils.MetaHdr].GetFirstElement(); el != nil; el = el.Next() {
		path := el.Value
		nmIt, _ := eeReq.OrdNavMP[utils.MetaHdr].Field(path)
		path = path[:len(path)-1] // remove the last index
		hdr.Set(strings.Join(path, utils.NestingSep), nmIt.String())
	}
	return
}
