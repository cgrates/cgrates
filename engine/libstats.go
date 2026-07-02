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
	"encoding/base64"
	"fmt"
	"time"

	"github.com/cgrates/cgrates/utils"
)

// NewStoredStatQueue initiates a StoredStatQueue out of StatQueue
func NewStoredStatQueue(sq *utils.StatQueue, ms utils.Marshaler, limit int) (sSQ *StoredStatQueue, err error) {
	sSQ = &StoredStatQueue{
		Tenant:     sq.Tenant,
		ID:         sq.ID,
		Compressed: sq.Compress(uint64(limit)),
		SQItems:    make([]utils.SQItem, len(sq.SQItems)),
		SQMetrics:  make(map[string][]byte, len(sq.SQMetrics)),
	}
	copy(sSQ.SQItems, sq.SQItems)
	for metricID, metric := range sq.SQMetrics {
		marshaled, err := ms.Marshal(metric)
		if err != nil {
			return nil, err
		}
		sSQ.SQMetrics[metricID] = marshaled
	}
	return sSQ, nil
}

// StoredStatQueue differs from StatQueue due to serialization of SQMetrics
type StoredStatQueue struct {
	Tenant     string
	ID         string
	SQItems    []utils.SQItem
	SQMetrics  map[string][]byte
	Compressed bool
}

// SqID will compose the unique identifier for the StatQueue out of Tenant and ID
func (ssq *StoredStatQueue) SqID() string {
	return utils.ConcatenatedKey(ssq.Tenant, ssq.ID)
}

// AsStatQueue converts into StatQueue unmarshaling SQMetrics
func (ssq *StoredStatQueue) AsStatQueue(ms utils.Marshaler) (sq *utils.StatQueue, err error) {
	if ssq == nil {
		return
	}
	sq = &utils.StatQueue{
		Tenant:    ssq.Tenant,
		ID:        ssq.ID,
		SQItems:   make([]utils.SQItem, len(ssq.SQItems)),
		SQMetrics: make(map[string]utils.StatMetric, len(ssq.SQMetrics)),
	}
	copy(sq.SQItems, ssq.SQItems)
	for metricID, marshaled := range ssq.SQMetrics {
		metric, err := utils.NewStatMetric(metricID, 0, []string{})
		if err != nil {
			return nil, err
		}
		if err := ms.Unmarshal(marshaled, metric); err != nil {
			return nil, err
		}
		sq.SQMetrics[metricID] = metric
	}
	if ssq.Compressed {
		sq.Expand()
	}
	return
}

// AsMapStringInterface converts StoredStatQueue struct to map[string]any
func (ssq *StoredStatQueue) AsMapStringInterface() map[string]any {
	if ssq == nil {
		return nil
	}
	return map[string]any{
		utils.Tenant:     ssq.Tenant,
		utils.ID:         ssq.ID,
		utils.SQItems:    ssq.SQItems,
		utils.SQMetrics:  ssq.SQMetrics,
		utils.Compressed: ssq.Compressed,
	}
}

// MapStringInterfaceToStoredStatQueue converts map[string]any to StoredStatQueue struct
func MapStringInterfaceToStoredStatQueue(m map[string]any) (*StoredStatQueue, error) {
	ssq := &StoredStatQueue{}
	if v, ok := m[utils.Tenant].(string); ok {
		ssq.Tenant = v
	}
	if v, ok := m[utils.ID].(string); ok {
		ssq.ID = v
	}
	if items, ok := m[utils.SQItems].([]any); ok {
		for _, item := range items {
			if itemMap, ok := item.(map[string]any); ok {
				sqItem := utils.SQItem{}
				if eventID, ok := itemMap[utils.EventID].(string); ok {
					sqItem.EventID = eventID
				}
				if expiryTime, ok := itemMap[utils.ExpiryTime].(*time.Time); ok {
					sqItem.ExpiryTime = expiryTime
				} else if expiryStr, ok := itemMap[utils.ExpiryTime].(string); ok {
					if parsedTime, err := time.Parse(time.RFC3339, expiryStr); err == nil {
						sqItem.ExpiryTime = &parsedTime
					} else {
						return nil, err
					}
				}
				ssq.SQItems = append(ssq.SQItems, sqItem)
			}
		}
	}
	if metrics, ok := m[utils.SQMetrics].(map[string]any); ok {
		ssq.SQMetrics = make(map[string][]byte)
		for key, value := range metrics {
			if metricBytes, ok := value.(string); ok {
				decodedBytes, err := base64.StdEncoding.DecodeString(metricBytes)
				if err != nil {
					return nil, fmt.Errorf("failed to decode base64 string: %v", err)
				}
				ssq.SQMetrics[key] = decodedBytes
			}
		}
	}
	if v, ok := m[utils.Compressed].(bool); ok {
		ssq.Compressed = v
	}
	return ssq, nil
}
