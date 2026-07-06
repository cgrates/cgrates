/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/

package engine

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/cgrates/cgrates/utils"
)

func csvLoad(s any, values []string) (any, error) {
	fieldValueMap := make(map[string]string)
	st := reflect.TypeOf(s)
	numFields := st.NumField()
	for i := 0; i < numFields; i++ {
		field := st.Field(i)
		re := field.Tag.Get("re")
		index := field.Tag.Get("index")
		if index != utils.EmptyString {
			idx, err := strconv.Atoi(index)
			if err != nil || len(values) <= idx {
				return nil, fmt.Errorf("invalid %v.%v index %v", st.Name(), field.Name, index)
			}
			if re != utils.EmptyString {
				if matched, err := regexp.MatchString(re, values[idx]); !matched || err != nil {
					return nil, fmt.Errorf("invalid %v.%v value %v", st.Name(), field.Name, values[idx])
				}
			}
			fieldValueMap[field.Name] = values[idx]
		}
	}
	elem := reflect.New(st).Elem()
	for fieldName, fieldValue := range fieldValueMap {
		field := elem.FieldByName(fieldName)
		if field.IsValid() {
			switch field.Kind() {
			case reflect.Float64:
				if fieldValue == utils.EmptyString {
					fieldValue = "0"
				}
				value, err := strconv.ParseFloat(fieldValue, 64)
				if err != nil {
					return nil, fmt.Errorf(`invalid value "%s" for field %s.%s`, fieldValue, st.Name(), fieldName)
				}
				field.SetFloat(value)
			case reflect.Int:
				if fieldValue == utils.EmptyString {
					fieldValue = "0"
				}
				value, err := strconv.Atoi(fieldValue)
				if err != nil {
					return nil, fmt.Errorf(`invalid value "%s" for field %s.%s`, fieldValue, st.Name(), fieldName)
				}
				field.SetInt(int64(value))
			case reflect.Bool:
				if fieldValue == utils.EmptyString {
					fieldValue = "false"
				}
				value, err := strconv.ParseBool(fieldValue)
				if err != nil {
					return nil, fmt.Errorf(`invalid value "%s" for field %s.%s`, fieldValue, st.Name(), fieldName)
				}
				field.SetBool(value)
			case reflect.String:
				field.SetString(fieldValue)
			}
		}
	}
	return elem.Interface(), nil
}

func getColumnCount(s any) int {
	st := reflect.TypeOf(s)
	numFields := st.NumField()
	count := 0
	for i := 0; i < numFields; i++ {
		field := st.Field(i)
		index := field.Tag.Get("index")
		if index != utils.EmptyString {
			count++
		}
	}
	return count
}

type ResourceMdls []*ResourceMdl

func (tps ResourceMdls) AsTPResources() (result []*utils.TPResourceProfile) {
	mrl := make(map[string]*utils.TPResourceProfile)
	filterMap := make(map[string]utils.StringSet)
	thresholdMap := make(map[string]utils.StringSet)
	for _, tp := range tps {
		tenID := (&utils.TenantID{Tenant: tp.Tenant, ID: tp.ID}).TenantID()
		rl, found := mrl[tenID]
		if !found {
			rl = &utils.TPResourceProfile{
				TPid:    tp.Tpid,
				Tenant:  tp.Tenant,
				ID:      tp.ID,
				Blocker: tp.Blocker,
				Stored:  tp.Stored,
			}
		}
		if tp.UsageTTL != utils.EmptyString {
			rl.UsageTTL = tp.UsageTTL
		}
		if tp.Weights != "" {
			rl.Weights = tp.Weights
		}
		if tp.Limit != utils.EmptyString {
			rl.Limit = tp.Limit
		}
		if tp.AllocationMessage != utils.EmptyString {
			rl.AllocationMessage = tp.AllocationMessage
		}
		rl.Blocker = tp.Blocker
		rl.Stored = tp.Stored
		if tp.ThresholdIDs != utils.EmptyString {
			if _, has := thresholdMap[tenID]; !has {
				thresholdMap[tenID] = make(utils.StringSet)
			}
			thresholdMap[tenID].AddSlice(strings.Split(tp.ThresholdIDs, utils.InfieldSep))
		}
		if tp.FilterIDs != utils.EmptyString {
			if _, has := filterMap[tenID]; !has {
				filterMap[tenID] = make(utils.StringSet)
			}
			filterMap[tenID].AddSlice(strings.Split(tp.FilterIDs, utils.InfieldSep))
		}
		mrl[tenID] = rl
	}
	result = make([]*utils.TPResourceProfile, len(mrl))
	i := 0
	for tntID, rl := range mrl {
		result[i] = rl
		result[i].FilterIDs = filterMap[tntID].AsSlice()
		result[i].ThresholdIDs = thresholdMap[tntID].AsSlice()
		i++
	}
	return
}

func APItoResource(tpRL *utils.TPResourceProfile, timezone string) (rp *utils.ResourceProfile, err error) {
	rp = &utils.ResourceProfile{
		Tenant:            tpRL.Tenant,
		ID:                tpRL.ID,
		Blocker:           tpRL.Blocker,
		Stored:            tpRL.Stored,
		AllocationMessage: tpRL.AllocationMessage,
		ThresholdIDs:      make([]string, len(tpRL.ThresholdIDs)),
		FilterIDs:         make([]string, len(tpRL.FilterIDs)),
	}
	if tpRL.Weights != utils.EmptyString {
		rp.Weights, err = utils.NewDynamicWeightsFromString(tpRL.Weights, utils.InfieldSep, utils.ANDSep)
		if err != nil {
			return
		}
	}
	if tpRL.UsageTTL != utils.EmptyString {
		if rp.UsageTTL, err = utils.ParseDurationWithNanosecs(tpRL.UsageTTL); err != nil {
			return nil, err
		}
	}
	copy(rp.FilterIDs, tpRL.FilterIDs)
	copy(rp.ThresholdIDs, tpRL.ThresholdIDs)
	if tpRL.Limit != utils.EmptyString {
		if rp.Limit, err = strconv.ParseFloat(tpRL.Limit, 64); err != nil {
			return nil, err
		}
	}
	return rp, nil
}

type IPMdls []*IPMdl

func (tps IPMdls) AsTPIPs() []*utils.TPIPProfile {
	filterMap := make(map[string]utils.StringSet)
	mst := make(map[string]*utils.TPIPProfile)
	poolMap := make(map[string]map[string]*utils.TPIPPool)
	for _, mdl := range tps {
		tenID := (&utils.TenantID{Tenant: mdl.Tenant, ID: mdl.ID}).TenantID()
		tpip, found := mst[tenID]
		if !found {
			tpip = &utils.TPIPProfile{
				TPid:   mdl.Tpid,
				Tenant: mdl.Tenant,
				ID:     mdl.ID,
				Stored: mdl.Stored,
			}
		}
		// Handle Pool
		if mdl.PoolID != utils.EmptyString {
			if _, has := poolMap[tenID]; !has {
				poolMap[tenID] = make(map[string]*utils.TPIPPool)
			}
			poolID := mdl.PoolID
			if mdl.PoolFilterIDs != utils.EmptyString {
				poolID = utils.ConcatenatedKey(poolID,
					utils.NewStringSet(strings.Split(mdl.PoolFilterIDs, utils.InfieldSep)).Sha1())
			}
			pool, found := poolMap[tenID][poolID]
			if !found {
				pool = &utils.TPIPPool{
					ID:       mdl.PoolID,
					Type:     mdl.PoolType,
					Range:    mdl.PoolRange,
					Strategy: mdl.PoolStrategy,
					Message:  mdl.PoolMessage,
					Weights:  mdl.PoolWeights,
					Blockers: mdl.PoolBlockers,
				}
			}
			if mdl.PoolFilterIDs != utils.EmptyString {
				poolFilterSplit := strings.Split(mdl.PoolFilterIDs, utils.InfieldSep)
				pool.FilterIDs = append(pool.FilterIDs, poolFilterSplit...)
			}
			poolMap[tenID][poolID] = pool
		}
		// Profile-level fields
		if mdl.TTL != utils.EmptyString {
			tpip.TTL = mdl.TTL
		}
		if mdl.Weights != "" {
			tpip.Weights = mdl.Weights
		}
		if mdl.Stored {
			tpip.Stored = mdl.Stored
		}

		if mdl.FilterIDs != utils.EmptyString {
			if _, has := filterMap[tenID]; !has {
				filterMap[tenID] = make(utils.StringSet)
			}
			filterMap[tenID].AddSlice(strings.Split(mdl.FilterIDs, utils.InfieldSep))
		}
		mst[tenID] = tpip
	}
	// Build result with Pools
	result := make([]*utils.TPIPProfile, len(mst))
	i := 0
	for tntID, tpip := range mst {
		result[i] = tpip
		for _, poolData := range poolMap[tntID] {
			result[i].Pools = append(result[i].Pools, poolData)
		}
		result[i].FilterIDs = filterMap[tntID].AsSlice()
		i++
	}
	return result
}

