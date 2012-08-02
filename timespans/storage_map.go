/*
Rating system designed to be used in VoIP Carriems World
Copyright (C) 2012  Radu Ioan Fericean

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either vemsion 3 of the License, or
(at your option) any later vemsion.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>
*/

package timespans

import (
	"strings"
)

type MapStorage struct {
	dict map[string][]byte
	ms   Marshaler
}

func NewMapStorage() (*MapStorage, error) {
	return &MapStorage{dict: make(map[string][]byte), ms: new(MyMarshaler)}, nil
}

func (ms *MapStorage) Close() {}

func (ms *MapStorage) Flush() error {
	ms.dict = make(map[string][]byte)
	return nil
}

func (ms *MapStorage) GetActivationPeriodsOrFallback(key string) (aps []*ActivationPeriod, fallbackKey string, err error) {
	elem, ok := ms.dict[key]
	if !ok {
		return
	}
	err = ms.ms.Unmarshal(elem, &aps)
	if err != nil {
		err = ms.ms.Unmarshal(elem, &fallbackKey)
	}
	return
}

func (ms *MapStorage) SetActivationPeriodsOrFallback(key string, aps []*ActivationPeriod, fallbackKey string) (err error) {
	var result []byte
	if len(aps) > 0 {
		result, err = ms.ms.Marshal(aps)
	} else {
		result, err = ms.ms.Marshal(fallbackKey)
	}
	ms.dict[key] = result
	return
}

func (ms *MapStorage) GetDestination(key string) (dest *Destination, err error) {
	if values, ok := ms.dict[key]; ok {
		dest = &Destination{Id: key}
		err = ms.ms.Unmarshal(values, dest)
	}
	return
}
func (ms *MapStorage) SetDestination(dest *Destination) (err error) {
	result, err := ms.ms.Marshal(dest)
	ms.dict[dest.Id] = result
	return
}

func (ms *MapStorage) GetActions(key string) (as []*Action, err error) {
	if values, ok := ms.dict[key]; ok {
		err = ms.ms.Unmarshal(values, &as)
	}
	return
}

func (ms *MapStorage) SetActions(key string, as []*Action) (err error) {
	result, err := ms.ms.Marshal(as)
	ms.dict[key] = result
	return
}

func (ms *MapStorage) GetUserBalance(key string) (ub *UserBalance, err error) {
	if values, ok := ms.dict[key]; ok {
		ub = &UserBalance{Id: key}
		err = ms.ms.Unmarshal(values, ub)
	}
	return
}

func (ms *MapStorage) SetUserBalance(ub *UserBalance) (err error) {
	result, err := ms.ms.Marshal(ub)
	ms.dict[ub.Id] = result
	return
}

func (ms *MapStorage) GetActionTimings(key string) (ats []*ActionTiming, err error) {
	if values, ok := ms.dict[key]; ok {
		err = ms.ms.Unmarshal(values, &ats)
	}
	return
}

func (ms *MapStorage) SetActionTimings(key string, ats []*ActionTiming) (err error) {
	result, err := ms.ms.Marshal(ats)
	ms.dict[key] = result
	return
}

func (ms *MapStorage) GetAllActionTimings() (ats map[string][]*ActionTiming, err error) {
	ats = make(map[string][]*ActionTiming)
	for key, value := range ms.dict {
		if !strings.Contains(key, ACTION_TIMING_PREFIX) {
			continue
		}
		var tempAts []*ActionTiming
		err = ms.ms.Unmarshal(value, &tempAts)
		ats[key] = tempAts
	}

	return
}
