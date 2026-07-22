// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package engine

import (
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
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
	cfg := config.NewDefaultCGRConfig()
	locker := NewGuardianLocker(cfg)
	cacheS := NewCacheS(cfg, nil, nil, nil, locker)
	data, _ := NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := NewDBConnManager(map[string]DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dmMatch = NewDataManager(dbCM, cfg, nil, locker)
	dmMatch.SetCache(cacheS)
	ctx := utils.MetaRating
	x, err := NewFilterRule(utils.MetaString, "~*req.Field", []string{"profile"})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	stringFilter = append(stringFilter, x)
	attribStringF := &Filter{
		Tenant: cfg.GeneralCfg().DefaultTenant,
		ID:     "stringFilter",
		Rules:  stringFilter}
	dmMatch.SetFilter(context.Background(), attribStringF, true)
	x, err = NewFilterRule(utils.MetaPrefix, "~*req.Field", []string{"profilePrefix"})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	prefixFilter = append(prefixFilter, x)
	attribPrefF := &Filter{
		Tenant: cfg.GeneralCfg().DefaultTenant,
		ID:     "prefFilter",
		Rules:  prefixFilter}
	dmMatch.SetFilter(context.Background(), attribPrefF, true)
	x, err = NewFilterRule(utils.MetaGreaterOrEqual, "~*req.Weight", []string{"200.00"})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	defaultFilter = append(defaultFilter, x)
	attribDefaultF := &Filter{
		Tenant: cfg.GeneralCfg().DefaultTenant,
		ID:     "defaultFilter",
		Rules:  defaultFilter}
	dmMatch.SetFilter(context.Background(), attribDefaultF, true)

	x, err = NewFilterRule(utils.MetaSuffix, "~*req.Field", []string{"Prefix"})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	suffixFilter = append(suffixFilter, x)
	attribSufF := &Filter{
		Tenant: cfg.GeneralCfg().DefaultTenant,
		ID:     "sufFilter",
		Rules:  suffixFilter}
	dmMatch.SetFilter(context.Background(), attribSufF, true)

	tnt := cfg.GeneralCfg().DefaultTenant
	if err = addItemToFilterIndex(context.TODO(), dmMatch, utils.CacheAttributeFilterIndexes,
		tnt, ctx, stringFilterID, []string{"stringFilter"}); err != nil {
		t.Error(err)
	}
	if err = addItemToFilterIndex(context.TODO(), dmMatch, utils.CacheAttributeFilterIndexes,
		tnt, ctx, prefixFilterID, []string{"prefFilter"}); err != nil {
		t.Error(err)
	}
	if err = addItemToFilterIndex(context.TODO(), dmMatch, utils.CacheAttributeFilterIndexes,
		tnt, ctx, suffixFilterID, []string{"sufFilter"}); err != nil {
		t.Error(err)
	}
	tntCtx := utils.ConcatenatedKey(tnt, ctx)

	matchEV = utils.MapStorage{utils.MetaReq: map[string]any{
		utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
		"Field":          "profile",
	}}
	aPrflIDs, err := MatchingItemIDsForEvent(context.TODO(), matchEV, nil, nil, nil, nil, nil,
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
	aPrflIDs, err = MatchingItemIDsForEvent(context.TODO(), matchEV, nil, nil, nil, nil, nil,
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
	aPrflIDs, err = MatchingItemIDsForEvent(context.TODO(), matchEV, nil, nil, nil, nil, nil,
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
	cfg := config.NewDefaultCGRConfig()
	locker := NewGuardianLocker(cfg)
	cacheS := NewCacheS(cfg, nil, nil, nil, locker)
	data, _ := NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := NewDBConnManager(map[string]DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dmMatch = NewDataManager(dbCM, cfg, nil, locker)
	dmMatch.SetCache(cacheS)
	ctx := utils.MetaRating
	x, err := NewFilterRule(utils.MetaString, "~*req.CallCost.Account", []string{"1001"})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	stringFilter = append(stringFilter, x)
	attribStringF := &Filter{
		Tenant: cfg.GeneralCfg().DefaultTenant,
		ID:     "stringFilter",
		Rules:  stringFilter}
	dmMatch.SetFilter(context.TODO(), attribStringF, true)

	x, err = NewFilterRule(utils.MetaPrefix, "~*req.CallCost.Field", []string{"profile"})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	prefixFilter = append(prefixFilter, x)
	attribPrefF := &Filter{
		Tenant: cfg.GeneralCfg().DefaultTenant,
		ID:     "prefFilter",
		Rules:  prefixFilter}
	dmMatch.SetFilter(context.TODO(), attribPrefF, true)

	x, err = NewFilterRule(utils.MetaGreaterOrEqual, "~*req.Weight", []string{"200.00"})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	defaultFilter = append(defaultFilter, x)
	attribDefaultF := &Filter{
		Tenant: cfg.GeneralCfg().DefaultTenant,
		ID:     "defaultFilter",
		Rules:  defaultFilter,
	}
	dmMatch.SetFilter(context.TODO(), attribDefaultF, true)

	tnt := cfg.GeneralCfg().DefaultTenant
	if err = addItemToFilterIndex(context.TODO(), dmMatch, utils.CacheAttributeFilterIndexes,
		tnt, ctx, stringFilterID, []string{"stringFilter"}); err != nil {
		t.Error(err)
	}
	if err = addItemToFilterIndex(context.TODO(), dmMatch, utils.CacheAttributeFilterIndexes,
		tnt, ctx, prefixFilterID, []string{"prefFilter"}); err != nil {
		t.Error(err)
	}
	tntCtx := utils.ConcatenatedKey(cfg.GeneralCfg().DefaultTenant, ctx)

	matchEV = utils.MapStorage{utils.MetaReq: map[string]any{
		utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
		"CallCost":       map[string]any{"Account": 1001},
	}}
	aPrflIDs, err := MatchingItemIDsForEvent(context.TODO(), matchEV, nil, nil, nil, nil, nil,
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
	aPrflIDs, err = MatchingItemIDsForEvent(context.TODO(), matchEV, nil, nil, nil, nil, nil,
		dmMatch, utils.CacheAttributeFilterIndexes, tntCtx, true, true)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	_, has = aPrflIDs[prefixFilterID]
	if !has {
		t.Errorf("Expecting: %+v, received: %+v", prefixFilterID, aPrflIDs)
	}
}
