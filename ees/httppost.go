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
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func NewHTTPPostEe(cgrCfg *config.CGRConfig, cfgIdx int, filterS *engine.FilterS,
	dc utils.MapStorage) (httpPost *HTTPPost, err error) {
	dc[utils.ExportID] = cgrCfg.EEsCfg().Exporters[cfgIdx].ID
	httpPost = &HTTPPost{id: cgrCfg.EEsCfg().Exporters[cfgIdx].ID,
		cgrCfg: cgrCfg, cfgIdx: cfgIdx, filterS: filterS, dc: dc}
	httpPost.httpPoster, err = engine.NewHTTPPoster(cgrCfg.GeneralCfg().HttpSkipTlsVerify,
		cgrCfg.GeneralCfg().ReplyTimeout, cgrCfg.EEsCfg().Exporters[cfgIdx].ExportPath,
		utils.PosterTransportContentTypes[cgrCfg.EEsCfg().Exporters[cfgIdx].Type], cgrCfg.EEsCfg().Exporters[cfgIdx].Attempts)
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
	defer httpPost.Unlock()

	httpPost.dc[utils.NumberOfEvents] = httpPost.dc[utils.NumberOfEvents].(int) + 1

	var body interface{}
	urlVals := url.Values{}
	req := utils.MapStorage{}
	for k, v := range cgrEv.Event {
		req[k] = v
	}
	eeReq := NewEventExporterRequest(req, httpPost.dc, cgrEv.Tenant, httpPost.cgrCfg.GeneralCfg().DefaultTimezone,
		httpPost.filterS)

	if err = eeReq.SetFields(httpPost.cgrCfg.EEsCfg().Exporters[httpPost.cfgIdx].ContentFields()); err != nil {
		httpPost.dc[utils.NegativeExports].(utils.StringSet).Add(cgrEv.ID)
		return
	}
	for el := eeReq.cnt.GetFirstElement(); el != nil; el = el.Next() {
		var nmIt utils.NMInterface
		if nmIt, err = eeReq.cnt.Field(el.Value); err != nil {
			return
		}
		itm, isNMItem := nmIt.(*config.NMItem)
		if !isNMItem {
			return fmt.Errorf("cannot encode reply value: %s, err: not NMItems", utils.ToJSON(el.Value))
		}
		if itm == nil {
			continue // all attributes, not writable to diameter packet
		}
		urlVals.Set(strings.Join(itm.Path, utils.NestingSep), utils.IfaceAsString(itm.Data))
	}
	if aTime, err := cgrEv.FieldAsTime(utils.AnswerTime, httpPost.cgrCfg.GeneralCfg().DefaultTimezone); err == nil {
		if httpPost.dc[utils.FirstEventATime].(time.Time).IsZero() || httpPost.dc[utils.FirstEventATime].(time.Time).Before(aTime) {
			httpPost.dc[utils.FirstEventATime] = aTime
		}
		if aTime.After(httpPost.dc[utils.LastEventATime].(time.Time)) {
			httpPost.dc[utils.LastEventATime] = aTime
		}
	}
	if oID, err := cgrEv.FieldAsInt64(utils.OrderID); err == nil {
		if httpPost.dc[utils.FirstExpOrderID].(int64) > oID || httpPost.dc[utils.FirstExpOrderID].(int64) == 0 {
			httpPost.dc[utils.FirstExpOrderID] = oID
		}
		if httpPost.dc[utils.LastExpOrderID].(int64) < oID {
			httpPost.dc[utils.LastExpOrderID] = oID
		}
	}
	if cost, err := cgrEv.FieldAsFloat64(utils.Cost); err == nil {
		httpPost.dc[utils.TotalCost] = httpPost.dc[utils.TotalCost].(float64) + cost
	}
	if tor, err := cgrEv.FieldAsString(utils.ToR); err == nil {
		if usage, err := cgrEv.FieldAsDuration(utils.Usage); err == nil {
			switch tor {
			case utils.VOICE:
				httpPost.dc[utils.TotalDuration] = httpPost.dc[utils.TotalDuration].(time.Duration) + usage
			case utils.SMS:
				httpPost.dc[utils.TotalSMSUsage] = httpPost.dc[utils.TotalSMSUsage].(time.Duration) + usage
			case utils.MMS:
				httpPost.dc[utils.TotalMMSUsage] = httpPost.dc[utils.TotalMMSUsage].(time.Duration) + usage
			case utils.GENERIC:
				httpPost.dc[utils.TotalGenericUsage] = httpPost.dc[utils.TotalGenericUsage].(time.Duration) + usage
			case utils.DATA:
				httpPost.dc[utils.TotalDataUsage] = httpPost.dc[utils.TotalDataUsage].(time.Duration) + usage
			}
		}
	}
	httpPost.dc[utils.PositiveExports].(utils.StringSet).Add(cgrEv.ID)
	body = urlVals
	if err = httpPost.httpPoster.Post(body, utils.EmptyString); err != nil &&
		httpPost.cgrCfg.GeneralCfg().FailedPostsDir != utils.META_NONE {
		engine.AddFailedPost(httpPost.cgrCfg.EEsCfg().Exporters[httpPost.cfgIdx].ExportPath,
			httpPost.cgrCfg.EEsCfg().Exporters[httpPost.cfgIdx].Type, utils.EventExporterS, body)
	}
	return
}