func APItoIP(tp *utils.TPIPProfile) (*utils.IPProfile, error) {
	ipp := &utils.IPProfile{
		Tenant:    tp.Tenant,
		ID:        tp.ID,
		Stored:    tp.Stored,
		FilterIDs: make([]string, len(tp.FilterIDs)),
		Pools:     make([]*utils.IPPool, len(tp.Pools)),
	}
	if tp.Weights != utils.EmptyString {
		var err error
		ipp.Weights, err = utils.NewDynamicWeightsFromString(tp.Weights, utils.InfieldSep, utils.ANDSep)
		if err != nil {
			return nil, err
		}
	}
	if tp.TTL != utils.EmptyString {
		var err error
		if ipp.TTL, err = utils.ParseDurationWithNanosecs(tp.TTL); err != nil {
			return nil, err
		}
	}

	copy(ipp.FilterIDs, tp.FilterIDs)

	for i, pool := range tp.Pools {
		ipp.Pools[i] = &utils.IPPool{
			ID:        pool.ID,
			FilterIDs: pool.FilterIDs,
			Type:      pool.Type,
			Range:     pool.Range,
			Strategy:  pool.Strategy,
			Message:   pool.Message,
		}
		if pool.Weights != utils.EmptyString {
			var err error
			ipp.Pools[i].Weights, err = utils.NewDynamicWeightsFromString(pool.Weights, utils.InfieldSep, utils.ANDSep)
			if err != nil {
				return nil, err
			}
		}
		if pool.Blockers != utils.EmptyString {
			var err error
			ipp.Pools[i].Blockers, err = utils.NewDynamicBlockersFromString(pool.Blockers, utils.InfieldSep, utils.ANDSep)
			if err != nil {
				return nil, err
			}
		}
	}
	return ipp, nil
}

type StatMdls []*StatMdl

func (tps StatMdls) AsTPStats() (result []*utils.TPStatProfile) {
	filterMap := make(map[string]utils.StringSet)
	thresholdMap := make(map[string]utils.StringSet)
	statMetricsMap := make(map[string]map[string]*utils.TPMetricWithFilters)
	mst := make(map[string]*utils.TPStatProfile)
	for _, model := range tps {
		key := &utils.TenantID{Tenant: model.Tenant, ID: model.ID}
		st, found := mst[key.TenantID()]
		if !found {
			st = &utils.TPStatProfile{
				Tenant:      model.Tenant,
				TPid:        model.Tpid,
				ID:          model.ID,
				Stored:      model.Stored,
				Weights:     model.Weights,
				MinItems:    model.MinItems,
				TTL:         model.TTL,
				QueueLength: model.QueueLength,
			}
		}
		if model.Blockers != utils.EmptyString {
			st.Blockers = model.Blockers
		}
		if model.Stored {
			st.Stored = model.Stored
		}
		if model.Weights != utils.EmptyString {
			st.Weights = model.Weights
		}
		if model.MinItems != 0 {
			st.MinItems = model.MinItems
		}
		if model.TTL != utils.EmptyString {
			st.TTL = model.TTL
		}
		if model.QueueLength != 0 {
			st.QueueLength = model.QueueLength
		}
		if model.ThresholdIDs != utils.EmptyString {
			if _, has := thresholdMap[key.TenantID()]; !has {
				thresholdMap[key.TenantID()] = make(utils.StringSet)
			}
			thresholdMap[key.TenantID()].AddSlice(strings.Split(model.ThresholdIDs, utils.InfieldSep))
		}
		if model.FilterIDs != utils.EmptyString {
			if _, has := filterMap[key.TenantID()]; !has {
				filterMap[key.TenantID()] = make(utils.StringSet)
			}
			filterMap[key.TenantID()].AddSlice(strings.Split(model.FilterIDs, utils.InfieldSep))
		}
		if model.MetricIDs != utils.EmptyString {
			if _, has := statMetricsMap[key.TenantID()]; !has {
				statMetricsMap[key.TenantID()] = make(map[string]*utils.TPMetricWithFilters)
			}
			metricIDsSplit := strings.Split(model.MetricIDs, utils.InfieldSep)
			for _, metricID := range metricIDsSplit {
				stsMetric, found := statMetricsMap[key.TenantID()][metricID]
				if !found {
					stsMetric = &utils.TPMetricWithFilters{
						MetricID: metricID,
					}
				}
				if model.MetricFilterIDs != utils.EmptyString {
					filterIDs := strings.Split(model.MetricFilterIDs, utils.InfieldSep)
					stsMetric.FilterIDs = append(stsMetric.FilterIDs, filterIDs...)
				}
				if model.MetricBlockers != utils.EmptyString {
					stsMetric.Blockers = model.MetricBlockers
				}
				statMetricsMap[key.TenantID()][metricID] = stsMetric
			}
		}
		mst[key.TenantID()] = st
	}
	result = make([]*utils.TPStatProfile, len(mst))
	i := 0
	for tntID, st := range mst {
		result[i] = st
		result[i].FilterIDs = filterMap[tntID].AsSlice()
		result[i].ThresholdIDs = thresholdMap[tntID].AsSlice()
		for _, metric := range statMetricsMap[tntID] {
			result[i].Metrics = append(result[i].Metrics, metric)
		}
		i++
	}
	return
}

func APItoStats(tpST *utils.TPStatProfile, timezone string) (st *utils.StatQueueProfile, err error) {
	st = &utils.StatQueueProfile{
		Tenant:       tpST.Tenant,
		ID:           tpST.ID,
		FilterIDs:    make([]string, len(tpST.FilterIDs)),
		QueueLength:  tpST.QueueLength,
		MinItems:     tpST.MinItems,
		Metrics:      make([]*utils.MetricWithFilters, len(tpST.Metrics)),
		Stored:       tpST.Stored,
		ThresholdIDs: make([]string, len(tpST.ThresholdIDs)),
	}
	if tpST.Weights != utils.EmptyString {
		if st.Weights, err = utils.NewDynamicWeightsFromString(tpST.Weights, utils.InfieldSep, utils.ANDSep); err != nil {
			return
		}
	}
	if tpST.Blockers != utils.EmptyString {
		if st.Blockers, err = utils.NewDynamicBlockersFromString(tpST.Blockers, utils.InfieldSep, utils.ANDSep); err != nil {
			return
		}
	}
	if tpST.TTL != utils.EmptyString {
		if st.TTL, err = utils.ParseDurationWithNanosecs(tpST.TTL); err != nil {
			return nil, err
		}
	}
	for i, metric := range tpST.Metrics {
		st.Metrics[i] = &utils.MetricWithFilters{
			MetricID:  metric.MetricID,
			FilterIDs: metric.FilterIDs,
		}
		if metric.Blockers != utils.EmptyString {
			if st.Metrics[i].Blockers, err = utils.NewDynamicBlockersFromString(metric.Blockers, utils.InfieldSep, utils.ANDSep); err != nil {
				return
			}
		}
	}
	copy(st.ThresholdIDs, tpST.ThresholdIDs)
	copy(st.FilterIDs, tpST.FilterIDs)
	return st, nil
}

type RankingMdls []*RankingMdl

