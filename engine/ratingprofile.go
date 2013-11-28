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
	"sort"
	"time"
)

type RatingProfile struct {
	Id                    string
	RatingPlanActivations RatingPlanActivations
}

type RatingPlanActivation struct {
	ActivationTime time.Time
	RatingPlanId   string
	FallbackKeys   []string
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

func (rpas RatingPlanActivations) GetActiveForCall(cd *CallDescriptor) RatingPlanActivations {
	rpas.Sort()
	lastBeforeCallStart := 0
	firstAfterCallEnd := len(rpas)
	for index, rpa := range rpas {
		if rpa.ActivationTime.Before(cd.TimeStart) || rpa.ActivationTime.Equal(cd.TimeStart) {
			lastBeforeCallStart = index
		}
		if rpa.ActivationTime.After(cd.TimeEnd) {
			firstAfterCallEnd = index
			break
		}
	}
	return rpas[lastBeforeCallStart:firstAfterCallEnd]
}

type RatingInfo struct {
	MatchedSubject string
	MatchedPrefix  string
	ActivationTime time.Time
	RateIntervals  RateIntervalList
	FallbackKeys   []string
}

type RatingInfos []*RatingInfo

func (ris RatingInfos) Len() int {
	return len(ris)
}

func (ris RatingInfos) Swap(i, j int) {
	ris[i], ris[j] = ris[j], ris[i]
}

func (ris RatingInfos) Less(i, j int) bool {
	return ris[i].ActivationTime.Before(ris[j].ActivationTime)
}

func (ris RatingInfos) Sort() {
	sort.Sort(ris)
}

func (rp *RatingProfile) GetRatingPlansForPrefix(cd *CallDescriptor) (err error) {
	var ris RatingInfos
	for index, rpa := range rp.RatingPlanActivations.GetActiveForCall(cd) {
		rpl, err := storageGetter.GetRatingPlan(rpa.RatingPlanId, false)
		if err != nil || rpl == nil {
			Logger.Err(fmt.Sprintf("Error checking destination: %v", err))
			continue
		}
		bestPrecision := 0
		var rps RateIntervalList
		for dId, _ := range rpl.DestinationRates {
			//precision, err := storageGetter.DestinationContainsPrefix(dId, cd.Destination)
			d, err := storageGetter.GetDestination(dId, false)
			if err != nil {
				Logger.Err(fmt.Sprintf("Error checking destination: %v", err))
				continue
			}
			precision := d.containsPrefix(cd.Destination)
			if precision > bestPrecision {
				bestPrecision = precision
				rps = rpl.RateIntervalList(dId)
			}
		}
		// check if it's the first ri and add a blank one for the initial part not covered
		if index == 0 && cd.TimeStart.Before(rpa.ActivationTime) {
			ris = append(ris, &RatingInfo{"", "", cd.TimeStart, nil, []string{cd.GetKey(FALLBACK_SUBJECT)}})
		}
		if bestPrecision > 0 {
			ris = append(ris, &RatingInfo{rp.Id, cd.Destination[:bestPrecision], rpa.ActivationTime, rps, rpa.FallbackKeys})
		} else {
			// add for fallback information
			ris = append(ris, &RatingInfo{"", "", rpa.ActivationTime, nil, rpa.FallbackKeys})
		}
	}
	if len(ris) > 0 {
		cd.addRatingInfos(ris)
		return
	}

	return errors.New("not found")
}
