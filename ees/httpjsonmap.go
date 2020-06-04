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
	"encoding/json"
	"sync"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func NewHTTPJsonMapEe(cgrCfg *config.CGRConfig, cfgIdx int, filterS *engine.FilterS,
	dc utils.MapStorage) (httpJson *HTTPJsonMapEe, err error) {
	dc[utils.ExportID] = cgrCfg.EEsCfg().Exporters[cfgIdx].ID
	httpJson = &HTTPJsonMapEe{id: cgrCfg.EEsCfg().Exporters[cfgIdx].ID,
		cgrCfg: cgrCfg, cfgIdx: cfgIdx, filterS: filterS, dc: dc}
	if cgrCfg.EEsCfg().Exporters[cfgIdx].Type == utils.MetaHTTPjsonMap {
		httpJson.httpPoster, err = engine.NewHTTPPoster(cgrCfg.GeneralCfg().HttpSkipTlsVerify,
			cgrCfg.GeneralCfg().ReplyTimeout, cgrCfg.EEsCfg().Exporters[cfgIdx].ExportPath,
			utils.PosterTransportContentTypes[cgrCfg.EEsCfg().Exporters[cfgIdx].Type], cgrCfg.EEsCfg().Exporters[cfgIdx].Attempts)
	}

	return
}

// FileCSVee implements EventExporter interface for .csv files
type HTTPJsonMapEe struct {
	id         string
	cgrCfg     *config.CGRConfig
	cfgIdx     int // index of config instance within ERsCfg.Readers
	filterS    *engine.FilterS
	httpPoster *engine.HTTPPoster
	sync.RWMutex
	dc utils.MapStorage
}

// ID returns the identificator of this exporter
func (httpJson *HTTPJsonMapEe) ID() string {
	return httpJson.id
}

// OnEvicted implements EventExporter, doing the cleanup before exit
func (httpJson *HTTPJsonMapEe) OnEvicted(_ string, _ interface{}) {
	return
}

// ExportEvent implements EventExporter
func (httpJson *HTTPJsonMapEe) ExportEvent(cgrEv *utils.CGREvent) (err error) {
	httpJson.Lock()
	defer httpJson.Unlock()

	httpJson.dc[utils.NumberOfEvents] = httpJson.dc[utils.NumberOfEvents].(int) + 1

	var body interface{}
	valMp := make(map[string]string)
	req := utils.MapStorage{}
	for k, v := range cgrEv.Event {
		req[k] = v
	}
	eeReq := NewEventExporterRequest(req, httpJson.dc, cgrEv.Tenant, httpJson.cgrCfg.GeneralCfg().DefaultTimezone,
		httpJson.filterS)

	if err = eeReq.SetFields(httpJson.cgrCfg.EEsCfg().Exporters[httpJson.cfgIdx].ContentFields()); err != nil {
		httpJson.dc[utils.NegativeExports].(utils.StringSet).Add(cgrEv.ID)
		return
	}
	for el := eeReq.cnt.GetFirstElement(); el != nil; el = el.Next() {
		var strVal string
		if strVal, err = eeReq.cnt.FieldAsString(el.Value.Slice()); err != nil {
			return
		}
		valMp[el.Value.Slice()[1]] = strVal
	}
	if aTime, err := cgrEv.FieldAsTime(utils.AnswerTime, httpJson.cgrCfg.GeneralCfg().DefaultTimezone); err == nil {
		if httpJson.dc[utils.FirstEventATime].(time.Time).IsZero() || httpJson.dc[utils.FirstEventATime].(time.Time).Before(aTime) {
			httpJson.dc[utils.FirstEventATime] = aTime
		}
		if aTime.After(httpJson.dc[utils.LastEventATime].(time.Time)) {
			httpJson.dc[utils.LastEventATime] = aTime
		}
	}
	if oID, err := cgrEv.FieldAsInt64(utils.OrderID); err == nil {
		if httpJson.dc[utils.FirstExpOrderID].(int64) > oID || httpJson.dc[utils.FirstExpOrderID].(int64) == 0 {
			httpJson.dc[utils.FirstExpOrderID] = oID
		}
		if httpJson.dc[utils.LastExpOrderID].(int64) < oID {
			httpJson.dc[utils.LastExpOrderID] = oID
		}
	}
	if cost, err := cgrEv.FieldAsFloat64(utils.Cost); err == nil {
		httpJson.dc[utils.TotalCost] = httpJson.dc[utils.TotalCost].(float64) + cost
	}
	if tor, err := cgrEv.FieldAsString(utils.ToR); err == nil {
		if usage, err := cgrEv.FieldAsDuration(utils.Usage); err == nil {
			switch tor {
			case utils.VOICE:
				httpJson.dc[utils.TotalDuration] = httpJson.dc[utils.TotalDuration].(time.Duration) + usage
			case utils.SMS:
				httpJson.dc[utils.TotalSMSUsage] = httpJson.dc[utils.TotalSMSUsage].(time.Duration) + usage
			case utils.MMS:
				httpJson.dc[utils.TotalMMSUsage] = httpJson.dc[utils.TotalMMSUsage].(time.Duration) + usage
			case utils.GENERIC:
				httpJson.dc[utils.TotalGenericUsage] = httpJson.dc[utils.TotalGenericUsage].(time.Duration) + usage
			case utils.DATA:
				httpJson.dc[utils.TotalDataUsage] = httpJson.dc[utils.TotalDataUsage].(time.Duration) + usage
			}
		}
	}
	cgrID := utils.GenUUID()
	cgrID, err = cgrEv.FieldAsString(utils.CGRID)
	var runID string
	runID, err = cgrEv.FieldAsString(utils.RunID)
	httpJson.dc[utils.PositiveExports].(utils.StringSet).Add(cgrEv.ID)
	if body, err = json.Marshal(valMp); err != nil {
		return
	}
	return httpJson.post(body, utils.ConcatenatedKey(cgrID, runID))
}

