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
	"fmt"
	"github.com/cgrates/cgrates/utils"
	"sort"
)

type ActionTrigger struct {
	Id             string // uniquely identify the trigger
	BalanceId      string
	Direction      string
	ThresholdType  string
	ThresholdValue float64
	DestinationId  string
	Weight         float64
	ActionsId      string
	Executed       bool
}

func (at *ActionTrigger) Execute(ub *UserBalance) (err error) {
	// does NOT need to Lock() because it is triggered from a method that took the Lock
	var aac Actions
	aac, err = storageGetter.GetActions(at.ActionsId)
	aac.Sort()
	if err != nil {
		Logger.Err(fmt.Sprintf("Failed to get actions: %v", err))
		return
	}
	for _, a := range aac {
		a.ExpirationDate, _ = utils.ParseDate(a.ExpirationString)
		actionFunction, exists := getActionFunc(a.ActionType)
		if !exists {
			Logger.Warning(fmt.Sprintf("Function type %v not available, aborting execution!", a.ActionType))
			return
		}
		go Logger.Info(fmt.Sprintf("Executing %v: %v", ub.Id, a))
		err = actionFunction(ub, a)
	}
	go storageLogger.LogActionTrigger(ub.Id, RATER_SOURCE, at, aac)
	at.Executed = true
	storageGetter.SetUserBalance(ub)
	return
}

// Structure to store actions according to weight
type ActionTriggerPriotityList []*ActionTrigger

func (atpl ActionTriggerPriotityList) Len() int {
	return len(atpl)
}

func (atpl ActionTriggerPriotityList) Swap(i, j int) {
	atpl[i], atpl[j] = atpl[j], atpl[i]
}

func (atpl ActionTriggerPriotityList) Less(i, j int) bool {
	return atpl[i].Weight < atpl[j].Weight
}

func (atpl ActionTriggerPriotityList) Sort() {
	sort.Sort(atpl)
}
