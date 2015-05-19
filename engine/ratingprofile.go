/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2012-2015 ITsysCOM

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
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/cgrates/cgrates/cache2go"
	"github.com/cgrates/cgrates/history"
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

func (rp *RatingProfile) GetRatingPlansForPrefix(cd *CallDescriptor) (err error) {
	var ris RatingInfos
	for index, rpa := range rp.RatingPlanActivations.GetActiveForCall(cd) {
		rpl, err := dataStorage.GetRatingPlan(rpa.RatingPlanId, false)
		if err != nil || rpl == nil {
			Logger.Err(fmt.Sprintf("Error checking destination: %v", err))
			continue
		}
		prefix := ""
		destinationId := ""
		var rps RateIntervalList
		//log.Printf("RPA: %+v", rpa)
		if cd.Destination == utils.ANY || cd.Destination == "" {
			cd.Destination = utils.ANY
			if _, ok := rpl.DestinationRates[utils.ANY]; ok {
				rps = rpl.RateIntervalList(utils.ANY)
				prefix = utils.ANY
				destinationId = utils.ANY
			}
		} else {
			for _, p := range utils.SplitPrefix(cd.Destination, MIN_PREFIX_MATCH) {
				if x, err := cache2go.GetCached(DESTINATION_PREFIX + p); err == nil {

					destIds := x.(map[interface{}]struct{})
					for idId := range destIds {
						dId := idId.(string)
						if _, ok := rpl.DestinationRates[dId]; ok {
							rps = rpl.RateIntervalList(dId)
							prefix = p
							destinationId = dId
							break
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
				MatchedSubject: rp.Id,
				RatingPlanId:   rpl.Id,
				MatchedPrefix:  prefix,
				MatchedDestId:  destinationId,
				ActivationTime: rpa.ActivationTime,
				RateIntervals:  rps,
				FallbackKeys:   rpa.FallbackKeys})
		} else {
			// add for fallback information
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

	if len(ris) > 0 {
		cd.addRatingInfos(ris)
		return
	}

	return errors.New(utils.ERR_NOT_FOUND)
}

// history record method
func (rpf *RatingProfile) GetHistoryRecord() history.Record {
	js, _ := json.Marshal(rpf)
	return history.Record{
		Id:       rpf.Id,
		Filename: history.RATING_PROFILES_FN,
		Payload:  js,
	}
}

type TenantRatingSubject struct {
	Tenant, Subject string
}
