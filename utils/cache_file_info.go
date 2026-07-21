// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package utils

import "time"

type LoadInstance struct {
	LoadID           string // Unique identifier for the load
	RatingLoadID     string
	AccountingLoadID string
	//TariffPlanID     string    // Tariff plan identificator for the data loaded
	LoadTime time.Time // Time of load
}

// LoadInstancesAsMapStringInterface converts []*LoadInstance struct to map[string]any
func LoadInstancesAsMapStringInterface(loadInstances []*LoadInstance) map[string]any {
	return map[string]any{
		LoadHistory: loadInstances,
	}
}

// MapStringInterfaceToLoadInstances converts map[string]any to *[]LoadInstance struct
func MapStringInterfaceToLoadInstances(m map[string]any) []*LoadInstance {
	lh, ok := m[LoadHistory]
	if !ok {
		return nil
	}
	items, ok := lh.([]any)
	if !ok {
		return nil
	}
	loadInstances := make([]*LoadInstance, 0, len(items))
	for _, item := range items {
		itemMap, ok := item.(map[string]interface{})
		if !ok {
			return nil
		}
		loadInst := &LoadInstance{}
		if v, ok := itemMap[LoadID].(string); ok {
			loadInst.LoadID = v
		}
		if v, ok := itemMap[RatingLoadID].(string); ok {
			loadInst.RatingLoadID = v
		}
		if v, ok := itemMap[AccountingLoadID].(string); ok {
			loadInst.AccountingLoadID = v
		}
		if v, ok := itemMap[LoadTime].(string); ok {
			if t, err := time.Parse(time.RFC3339, v); err == nil {
				loadInst.LoadTime = t
			}
		}
		loadInstances = append(loadInstances, loadInst)
	}
	return loadInstances
}
