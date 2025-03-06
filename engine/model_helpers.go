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

package engine

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

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

// CsvDump receive and interface and convert it to a slice of string
func CsvDump(s any) ([]string, error) {
	fieldIndexMap := make(map[string]int)
	st := reflect.ValueOf(s)
	if st.Kind() == reflect.Ptr {
		st = st.Elem()
		s = st.Interface()
	}
	numFields := st.NumField()
	stCopy := reflect.TypeOf(s)
	for i := 0; i < numFields; i++ {
		field := stCopy.Field(i)
		index := field.Tag.Get("index")
		if index != utils.EmptyString {
			if idx, err := strconv.Atoi(index); err != nil {
				return nil, fmt.Errorf("invalid %v.%v index %v", stCopy.Name(), field.Name, index)
			} else {
				fieldIndexMap[field.Name] = idx
			}
		}
	}
	result := make([]string, len(fieldIndexMap))
	for fieldName, fieldIndex := range fieldIndexMap {
		field := st.FieldByName(fieldName)
		if field.IsValid() && fieldIndex < len(result) {
			switch field.Kind() {
			case reflect.Float64:
				result[fieldIndex] = strconv.FormatFloat(field.Float(), 'f', -1, 64)
			case reflect.Int:
				result[fieldIndex] = strconv.FormatInt(field.Int(), 10)
			case reflect.Bool:
				result[fieldIndex] = strconv.FormatBool(field.Bool())
			case reflect.String:
				result[fieldIndex] = field.String()
			}
		}
	}
	return result, nil
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

// CSVHeader return the header for csv fields as a slice of string
func (tps ResourceMdls) CSVHeader() (result []string) {
	return []string{"#" + utils.Tenant, utils.ID, utils.FilterIDs, utils.Weights,
		utils.UsageTTL, utils.Limit, utils.AllocationMessage, utils.Blocker, utils.Stored,
		utils.ThresholdIDs}
}

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

func APItoModelResource(rl *utils.TPResourceProfile) (mdls ResourceMdls) {
	if rl == nil {
		return
	}
	// In case that TPResourceProfile don't have filter
	if len(rl.FilterIDs) == 0 {
		mdl := &ResourceMdl{
			Tpid:              rl.TPid,
			Tenant:            rl.Tenant,
			ID:                rl.ID,
			Blocker:           rl.Blocker,
			Stored:            rl.Stored,
			UsageTTL:          rl.UsageTTL,
			Weights:           rl.Weights,
			Limit:             rl.Limit,
			AllocationMessage: rl.AllocationMessage,
		}
		for i, val := range rl.ThresholdIDs {
			if i != 0 {
				mdl.ThresholdIDs += utils.InfieldSep
			}
			mdl.ThresholdIDs += val
		}
		mdls = append(mdls, mdl)
	}
	for i, fltr := range rl.FilterIDs {
		mdl := &ResourceMdl{
			Tpid:    rl.TPid,
			Tenant:  rl.Tenant,
			ID:      rl.ID,
			Blocker: rl.Blocker,
			Stored:  rl.Stored,
		}
		if i == 0 {
			mdl.UsageTTL = rl.UsageTTL
			mdl.Weights = rl.Weights
			mdl.Limit = rl.Limit
			mdl.AllocationMessage = rl.AllocationMessage
			for i, val := range rl.ThresholdIDs {
				if i != 0 {
					mdl.ThresholdIDs += utils.InfieldSep
				}
				mdl.ThresholdIDs += val
			}
		}
		mdl.FilterIDs = fltr
		mdls = append(mdls, mdl)
	}
	return
}

func APItoResource(tpRL *utils.TPResourceProfile, timezone string) (rp *ResourceProfile, err error) {
	rp = &ResourceProfile{
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

func ResourceProfileToAPI(rp *ResourceProfile) (tpRL *utils.TPResourceProfile) {
	tpRL = &utils.TPResourceProfile{
		Tenant:            rp.Tenant,
		ID:                rp.ID,
		FilterIDs:         make([]string, len(rp.FilterIDs)),
		Limit:             strconv.FormatFloat(rp.Limit, 'f', -1, 64),
		AllocationMessage: rp.AllocationMessage,
		Blocker:           rp.Blocker,
		Stored:            rp.Stored,
		Weights:           rp.Weights.String(utils.InfieldSep, utils.ANDSep),
		ThresholdIDs:      make([]string, len(rp.ThresholdIDs)),
	}
	if rp.UsageTTL != time.Duration(0) {
		tpRL.UsageTTL = rp.UsageTTL.String()
	}
	copy(tpRL.FilterIDs, rp.FilterIDs)
	copy(tpRL.ThresholdIDs, rp.ThresholdIDs)
	return
}

type StatMdls []*StatMdl

// CSVHeader return the header for csv fields as a slice of string
func (tps StatMdls) CSVHeader() (result []string) {
	return []string{"#" + utils.Tenant, utils.ID, utils.FilterIDs, utils.Weights, utils.Blockers, utils.QueueLength, utils.TTL, utils.MinItems, utils.Stored, utils.ThresholdIDs, utils.MetricIDs, utils.MetricFilterIDs, utils.MetricBlockers}
}

func (tps StatMdls) AsTPStats() (result []*utils.TPStatProfile) {
	filterMap := make(map[string]utils.StringSet)
	thresholdMap := make(map[string]utils.StringSet)
	statMetricsMap := make(map[string]map[string]*utils.MetricWithFilters)
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
				statMetricsMap[key.TenantID()] = make(map[string]*utils.MetricWithFilters)
			}
			metricIDsSplit := strings.Split(model.MetricIDs, utils.InfieldSep)
			for _, metricID := range metricIDsSplit {
				stsMetric, found := statMetricsMap[key.TenantID()][metricID]
				if !found {
					stsMetric = &utils.MetricWithFilters{
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

func APItoModelStats(st *utils.TPStatProfile) (mdls StatMdls) {
	if st != nil && len(st.Metrics) != 0 {
		for i, metric := range st.Metrics {
			mdl := &StatMdl{
				Tpid:   st.TPid,
				Tenant: st.Tenant,
				ID:     st.ID,
			}
			if i == 0 {
				for i, val := range st.FilterIDs {
					if i != 0 {
						mdl.FilterIDs += utils.InfieldSep
					}
					mdl.FilterIDs += val
				}
				mdl.QueueLength = st.QueueLength
				mdl.TTL = st.TTL
				mdl.MinItems = st.MinItems
				mdl.Stored = st.Stored
				mdl.Blockers = st.Blockers
				mdl.Weights = st.Weights
				for i, val := range st.ThresholdIDs {
					if i != 0 {
						mdl.ThresholdIDs += utils.InfieldSep
					}
					mdl.ThresholdIDs += val
				}
			}
			for i, val := range metric.FilterIDs {
				if i != 0 {
					mdl.MetricFilterIDs += utils.InfieldSep
				}
				mdl.MetricFilterIDs += val
			}
			mdl.MetricBlockers = metric.Blockers
			mdl.MetricIDs = metric.MetricID
			mdls = append(mdls, mdl)
		}
	}
	return
}

func APItoStats(tpST *utils.TPStatProfile, timezone string) (st *StatQueueProfile, err error) {
	st = &StatQueueProfile{
		Tenant:       tpST.Tenant,
		ID:           tpST.ID,
		FilterIDs:    make([]string, len(tpST.FilterIDs)),
		QueueLength:  tpST.QueueLength,
		MinItems:     tpST.MinItems,
		Metrics:      make([]*MetricWithFilters, len(tpST.Metrics)),
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
		st.Metrics[i] = &MetricWithFilters{
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

func StatQueueProfileToAPI(st *StatQueueProfile) (tpST *utils.TPStatProfile) {
	tpST = &utils.TPStatProfile{
		Tenant:       st.Tenant,
		ID:           st.ID,
		FilterIDs:    make([]string, len(st.FilterIDs)),
		QueueLength:  st.QueueLength,
		Metrics:      make([]*utils.MetricWithFilters, len(st.Metrics)),
		Blockers:     st.Blockers.String(utils.InfieldSep, utils.ANDSep),
		Stored:       st.Stored,
		Weights:      st.Weights.String(utils.InfieldSep, utils.ANDSep),
		MinItems:     st.MinItems,
		ThresholdIDs: make([]string, len(st.ThresholdIDs)),
	}
	for i, metric := range st.Metrics {
		tpST.Metrics[i] = &utils.MetricWithFilters{
			MetricID: metric.MetricID,
			Blockers: metric.Blockers.String(utils.InfieldSep, utils.ANDSep),
		}
		if len(metric.FilterIDs) != 0 {
			tpST.Metrics[i].FilterIDs = make([]string, len(metric.FilterIDs))
			copy(tpST.Metrics[i].FilterIDs, metric.FilterIDs)
		}
	}
	if st.TTL != time.Duration(0) {
		tpST.TTL = st.TTL.String()
	}
	copy(tpST.FilterIDs, st.FilterIDs)
	copy(tpST.ThresholdIDs, st.ThresholdIDs)
	return
}

type RankingMdls []*RankingMdl

func (tps RankingMdls) CSVHeader() (result []string) {
	return []string{"#" + utils.Tenant, utils.ID, utils.Schedule, utils.StatIDs,
		utils.MetricIDs, utils.Sorting, utils.SortingParameters, utils.Stored,
		utils.ThresholdIDs}
}

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

func APItoModelTPRanking(tpRG *utils.TPRankingProfile) (mdls RankingMdls) {
	if tpRG == nil {
		return
	}
	if len(tpRG.StatIDs) == 0 {
		mdl := &RankingMdl{
			Tpid:     tpRG.TPid,
			Tenant:   tpRG.Tenant,
			ID:       tpRG.ID,
			Schedule: tpRG.Schedule,
			Sorting:  tpRG.Sorting,
			Stored:   tpRG.Stored,
		}

		for i, val := range tpRG.ThresholdIDs {
			if i != 0 {
				mdl.ThresholdIDs += utils.InfieldSep
			}
			mdl.ThresholdIDs += val
		}
		for i, metric := range tpRG.MetricIDs {
			if i != 0 {
				mdl.MetricIDs += utils.InfieldSep
			}
			mdl.MetricIDs += metric
		}
		for i, sorting := range tpRG.SortingParameters {
			if i != 0 {
				mdl.SortingParameters += utils.InfieldSep
			}
			mdl.SortingParameters += sorting
		}

		mdls = append(mdls, mdl)
	}
	for i, stat := range tpRG.StatIDs {
		mdl := &RankingMdl{
			Tpid:   tpRG.TPid,
			Tenant: tpRG.Tenant,
			ID:     tpRG.ID,
		}
		if i == 0 {
			mdl.Schedule = tpRG.Schedule
			mdl.Sorting = tpRG.Sorting
			for i, val := range tpRG.ThresholdIDs {
				if i != 0 {
					mdl.ThresholdIDs += utils.InfieldSep
				}
				mdl.ThresholdIDs += val
			}
			for i, metric := range tpRG.MetricIDs {
				if i != 0 {
					mdl.MetricIDs += utils.InfieldSep
				}
				mdl.MetricIDs += metric
			}
			for i, sorting := range tpRG.SortingParameters {
				if i != 0 {
					mdl.SortingParameters += utils.InfieldSep
				}
				mdl.SortingParameters += sorting
			}
		}
		mdl.StatIDs = stat
		mdls = append(mdls, mdl)
	}
	return
}

func APItoRanking(tpRG *utils.TPRankingProfile) (rg *RankingProfile, err error) {
	rg = &RankingProfile{
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

func RankingProfileToAPI(rg *RankingProfile) (tpRG *utils.TPRankingProfile) {
	tpRG = &utils.TPRankingProfile{
		Tenant:            rg.Tenant,
		ID:                rg.ID,
		Schedule:          rg.Schedule,
		Sorting:           rg.Sorting,
		Stored:            rg.Stored,
		StatIDs:           make([]string, len(rg.StatIDs)),
		MetricIDs:         make([]string, len(rg.MetricIDs)),
		SortingParameters: make([]string, len(rg.SortingParameters)),
		ThresholdIDs:      make([]string, len(rg.ThresholdIDs)),
	}
	copy(tpRG.StatIDs, rg.StatIDs)
	copy(tpRG.ThresholdIDs, rg.ThresholdIDs)
	copy(tpRG.MetricIDs, rg.MetricIDs)
	copy(tpRG.SortingParameters, rg.SortingParameters)
	return
}

type TrendMdls []*TrendMdl

func (tps TrendMdls) CSVHeader() (result []string) {
	return []string{"#" + utils.Tenant, utils.ID, utils.Schedule, utils.StatID,
		utils.Metrics, utils.TTL, utils.QueueLength, utils.MinItems, utils.CorrelationType, utils.Tolerance, utils.Stored, utils.ThresholdIDs}
}

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

func APItoModelTrends(tr *utils.TPTrendsProfile) (mdls TrendMdls) {
	if tr != nil {
		mdl := &TrendMdl{
			Tpid:   tr.TPid,
			Tenant: tr.Tenant,
			ID:     tr.ID,
		}
		mdl.Schedule = tr.Schedule
		mdl.QueueLength = tr.QueueLength
		mdl.StatID = tr.StatID
		mdl.TTL = tr.TTL
		mdl.MinItems = tr.MinItems
		mdl.CorrelationType = tr.CorrelationType
		mdl.Tolerance = tr.Tolerance
		mdl.Stored = tr.Stored
		for i, val := range tr.ThresholdIDs {
			if i != 0 {
				mdl.ThresholdIDs += utils.InfieldSep
			}
			mdl.ThresholdIDs += val
		}
		for i, val := range tr.Metrics {
			if i != 0 {
				mdl.Metrics += utils.InfieldSep
			}
			mdl.Metrics += val
		}
		mdls = append(mdls, mdl)
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

func TrendProfileToAPI(tr *utils.TrendProfile) (tpTR *utils.TPTrendsProfile) {
	tpTR = &utils.TPTrendsProfile{
		Tenant:          tr.Tenant,
		ID:              tr.ID,
		Schedule:        tr.Schedule,
		StatID:          tr.StatID,
		ThresholdIDs:    make([]string, len(tr.ThresholdIDs)),
		Metrics:         make([]string, len(tr.Metrics)),
		QueueLength:     tr.QueueLength,
		MinItems:        tr.MinItems,
		CorrelationType: tr.CorrelationType,
		Tolerance:       tr.Tolerance,
		Stored:          tr.Stored,
	}
	if tr.TTL != time.Duration(0) {
		tpTR.TTL = tr.TTL.String()
	}
	copy(tpTR.ThresholdIDs, tr.ThresholdIDs)
	copy(tpTR.Metrics, tr.Metrics)
	return
}

type ThresholdMdls []*ThresholdMdl

// CSVHeader return the header for csv fields as a slice of string
func (tps ThresholdMdls) CSVHeader() (result []string) {
	return []string{"#" + utils.Tenant, utils.ID, utils.FilterIDs, utils.Weights,
		utils.MaxHits, utils.MinHits, utils.MinSleep,
		utils.Blocker, utils.ActionProfileIDs, utils.Async}
}

func (tps ThresholdMdls) AsTPThreshold() (result []*utils.TPThresholdProfile) {
	mst := make(map[string]*utils.TPThresholdProfile)
	filterMap := make(map[string]utils.StringSet)
	actionMap := make(map[string]utils.StringSet)
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

		mst[tenID] = th
	}
	result = make([]*utils.TPThresholdProfile, len(mst))
	i := 0
	for tntID, th := range mst {
		result[i] = th
		result[i].FilterIDs = filterMap[tntID].AsSlice()
		result[i].ActionProfileIDs = actionMap[tntID].AsSlice()
		i++
	}
	return
}

func APItoModelTPThreshold(th *utils.TPThresholdProfile) (mdls ThresholdMdls) {
	if th != nil {
		if len(th.ActionProfileIDs) == 0 {
			return
		}
		min := len(th.FilterIDs)
		if min > len(th.ActionProfileIDs) {
			min = len(th.ActionProfileIDs)
		}
		for i := 0; i < min; i++ {
			mdl := &ThresholdMdl{
				Tpid:   th.TPid,
				Tenant: th.Tenant,
				ID:     th.ID,
			}
			if i == 0 {
				mdl.Blocker = th.Blocker
				mdl.Weights = th.Weights
				mdl.MaxHits = th.MaxHits
				mdl.MinHits = th.MinHits
				mdl.MinSleep = th.MinSleep
				mdl.Async = th.Async
			}
			mdl.FilterIDs = th.FilterIDs[i]
			mdl.ActionProfileIDs = th.ActionProfileIDs[i]
			mdls = append(mdls, mdl)
		}

		if len(th.FilterIDs)-min > 0 {
			for i := min; i < len(th.FilterIDs); i++ {
				mdl := &ThresholdMdl{
					Tpid:   th.TPid,
					Tenant: th.Tenant,
					ID:     th.ID,
				}
				mdl.FilterIDs = th.FilterIDs[i]
				mdls = append(mdls, mdl)
			}
		}
		if len(th.ActionProfileIDs)-min > 0 {
			for i := min; i < len(th.ActionProfileIDs); i++ {
				mdl := &ThresholdMdl{
					Tpid:   th.TPid,
					Tenant: th.Tenant,
					ID:     th.ID,
				}
				if min == 0 && i == 0 {
					mdl.Blocker = th.Blocker
					mdl.Weights = th.Weights
					mdl.MaxHits = th.MaxHits
					mdl.MinHits = th.MinHits
					mdl.MinSleep = th.MinSleep
					mdl.Async = th.Async
				}
				mdl.ActionProfileIDs = th.ActionProfileIDs[i]
				mdls = append(mdls, mdl)
			}
		}
	}
	return
}

func APItoThresholdProfile(tpTH *utils.TPThresholdProfile, timezone string) (th *ThresholdProfile, err error) {
	th = &ThresholdProfile{
		Tenant:           tpTH.Tenant,
		ID:               tpTH.ID,
		MaxHits:          tpTH.MaxHits,
		MinHits:          tpTH.MinHits,
		Blocker:          tpTH.Blocker,
		Async:            tpTH.Async,
		ActionProfileIDs: make([]string, len(tpTH.ActionProfileIDs)),
		FilterIDs:        make([]string, len(tpTH.FilterIDs)),
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
	copy(th.ActionProfileIDs, tpTH.ActionProfileIDs)
	copy(th.FilterIDs, tpTH.FilterIDs)
	return th, nil
}

func ThresholdProfileToAPI(th *ThresholdProfile) (tpTH *utils.TPThresholdProfile) {
	tpTH = &utils.TPThresholdProfile{
		Tenant:           th.Tenant,
		ID:               th.ID,
		FilterIDs:        make([]string, len(th.FilterIDs)),
		MaxHits:          th.MaxHits,
		MinHits:          th.MinHits,
		Blocker:          th.Blocker,
		Weights:          th.Weights.String(utils.InfieldSep, utils.ANDSep),
		ActionProfileIDs: make([]string, len(th.ActionProfileIDs)),
		Async:            th.Async,
	}
	if th.MinSleep != time.Duration(0) {
		tpTH.MinSleep = th.MinSleep.String()
	}
	copy(tpTH.FilterIDs, th.FilterIDs)
	copy(tpTH.ActionProfileIDs, th.ActionProfileIDs)
	return
}

type FilterMdls []*FilterMdl

// CSVHeader return the header for csv fields as a slice of string
func (tps FilterMdls) CSVHeader() (result []string) {
	return []string{"#" + utils.Tenant, utils.ID, utils.Type, utils.Element,
		utils.Values}
}

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

func APItoModelTPFilter(th *utils.TPFilterProfile) (mdls FilterMdls) {
	if th == nil || len(th.Filters) == 0 {
		return
	}
	for _, fltr := range th.Filters {
		mdl := &FilterMdl{
			Tpid:   th.TPid,
			Tenant: th.Tenant,
			ID:     th.ID,
		}
		mdl.Type = fltr.Type
		mdl.Element = fltr.Element
		for i, val := range fltr.Values {
			if i != 0 {
				mdl.Values += utils.InfieldSep
			}
			mdl.Values += val
		}
		mdls = append(mdls, mdl)
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

func FilterToTPFilter(f *Filter) (tpFltr *utils.TPFilterProfile) {
	tpFltr = &utils.TPFilterProfile{
		Tenant:  f.Tenant,
		ID:      f.ID,
		Filters: make([]*utils.TPFilter, len(f.Rules)),
	}
	for i, reqFltr := range f.Rules {
		tpFltr.Filters[i] = &utils.TPFilter{
			Type:    reqFltr.Type,
			Element: reqFltr.Element,
			Values:  make([]string, len(reqFltr.Values)),
		}
		copy(tpFltr.Filters[i].Values, reqFltr.Values)
	}
	return
}

type RouteMdls []*RouteMdl

// CSVHeader return the header for csv fields as a slice of string
func (tps RouteMdls) CSVHeader() (result []string) {
	return []string{"#" + utils.Tenant, utils.ID, utils.FilterIDs, utils.Weights,
		utils.Sorting, utils.SortingParameters, utils.RouteID, utils.RouteFilterIDs,
		utils.RouteAccountIDs, utils.RouteRateProfileIDs, utils.RouteRateProfileIDs,
		utils.RouteResourceIDs, utils.RouteStatIDs, utils.RouteWeights, utils.RouteBlockers,
		utils.RouteParameters,
	}
}

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

func APItoModelTPRoutes(st *utils.TPRouteProfile) (mdls RouteMdls) {
	if len(st.Routes) == 0 {
		return
	}
	for i, route := range st.Routes {
		mdl := &RouteMdl{
			Tenant: st.Tenant,
			Tpid:   st.TPid,
			ID:     st.ID,
		}
		if i == 0 {
			mdl.Sorting = st.Sorting
			mdl.Weights = st.Weights
			mdl.Blockers = st.Blockers
			for i, val := range st.FilterIDs {
				if i != 0 {
					mdl.FilterIDs += utils.InfieldSep
				}
				mdl.FilterIDs += val
			}
			for i, val := range st.SortingParameters {
				if i != 0 {
					mdl.SortingParameters += utils.InfieldSep
				}
				mdl.SortingParameters += val
			}
		}
		mdl.RouteID = route.ID
		for i, val := range route.AccountIDs {
			if i != 0 {
				mdl.RouteAccountIDs += utils.InfieldSep
			}
			mdl.RouteAccountIDs += val
		}
		for i, val := range route.RateProfileIDs {
			if i != 0 {
				mdl.RouteRateProfileIDs += utils.InfieldSep
			}
			mdl.RouteRateProfileIDs += val
		}
		for i, val := range route.FilterIDs {
			if i != 0 {
				mdl.RouteFilterIDs += utils.InfieldSep
			}
			mdl.RouteFilterIDs += val
		}
		for i, val := range route.ResourceIDs {
			if i != 0 {
				mdl.RouteResourceIDs += utils.InfieldSep
			}
			mdl.RouteResourceIDs += val
		}
		for i, val := range route.StatIDs {
			if i != 0 {
				mdl.RouteStatIDs += utils.InfieldSep
			}
			mdl.RouteStatIDs += val
		}
		mdl.RouteWeights = route.Weights
		mdl.RouteParameters = route.RouteParameters
		mdl.RouteBlockers = route.Blockers
		mdls = append(mdls, mdl)
	}
	return
}

func APItoRouteProfile(tpRp *utils.TPRouteProfile, timezone string) (rp *RouteProfile, err error) {
	rp = &RouteProfile{
		Tenant:            tpRp.Tenant,
		ID:                tpRp.ID,
		Sorting:           tpRp.Sorting,
		Routes:            make([]*Route, len(tpRp.Routes)),
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
		rp.Routes[i] = &Route{
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

func RouteProfileToAPI(rp *RouteProfile) (tpRp *utils.TPRouteProfile) {
	tpRp = &utils.TPRouteProfile{
		Tenant:            rp.Tenant,
		ID:                rp.ID,
		FilterIDs:         make([]string, len(rp.FilterIDs)),
		Weights:           rp.Weights.String(utils.InfieldSep, utils.ANDSep),
		Blockers:          rp.Blockers.String(utils.InfieldSep, utils.ANDSep),
		Sorting:           rp.Sorting,
		SortingParameters: make([]string, len(rp.SortingParameters)),
		Routes:            make([]*utils.TPRoute, len(rp.Routes)),
	}

	for i, route := range rp.Routes {
		tpRp.Routes[i] = &utils.TPRoute{
			ID:              route.ID,
			FilterIDs:       route.FilterIDs,
			AccountIDs:      route.AccountIDs,
			RateProfileIDs:  route.RateProfileIDs,
			ResourceIDs:     route.ResourceIDs,
			StatIDs:         route.StatIDs,
			Weights:         route.Weights.String(utils.InfieldSep, utils.ANDSep),
			Blockers:        route.Blockers.String(utils.InfieldSep, utils.ANDSep),
			RouteParameters: route.RouteParameters,
		}
	}
	copy(tpRp.FilterIDs, rp.FilterIDs)
	copy(tpRp.SortingParameters, rp.SortingParameters)
	return
}

type AttributeMdls []*AttributeMdl

// CSVHeader return the header for csv fields as a slice of string
func (tps AttributeMdls) CSVHeader() (result []string) {
	return []string{"#" + utils.Tenant, utils.ID, utils.FilterIDs, utils.Weights, utils.Blockers, utils.AttributeFilterIDs, utils.AttributeBlockers, utils.Path, utils.Type, utils.Value}
}

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

func APItoModelTPAttribute(ap *utils.TPAttributeProfile) (mdls AttributeMdls) {
	if len(ap.Attributes) == 0 {
		return
	}
	for i, reqAttribute := range ap.Attributes {
		mdl := &AttributeMdl{
			Tpid:   ap.TPid,
			Tenant: ap.Tenant,
			ID:     ap.ID,
		}
		if i == 0 {
			for i, val := range ap.FilterIDs {
				if i != 0 {
					mdl.FilterIDs += utils.InfieldSep
				}
				mdl.FilterIDs += val
			}
			if ap.Blockers != utils.EmptyString {
				mdl.Blockers = ap.Blockers
			}
			if ap.Weights != utils.EmptyString {
				mdl.Weights = ap.Weights
			}
		}
		mdl.AttributeBlockers = reqAttribute.Blockers
		mdl.Path = reqAttribute.Path
		mdl.Value = reqAttribute.Value
		mdl.Type = reqAttribute.Type
		mdl.AttributeFilterIDs = strings.Join(reqAttribute.FilterIDs, utils.InfieldSep)
		mdls = append(mdls, mdl)
	}
	return
}

func APItoAttributeProfile(tpAttr *utils.TPAttributeProfile, timezone string) (attrPrf *AttributeProfile, err error) {
	attrPrf = &AttributeProfile{
		Tenant:     tpAttr.Tenant,
		ID:         tpAttr.ID,
		FilterIDs:  make([]string, len(tpAttr.FilterIDs)),
		Attributes: make([]*Attribute, len(tpAttr.Attributes)),
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
		attrPrf.Attributes[i] = &Attribute{
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

func AttributeProfileToAPI(attrPrf *AttributeProfile) (tpAttr *utils.TPAttributeProfile) {
	tpAttr = &utils.TPAttributeProfile{
		Tenant:     attrPrf.Tenant,
		ID:         attrPrf.ID,
		FilterIDs:  make([]string, len(attrPrf.FilterIDs)),
		Attributes: make([]*utils.TPAttribute, len(attrPrf.Attributes)),
		Blockers:   attrPrf.Blockers.String(utils.InfieldSep, utils.ANDSep),
		Weights:    attrPrf.Weights.String(utils.InfieldSep, utils.ANDSep),
	}
	copy(tpAttr.FilterIDs, attrPrf.FilterIDs)
	for i, attr := range attrPrf.Attributes {
		tpAttr.Attributes[i] = &utils.TPAttribute{
			FilterIDs: attr.FilterIDs,
			Blockers:  attr.Blockers.String(utils.InfieldSep, utils.ANDSep),
			Path:      attr.Path,
			Type:      attr.Type,
			Value:     attr.Value.GetRule(),
		}
	}
	return
}

type ChargerMdls []*ChargerMdl

// CSVHeader return the header for csv fields as a slice of string
func (tps ChargerMdls) CSVHeader() (result []string) {
	return []string{"#" + utils.Tenant, utils.ID, utils.FilterIDs, utils.Weights,
		utils.Blockers, utils.RunID, utils.AttributeIDs}
}

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

func APItoModelTPCharger(tpCPP *utils.TPChargerProfile) (mdls ChargerMdls) {
	if tpCPP != nil {
		min := len(tpCPP.FilterIDs)
		isFilter := true
		if min > len(tpCPP.AttributeIDs) {
			min = len(tpCPP.AttributeIDs)
			isFilter = false
		}
		if min == 0 {
			mdl := &ChargerMdl{
				Tenant:   tpCPP.Tenant,
				Tpid:     tpCPP.TPid,
				ID:       tpCPP.ID,
				Weights:  tpCPP.Weights,
				Blockers: tpCPP.Blockers,
				RunID:    tpCPP.RunID,
			}
			if isFilter && len(tpCPP.AttributeIDs) > 0 {
				mdl.AttributeIDs = tpCPP.AttributeIDs[0]
			} else if len(tpCPP.FilterIDs) > 0 {
				mdl.FilterIDs = tpCPP.FilterIDs[0]
			}
			min = 1
			mdls = append(mdls, mdl)
		} else {
			for i := 0; i < min; i++ {
				mdl := &ChargerMdl{
					Tenant: tpCPP.Tenant,
					Tpid:   tpCPP.TPid,
					ID:     tpCPP.ID,
				}
				if i == 0 {
					mdl.Weights = tpCPP.Weights
					mdl.Blockers = tpCPP.Blockers
					mdl.RunID = tpCPP.RunID
				}
				mdl.AttributeIDs = tpCPP.AttributeIDs[i]
				mdl.FilterIDs = tpCPP.FilterIDs[i]
				mdls = append(mdls, mdl)
			}
		}
		if len(tpCPP.FilterIDs)-min > 0 {
			for i := min; i < len(tpCPP.FilterIDs); i++ {
				mdl := &ChargerMdl{
					Tenant: tpCPP.Tenant,
					Tpid:   tpCPP.TPid,
					ID:     tpCPP.ID,
				}
				mdl.FilterIDs = tpCPP.FilterIDs[i]
				mdls = append(mdls, mdl)
			}
		}
		if len(tpCPP.AttributeIDs)-min > 0 {
			for i := min; i < len(tpCPP.AttributeIDs); i++ {
				mdl := &ChargerMdl{
					Tenant: tpCPP.Tenant,
					Tpid:   tpCPP.TPid,
					ID:     tpCPP.ID,
				}
				mdl.AttributeIDs = tpCPP.AttributeIDs[i]
				mdls = append(mdls, mdl)
			}
		}
	}
	return
}

func APItoChargerProfile(tpCPP *utils.TPChargerProfile, timezone string) (cpp *ChargerProfile) {
	cpp = &ChargerProfile{
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

func ChargerProfileToAPI(chargerPrf *ChargerProfile) (tpCharger *utils.TPChargerProfile) {
	tpCharger = &utils.TPChargerProfile{
		Tenant:       chargerPrf.Tenant,
		ID:           chargerPrf.ID,
		FilterIDs:    make([]string, len(chargerPrf.FilterIDs)),
		Weights:      chargerPrf.Weights.String(utils.InfieldSep, utils.ANDSep),
		Blockers:     chargerPrf.Blockers.String(utils.InfieldSep, utils.ANDSep),
		RunID:        chargerPrf.RunID,
		AttributeIDs: make([]string, len(chargerPrf.AttributeIDs)),
	}
	copy(tpCharger.FilterIDs, chargerPrf.FilterIDs)
	copy(tpCharger.AttributeIDs, chargerPrf.AttributeIDs)
	return
}

// CSVHeader return the header for csv fields as a slice of string

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

// CSVHeader return the header for csv fields as a slice of string
func (tps RateProfileMdls) CSVHeader() (result []string) {
	return []string{"#" + utils.Tenant, utils.ID, utils.FilterIDs,
		utils.Weights, utils.ConnectFee, utils.MinCost, utils.MaxCost, utils.MaxCostStrategy,
		utils.RateID, utils.RateFilterIDs, utils.RateActivationStart, utils.RateWeights,
		utils.RateBlocker, utils.RateIntervalStart, utils.RateFixedFee, utils.RateRecurrentFee,
		utils.RateUnit, utils.RateIncrement,
	}
}

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

func APItoModelTPRateProfile(tPrf *utils.TPRateProfile) (mdls RateProfileMdls) {
	if len(tPrf.Rates) == 0 {
		return
	}
	i := 0
	for _, rate := range tPrf.Rates {
		for j, intervalRate := range rate.IntervalRates {
			mdl := &RateProfileMdl{
				Tenant: tPrf.Tenant,
				Tpid:   tPrf.TPid,
				ID:     tPrf.ID,
			}
			if i == 0 {
				for i, val := range tPrf.FilterIDs {
					if i != 0 {
						mdl.FilterIDs += utils.InfieldSep
					}
					mdl.FilterIDs += val
				}

				mdl.Weights = tPrf.Weights
				mdl.MinCost = tPrf.MinCost
				mdl.MaxCost = tPrf.MaxCost
				mdl.MaxCostStrategy = tPrf.MaxCostStrategy
			}
			mdl.RateID = rate.ID
			if j == 0 {
				for i, val := range rate.FilterIDs {
					if i != 0 {
						mdl.RateFilterIDs += utils.InfieldSep
					}
					mdl.RateFilterIDs += val
				}
				mdl.RateWeights = rate.Weights
				mdl.RateActivationTimes = rate.ActivationTimes
				mdl.RateBlocker = rate.Blocker

			}
			mdl.RateRecurrentFee = intervalRate.RecurrentFee
			mdl.RateFixedFee = intervalRate.FixedFee
			mdl.RateUnit = intervalRate.Unit
			mdl.RateIncrement = intervalRate.Increment
			mdl.RateIntervalStart = intervalRate.IntervalStart
			mdls = append(mdls, mdl)
			i++
		}

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

func RateProfileToAPI(rp *utils.RateProfile) (tpRp *utils.TPRateProfile) {
	tpRp = &utils.TPRateProfile{
		Tenant:          rp.Tenant,
		ID:              rp.ID,
		FilterIDs:       make([]string, len(rp.FilterIDs)),
		Weights:         rp.Weights.String(utils.InfieldSep, utils.ANDSep),
		MaxCostStrategy: rp.MaxCostStrategy,
		Rates:           make(map[string]*utils.TPRate),
	}
	if rp.MinCost != nil {
		//there should not be an invalid value of converting from Decimal into float64
		minCostF, _ := rp.MinCost.Float64()
		tpRp.MinCost = minCostF
	}
	if rp.MaxCost != nil {
		//there should not be an invalid value of converting from Decimal into float64
		maxCostF, _ := rp.MaxCost.Float64()
		tpRp.MaxCost = maxCostF
	}

	for key, rate := range rp.Rates {
		tpRp.Rates[key] = &utils.TPRate{
			ID:              rate.ID,
			Weights:         rate.Weights.String(utils.InfieldSep, utils.ANDSep),
			Blocker:         rate.Blocker,
			FilterIDs:       rate.FilterIDs,
			ActivationTimes: rate.ActivationTimes,
			IntervalRates:   make([]*utils.TPIntervalRate, len(rate.IntervalRates)),
		}
		for i, iRate := range rate.IntervalRates {
			tpRp.Rates[key].IntervalRates[i] = &utils.TPIntervalRate{
				IntervalStart: iRate.IntervalStart.String(),
			}
			if iRate.FixedFee != nil {
				//there should not be an invalid value of converting from Decimal into float64
				fixedFeeF, _ := iRate.FixedFee.Float64()
				tpRp.Rates[key].IntervalRates[i].FixedFee = fixedFeeF
			}
			if iRate.Unit != nil {
				tpRp.Rates[key].IntervalRates[i].Unit = iRate.Unit.String()
			}
			if iRate.Increment != nil {
				tpRp.Rates[key].IntervalRates[i].Increment = iRate.Increment.String()
			}
			if iRate.RecurrentFee != nil {
				//there should not be an invalid value of converting from Decimal into float64
				recFeeF, _ := iRate.RecurrentFee.Float64()
				tpRp.Rates[key].IntervalRates[i].RecurrentFee = recFeeF
			}
		}
	}
	copy(tpRp.FilterIDs, rp.FilterIDs)
	return
}

type ActionProfileMdls []*ActionProfileMdl

// CSVHeader return the header for csv fields as a slice of string
func (apm ActionProfileMdls) CSVHeader() (result []string) {
	return []string{"#" + utils.Tenant, utils.ID, utils.FilterIDs,
		utils.Weights, utils.Blockers, utils.Schedule, utils.TargetType, utils.TargetIDs,
		utils.ActionID, utils.ActionFilterIDs, utils.ActionTTL,
		utils.ActionType, utils.ActionOpts, utils.ActionPath, utils.ActionValue,
	}
}

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
						Path:  tp.ActionPath,
						Value: tp.ActionValue,
					}},
				}
				if tp.ActionFilterIDs != utils.EmptyString {
					tpAAction.FilterIDs = utils.NewStringSet(strings.Split(tp.ActionFilterIDs, utils.InfieldSep)).AsSlice()
				}
				aPrf.Actions = append(aPrf.Actions, tpAAction)
			} else {
				aPrf.Actions[lacts-1].Diktats = append(aPrf.Actions[lacts-1].Diktats, &utils.TPAPDiktat{
					Path:  tp.ActionPath,
					Value: tp.ActionValue,
				})
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

func APItoModelTPActionProfile(tPrf *utils.TPActionProfile) (mdls ActionProfileMdls) {
	if len(tPrf.Actions) == 0 {
		return
	}
	for i, action := range tPrf.Actions {
		mdl := &ActionProfileMdl{
			Tenant: tPrf.Tenant,
			Tpid:   tPrf.TPid,
			ID:     tPrf.ID,
		}
		if i == 0 {
			mdl.FilterIDs = strings.Join(tPrf.FilterIDs, utils.InfieldSep)
			mdl.Weights = tPrf.Weights
			mdl.Blockers = tPrf.Blockers
			mdl.Schedule = tPrf.Schedule
			for _, target := range tPrf.Targets {
				mdl.TargetType = target.TargetType
				mdl.TargetIDs = strings.Join(target.TargetIDs, utils.InfieldSep)
			}
		}
		mdl.ActionID = action.ID
		mdl.ActionFilterIDs = strings.Join(action.FilterIDs, utils.InfieldSep)
		mdl.ActionTTL = action.TTL
		mdl.ActionType = action.Type
		mdl.ActionOpts = action.Opts
		for j, actD := range action.Diktats {
			if j != 0 {
				mdl = &ActionProfileMdl{
					Tenant:     mdl.Tenant,
					Tpid:       mdl.Tpid,
					ID:         mdl.ID,
					ActionID:   mdl.ActionID,
					ActionType: mdl.ActionType,
				}
			}
			mdl.ActionPath = actD.Path
			mdl.ActionValue = actD.Value
		}
		mdls = append(mdls, mdl)
	}
	return
}

func APItoActionProfile(tpAp *utils.TPActionProfile, timezone string) (ap *ActionProfile, err error) {
	ap = &ActionProfile{
		Tenant:    tpAp.Tenant,
		ID:        tpAp.ID,
		FilterIDs: make([]string, len(tpAp.FilterIDs)),
		Schedule:  tpAp.Schedule,
		Targets:   make(map[string]utils.StringSet),
		Actions:   make([]*APAction, len(tpAp.Actions)),
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
		actDs := make([]*APDiktat, len(act.Diktats))
		for j, actD := range act.Diktats {
			actDs[j] = &APDiktat{
				Path:  actD.Path,
				Value: actD.Value,
			}
		}
		ap.Actions[i] = &APAction{
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
			for _, opt := range strings.Split(act.Opts, utils.InfieldSep) { // example of opts: key1:val1;key2:val2;key3:val3
				keyValSls := utils.SplitConcatenatedKey(opt)
				if len(keyValSls) != 2 {
					err = fmt.Errorf("malformed option for ActionProfile <%s> for action <%s>", ap.TenantID(), act.ID)
					return
				}
				ap.Actions[i].Opts[keyValSls[0]] = keyValSls[1]
			}
		}

	}
	return
}

func ActionProfileToAPI(ap *ActionProfile) (tpAp *utils.TPActionProfile) {
	tpAp = &utils.TPActionProfile{
		Tenant:    ap.Tenant,
		ID:        ap.ID,
		FilterIDs: make([]string, len(ap.FilterIDs)),
		Weights:   ap.Weights.String(utils.InfieldSep, utils.ANDSep),
		Blockers:  ap.Blockers.String(utils.InfieldSep, utils.ANDSep),
		Schedule:  ap.Schedule,
		Targets:   make([]*utils.TPActionTarget, 0, len(ap.Targets)),
		Actions:   make([]*utils.TPAPAction, len(ap.Actions)),
	}
	copy(tpAp.FilterIDs, ap.FilterIDs)
	for targetType, targetIDs := range ap.Targets {
		tpAp.Targets = append(tpAp.Targets, &utils.TPActionTarget{TargetType: targetType, TargetIDs: targetIDs.AsSlice()})
	}
	for i, act := range ap.Actions {
		actDs := make([]*utils.TPAPDiktat, len(act.Diktats))
		for j, actD := range act.Diktats {
			actDs[j] = &utils.TPAPDiktat{
				Path:  actD.Path,
				Value: actD.Value,
			}
		}

		elems := make([]string, 0, len(act.Opts))
		for k, v := range act.Opts {
			elems = append(elems, utils.ConcatenatedKey(k, utils.IfaceAsString(v)))
		}
		tpAp.Actions[i] = &utils.TPAPAction{
			ID:        act.ID,
			FilterIDs: act.FilterIDs,
			TTL:       act.TTL.String(),
			Type:      act.Type,
			Diktats:   actDs,
			Opts:      strings.Join(elems, utils.InfieldSep),
		}
	}
	return
}

type AccountMdls []*AccountMdl

// CSVHeader return the header for csv fields as a slice of string
func (apm AccountMdls) CSVHeader() (result []string) {
	return []string{"#" + utils.Tenant, utils.ID, utils.FilterIDs,
		utils.Weights, utils.Blockers, utils.Opts, utils.BalanceID, utils.BalanceFilterIDs, utils.BalanceWeights, utils.BalanceBlockers, utils.BalanceType, utils.BalanceUnits, utils.BalanceUnitFactors, utils.BalanceOpts, utils.BalanceCostIncrements, utils.BalanceAttributeIDs, utils.BalanceRateProfileIDs,
		utils.ThresholdIDs}
}

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

func APItoModelTPAccount(tPrf *utils.TPAccount) (mdls AccountMdls) {
	if len(tPrf.Balances) == 0 {
		return
	}
	i := 0
	for _, balance := range tPrf.Balances {
		mdl := &AccountMdl{
			Tenant: tPrf.Tenant,
			Tpid:   tPrf.TPid,
			ID:     tPrf.ID,
		}
		if i == 0 {
			for i, val := range tPrf.FilterIDs {
				if i != 0 {
					mdl.FilterIDs += utils.InfieldSep
				}
				mdl.FilterIDs += val
			}
			for i, val := range tPrf.ThresholdIDs {
				if i != 0 {
					mdl.ThresholdIDs += utils.InfieldSep
				}
				mdl.ThresholdIDs += val
			}
			mdl.Weights = tPrf.Weights
			mdl.Blockers = tPrf.Blockers
		}
		mdl.BalanceID = balance.ID
		for i, val := range balance.FilterIDs {
			if i != 0 {
				mdl.BalanceFilterIDs += utils.InfieldSep
			}
			mdl.BalanceFilterIDs += val
		}
		mdl.BalanceWeights = balance.Weights
		mdl.BalanceBlockers = balance.Blockers
		mdl.BalanceType = balance.Type
		mdl.BalanceOpts = balance.Opts
		for i, costIncr := range balance.CostIncrement {
			if i != 0 {
				mdl.BalanceCostIncrements += utils.InfieldSep
			}
			mdl.BalanceCostIncrements += costIncr.AsString()
		}
		for i, attrID := range balance.AttributeIDs {
			if i != 0 {
				mdl.BalanceAttributeIDs += utils.InfieldSep
			}
			mdl.BalanceAttributeIDs += attrID
		}
		for i, ratePrfID := range balance.RateProfileIDs {
			if i != 0 {
				mdl.BalanceRateProfileIDs += utils.InfieldSep
			}
			mdl.BalanceRateProfileIDs += ratePrfID
		}
		for i, unitFactor := range balance.UnitFactors {
			if i != 0 {
				mdl.BalanceUnitFactors += utils.InfieldSep
			}
			mdl.BalanceUnitFactors += unitFactor.AsString()
		}
		mdl.BalanceUnits = balance.Units
		mdls = append(mdls, mdl)
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

func AccountToAPI(acc *utils.Account) (tpAcc *utils.TPAccount) {
	tpAcc = &utils.TPAccount{
		Tenant:       acc.Tenant,
		ID:           acc.ID,
		Weights:      acc.Weights.String(utils.InfieldSep, utils.ANDSep),
		Blockers:     acc.Blockers.String(utils.InfieldSep, utils.ANDSep),
		FilterIDs:    make([]string, len(acc.FilterIDs)),
		Balances:     make(map[string]*utils.TPAccountBalance, len(acc.Balances)),
		ThresholdIDs: make([]string, len(acc.ThresholdIDs)),
	}
	copy(tpAcc.FilterIDs, acc.FilterIDs)
	for i, bal := range acc.Balances {
		tpAcc.Balances[i] = &utils.TPAccountBalance{
			ID:             bal.ID,
			FilterIDs:      make([]string, len(bal.FilterIDs)),
			Weights:        bal.Weights.String(utils.InfieldSep, utils.ANDSep),
			Blockers:       bal.Blockers.String(utils.InfieldSep, utils.ANDSep),
			Type:           bal.Type,
			Units:          bal.Units.String(),
			CostIncrement:  make([]*utils.TPBalanceCostIncrement, len(bal.CostIncrements)),
			AttributeIDs:   make([]string, len(bal.AttributeIDs)),
			RateProfileIDs: make([]string, len(bal.RateProfileIDs)),
			UnitFactors:    make([]*utils.TPBalanceUnitFactor, len(bal.UnitFactors)),
		}
		copy(tpAcc.Balances[i].FilterIDs, bal.FilterIDs)
		//there should not be an invalid value of converting into float64
		elems := make([]string, 0, len(bal.Opts))
		for k, v := range bal.Opts {
			elems = append(elems, utils.ConcatenatedKey(k, utils.IfaceAsString(v)))
		}
		for k, cIncrement := range bal.CostIncrements {
			tpAcc.Balances[i].CostIncrement[k] = &utils.TPBalanceCostIncrement{
				FilterIDs: make([]string, len(cIncrement.FilterIDs)),
				Increment: cIncrement.Increment.String(),
			}
			copy(tpAcc.Balances[i].CostIncrement[k].FilterIDs, cIncrement.FilterIDs)
			if cIncrement.FixedFee != nil {
				//there should not be an invalid value of converting from Decimal into float64
				fxdFee, _ := cIncrement.FixedFee.Float64()
				tpAcc.Balances[i].CostIncrement[k].FixedFee = &fxdFee
			}
			if cIncrement.RecurrentFee != nil {
				//there should not be an invalid value of converting from Decimal into float64
				rcrFee, _ := cIncrement.RecurrentFee.Float64()
				tpAcc.Balances[i].CostIncrement[k].RecurrentFee = &rcrFee
			}
		}
		copy(tpAcc.Balances[i].AttributeIDs, bal.AttributeIDs)
		copy(tpAcc.Balances[i].RateProfileIDs, bal.RateProfileIDs)
		for k, uFactor := range bal.UnitFactors {
			tpAcc.Balances[i].UnitFactors[k] = &utils.TPBalanceUnitFactor{
				FilterIDs: make([]string, len(uFactor.FilterIDs)),
			}
			copy(tpAcc.Balances[i].UnitFactors[k].FilterIDs, uFactor.FilterIDs)
			if uFactor.Factor != nil {
				//there should not be an invalid value of converting from Decimal into float64
				untFctr, _ := uFactor.Factor.Float64()
				tpAcc.Balances[i].UnitFactors[k].Factor = untFctr
			}
		}
		tpAcc.Balances[i].Opts = strings.Join(elems, utils.InfieldSep)
	}
	copy(tpAcc.ThresholdIDs, acc.ThresholdIDs)
	return
}