func (models RankingMdls) AsTPRanking() (result []*utils.TPRankingProfile) {
	thresholdMap := make(map[string]utils.StringSet)
	metricsMap := make(map[string]utils.StringSet)
	sortingParameterMap := make(map[string]utils.StringSet)
	sortingParameterSlice := make(map[string][]string)
	statsMap := make(map[string]utils.StringSet)
	mrg := make(map[string]*utils.TPRankingProfile)
	for _, model := range models {
		key := &utils.TenantID{Tenant: model.Tenant, ID: model.ID}
		rg, found := mrg[key.TenantID()]
		if !found {
			rg = &utils.TPRankingProfile{
				Tenant:   model.Tenant,
				TPid:     model.Tpid,
				ID:       model.ID,
				Schedule: model.Schedule,
				Sorting:  model.Sorting,
				Stored:   model.Stored,
			}
		}
		if model.Schedule != utils.EmptyString {
			rg.Schedule = model.Schedule
		}
		if model.Sorting != utils.EmptyString {
			rg.Sorting = model.Sorting
		}
		if model.Stored {
			rg.Stored = model.Stored
		}
		if model.StatIDs != utils.EmptyString {
			if _, has := statsMap[key.TenantID()]; !has {
				statsMap[key.TenantID()] = make(utils.StringSet)
			}
			statsMap[key.TenantID()].AddSlice(strings.Split(model.StatIDs, utils.InfieldSep))
		}
		if model.ThresholdIDs != utils.EmptyString {
			if _, has := thresholdMap[key.TenantID()]; !has {
				thresholdMap[key.TenantID()] = make(utils.StringSet)
			}
			thresholdMap[key.TenantID()].AddSlice(strings.Split(model.ThresholdIDs, utils.InfieldSep))
		}
		if model.SortingParameters != utils.EmptyString {
			if _, has := sortingParameterMap[key.TenantID()]; !has {
				sortingParameterMap[key.TenantID()] = make(utils.StringSet)
				sortingParameterSlice[key.TenantID()] = make([]string, 0)
			}
			spltSl := strings.Split(model.SortingParameters, utils.InfieldSep)
			for _, splt := range spltSl {
				if _, has := sortingParameterMap[key.TenantID()][splt]; !has {
					sortingParameterMap[key.TenantID()].Add(splt)
					sortingParameterSlice[key.TenantID()] = append(sortingParameterSlice[key.TenantID()], splt)
				}
			}
		}
		if model.MetricIDs != utils.EmptyString {
			if _, has := metricsMap[key.TenantID()]; !has {
				metricsMap[key.TenantID()] = make(utils.StringSet)
			}
			metricsMap[key.TenantID()].AddSlice(strings.Split(model.MetricIDs, utils.InfieldSep))
		}
		mrg[key.TenantID()] = rg
	}
	result = make([]*utils.TPRankingProfile, len(mrg))
	i := 0
	for tntID, rg := range mrg {
		result[i] = rg
		result[i].StatIDs = statsMap[tntID].AsSlice()
		result[i].MetricIDs = metricsMap[tntID].AsSlice()
		result[i].SortingParameters = sortingParameterSlice[tntID]
		result[i].ThresholdIDs = thresholdMap[tntID].AsOrderedSlice()
		i++
	}
	return
}

func APItoRanking(tpRG *utils.TPRankingProfile) (rg *utils.RankingProfile, err error) {
	rg = &utils.RankingProfile{
		Tenant:            tpRG.Tenant,
		ID:                tpRG.ID,
		Schedule:          tpRG.Schedule,
		Sorting:           tpRG.Sorting,
		Stored:            tpRG.Stored,
		StatIDs:           make([]string, len(tpRG.StatIDs)),
		MetricIDs:         make([]string, len(tpRG.MetricIDs)),
		SortingParameters: make([]string, len(tpRG.SortingParameters)),
		ThresholdIDs:      make([]string, len(tpRG.ThresholdIDs)),
	}
	copy(rg.StatIDs, tpRG.StatIDs)
	copy(rg.ThresholdIDs, tpRG.ThresholdIDs)
	copy(rg.SortingParameters, tpRG.SortingParameters)
	copy(rg.MetricIDs, tpRG.MetricIDs)
	return rg, nil
}

type TrendMdls []*TrendMdl

func (models TrendMdls) AsTPTrends() (result []*utils.TPTrendsProfile) {
	thresholdsMap := make(map[string]utils.StringSet)
	trendMetricsMap := make(map[string]utils.StringSet)
	mtr := make(map[string]*utils.TPTrendsProfile)
	for _, model := range models {
		key := &utils.TenantID{Tenant: model.Tenant, ID: model.ID}
		tr, found := mtr[key.TenantID()]
		if !found {
			tr = &utils.TPTrendsProfile{
				Tenant:          model.Tenant,
				TPid:            model.Tpid,
				ID:              model.ID,
				Schedule:        model.Schedule,
				StatID:          model.StatID,
				TTL:             model.TTL,
				QueueLength:     model.QueueLength,
				MinItems:        model.MinItems,
				Tolerance:       model.Tolerance,
				Stored:          model.Stored,
				CorrelationType: model.CorrelationType,
			}
		}
		if model.Schedule != utils.EmptyString {
			tr.Schedule = model.Schedule
		}
		if model.StatID != utils.EmptyString {
			tr.StatID = model.StatID
		}
		if model.TTL != utils.EmptyString {
			tr.TTL = model.TTL
		}
		if model.QueueLength != 0 {
			tr.QueueLength = model.QueueLength
		}
		if model.MinItems != 0 {
			tr.MinItems = model.MinItems
		}
		if model.CorrelationType != utils.EmptyString {
			tr.CorrelationType = model.CorrelationType
		}
		if model.Tolerance != 0 {
			tr.Tolerance = model.Tolerance
		}
		if model.Stored {
			tr.Stored = true
		}
		if model.ThresholdIDs != utils.EmptyString {
			if _, has := thresholdsMap[key.TenantID()]; !has {
				thresholdsMap[key.TenantID()] = make(utils.StringSet)
			}
			thresholdsMap[key.TenantID()].AddSlice(strings.Split(model.ThresholdIDs, utils.InfieldSep))
		}
		if model.Metrics != utils.EmptyString {
			if _, has := trendMetricsMap[key.TenantID()]; !has {
				trendMetricsMap[key.TenantID()] = make(utils.StringSet)
			}
			trendMetricsMap[key.TenantID()].AddSlice(strings.Split(model.Metrics, utils.InfieldSep))
		}
		mtr[key.TenantID()] = tr
	}
	result = make([]*utils.TPTrendsProfile, len(mtr))
	i := 0
	for tntId, sr := range mtr {
		result[i] = sr
		result[i].ThresholdIDs = thresholdsMap[tntId].AsSlice()
		result[i].Metrics = trendMetricsMap[tntId].AsSlice()
		i++
	}
	return
}

func APItoTrends(tpTR *utils.TPTrendsProfile) (tr *utils.TrendProfile, err error) {
	tr = &utils.TrendProfile{
		Tenant:          tpTR.Tenant,
		ID:              tpTR.ID,
		StatID:          tpTR.StatID,
		Schedule:        tpTR.Schedule,
		QueueLength:     tpTR.QueueLength,
		ThresholdIDs:    make([]string, len(tpTR.ThresholdIDs)),
		Metrics:         make([]string, len(tpTR.Metrics)),
		MinItems:        tpTR.MinItems,
		CorrelationType: tpTR.CorrelationType,
		Tolerance:       tpTR.Tolerance,
	}
	if tpTR.TTL != utils.EmptyString {
		if tr.TTL, err = utils.ParseDurationWithNanosecs(tpTR.TTL); err != nil {
			return
		}
	}
	copy(tr.ThresholdIDs, tpTR.ThresholdIDs)
	copy(tr.Metrics, tpTR.Metrics)

	return
}

type ThresholdMdls []*ThresholdMdl

func (tps ThresholdMdls) AsTPThreshold() (result []*utils.TPThresholdProfile) {
	mst := make(map[string]*utils.TPThresholdProfile)
	filterMap := make(map[string]utils.StringSet)
	actionMap := make(map[string]utils.StringSet)
	eesIDsMap := make(map[string]utils.StringSet)
	attributeMap := make(map[string][]string)
	for _, tp := range tps {
		tenID := (&utils.TenantID{Tenant: tp.Tenant, ID: tp.ID}).TenantID()
		th, found := mst[tenID]
		if !found {
			th = &utils.TPThresholdProfile{
				TPid:     tp.Tpid,
				Tenant:   tp.Tenant,
				ID:       tp.ID,
				Blocker:  tp.Blocker,
				MaxHits:  tp.MaxHits,
				MinHits:  tp.MinHits,
				MinSleep: tp.MinSleep,
				Async:    tp.Async,
			}
		}
		if tp.ActionProfileIDs != utils.EmptyString {
			if _, has := actionMap[tenID]; !has {
				actionMap[tenID] = make(utils.StringSet)
			}
			actionMap[tenID].AddSlice(strings.Split(tp.ActionProfileIDs, utils.InfieldSep))
		}
		if tp.Weights != "" {
			th.Weights = tp.Weights
		}
		if tp.FilterIDs != utils.EmptyString {
			if _, has := filterMap[tenID]; !has {
				filterMap[tenID] = make(utils.StringSet)
			}
			filterMap[tenID].AddSlice(strings.Split(tp.FilterIDs, utils.InfieldSep))
		}
		if tp.EeIDs != utils.EmptyString {
			if _, has := eesIDsMap[tenID]; !has {
				eesIDsMap[tenID] = make(utils.StringSet)
			}
			eesIDsMap[tenID].AddSlice(strings.Split(tp.EeIDs, utils.InfieldSep))
		}
		if tp.AttributeIDs != utils.EmptyString {
			attributeSplit := strings.Split(tp.AttributeIDs, utils.InfieldSep)
			var inlineAttribute string
			var dynam bool
			for _, attribute := range attributeSplit {
				if !dynam && !strings.HasPrefix(attribute, utils.Meta) {
					if inlineAttribute != utils.EmptyString {
						attributeMap[tenID] = append(attributeMap[tenID], inlineAttribute[1:])
						inlineAttribute = utils.EmptyString
					}
					attributeMap[tenID] = append(attributeMap[tenID], attribute)
					continue
				}
				if dynam {
					dynam = !strings.Contains(attribute, string(utils.RSRDynEndChar))
				} else {
					dynam = strings.Contains(attribute, string(utils.RSRDynStartChar))
				}
				inlineAttribute += utils.InfieldSep + attribute
			}
			if inlineAttribute != utils.EmptyString {
				attributeMap[tenID] = append(attributeMap[tenID], inlineAttribute[1:])
				inlineAttribute = utils.EmptyString
			}
		}

		mst[tenID] = th
	}
	result = make([]*utils.TPThresholdProfile, len(mst))
	i := 0
	for tntID, th := range mst {
		result[i] = th
		result[i].FilterIDs = filterMap[tntID].AsSlice()
		result[i].ActionProfileIDs = actionMap[tntID].AsSlice()
		result[i].EeIDs = eesIDsMap[tntID].AsSlice()
		result[i].AttributeIDs = attributeMap[tntID]
		i++
	}
	return
}

