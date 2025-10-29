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
	"bytes"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

var (
	matchEV map[string]any
	dmMatch *DataManager
)

func TestFilterMatchingItemIDsForEvent(t *testing.T) {
	var stringFilter, prefixFilter, suffixFilter, defaultFilter []*FilterRule
	stringFilterID := "stringFilterID"
	prefixFilterID := "prefixFilterID"
	suffixFilterID := "suffixFilterID"
	data, dErr := NewInternalDB(nil, nil, true, nil, config.CgrConfig().DataDbCfg().Items)
	if dErr != nil {
		t.Error(dErr)
	}
	dmMatch = NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	Cache.Clear(nil)
	context := utils.MetaRating
	x, err := NewFilterRule(utils.MetaString, "~*req.Field", []string{"profile"})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	stringFilter = append(stringFilter, x)
	attribStringF := &Filter{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "stringFilter",
		Rules:  stringFilter}
	dmMatch.SetFilter(attribStringF, true)
	x, err = NewFilterRule(utils.MetaPrefix, "~*req.Field", []string{"profilePrefix"})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	prefixFilter = append(prefixFilter, x)
	attribPrefF := &Filter{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "prefFilter",
		Rules:  prefixFilter}
	dmMatch.SetFilter(attribPrefF, true)
	x, err = NewFilterRule(utils.MetaGreaterOrEqual, "~*req.Weight", []string{"200.00"})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	defaultFilter = append(defaultFilter, x)
	attribDefaultF := &Filter{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "defaultFilter",
		Rules:  defaultFilter}
	dmMatch.SetFilter(attribDefaultF, true)

	x, err = NewFilterRule(utils.MetaSuffix, "~*req.Field", []string{"Prefix"})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	suffixFilter = append(suffixFilter, x)
	attribSufF := &Filter{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "sufFilter",
		Rules:  suffixFilter}
	dmMatch.SetFilter(attribSufF, true)

	tnt := config.CgrConfig().GeneralCfg().DefaultTenant
	if err = addItemToFilterIndex(dmMatch, utils.CacheAttributeFilterIndexes,
		tnt, context, stringFilterID, []string{"stringFilter"}); err != nil {
		t.Error(err)
	}
	if err = addItemToFilterIndex(dmMatch, utils.CacheAttributeFilterIndexes,
		tnt, context, prefixFilterID, []string{"prefFilter"}); err != nil {
		t.Error(err)
	}
	if err = addItemToFilterIndex(dmMatch, utils.CacheAttributeFilterIndexes,
		tnt, context, suffixFilterID, []string{"sufFilter"}); err != nil {
		t.Error(err)
	}
	tntCtx := utils.ConcatenatedKey(tnt, context)

	matchEV = utils.MapStorage{utils.MetaReq: map[string]any{
		utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
		"Field":          "profile",
	}}
	aPrflIDs, err := MatchingItemIDsForEvent(matchEV, nil, nil, nil, nil,
		dmMatch, utils.CacheAttributeFilterIndexes, tntCtx, true, false)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	_, has := aPrflIDs[stringFilterID]
	if !has {
		t.Errorf("Expecting: %+v, received: %+v", stringFilterID, aPrflIDs)
	}

	matchEV = utils.MapStorage{utils.MetaReq: map[string]any{
		"Field": "profilePrefix",
	}}
	aPrflIDs, err = MatchingItemIDsForEvent(matchEV, nil, nil, nil, nil,
		dmMatch, utils.CacheAttributeFilterIndexes, tntCtx, true, false)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	_, has = aPrflIDs[prefixFilterID]
	if !has {
		t.Errorf("Expecting: %+v, received: %+v", prefixFilterID, aPrflIDs)
	}

	matchEV = utils.MapStorage{utils.MetaReq: map[string]any{
		"Field": "profilePrefix",
	}}
	aPrflIDs, err = MatchingItemIDsForEvent(matchEV, nil, nil, nil, nil,
		dmMatch, utils.CacheAttributeFilterIndexes, tntCtx, true, false)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	_, has = aPrflIDs[suffixFilterID]
	if !has {
		t.Errorf("Expecting: %+v, received: %+v", suffixFilterID, aPrflIDs)
	}
}

