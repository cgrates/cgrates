/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2012  Radu Ioan Fericean

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

package rater

import (
	"errors"
	"fmt"
	"strings"
)

const (
	// the minimum length for a destination prefix to be matched.
	MIN_PREFIX_LENGTH = 2
)

type RatingProfile struct {
	Id             string
	FallbackKey    string
	DestinationMap map[string][]*ActivationPeriod
}

func (rp *RatingProfile) store() (result string) {
	result += rp.FallbackKey + ">"
	for k, aps := range rp.DestinationMap {
		result += k + "="
		for _, ap := range aps {
			result += ap.store() + "<"
		}
		result = strings.TrimRight(result, "<")
		result += ">"
	}
	result = strings.TrimRight(result, ">")
	return
}

func (rp *RatingProfile) restore(input string) {
	if rp.DestinationMap == nil {
		rp.DestinationMap = make(map[string][]*ActivationPeriod, 1)
	}
	elements := strings.Split(input, ">")
	rp.FallbackKey = elements[0]
	for _, kv := range elements[1:] {
		pair := strings.SplitN(kv, "=", 2)
		apList := strings.Split(pair[1], "<")
		var newAps []*ActivationPeriod
		for _, aps := range apList {
			ap := new(ActivationPeriod)
			ap.restore(aps)
			newAps = append(newAps, ap)
		}
		rp.DestinationMap[pair[0]] = newAps
	}
}

// Adds an activation period that applyes to current rating profile if not already present.
func (rp *RatingProfile) AddActivationPeriodIfNotPresent(destInfo string, aps ...*ActivationPeriod) {
	if rp.DestinationMap == nil {
		rp.DestinationMap = make(map[string][]*ActivationPeriod, 1)
	}
	for _, ap := range aps {
		found := false
		for _, eap := range rp.DestinationMap[destInfo] {
			if ap.Equal(eap) {
				found = true
				break
			}
		}
		if !found {
			rp.DestinationMap[destInfo] = append(rp.DestinationMap[destInfo], ap)
		}
	}
}

func (rp *RatingProfile) GetActivationPeriodsForPrefix(destPrefix string) (foundPrefix string, aps []*ActivationPeriod, err error) {
	bestPrecision := 0
	for k, v := range rp.DestinationMap {
		d, err := GetDestination(k)
		if err != nil {
			Logger.Err(fmt.Sprintf("Cannot find destination with id: %s", k))
			continue
		}
		if precision, ok := d.containsPrefix(destPrefix); ok && precision > bestPrecision {
			bestPrecision = precision
			aps = v
		}
	}

	if bestPrecision > 0 {
		return destPrefix[:bestPrecision], aps, nil
	}

	return "", nil, errors.New("not found")
}