func APItoThresholdProfile(tpTH *utils.TPThresholdProfile, timezone string) (th *utils.ThresholdProfile, err error) {
	th = &utils.ThresholdProfile{
		Tenant:           tpTH.Tenant,
		ID:               tpTH.ID,
		MaxHits:          tpTH.MaxHits,
		MinHits:          tpTH.MinHits,
		Blocker:          tpTH.Blocker,
		Async:            tpTH.Async,
		ActionProfileIDs: make([]string, len(tpTH.ActionProfileIDs)),
		FilterIDs:        make([]string, len(tpTH.FilterIDs)),
		EeIDs:            make([]string, len(tpTH.EeIDs)),
		AttributeIDs:     make([]string, len(tpTH.AttributeIDs)),
	}
	if tpTH.Weights != utils.EmptyString {
		if th.Weights, err = utils.NewDynamicWeightsFromString(tpTH.Weights, utils.InfieldSep, utils.ANDSep); err != nil {
			return
		}
	}
	if tpTH.MinSleep != utils.EmptyString {
		if th.MinSleep, err = utils.ParseDurationWithNanosecs(tpTH.MinSleep); err != nil {
			return nil, err
		}
	}
	copy(th.EeIDs, tpTH.EeIDs)
	copy(th.ActionProfileIDs, tpTH.ActionProfileIDs)
	copy(th.FilterIDs, tpTH.FilterIDs)
	copy(th.AttributeIDs, tpTH.AttributeIDs)
	return th, nil
}

type FilterMdls []*FilterMdl

func (tps FilterMdls) AsTPFilter() (result []*utils.TPFilterProfile) {
	mst := make(map[string]*utils.TPFilterProfile)
	filterRules := make(map[string]*utils.TPFilter)
	for _, tp := range tps {
		tenID := (&utils.TenantID{Tenant: tp.Tenant, ID: tp.ID}).TenantID()
		th, found := mst[tenID]
		if !found {
			th = &utils.TPFilterProfile{
				TPid:   tp.Tpid,
				Tenant: tp.Tenant,
				ID:     tp.ID,
			}
		}
		if tp.Type != utils.EmptyString {
			var vals []string
			if tp.Values != utils.EmptyString {
				vals = splitDynFltrValues(tp.Values, utils.InfieldSep)
			}
			key := utils.ConcatenatedKey(tenID, tp.Type, tp.Element)
			if f, has := filterRules[key]; has {
				f.Values = append(f.Values, vals...)
			} else {
				f = &utils.TPFilter{
					Type:    tp.Type,
					Element: tp.Element,
					Values:  vals,
				}
				th.Filters = append(th.Filters, f)
				filterRules[key] = f
			}
		}
		mst[tenID] = th
	}
	result = make([]*utils.TPFilterProfile, len(mst))
	i := 0
	for _, th := range mst {
		result[i] = th
		i++
	}
	return
}

func APItoFilter(tpTH *utils.TPFilterProfile, timezone string) (th *Filter, err error) {
	th = &Filter{
		Tenant: tpTH.Tenant,
		ID:     tpTH.ID,
		Rules:  make([]*FilterRule, len(tpTH.Filters)),
	}
	for i, f := range tpTH.Filters {
		rf := &FilterRule{Type: f.Type, Element: f.Element, Values: f.Values}
		if err := rf.CompileValues(); err != nil {
			return nil, err
		}
		th.Rules[i] = rf
	}
	return th, nil
}

type RouteMdls []*RouteMdl

func (tps RouteMdls) AsTPRouteProfile() (result []*utils.TPRouteProfile) {
	filterMap := make(map[string]utils.StringSet)
	tpRouteProfileMap := make(map[string]*utils.TPRouteProfile)
	routeMap := make(map[string]map[string]*utils.TPRoute)
	sortingParameterMap := make(map[string]utils.StringSet)
	for _, tp := range tps {
		tenID := (&utils.TenantID{Tenant: tp.Tenant, ID: tp.ID}).TenantID()
		tpRouteProfile, found := tpRouteProfileMap[tenID]
		if !found {
			tpRouteProfile = &utils.TPRouteProfile{
				TPid:              tp.Tpid,
				Tenant:            tp.Tenant,
				ID:                tp.ID,
				SortingParameters: []string{},
			}
		}
		if tp.RouteID != utils.EmptyString {
			if _, has := routeMap[tenID]; !has {
				routeMap[tenID] = make(map[string]*utils.TPRoute)
			}
			routeID := tp.RouteID
			if tp.RouteFilterIDs != utils.EmptyString {
				routeID = utils.ConcatenatedKey(routeID,
					utils.NewStringSet(strings.Split(tp.RouteFilterIDs, utils.InfieldSep)).Sha1())
			}
			tpRoute, found := routeMap[tenID][routeID]
			if !found {
				tpRoute = &utils.TPRoute{
					ID:              tp.RouteID,
					Weights:         tp.RouteWeights,
					Blockers:        tp.RouteBlockers,
					RouteParameters: tp.RouteParameters,
				}
			}
			if tp.RouteFilterIDs != utils.EmptyString {
				routeFilterSplit := strings.Split(tp.RouteFilterIDs, utils.InfieldSep)
				tpRoute.FilterIDs = append(tpRoute.FilterIDs, routeFilterSplit...)
			}
			if tp.RouteRateProfileIDs != utils.EmptyString {
				ratingPlanSplit := strings.Split(tp.RouteRateProfileIDs, utils.InfieldSep)
				tpRoute.RateProfileIDs = append(tpRoute.RateProfileIDs, ratingPlanSplit...)
			}
			if tp.RouteResourceIDs != utils.EmptyString {
				resSplit := strings.Split(tp.RouteResourceIDs, utils.InfieldSep)
				tpRoute.ResourceIDs = append(tpRoute.ResourceIDs, resSplit...)
			}
			if tp.RouteStatIDs != utils.EmptyString {
				statSplit := strings.Split(tp.RouteStatIDs, utils.InfieldSep)
				tpRoute.StatIDs = append(tpRoute.StatIDs, statSplit...)
			}
			if tp.RouteAccountIDs != utils.EmptyString {
				accSplit := strings.Split(tp.RouteAccountIDs, utils.InfieldSep)
				tpRoute.AccountIDs = append(tpRoute.AccountIDs, accSplit...)
			}
			routeMap[tenID][routeID] = tpRoute
		}
		if tp.Sorting != utils.EmptyString {
			tpRouteProfile.Sorting = tp.Sorting
		}
		if tp.SortingParameters != utils.EmptyString {
			if _, has := sortingParameterMap[tenID]; !has {
				sortingParameterMap[tenID] = make(utils.StringSet)
			}
			sortingParameterMap[tenID].AddSlice(strings.Split(tp.SortingParameters, utils.InfieldSep))
		}
		if tp.Weights != utils.EmptyString {
			tpRouteProfile.Weights = tp.Weights
		}
		if tp.Blockers != utils.EmptyString {
			tpRouteProfile.Blockers = tp.Blockers
		}
		if tp.FilterIDs != utils.EmptyString {
			if _, has := filterMap[tenID]; !has {
				filterMap[tenID] = make(utils.StringSet)
			}
			filterMap[tenID].AddSlice(strings.Split(tp.FilterIDs, utils.InfieldSep))
		}
		tpRouteProfileMap[tenID] = tpRouteProfile
	}
	result = make([]*utils.TPRouteProfile, len(tpRouteProfileMap))
	i := 0
	for tntID, tpRouteProfile := range tpRouteProfileMap {
		result[i] = tpRouteProfile
		for _, routeData := range routeMap[tntID] {
			result[i].Routes = append(result[i].Routes, routeData)
		}
		result[i].FilterIDs = filterMap[tntID].AsSlice()
		result[i].SortingParameters = sortingParameterMap[tntID].AsSlice()
		i++
	}
	return
}