func TestFilterMatchingItemIDsForEvent2(t *testing.T) {
	var stringFilter, prefixFilter, defaultFilter []*FilterRule
	stringFilterID := "stringFilterID"
	prefixFilterID := "prefixFilterID"
	data, dErr := NewInternalDB(nil, nil, true, nil, config.CgrConfig().DataDbCfg().Items)
	if dErr != nil {
		t.Error(dErr)
	}
	dmMatch = NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	context := utils.MetaRating
	x, err := NewFilterRule(utils.MetaString, "~*req.CallCost.Account", []string{"1001"})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	stringFilter = append(stringFilter, x)
	attribStringF := &Filter{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "stringFilter",
		Rules:  stringFilter}
	dmMatch.SetFilter(attribStringF, true)
	x, err = NewFilterRule(utils.MetaPrefix, "~*req.CallCost.Field", []string{"profile"})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	prefixFilter = append(prefixFilter, x)
	attribPrefF := &Filter{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "prefFilter",
		Rules:  prefixFilter}
	dmMatch.SetFilter(attribPrefF, true)
	x, err = NewFilterRule(utils.MetaGreaterOrEqual, "~*req.Weight", []string{"200.00"})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	defaultFilter = append(defaultFilter, x)
	attribDefaultF := &Filter{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "defaultFilter",
		Rules:  defaultFilter}
	dmMatch.SetFilter(attribDefaultF, true)

	tnt := config.CgrConfig().GeneralCfg().DefaultTenant
	if err = addItemToFilterIndex(dmMatch, utils.CacheAttributeFilterIndexes,
		tnt, context, stringFilterID, []string{"stringFilter"}); err != nil {
		t.Error(err)
	}
	if err = addItemToFilterIndex(dmMatch, utils.CacheAttributeFilterIndexes,
		tnt, context, prefixFilterID, []string{"prefFilter"}); err != nil {
		t.Error(err)
	}
	tntCtx := utils.ConcatenatedKey(config.CgrConfig().GeneralCfg().DefaultTenant, context)

	matchEV = utils.MapStorage{utils.MetaReq: map[string]any{
		utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
		"CallCost":       map[string]any{"Account": 1001},
	}}
	aPrflIDs, err := MatchingItemIDsForEvent(matchEV, nil, nil, nil, nil,
		dmMatch, utils.CacheAttributeFilterIndexes, tntCtx, true, true)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	_, has := aPrflIDs[stringFilterID]
	if !has {
		t.Errorf("Expecting: %+v, received: %+v", stringFilterID, aPrflIDs)
	}
	matchEV = utils.MapStorage{utils.MetaReq: map[string]any{
		"CallCost": map[string]any{"Field": "profilePrefix"},
	}}
	aPrflIDs, err = MatchingItemIDsForEvent(matchEV, nil, nil, nil, nil,
		dmMatch, utils.CacheAttributeFilterIndexes, tntCtx, true, true)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	_, has = aPrflIDs[prefixFilterID]
	if !has {
		t.Errorf("Expecting: %+v, received: %+v", prefixFilterID, aPrflIDs)
	}
}

func TestMatchingItemIDsForEventWarningThresholds(t *testing.T) {
	utils.Logger.SetLogLevel(4)
	utils.Logger.SetSyslog(nil)
	var buf bytes.Buffer
	log.SetOutput(&buf)
	t.Cleanup(func() {
		utils.Logger.SetLogLevel(0)
		log.SetOutput(os.Stderr)
	})
	cfg := config.NewDefaultCGRConfig()
	data, dErr := NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	if dErr != nil {
		t.Fatal(dErr)
	}
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	tnt := cfg.GeneralCfg().DefaultTenant
	contextKey := utils.MetaRating
	tntCtx := utils.ConcatenatedKey(tnt, contextKey)
	for i := 0; i < matchedItemsWarningThreshold+5; i++ {
		fid := fmt.Sprintf("filter_%d", i)
		rule := &FilterRule{
			Type:    utils.MetaString,
			Element: "~*req.Field",
			Values:  []string{"trigger"},
		}
		if err := rule.CompileValues(); err != nil {
			t.Fatal(err)
		}
		f := &Filter{
			Tenant: tnt,
			ID:     fid,
			Rules:  []*FilterRule{rule},
		}
		if err := dm.SetFilter(f, true); err != nil {
			t.Fatal(err)
		}
		if err := addItemToFilterIndex(dm, utils.CacheAttributeFilterIndexes,
			tnt, contextKey, fid, []string{fid}); err != nil {
			t.Fatal(err)
		}
	}
	ev := utils.MapStorage{
		utils.MetaReq: map[string]any{
			"Field": "trigger",
		},
	}
	ids, err := MatchingItemIDsForEvent(ev, nil, nil, nil, nil,
		dm, utils.CacheAttributeFilterIndexes, tntCtx, true, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(ids) <= matchedItemsWarningThreshold {
		t.Errorf("Expected more than %d matched items, got %d",
			matchedItemsWarningThreshold, len(ids))
	}
	expectedLog := "Matched 105 *attribute_filter_indexes items. Performance may be affected."
	logContent := buf.String()
	if !strings.Contains(logContent, expectedLog) {
		t.Errorf("Expected warning not found in logs:\n%s", logContent)
	}
}
