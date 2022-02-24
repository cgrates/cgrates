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
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

type EventExporter interface {
	Cfg() *config.EventExporterCfg         // return the config
	Connect() error                        // called before exporting an event to make sure it is connected
	ExportEvent(interface{}, string) error // called on each event to be exported
	Close() error                          // called when the exporter needs to terminate
	GetMetrics() *utils.SafeMapStorage     // called to get metrics
	PrepareMap(*utils.CGREvent) (interface{}, error)
	PrepareOrderMap(*utils.OrderedNavigableMap) (interface{}, error)
}

// NewEventExporter produces exporters
func NewEventExporter(cfg *config.EventExporterCfg, cgrCfg *config.CGRConfig, filterS *engine.FilterS,
	connMngr *engine.ConnManager) (ee EventExporter, err error) {
	var dc *utils.SafeMapStorage
	if dc, err = newEEMetrics(utils.FirstNonEmpty(
		cfg.Timezone,
		cgrCfg.GeneralCfg().DefaultTimezone)); err != nil {
		return
	}
	switch cfg.Type {
	case utils.MetaFileCSV:
		return NewFileCSVee(cfg, cgrCfg, filterS, dc)
	case utils.MetaFileFWV:
		return NewFileFWVee(cfg, cgrCfg, filterS, dc)
	case utils.MetaHTTPPost:
		return NewHTTPPostEE(cfg, cgrCfg, filterS, dc)
	case utils.MetaHTTPjsonMap:
		return NewHTTPjsonMapEE(cfg, cgrCfg, filterS, dc)
	case utils.MetaNatsjsonMap:
		return NewNatsEE(cfg, cgrCfg.GeneralCfg().NodeID,
			cgrCfg.GeneralCfg().ConnectTimeout, dc)
	case utils.MetaAMQPjsonMap:
		return NewAMQPee(cfg, dc), nil
	case utils.MetaAMQPV1jsonMap:
		return NewAMQPv1EE(cfg, dc), nil
	case utils.MetaS3jsonMap:
		return NewS3EE(cfg, dc), nil
	case utils.MetaSQSjsonMap:
		return NewSQSee(cfg, dc), nil
	case utils.MetaKafkajsonMap:
		return NewKafkaEE(cfg, dc), nil
	case utils.MetaVirt:
		return NewVirtualEE(cfg, dc), nil
	case utils.MetaElastic:
		return NewElasticEE(cfg, dc)
	case utils.MetaSQL:
		return NewSQLEe(cfg, dc)
	case utils.MetaLog:
		return NewLogEE(cfg, dc), nil
	case utils.MetaRPC:
		return NewRpcEE(cfg, dc, connMngr)
	default:
		return nil, fmt.Errorf("unsupported exporter type: <%s>", cfg.Type)
	}
}

func newConcReq(limit int) (c *concReq) {
	c = &concReq{limit: limit}
	if limit > 0 {
		c.reqs = make(chan struct{}, limit)
		for i := 0; i < limit; i++ {
			c.reqs <- struct{}{}
		}
	}
	return
}

type concReq struct {
	reqs  chan struct{}
	limit int
}

func (c *concReq) get() {
	if c.limit > 0 {
		<-c.reqs
	}
}
func (c *concReq) done() {
	if c.limit > 0 {
		c.reqs <- struct{}{}
	}
}

// composeHeaderTrailer will return the orderNM for *hdr or *trl
func composeHeaderTrailer(prfx string, fields []*config.FCTemplate, dc utils.DataStorage, cfg *config.CGRConfig, fltS *engine.FilterS) (r *utils.OrderedNavigableMap, err error) {
	r = utils.NewOrderedNavigableMap()
	err = engine.NewExportRequest(map[string]utils.DataStorage{
		utils.MetaDC:  dc,
		utils.MetaCfg: cfg.GetDataProvider(),
	}, cfg.GeneralCfg().DefaultTenant, fltS,
		map[string]*utils.OrderedNavigableMap{prfx: r}).SetFields(fields)
	return
}

func composeExp(fields []*config.FCTemplate, cgrEv *utils.CGREvent, dc utils.DataStorage, cfg *config.CGRConfig, fltS *engine.FilterS) (r *utils.OrderedNavigableMap, err error) {
	r = utils.NewOrderedNavigableMap()
	err = engine.NewExportRequest(map[string]utils.DataStorage{
		utils.MetaReq:  utils.MapStorage(cgrEv.Event),
		utils.MetaDC:   dc,
		utils.MetaOpts: utils.MapStorage(cgrEv.APIOpts),
		utils.MetaCfg:  cfg.GetDataProvider(),
	}, utils.FirstNonEmpty(cgrEv.Tenant, cfg.GeneralCfg().DefaultTenant),
		fltS,
		map[string]*utils.OrderedNavigableMap{utils.MetaExp: r}).SetFields(fields)
	return
}

func newEEMetrics(location string) (*utils.SafeMapStorage, error) {
	tNow := time.Now()
	loc, err := time.LoadLocation(location)
	if err != nil {
		return nil, err
	}
	return &utils.SafeMapStorage{MapStorage: utils.MapStorage{
		utils.NumberOfEvents:  int64(0),
		utils.PositiveExports: utils.StringSet{},
		utils.NegativeExports: utils.StringSet{},
		utils.TimeNow: time.Date(tNow.Year(), tNow.Month(), tNow.Day(),
			tNow.Hour(), tNow.Minute(), tNow.Second(), tNow.Nanosecond(), loc),
	}}, nil
}