func APItoRouteProfile(tpRp *utils.TPRouteProfile, timezone string) (rp *utils.RouteProfile, err error) {
	rp = &utils.RouteProfile{
		Tenant:            tpRp.Tenant,
		ID:                tpRp.ID,
		Sorting:           tpRp.Sorting,
		Routes:            make([]*utils.Route, len(tpRp.Routes)),
		SortingParameters: make([]string, len(tpRp.SortingParameters)),
		FilterIDs:         make([]string, len(tpRp.FilterIDs)),
	}
	if tpRp.Weights != utils.EmptyString {
		rp.Weights, err = utils.NewDynamicWeightsFromString(tpRp.Weights, utils.InfieldSep, utils.ANDSep)
		if err != nil {
			return nil, err
		}
	}
	if tpRp.Blockers != utils.EmptyString {
		rp.Blockers, err = utils.NewDynamicBlockersFromString(tpRp.Blockers, utils.InfieldSep, utils.ANDSep)
		if err != nil {
			return nil, err
		}
	}
	copy(rp.SortingParameters, tpRp.SortingParameters)
	copy(rp.FilterIDs, tpRp.FilterIDs)
	for i, route := range tpRp.Routes {
		rp.Routes[i] = &utils.Route{
			ID:              route.ID,
			RateProfileIDs:  route.RateProfileIDs,
			AccountIDs:      route.AccountIDs,
			FilterIDs:       route.FilterIDs,
			ResourceIDs:     route.ResourceIDs,
			StatIDs:         route.StatIDs,
			RouteParameters: route.RouteParameters,
		}
		if route.Weights != utils.EmptyString {
			rp.Routes[i].Weights, err = utils.NewDynamicWeightsFromString(route.Weights, utils.InfieldSep, utils.ANDSep)
			if err != nil {
				return nil, err
			}
		}
		if route.Blockers != utils.EmptyString {
			rp.Routes[i].Blockers, err = utils.NewDynamicBlockersFromString(route.Blockers, utils.InfieldSep, utils.ANDSep)
			if err != nil {
				return nil, err
			}
		}
	}
	return rp, nil
}

type AttributeMdls []*AttributeMdl

func (tps AttributeMdls) AsTPAttributes() (result []*utils.TPAttributeProfile) {
	mst := make(map[string]*utils.TPAttributeProfile)
	filterMap := make(map[string]utils.StringSet)
	for _, tp := range tps {
		key := &utils.TenantID{Tenant: tp.Tenant, ID: tp.ID}
		tenID := (&utils.TenantID{Tenant: tp.Tenant, ID: tp.ID}).TenantID()
		th, found := mst[tenID]
		if !found {
			th = &utils.TPAttributeProfile{
				TPid:   tp.Tpid,
				Tenant: tp.Tenant,
				ID:     tp.ID,
			}
		}
		if tp.Blockers != utils.EmptyString {
			th.Blockers = tp.Blockers
		}
		if tp.Weights != utils.EmptyString {
			th.Weights = tp.Weights
		}
		if tp.FilterIDs != utils.EmptyString {
			if _, has := filterMap[key.TenantID()]; !has {
				filterMap[key.TenantID()] = make(utils.StringSet)
			}
			filterMap[key.TenantID()].AddSlice(strings.Split(tp.FilterIDs, utils.InfieldSep))
		}
		if tp.Path != utils.EmptyString {
			filterIDs := make([]string, 0)
			if tp.AttributeFilterIDs != utils.EmptyString {
				filterIDs = append(filterIDs, strings.Split(tp.AttributeFilterIDs, utils.InfieldSep)...)
			}
			th.Attributes = append(th.Attributes, &utils.TPAttribute{
				FilterIDs: filterIDs,
				Blockers:  tp.AttributeBlockers,
				Type:      tp.Type,
				Path:      tp.Path,
				Value:     tp.Value,
			})
		}
		mst[key.TenantID()] = th
	}
	result = make([]*utils.TPAttributeProfile, len(mst))
	i := 0
	for tntID, th := range mst {
		result[i] = th
		result[i].FilterIDs = filterMap[tntID].AsSlice()
		i++
	}
	return
}

func APItoAttributeProfile(tpAttr *utils.TPAttributeProfile, timezone string) (attrPrf *utils.AttributeProfile, err error) {
	attrPrf = &utils.AttributeProfile{
		Tenant:     tpAttr.Tenant,
		ID:         tpAttr.ID,
		FilterIDs:  make([]string, len(tpAttr.FilterIDs)),
		Attributes: make([]*utils.Attribute, len(tpAttr.Attributes)),
	}
	if tpAttr.Blockers != utils.EmptyString {
		if attrPrf.Blockers, err = utils.NewDynamicBlockersFromString(tpAttr.Blockers, utils.InfieldSep, utils.ANDSep); err != nil {
			return
		}
	}
	if tpAttr.Weights != utils.EmptyString {
		if attrPrf.Weights, err = utils.NewDynamicWeightsFromString(tpAttr.Weights, utils.InfieldSep, utils.ANDSep); err != nil {
			return
		}
	}
	copy(attrPrf.FilterIDs, tpAttr.FilterIDs)
	for i, reqAttr := range tpAttr.Attributes {
		if reqAttr.Path == utils.EmptyString { // we do not suppot empty Path in Attributes
			err = fmt.Errorf("empty path in AttributeProfile <%s>", attrPrf.TenantID())
			return
		}
		sbstPrsr, err := utils.NewRSRParsers(reqAttr.Value, utils.RSRSep)
		if err != nil {
			return nil, err
		}
		attrPrf.Attributes[i] = &utils.Attribute{
			FilterIDs: reqAttr.FilterIDs,
			Path:      reqAttr.Path,
			Type:      reqAttr.Type,
			Value:     sbstPrsr,
		}
		if reqAttr.Blockers != utils.EmptyString {
			if attrPrf.Attributes[i].Blockers, err = utils.NewDynamicBlockersFromString(reqAttr.Blockers, utils.InfieldSep, utils.ANDSep); err != nil {
				return nil, err
			}
		}
	}
	return attrPrf, nil
}

type ChargerMdls []*ChargerMdl

func (tps ChargerMdls) AsTPChargers() (result []*utils.TPChargerProfile) {
	mst := make(map[string]*utils.TPChargerProfile)
	filterMap := make(map[string]utils.StringSet)
	attributeMap := make(map[string][]string)
	for _, tp := range tps {
		tntID := (&utils.TenantID{Tenant: tp.Tenant, ID: tp.ID}).TenantID()
		tpCPP, found := mst[tntID]
		if !found {
			tpCPP = &utils.TPChargerProfile{
				TPid:   tp.Tpid,
				Tenant: tp.Tenant,
				ID:     tp.ID,
			}
		}
		if tp.Weights != utils.EmptyString {
			tpCPP.Weights = tp.Weights
		}
		if tp.Blockers != utils.EmptyString {
			tpCPP.Blockers = tp.Blockers
		}
		if tp.FilterIDs != utils.EmptyString {
			if _, has := filterMap[tntID]; !has {
				filterMap[tntID] = make(utils.StringSet)
			}
			filterMap[tntID].AddSlice(strings.Split(tp.FilterIDs, utils.InfieldSep))
		}
		if tp.RunID != utils.EmptyString {
			tpCPP.RunID = tp.RunID
		}
		if tp.AttributeIDs != utils.EmptyString {
			attributeSplit := strings.Split(tp.AttributeIDs, utils.InfieldSep)
			var inlineAttribute string
			var dynam bool
			for _, attribute := range attributeSplit {
				if !dynam && !strings.HasPrefix(attribute, utils.Meta) {
					if inlineAttribute != utils.EmptyString {
						attributeMap[tntID] = append(attributeMap[tntID], inlineAttribute[1:])
						inlineAttribute = utils.EmptyString
					}
					attributeMap[tntID] = append(attributeMap[tntID], attribute)
					continue
				}
				if dynam {
					dynam = !strings.Contains(attribute, string(utils.RSRDynEndChar))
				} else {
					dynam = strings.Contains(attribute, string(utils.RSRDynStartChar))
				}
				inlineAttribute += utils.InfieldSep + attribute
			}
			if inlineAttribute != utils.EmptyString {
				attributeMap[tntID] = append(attributeMap[tntID], inlineAttribute[1:])
				inlineAttribute = utils.EmptyString
			}
		}
		mst[tntID] = tpCPP
	}
	result = make([]*utils.TPChargerProfile, len(mst))
	i := 0
	for tntID, tp := range mst {
		result[i] = tp
		result[i].FilterIDs = filterMap[tntID].AsSlice()
		result[i].AttributeIDs = make([]string, 0, len(attributeMap[tntID]))
		result[i].AttributeIDs = append(result[i].AttributeIDs, attributeMap[tntID]...)
		i++
	}
	return
}

