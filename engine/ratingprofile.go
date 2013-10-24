/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2013 ITsysCOM

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
	"errors"
	"fmt"
	"github.com/cgrates/cgrates/cache2go"
	"sort"
	"time"
)

type RatingProfile struct {
	Id                                                       string
	FallbackKey                                              string // FallbackKey is used as complete combination of Tenant:TOR:Direction:Subject
	RatingPlanActivations                                    RatingPlanActivations
	Tag, Tenant, TOR, Direction, Subject                     string // used only for loading
	DestRatesTimingTag, RatesFallbackSubject, ActivationTime string // used only for loading
}

type RatingPlanActivation struct {
	ActivationTime time.Time
	RatingPlanId   string
}

func (rpa *RatingPlanActivation) GetRatingPlan() (rp *RatingPlan, err error) {
	if x, err := cache2go.GetCached(rpa.RatingPlanId); err != nil {
		rp, err = storageGetter.GetRatingPlan(rpa.RatingPlanId)
		if err == nil && rp != nil {
			cache2go.Cache(rpa.RatingPlanId, rp)
		}
	} else {
		rp = x.(*RatingPlan)
	}
	return
}

func (rpa *RatingPlanActivation) Equal(orpa *RatingPlanActivation) bool {
	return rpa.ActivationTime == orpa.ActivationTime && rpa.RatingPlanId == orpa.RatingPlanId
}

type RatingPlanActivations []*RatingPlanActivation

func (rpas RatingPlanActivations) Len() int {
	return len(rpas)
}

func (rpas RatingPlanActivations) Swap(i, j int) {
	rpas[i], rpas[j] = rpas[j], rpas[i]
}

func (rpas RatingPlanActivations) Less(i, j int) bool {
	return rpas[i].ActivationTime.Before(rpas[j].ActivationTime)
}

func (rpas RatingPlanActivations) Sort() {
	sort.Sort(rpas)
}

type RatingInfo struct {
	ActivationTime time.Time
	RateIntervals  RateIntervalList
}

func (rp *RatingProfile) GetRatingPlansForPrefix(cd *CallDescriptor) (foundPrefixes []string, ris []*RatingInfo, err error) {
	rp.RatingPlanActivations.Sort()
	for _, rpa := range rp.RatingPlanActivations {
		if rpa.ActivationTime.Before(cd.TimeEnd) {
			rpl, err := rpa.GetRatingPlan()
			if err != nil || rpl == nil {
				Logger.Err(fmt.Sprintf("Error checking destination: %v", err))
				continue
			}
			bestPrecision := 0
			var rps RateIntervalList
			for dId, _ := range rpl.DestinationRates {
				precision, err := storageGetter.DestinationContainsPrefix(dId, cd.Destination)
				if err != nil {
					Logger.Err(fmt.Sprintf("Error checking destination: %v", err))
					continue
				}
				if precision > bestPrecision {
					bestPrecision = precision
					rps = rpl.RateIntervalList(dId)
				}
			}
			if bestPrecision > 0 {
				ris = append(ris, &RatingInfo{rpa.ActivationTime, rps})
				foundPrefixes = append(foundPrefixes, cd.Destination[:bestPrecision])
			}
		}
	}
	if len(ris) > 0 {
		return foundPrefixes, ris, nil
	}

	return nil, nil, errors.New("not found")
}