func updateEEMetrics(dc *utils.SafeMapStorage, cgrID string, ev engine.MapEvent, hasError bool, timezone string) {
	dc.Lock()
	defer dc.Unlock()
	if hasError {
		dc.MapStorage[utils.NegativeExports].(utils.StringSet).Add(cgrID)
	} else {
		dc.MapStorage[utils.PositiveExports].(utils.StringSet).Add(cgrID)
	}
	if aTime, err := ev.GetTime(utils.AnswerTime, timezone); err == nil {
		if _, has := dc.MapStorage[utils.FirstEventATime]; !has {
			dc.MapStorage[utils.FirstEventATime] = time.Time{}
		}
		if _, has := dc.MapStorage[utils.LastEventATime]; !has {
			dc.MapStorage[utils.LastEventATime] = time.Time{}
		}
		if dc.MapStorage[utils.FirstEventATime].(time.Time).IsZero() ||
			aTime.Before(dc.MapStorage[utils.FirstEventATime].(time.Time)) {
			dc.MapStorage[utils.FirstEventATime] = aTime
		}
		if aTime.After(dc.MapStorage[utils.LastEventATime].(time.Time)) {
			dc.MapStorage[utils.LastEventATime] = aTime
		}
	}
	if oID, err := ev.GetTInt64(utils.OrderID); err == nil {
		if _, has := dc.MapStorage[utils.FirstExpOrderID]; !has {
			dc.MapStorage[utils.FirstExpOrderID] = int64(0)
		}
		if _, has := dc.MapStorage[utils.LastExpOrderID]; !has {
			dc.MapStorage[utils.LastExpOrderID] = int64(0)
		}
		if dc.MapStorage[utils.FirstExpOrderID].(int64) == 0 ||
			dc.MapStorage[utils.FirstExpOrderID].(int64) > oID {
			dc.MapStorage[utils.FirstExpOrderID] = oID
		}
		if dc.MapStorage[utils.LastExpOrderID].(int64) < oID {
			dc.MapStorage[utils.LastExpOrderID] = oID
		}
	}
	if cost, err := ev.GetFloat64(utils.Cost); err == nil {
		if _, has := dc.MapStorage[utils.TotalCost]; !has {
			dc.MapStorage[utils.TotalCost] = float64(0.0)
		}
		dc.MapStorage[utils.TotalCost] = dc.MapStorage[utils.TotalCost].(float64) + cost
	}
	if tor, err := ev.GetString(utils.ToR); err == nil {
		if usage, err := ev.GetDuration(utils.Usage); err == nil {
			switch tor {
			case utils.MetaVoice:
				if _, has := dc.MapStorage[utils.TotalDuration]; !has {
					dc.MapStorage[utils.TotalDuration] = time.Duration(0)
				}
				dc.MapStorage[utils.TotalDuration] = dc.MapStorage[utils.TotalDuration].(time.Duration) + usage
			case utils.MetaSMS:
				if _, has := dc.MapStorage[utils.TotalSMSUsage]; !has {
					dc.MapStorage[utils.TotalSMSUsage] = time.Duration(0)
				}
				dc.MapStorage[utils.TotalSMSUsage] = dc.MapStorage[utils.TotalSMSUsage].(time.Duration) + usage
			case utils.MetaMMS:
				if _, has := dc.MapStorage[utils.TotalMMSUsage]; !has {
					dc.MapStorage[utils.TotalMMSUsage] = time.Duration(0)
				}
				dc.MapStorage[utils.TotalMMSUsage] = dc.MapStorage[utils.TotalMMSUsage].(time.Duration) + usage
			case utils.MetaGeneric:
				if _, has := dc.MapStorage[utils.TotalGenericUsage]; !has {
					dc.MapStorage[utils.TotalGenericUsage] = time.Duration(0)
				}
				dc.MapStorage[utils.TotalGenericUsage] = dc.MapStorage[utils.TotalGenericUsage].(time.Duration) + usage
			case utils.MetaData:
				if _, has := dc.MapStorage[utils.TotalDataUsage]; !has {
					dc.MapStorage[utils.TotalDataUsage] = time.Duration(0)
				}
				dc.MapStorage[utils.TotalDataUsage] = dc.MapStorage[utils.TotalDataUsage].(time.Duration) + usage
			}
		}
	}
}

type bytePreparing struct{}

func (bytePreparing) PrepareMap(mp *utils.CGREvent) (interface{}, error) {
	return json.Marshal(mp.Event)
}
func (bytePreparing) PrepareOrderMap(mp *utils.OrderedNavigableMap) (interface{}, error) {
	valMp := make(map[string]interface{})
	for el := mp.GetFirstElement(); el != nil; el = el.Next() {
		path := el.Value
		nmIt, _ := mp.Field(path)
		path = path[:len(path)-1] // remove the last index
		valMp[strings.Join(path, utils.NestingSep)] = nmIt.String()
	}
	return json.Marshal(valMp)
}

type slicePreparing struct{}

func (slicePreparing) PrepareMap(mp *utils.CGREvent) (interface{}, error) {
	csvRecord := make([]string, 0, len(mp.Event))
	for _, val := range mp.Event {
		csvRecord = append(csvRecord, utils.IfaceAsString(val))
	}
	return csvRecord, nil
}
func (slicePreparing) PrepareOrderMap(mp *utils.OrderedNavigableMap) (interface{}, error) {
	return mp.OrderedFieldsAsStrings(), nil
}