func APItoChargerProfile(tpCPP *utils.TPChargerProfile, timezone string) (cpp *utils.ChargerProfile) {
	cpp = &utils.ChargerProfile{
		Tenant:       tpCPP.Tenant,
		ID:           tpCPP.ID,
		RunID:        tpCPP.RunID,
		FilterIDs:    make([]string, len(tpCPP.FilterIDs)),
		AttributeIDs: make([]string, len(tpCPP.AttributeIDs)),
	}
	if tpCPP.Weights != utils.EmptyString {
		var err error
		cpp.Weights, err = utils.NewDynamicWeightsFromString(tpCPP.Weights, utils.InfieldSep, utils.ANDSep)
		if err != nil {
			return
		}
	}
	if tpCPP.Blockers != utils.EmptyString {
		var err error
		cpp.Blockers, err = utils.NewDynamicBlockersFromString(tpCPP.Blockers, utils.InfieldSep, utils.ANDSep)
		if err != nil {
			return
		}
	}
	copy(cpp.FilterIDs, tpCPP.FilterIDs)
	copy(cpp.AttributeIDs, tpCPP.AttributeIDs)
	return cpp
}

func paramsToString(sp []any) (strategy string) {
	if len(sp) != 0 {
		strategy = sp[0].(string)
		for i := 1; i < len(sp); i++ {
			strategy += utils.InfieldSep + sp[i].(string)
		}
	}
	return
}

// RateProfileMdls is used
type RateProfileMdls []*RateProfileMdl

func (tps RateProfileMdls) AsTPRateProfile() (result []*utils.TPRateProfile) {
	filterMap := make(map[string]utils.StringSet)
	mst := make(map[string]*utils.TPRateProfile)
	rateMap := make(map[string]map[string]*utils.TPRate)
	for _, tp := range tps {
		tenID := (&utils.TenantID{Tenant: tp.Tenant, ID: tp.ID}).TenantID()
		rPrf, found := mst[tenID]
		if !found {
			rPrf = &utils.TPRateProfile{
				TPid:   tp.Tpid,
				Tenant: tp.Tenant,
				ID:     tp.ID,
			}
		}
		if tp.RateID != utils.EmptyString {
			if _, has := rateMap[tenID]; !has {
				rateMap[tenID] = make(map[string]*utils.TPRate)
			}
			rate, found := rateMap[tenID][tp.RateID]
			if !found {
				rate = &utils.TPRate{
					ID:            tp.RateID,
					IntervalRates: make([]*utils.TPIntervalRate, 0),
					Blocker:       tp.RateBlocker,
				}
			}
			if tp.RateFilterIDs != utils.EmptyString {
				rateFilterSplit := strings.Split(tp.RateFilterIDs, utils.InfieldSep)
				rate.FilterIDs = append(rate.FilterIDs, rateFilterSplit...)
			}
			if tp.RateActivationTimes != utils.EmptyString {
				rate.ActivationTimes = tp.RateActivationTimes
			}
			if tp.RateWeights != utils.EmptyString {
				rate.Weights = tp.RateWeights
			}
			// create new interval rate and append to the slice
			intervalRate := new(utils.TPIntervalRate)
			if tp.RateIntervalStart != utils.EmptyString {
				intervalRate.IntervalStart = tp.RateIntervalStart
			}
			if tp.RateFixedFee != 0 {
				intervalRate.FixedFee = tp.RateFixedFee
			}
			if tp.RateRecurrentFee != 0 {
				intervalRate.RecurrentFee = tp.RateRecurrentFee
			}
			if tp.RateIncrement != utils.EmptyString {
				intervalRate.Increment = tp.RateIncrement
			}
			if tp.RateUnit != utils.EmptyString {
				intervalRate.Unit = tp.RateUnit
			}
			rate.IntervalRates = append(rate.IntervalRates, intervalRate)
			rateMap[tenID][tp.RateID] = rate
		}

		if tp.Weights != utils.EmptyString {
			rPrf.Weights = tp.Weights
		}
		if tp.MinCost != 0 {
			rPrf.MinCost = tp.MinCost
		}
		if tp.MaxCost != 0 {
			rPrf.MaxCost = tp.MaxCost
		}
		if tp.MaxCostStrategy != utils.EmptyString {
			rPrf.MaxCostStrategy = tp.MaxCostStrategy
		}
		if tp.FilterIDs != utils.EmptyString {
			if _, has := filterMap[tenID]; !has {
				filterMap[tenID] = make(utils.StringSet)
			}
			filterMap[tenID].AddSlice(strings.Split(tp.FilterIDs, utils.InfieldSep))
		}
		mst[tenID] = rPrf
	}
	result = make([]*utils.TPRateProfile, len(mst))
	i := 0
	for tntID, th := range mst {
		result[i] = th
		result[i].Rates = rateMap[tntID]
		result[i].FilterIDs = filterMap[tntID].AsSlice()
		i++
	}
	return
}

func APItoRateProfile(tpRp *utils.TPRateProfile, timezone string) (rp *utils.RateProfile, err error) {
	rp = &utils.RateProfile{
		Tenant:          tpRp.Tenant,
		ID:              tpRp.ID,
		FilterIDs:       make([]string, len(tpRp.FilterIDs)),
		MaxCostStrategy: tpRp.MaxCostStrategy,
		Rates:           make(map[string]*utils.Rate),
		MinCost:         utils.NewDecimalFromFloat64(tpRp.MinCost),
		MaxCost:         utils.NewDecimalFromFloat64(tpRp.MaxCost),
	}
	if tpRp.Weights != utils.EmptyString {
		weight, err := utils.NewDynamicWeightsFromString(tpRp.Weights, utils.InfieldSep, utils.ANDSep)
		if err != nil {
			return nil, err
		}
		rp.Weights = weight
	}
	copy(rp.FilterIDs, tpRp.FilterIDs)
	for key, rate := range tpRp.Rates {
		rp.Rates[key] = &utils.Rate{
			ID:              rate.ID,
			Blocker:         rate.Blocker,
			FilterIDs:       rate.FilterIDs,
			ActivationTimes: rate.ActivationTimes,
			IntervalRates:   make([]*utils.IntervalRate, len(rate.IntervalRates)),
		}
		if rate.Weights != utils.EmptyString {
			weight, err := utils.NewDynamicWeightsFromString(rate.Weights, utils.InfieldSep, utils.ANDSep)
			if err != nil {
				return nil, err
			}
			rp.Rates[key].Weights = weight
		}
		for i, iRate := range rate.IntervalRates {
			rp.Rates[key].IntervalRates[i] = new(utils.IntervalRate)
			if rp.Rates[key].IntervalRates[i].IntervalStart, err = utils.NewDecimalFromUsage(iRate.IntervalStart); err != nil {
				return nil, err
			}
			rp.Rates[key].IntervalRates[i].FixedFee = utils.NewDecimalFromFloat64(iRate.FixedFee)
			rp.Rates[key].IntervalRates[i].RecurrentFee = utils.NewDecimalFromFloat64(iRate.RecurrentFee)
			if rp.Rates[key].IntervalRates[i].Unit, err = utils.NewDecimalFromUsage(iRate.Unit); err != nil {
				return nil, err
			}
			if rp.Rates[key].IntervalRates[i].Increment, err = utils.NewDecimalFromUsage(iRate.Increment); err != nil {
				return nil, err
			}
		}
	}
	return rp, nil
}

type ActionProfileMdls []*ActionProfileMdl