func (httpJson *HTTPJsonMapEe) post(body interface{}, key string) (err error) {
	switch httpJson.cgrCfg.EEsCfg().Exporters[httpJson.cfgIdx].Type {
	case utils.MetaHTTPjsonMap:
		err = httpJson.httpPoster.Post(body, utils.EmptyString)
	case utils.MetaAMQPjsonMap:
		err = engine.PostersCache.PostAMQP(httpJson.cgrCfg.EEsCfg().Exporters[httpJson.cfgIdx].ExportPath,
			httpJson.cgrCfg.EEsCfg().Exporters[httpJson.cfgIdx].Attempts, body.([]byte))
	case utils.MetaAMQPV1jsonMap:
		err = engine.PostersCache.PostAMQPv1(httpJson.cgrCfg.EEsCfg().Exporters[httpJson.cfgIdx].ExportPath,
			httpJson.cgrCfg.EEsCfg().Exporters[httpJson.cfgIdx].Attempts, body.([]byte))
	case utils.MetaSQSjsonMap:
		err = engine.PostersCache.PostSQS(httpJson.cgrCfg.EEsCfg().Exporters[httpJson.cfgIdx].ExportPath,
			httpJson.cgrCfg.EEsCfg().Exporters[httpJson.cfgIdx].Attempts, body.([]byte))
	case utils.MetaKafkajsonMap:
		err = engine.PostersCache.PostKafka(httpJson.cgrCfg.EEsCfg().Exporters[httpJson.cfgIdx].ExportPath,
			httpJson.cgrCfg.EEsCfg().Exporters[httpJson.cfgIdx].Attempts, body.([]byte), key)
	case utils.MetaS3jsonMap:
		err = engine.PostersCache.PostS3(httpJson.cgrCfg.EEsCfg().Exporters[httpJson.cfgIdx].ExportPath,
			httpJson.cgrCfg.EEsCfg().Exporters[httpJson.cfgIdx].Attempts, body.([]byte), key)
	}
	if err != nil && httpJson.cgrCfg.GeneralCfg().FailedPostsDir != utils.META_NONE {
		engine.AddFailedPost(httpJson.cgrCfg.EEsCfg().Exporters[httpJson.cfgIdx].ExportPath,
			httpJson.cgrCfg.EEsCfg().Exporters[httpJson.cfgIdx].Type, utils.EventExporterS, body)
	}
	return
}
