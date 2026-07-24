// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package engine

import (
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

var (
	matchEV             map[string]any
	dmMatch             *DataManager
	stringIndexedFields *[]string
	prefixIndexedFields *[]string
)

func TestFilterMatchingItemIDsForEvent(t *testing.T) {
	var stringFilter, prefixFilter, defaultFilter []*FilterRule
	stringFilterID := "stringFilterID"
	prefixFilterID := "prefixFilterID"
	data := NewInternalDB(nil, nil, true, config.CgrConfig().DataDbCfg().Items)
	dmMatch = NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
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
	dmMatch.SetFilter(attribStringF)
	x, err = NewFilterRule(utils.MetaPrefix, "~*req.Field", []string{"profilePrefix"})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	prefixFilter = append(prefixFilter, x)
	attribPrefF := &Filter{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "prefFilter",
		Rules:  prefixFilter}
	dmMatch.SetFilter(attribPrefF)
	x, err = NewFilterRule(utils.MetaGreaterOrEqual, "~*req.Weight", []string{"200.00"})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	defaultFilter = append(defaultFilter, x)
	attribDefaultF := &Filter{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "defaultFilter",
		Rules:  defaultFilter}
	dmMatch.SetFilter(attribDefaultF)
	prefix := utils.ConcatenatedKey(config.CgrConfig().GeneralCfg().DefaultTenant, context)
	atrRFI := NewFilterIndexer(dmMatch, utils.AttributeProfilePrefix, prefix)
	atrRFI.IndexTPFilter(FilterToTPFilter(attribStringF), stringFilterID)
	atrRFI.IndexTPFilter(FilterToTPFilter(attribPrefF), prefixFilterID)
	err = atrRFI.StoreIndexes(true, utils.NonTransactional)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	matchEV = map[string]any{
		utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
		"Field":          "profile",
	}
	aPrflIDs, err := MatchingItemIDsForEvent(matchEV, nil, nil,
		dmMatch, utils.CacheAttributeFilterIndexes, prefix, true, false)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	_, has := aPrflIDs[stringFilterID]
	if !has {
		t.Errorf("Expecting: %+v, received: %+v", stringFilterID, aPrflIDs)
	}
	matchEV = map[string]any{
		"Field": "profilePrefix",
	}
	aPrflIDs, err = MatchingItemIDsForEvent(matchEV, nil, nil,
		dmMatch, utils.CacheAttributeFilterIndexes, prefix, true, false)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	_, has = aPrflIDs[prefixFilterID]
	if !has {
		t.Errorf("Expecting: %+v, received: %+v", prefixFilterID, aPrflIDs)
	}
}

func TestFilterMatchingItemIDsForEvent2(t *testing.T) {
	var stringFilter, prefixFilter, defaultFilter []*FilterRule
	stringFilterID := "stringFilterID"
	prefixFilterID := "prefixFilterID"
	data := NewInternalDB(nil, nil, true, config.CgrConfig().DataDbCfg().Items)
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
	dmMatch.SetFilter(attribStringF)
	x, err = NewFilterRule(utils.MetaPrefix, "~*req.CallCost.Field", []string{"profile"})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	prefixFilter = append(prefixFilter, x)
	attribPrefF := &Filter{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "prefFilter",
		Rules:  prefixFilter}
	dmMatch.SetFilter(attribPrefF)
	x, err = NewFilterRule(utils.MetaGreaterOrEqual, "~*req.Weight", []string{"200.00"})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	defaultFilter = append(defaultFilter, x)
	attribDefaultF := &Filter{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "defaultFilter",
		Rules:  defaultFilter}
	dmMatch.SetFilter(attribDefaultF)
	prefix := utils.ConcatenatedKey(config.CgrConfig().GeneralCfg().DefaultTenant, context)
	atrRFI := NewFilterIndexer(dmMatch, utils.AttributeProfilePrefix, prefix)
	atrRFI.IndexTPFilter(FilterToTPFilter(attribStringF), stringFilterID)
	atrRFI.IndexTPFilter(FilterToTPFilter(attribPrefF), prefixFilterID)
	err = atrRFI.StoreIndexes(true, utils.NonTransactional)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	matchEV = map[string]any{
		utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
		"CallCost":       map[string]any{"Account": 1001},
	}
	aPrflIDs, err := MatchingItemIDsForEvent(matchEV, nil, nil,
		dmMatch, utils.CacheAttributeFilterIndexes, prefix, true, true)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	_, has := aPrflIDs[stringFilterID]
	if !has {
		t.Errorf("Expecting: %+v, received: %+v", stringFilterID, aPrflIDs)
	}
	matchEV = map[string]any{
		"CallCost": map[string]any{"Field": "profilePrefix"},
	}
	aPrflIDs, err = MatchingItemIDsForEvent(matchEV, nil, nil,
		dmMatch, utils.CacheAttributeFilterIndexes, prefix, true, true)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	_, has = aPrflIDs[prefixFilterID]
	if !has {
		t.Errorf("Expecting: %+v, received: %+v", prefixFilterID, aPrflIDs)
	}
}