func (apm ActionProfileMdls) AsTPActionProfile() (result []*utils.TPActionProfile) {
	filterIDsMap := make(map[string]utils.StringSet)
	targetIDsMap := make(map[string]map[string]utils.StringSet)
	actPrfMap := make(map[string]*utils.TPActionProfile)
	for _, tp := range apm {
		tenID := (&utils.TenantID{Tenant: tp.Tenant, ID: tp.ID}).TenantID()
		aPrf, found := actPrfMap[tenID]
		if !found {
			aPrf = &utils.TPActionProfile{
				TPid:   tp.Tpid,
				Tenant: tp.Tenant,
				ID:     tp.ID,
			}
		}
		if tp.FilterIDs != utils.EmptyString {
			if _, has := filterIDsMap[tenID]; !has {
				filterIDsMap[tenID] = make(utils.StringSet)
			}
			filterIDsMap[tenID].AddSlice(strings.Split(tp.FilterIDs, utils.InfieldSep))
		}
		if tp.Weights != utils.EmptyString {
			aPrf.Weights = tp.Weights
		}
		if tp.Blockers != utils.EmptyString {
			aPrf.Blockers = tp.Blockers
		}
		if tp.Schedule != utils.EmptyString {
			aPrf.Schedule = tp.Schedule
		}
		if tp.TargetType != utils.EmptyString {
			if _, has := targetIDsMap[tenID]; !has {
				targetIDsMap[tenID] = make(map[string]utils.StringSet)
			}
			targetIDsMap[tenID][tp.TargetType] = utils.NewStringSet(strings.Split(tp.TargetIDs, utils.InfieldSep))
		}

		if tp.ActionID != utils.EmptyString {
			var tpAAction *utils.TPAPAction
			if lacts := len(aPrf.Actions); lacts == 0 ||
				aPrf.Actions[lacts-1].ID != tp.ActionID {
				tpAAction = &utils.TPAPAction{
					ID:   tp.ActionID,
					TTL:  tp.ActionTTL,
					Type: tp.ActionType,
					Opts: tp.ActionOpts,
					Diktats: []*utils.TPAPDiktat{{
						ID:   tp.ActionDiktatsID,
						Opts: tp.ActionDiktatsOpts,
					}},
				}
				if tp.ActionFilterIDs != utils.EmptyString {
					tpAAction.FilterIDs = utils.NewStringSet(strings.Split(tp.ActionFilterIDs, utils.InfieldSep)).AsSlice()
				}
				if tp.ActionWeights != utils.EmptyString {
					tpAAction.Weights = tp.ActionWeights
				}
				if tp.ActionBlockers != utils.EmptyString {
					tpAAction.Blockers = tp.ActionBlockers
				}
				if tp.ActionDiktatsFilterIDs != utils.EmptyString {
					tpAAction.Diktats[0].FilterIDs = utils.NewStringSet(strings.Split(tp.ActionDiktatsFilterIDs, utils.InfieldSep)).AsSlice()
				}
				if tp.ActionDiktatsWeights != utils.EmptyString {
					tpAAction.Diktats[0].Weights = tp.ActionDiktatsWeights
				}
				if tp.ActionDiktatsBlockers != utils.EmptyString {
					tpAAction.Diktats[0].Blockers = tp.ActionDiktatsBlockers
				}
				aPrf.Actions = append(aPrf.Actions, tpAAction)
			} else {
				diktat := &utils.TPAPDiktat{
					ID:   tp.ActionDiktatsID,
					Opts: tp.ActionDiktatsOpts,
				}
				if tp.ActionDiktatsFilterIDs != utils.EmptyString {
					diktat.FilterIDs = utils.NewStringSet(strings.Split(tp.ActionDiktatsFilterIDs, utils.InfieldSep)).AsSlice()
				}
				if tp.ActionDiktatsWeights != utils.EmptyString {
					diktat.Weights = tp.ActionDiktatsWeights
				}
				if tp.ActionDiktatsBlockers != utils.EmptyString {
					diktat.Blockers = tp.ActionDiktatsBlockers
				}
				aPrf.Actions[lacts-1].Diktats = append(aPrf.Actions[lacts-1].Diktats,
					diktat)
			}
		}
		actPrfMap[tenID] = aPrf
	}
	result = make([]*utils.TPActionProfile, len(actPrfMap))
	i := 0
	for tntID, th := range actPrfMap {
		result[i] = th
		result[i].FilterIDs = filterIDsMap[tntID].AsSlice()
		for targetType, targetIDs := range targetIDsMap[tntID] {
			result[i].Targets = append(result[i].Targets, &utils.TPActionTarget{TargetType: targetType, TargetIDs: targetIDs.AsSlice()})
		}
		i++
	}
	return
}

func APItoActionProfile(tpAp *utils.TPActionProfile, timezone string) (ap *utils.ActionProfile, err error) {
	ap = &utils.ActionProfile{
		Tenant:    tpAp.Tenant,
		ID:        tpAp.ID,
		FilterIDs: make([]string, len(tpAp.FilterIDs)),
		Schedule:  tpAp.Schedule,
		Targets:   make(map[string]utils.StringSet),
		Actions:   make([]*utils.APAction, len(tpAp.Actions)),
	}
	if tpAp.Weights != utils.EmptyString {
		if ap.Weights, err = utils.NewDynamicWeightsFromString(tpAp.Weights, utils.InfieldSep, utils.ANDSep); err != nil {
			return
		}
	}
	if tpAp.Blockers != utils.EmptyString {
		if ap.Blockers, err = utils.NewDynamicBlockersFromString(tpAp.Blockers, utils.InfieldSep, utils.ANDSep); err != nil {
			return
		}
	}
	copy(ap.FilterIDs, tpAp.FilterIDs)
	for _, target := range tpAp.Targets {
		ap.Targets[target.TargetType] = utils.NewStringSet(target.TargetIDs)
	}
	for i, act := range tpAp.Actions {
		actDs := make([]*utils.APDiktat, len(act.Diktats))
		for j, actD := range act.Diktats {
			if actD.ID == utils.EmptyString {
				return nil, fmt.Errorf("missing ID from Diktats of ActionProfile <%s> Action <%s>", ap.TenantID(), actD.ID)
			}
			actDs[j] = &utils.APDiktat{
				ID:        actD.ID,
				FilterIDs: actD.FilterIDs,
			}
			if actD.Opts != utils.EmptyString {
				actDs[j].Opts = make(map[string]any)
				for opt := range strings.SplitSeq(actD.Opts, utils.InfieldSep) { // example of opts: key1:val1;key2:val2;key3:val3
					keyValSls := utils.SplitPath(opt, utils.InInFieldSep[0], 2)
					if len(keyValSls) != 2 {
						return nil, fmt.Errorf("malformed option for ActionProfile <%s> for action <%s> for diktat <%s>", ap.TenantID(), actD.ID, actD.ID)

					}
					actDs[j].Opts[keyValSls[0]] = keyValSls[1]
				}
			}
			if actD.Weights != utils.EmptyString {
				if actDs[j].Weights, err = utils.NewDynamicWeightsFromString(actD.Weights, utils.InfieldSep, utils.ANDSep); err != nil {
					return
				}
			}
			if actD.Blockers != utils.EmptyString {
				if actDs[j].Blockers, err = utils.NewDynamicBlockersFromString(actD.Blockers, utils.InfieldSep, utils.ANDSep); err != nil {
					return
				}
			}
		}
		ap.Actions[i] = &utils.APAction{
			ID:        act.ID,
			FilterIDs: act.FilterIDs,
			Type:      act.Type,
			Diktats:   actDs,
		}
		if ap.Actions[i].TTL, err = utils.ParseDurationWithNanosecs(act.TTL); err != nil {
			return
		}
		if act.Opts != utils.EmptyString {
			ap.Actions[i].Opts = make(map[string]any)
			for opt := range strings.SplitSeq(act.Opts, utils.InfieldSep) { // example of opts: key1:val1;key2:val2;key3:val3
				keyValSls := utils.SplitPath(opt, utils.InInFieldSep[0], 2)
				if len(keyValSls) != 2 {
					err = fmt.Errorf("malformed option for ActionProfile <%s> for action <%s>", ap.TenantID(), act.ID)
					return
				}
				ap.Actions[i].Opts[keyValSls[0]] = keyValSls[1]
			}
		}
		if act.Weights != utils.EmptyString {
			if ap.Actions[i].Weights, err = utils.NewDynamicWeightsFromString(act.Weights, utils.InfieldSep, utils.ANDSep); err != nil {
				return
			}
		}
		if act.Blockers != utils.EmptyString {
			if ap.Actions[i].Blockers, err = utils.NewDynamicBlockersFromString(act.Blockers, utils.InfieldSep, utils.ANDSep); err != nil {
				return
			}
		}
	}
	return
}

type AccountMdls []*AccountMdl

