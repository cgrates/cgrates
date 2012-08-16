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

package timespans

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

type ActionTrigger struct {
	Id             string // uniquely identify the timing
	BalanceId      string
	Direction      string
	ThresholdValue float64
	DestinationId  string
	Weight         float64
	ActionsId      string
	Executed       bool
}

func (at *ActionTrigger) Execute(ub *UserBalance) (err error) {
	// does NOT need to Lock() because it is triggered from a method that took the Lock
	var aac ActionPriotityList
	aac, err = storageGetter.GetActions(at.ActionsId)
	aac.Sort()
	if err != nil {
		Logger.Err(fmt.Sprintf("Failed to get actions: ", err))
		return
	}
	for _, a := range aac {
		actionFunction, exists := actionTypeFuncMap[a.ActionType]
		if !exists {
			Logger.Warning(fmt.Sprintf("Function type %v not available, aborting execution!", a.ActionType))
			return
		}
		go Logger.Info(fmt.Sprintf("Executing %v: %v", ub.Id, a))
		err = actionFunction(ub, a)
	}
	go storageLogger.LogActionTrigger(ub.Id, at, aac)
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

/*
Serializes the action trigger for the storage. Used for key-value storages.
*/
func (at *ActionTrigger) store() (result string) {
	result += at.Id + ";"
	result += at.BalanceId + ";"
	result += at.Direction + ";"
	result += at.DestinationId + ";"
	result += at.ActionsId + ";"
	result += strconv.FormatFloat(at.ThresholdValue, 'f', -1, 64) + ";"
	result += strconv.FormatFloat(at.Weight, 'f', -1, 64) + ";"
	result += strconv.FormatBool(at.Executed)
	return
}

/*
De-serializes the action timing for the storage. Used for key-value storages.
*/
func (at *ActionTrigger) restore(input string) {
	elements := strings.Split(input, ";")
	if len(elements) != 8 {
		return
	}
	at.Id = elements[0]
	at.BalanceId = elements[1]
	at.Direction = elements[2]
	at.DestinationId = elements[3]
	at.ActionsId = elements[4]
	at.ThresholdValue, _ = strconv.ParseFloat(elements[5], 64)
	at.Weight, _ = strconv.ParseFloat(elements[6], 64)
	at.Executed, _ = strconv.ParseBool(elements[7])
}
