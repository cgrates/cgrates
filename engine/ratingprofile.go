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
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/cgrates/cgrates/utils"
)

type RatingProfile struct {
	Id                    string
	RatingPlanActivations RatingPlanActivations
}

type RatingPlanActivation struct {
	ActivationTime  time.Time
	RatingPlanId    string
	FallbackKeys    []string
	CdrStatQueueIds []string
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
	RatingPlanId   string
	MatchedPrefix  string
	MatchedDestId  string
	ActivationTime time.Time
	RateIntervals  RateIntervalList
	FallbackKeys   []string
}

// SelectRatingIntevalsForTimespan orders rate intervals in time preserving only those which aply to the specified timestamp
func (ri RatingInfo) SelectRatingIntevalsForTimespan(ts *TimeSpan) (result RateIntervalList) {
	sorter := &RateIntervalTimeSorter{referenceTime: ts.TimeStart, ris: ri.RateIntervals}
	rateIntervals := sorter.Sort()
	// get the rating interval closest to begining of timespan
	var delta time.Duration = -1
	var bestRateIntervalIndex int
	var bestIntervalWeight float64
	for index, rateInterval := range rateIntervals {
		if !rateInterval.Contains(ts.TimeStart, false) {
			continue
		}
		if rateInterval.Weight < bestIntervalWeight {
			break // don't consider lower weights'
		}
		startTime := rateInterval.Timing.getLeftMargin(ts.TimeStart)
		tmpDelta := ts.TimeStart.Sub(startTime)
		if (startTime.Before(ts.TimeStart) ||
			startTime.Equal(ts.TimeStart)) &&
			(delta == -1 || tmpDelta < delta) {
			bestRateIntervalIndex = index
			bestIntervalWeight = rateInterval.Weight
			delta = tmpDelta
		}
	}
	result = append(result, rateIntervals[bestRateIntervalIndex])
	// check if later rating intervals influence this timespan
	//log.Print("RIS: ", utils.ToIJSON(rateIntervals))
	for i := bestRateIntervalIndex + 1; i < len(rateIntervals); i++ {
		if rateIntervals[i].Weight < bestIntervalWeight {
			break // don't consider lower weights'
		}
		startTime := rateIntervals[i].Timing.getLeftMargin(ts.TimeStart)
		if startTime.Before(ts.TimeEnd) {
			result = append(result, rateIntervals[i])
		}
	}
	return
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

func (ris RatingInfos) String() string {
	b, _ := json.MarshalIndent(ris, "", " ")
	return string(b)
}

func (rpf *RatingProfile) GetRatingPlansForPrefix(cd *CallDescriptor) (err error) {
	var ris RatingInfos
	for index, rpa := range rpf.RatingPlanActivations.GetActiveForCall(cd) {
		rpl, err := dm.GetRatingPlan(rpa.RatingPlanId, false, utils.NonTransactional)
		if err != nil || rpl == nil {
			utils.Logger.Err(fmt.Sprintf("Error checking destination: %v", err))
			continue
		}
		prefix := ""
		destinationId := ""
		var rps RateIntervalList
		if cd.Destination == utils.ANY || cd.Destination == "" {
			cd.Destination = utils.ANY
			if _, ok := rpl.DestinationRates[utils.ANY]; ok {
				rps = rpl.RateIntervalList(utils.ANY)
				prefix = utils.ANY
				destinationId = utils.ANY
			}
		} else {
			for _, p := range utils.SplitPrefix(cd.Destination, MIN_PREFIX_MATCH) {
				if destIDs, err := dm.DataDB().GetReverseDestination(p, false, utils.NonTransactional); err == nil {
					var bestWeight *float64
					for _, dID := range destIDs {
						if _, ok := rpl.DestinationRates[dID]; ok {
							ril := rpl.RateIntervalList(dID)
							currentWeight := ril.GetWeight()
							if bestWeight == nil || currentWeight > *bestWeight {
								bestWeight = utils.Float64Pointer(currentWeight)
								rps = ril
								prefix = p
								destinationId = dID
							}
						}
					}
				}
				if rps != nil {
					break
				}
			}
			if rps == nil { // fallback on *any destination
				if _, ok := rpl.DestinationRates[utils.ANY]; ok {
					rps = rpl.RateIntervalList(utils.ANY)
					prefix = utils.ANY
					destinationId = utils.ANY
				}
			}
		}
		// check if it's the first ri and add a blank one for the initial part not covered
		if index == 0 && cd.TimeStart.Before(rpa.ActivationTime) {
			ris = append(ris, &RatingInfo{
				MatchedSubject: "",
				MatchedPrefix:  "",
				MatchedDestId:  "",
				ActivationTime: cd.TimeStart,
				RateIntervals:  nil,
				FallbackKeys:   []string{cd.GetKey(FALLBACK_SUBJECT)}})
		}
		if len(prefix) > 0 {
			ris = append(ris, &RatingInfo{
				MatchedSubject: rpf.Id,
				RatingPlanId:   rpl.Id,
				MatchedPrefix:  prefix,
				MatchedDestId:  destinationId,
				ActivationTime: rpa.ActivationTime,
				RateIntervals:  rps,
				FallbackKeys:   rpa.FallbackKeys})
		} else {
			// add for fallback information
			if len(rpa.FallbackKeys) > 0 {
				ris = append(ris, &RatingInfo{
					MatchedSubject: "",
					MatchedPrefix:  "",
					MatchedDestId:  "",
					ActivationTime: rpa.ActivationTime,
					RateIntervals:  nil,
					FallbackKeys:   rpa.FallbackKeys,
				})
			}
		}
	}
	if len(ris) > 0 {
		cd.addRatingInfos(ris)
		return
	}
	return utils.ErrNotFound
}

type TenantRatingSubject struct {
	Tenant, Subject string
}

func RatingProfileSubjectPrefixMatching(key string) (rp *RatingProfile, err error) {
	if !rpSubjectPrefixMatching || strings.HasSuffix(key, utils.ANY) {
		return dm.GetRatingProfile(key, false, utils.NonTransactional)
	}
	if rp, err = dm.GetRatingProfile(key, false, utils.NonTransactional); err == nil && rp != nil { // rp nil represents cached no-result
		return
	}
	lastIndex := strings.LastIndex(key, utils.CONCATENATED_KEY_SEP)
	baseKey := key[:lastIndex]
	subject := key[lastIndex:]
	lenSubject := len(subject)
	for i := 1; i < lenSubject-1; i++ {
		if rp, err = dm.GetRatingProfile(baseKey+subject[:lenSubject-i], false, utils.NonTransactional); err == nil && rp != nil {
			return
		}
	}
	return
}