func (apm AccountMdls) AsTPAccount() (result []*utils.TPAccount, err error) {
	filterIDsMap := make(map[string]utils.StringSet)
	thresholdIDsMap := make(map[string]utils.StringSet)
	actPrfMap := make(map[string]*utils.TPAccount)
	for _, tp := range apm {
		tenID := (&utils.TenantID{Tenant: tp.Tenant, ID: tp.ID}).TenantID()
		aPrf, found := actPrfMap[tenID]
		if !found {
			aPrf = &utils.TPAccount{
				TPid:     tp.Tpid,
				Tenant:   tp.Tenant,
				ID:       tp.ID,
				Weights:  tp.Weights,
				Blockers: tp.Blockers,
				Balances: make(map[string]*utils.TPAccountBalance),
			}
		}
		if tp.FilterIDs != utils.EmptyString {
			if _, has := filterIDsMap[tenID]; !has {
				filterIDsMap[tenID] = make(utils.StringSet)
			}
			filterIDsMap[tenID].AddSlice(strings.Split(tp.FilterIDs, utils.InfieldSep))
		}
		if tp.ThresholdIDs != utils.EmptyString {
			if _, has := thresholdIDsMap[tenID]; !has {
				thresholdIDsMap[tenID] = make(utils.StringSet)
			}
			thresholdIDsMap[tenID].AddSlice(strings.Split(tp.ThresholdIDs, utils.InfieldSep))
		}
		if tp.BalanceID != utils.EmptyString {
			aPrf.Balances[tp.BalanceID] = &utils.TPAccountBalance{
				ID:       tp.BalanceID,
				Weights:  tp.BalanceWeights,
				Blockers: tp.BalanceBlockers,
				Type:     tp.BalanceType,
				Opts:     tp.BalanceOpts,
				Units:    tp.BalanceUnits,
			}

			if tp.BalanceFilterIDs != utils.EmptyString {
				aPrf.Balances[tp.BalanceID].FilterIDs = utils.NewStringSet(strings.Split(tp.BalanceFilterIDs, utils.InfieldSep)).AsSlice()
			}
			// cost increment mdl: fltr1&fltr2;incr;fixed;recurrent
			if tp.BalanceCostIncrements != utils.EmptyString {
				costIncrements := make([]*utils.TPBalanceCostIncrement, 0)
				sls := strings.Split(tp.BalanceCostIncrements, utils.InfieldSep)
				if len(sls)%4 != 0 {
					return nil, fmt.Errorf("invalid key: <%s> for BalanceCostIncrements", tp.BalanceCostIncrements)
				}
				for j := 0; j < len(sls); j = j + 4 {
					costIncrement, err := utils.NewTPBalanceCostIncrement(sls[j], sls[j+1], sls[j+2], sls[j+3])
					if err != nil {
						return nil, err
					}
					costIncrements = append(costIncrements, costIncrement)
				}
				aPrf.Balances[tp.BalanceID].CostIncrement = costIncrements
			}
			if tp.BalanceAttributeIDs != utils.EmptyString {
				// the order for attributes is important
				// also no duplicate check as we would
				// need to let the user execute the same
				// attribute twice if needed
				aPrf.Balances[tp.BalanceID].AttributeIDs = strings.Split(tp.BalanceAttributeIDs, utils.InfieldSep)
			}
			if tp.BalanceRateProfileIDs != utils.EmptyString {
				aPrf.Balances[tp.BalanceID].RateProfileIDs = utils.NewStringSet(strings.Split(tp.BalanceRateProfileIDs, utils.InfieldSep)).AsSlice()
			}
			if tp.BalanceUnitFactors != utils.EmptyString {
				unitFactors := make([]*utils.TPBalanceUnitFactor, 0)
				sls := strings.Split(tp.BalanceUnitFactors, utils.InfieldSep)
				if len(sls)%2 != 0 {
					return nil, fmt.Errorf("invalid key: <%s> for BalanceUnitFactors", tp.BalanceUnitFactors)
				}

				for j := 0; j < len(sls); j = j + 2 {
					unitFactor, err := utils.NewTPBalanceUnitFactor(sls[j], sls[j+1])
					if err != nil {
						return nil, err
					}
					unitFactors = append(unitFactors, unitFactor)
				}
				aPrf.Balances[tp.BalanceID].UnitFactors = unitFactors
			}

		}
		actPrfMap[tenID] = aPrf
	}
	result = make([]*utils.TPAccount, len(actPrfMap))
	i := 0
	for tntID, th := range actPrfMap {
		result[i] = th
		result[i].FilterIDs = filterIDsMap[tntID].AsSlice()
		result[i].ThresholdIDs = thresholdIDsMap[tntID].AsSlice()
		i++
	}
	return
}

func APItoAccount(tpAcc *utils.TPAccount, timezone string) (acc *utils.Account, err error) {
	acc = &utils.Account{
		Tenant:       tpAcc.Tenant,
		ID:           tpAcc.ID,
		FilterIDs:    make([]string, len(tpAcc.FilterIDs)),
		Balances:     make(map[string]*utils.Balance, len(tpAcc.Balances)),
		ThresholdIDs: make([]string, len(tpAcc.ThresholdIDs)),
	}
	if tpAcc.Weights != utils.EmptyString {
		weight, err := utils.NewDynamicWeightsFromString(tpAcc.Weights, utils.InfieldSep, utils.ANDSep)
		if err != nil {
			return nil, err
		}
		acc.Weights = weight
	}
	if tpAcc.Blockers != utils.EmptyString {
		blockers, err := utils.NewDynamicBlockersFromString(tpAcc.Blockers, utils.InfieldSep, utils.ANDSep)
		if err != nil {
			return nil, err
		}
		acc.Blockers = blockers
	}
	copy(acc.FilterIDs, tpAcc.FilterIDs)
	for id, bal := range tpAcc.Balances {
		acc.Balances[id] = &utils.Balance{
			ID:        bal.ID,
			FilterIDs: bal.FilterIDs,
			Type:      bal.Type,
		}
		if bal.Units != utils.EmptyString {
			units, err := utils.NewDecimalFromUsage(bal.Units)
			if err != nil {
				return nil, err
			}
			acc.Balances[id].Units = units
		}
		if bal.Weights != utils.EmptyString {
			weights, err := utils.NewDynamicWeightsFromString(bal.Weights, utils.InfieldSep, utils.ANDSep)
			if err != nil {
				return nil, err
			}
			acc.Balances[id].Weights = weights
		}
		if bal.Blockers != utils.EmptyString {
			blockers, err := utils.NewDynamicBlockersFromString(bal.Blockers, utils.InfieldSep, utils.ANDSep)
			if err != nil {
				return nil, err
			}
			acc.Balances[id].Blockers = blockers
		}
		if bal.UnitFactors != nil {
			acc.Balances[id].UnitFactors = make([]*utils.UnitFactor, len(bal.UnitFactors))
			for j, unitFactor := range bal.UnitFactors {
				acc.Balances[id].UnitFactors[j] = &utils.UnitFactor{
					FilterIDs: unitFactor.FilterIDs,
					Factor:    utils.NewDecimalFromFloat64(unitFactor.Factor),
				}
			}
		}
		if bal.Opts != utils.EmptyString {
			acc.Balances[id].Opts = make(map[string]any)
			for _, opt := range strings.Split(bal.Opts, utils.InfieldSep) { // example of opts: key1:val1;key2:val2;key3:val3
				keyValSls := utils.SplitConcatenatedKey(opt)
				if len(keyValSls) != 2 {
					err = fmt.Errorf("malformed option for ActionProfile <%s> for action <%s>", acc.TenantID(), bal.ID)
					return
				}
				acc.Balances[id].Opts[keyValSls[0]] = keyValSls[1]
			}
		}
		if bal.CostIncrement != nil {
			acc.Balances[id].CostIncrements = make([]*utils.CostIncrement, len(bal.CostIncrement))
			for j, costIncrement := range bal.CostIncrement {
				acc.Balances[id].CostIncrements[j] = &utils.CostIncrement{
					FilterIDs: costIncrement.FilterIDs,
				}
				if costIncrement.Increment != utils.EmptyString {
					acc.Balances[id].CostIncrements[j].Increment, err = utils.NewDecimalFromUsage(costIncrement.Increment)
				}
				if costIncrement.FixedFee != nil {
					acc.Balances[id].CostIncrements[j].FixedFee = utils.NewDecimalFromFloat64(*costIncrement.FixedFee)
				}
				if costIncrement.RecurrentFee != nil {
					acc.Balances[id].CostIncrements[j].RecurrentFee = utils.NewDecimalFromFloat64(*costIncrement.RecurrentFee)
				}
			}
		}
		if bal.AttributeIDs != nil {
			acc.Balances[id].AttributeIDs = make([]string, len(bal.AttributeIDs))
			copy(acc.Balances[id].AttributeIDs, bal.AttributeIDs)
		}
		if bal.RateProfileIDs != nil {
			acc.Balances[id].RateProfileIDs = make([]string, len(bal.RateProfileIDs))
			copy(acc.Balances[id].RateProfileIDs, bal.RateProfileIDs)
		}
	}
	copy(acc.ThresholdIDs, tpAcc.ThresholdIDs)
	return
}
