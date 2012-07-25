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
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	FORMAT     = "2006-1-2 15:04:05 MST"
	ASAP       = "*asap"
	ASAP_DELAY = "5m"
)

type ActionTiming struct {
	Id             string // identify the timing
	Tag            string // informative purpos only
	UserBalanceIds []string
	Timing         *Interval
	Weight         float64
	ActionsId      string
	actions        ActionPriotityList
}

func (at *ActionTiming) GetNextStartTime() (t time.Time) {
	i := at.Timing
	if i == nil {
		return
	}
	now := time.Now()
	y, m, d := now.Date()
	z, _ := now.Zone()
	if i.StartTime != "" {
		l := fmt.Sprintf("%d-%d-%d %s %s", y, m, d, i.StartTime, z)
		var err error
		t, err = time.Parse(FORMAT, l)
		if err != nil {
			log.Printf("Cannot parse action timing's StartTime %v", l)
			return
		}
	}
	// weekdays
	if i.WeekDays != nil && len(i.WeekDays) > 0 {
		i.WeekDays.Sort()
		if t.IsZero() {
			t = time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second(), 0, now.Location())
		}
		d := t.Day()
		for _, j := range []int{0, 1, 2, 3, 4, 5, 6, 7} {
			t = time.Date(t.Year(), t.Month(), d+j, t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), t.Location())
			for _, wd := range i.WeekDays {
				if t.Weekday() == wd && (t.Equal(now) || t.After(now)) {
					return
				}
			}
		}
	}
	// monthdays
	if i.MonthDays != nil && len(i.MonthDays) > 0 {
		i.MonthDays.Sort()
		now := time.Now()
		x := sort.SearchInts(i.MonthDays, now.Day())
		d = i.MonthDays[0]
		if x < len(i.MonthDays) {
			if i.MonthDays[x] == now.Day() {
				if t.Equal(now) || t.After(now) {
					h, m, s := t.Clock()
					t = time.Date(now.Year(), now.Month(), now.Day(), h, m, s, 0, time.Local)
					goto MONTHS
				}
				if x+1 < len(i.MonthDays) { // today was found in the list, jump to the next grater day
					d = i.MonthDays[x+1]
				}
			} else { // today was not found in the list, x is the first greater day
				d = i.MonthDays[x]
			}
		}
		h, m, s := t.Clock()
		t = time.Date(now.Year(), now.Month(), d, h, m, s, 0, time.Local)
	}
MONTHS:
	if i.Months != nil && len(i.Months) > 0 {
		i.Months.Sort()
		now := time.Now()
		x := sort.Search(len(i.Months), func(x int) bool { return i.Months[x] >= now.Month() })
		m = i.Months[0]
		if x < len(i.Months) {
			if i.Months[x] == now.Month() {
				if t.Equal(now) || t.After(now) {
					h, m, s := t.Clock()
					t = time.Date(now.Year(), now.Month(), t.Day(), h, m, s, 0, time.Local)
					return
				}
				if x+1 < len(i.Months) { // today was found in the list, jump to the next grater day
					m = i.Months[x+1]
					// reset the monthday
					if i.MonthDays != nil {
						t = time.Date(t.Year(), t.Month(), i.MonthDays[0], t.Hour(), t.Minute(), t.Second(), 0, t.Location())
					}
				}
			} else { // today was not found in the list, x is the first greater day
				m = i.Months[x]
				// reset the monthday
				if i.MonthDays != nil {
					t = time.Date(t.Year(), t.Month(), i.MonthDays[0], t.Hour(), t.Minute(), t.Second(), 0, t.Location())
				}
			}
		}
		h, min, s := t.Clock()
		t = time.Date(now.Year(), m, t.Day(), h, min, s, 0, time.Local)
	}
	return
}

func (at *ActionTiming) getActions() (as []*Action, err error) {
	if at.actions == nil {
		at.actions, err = storageGetter.GetActions(at.ActionsId)
	}
	at.actions.Sort()
	return at.actions, err
}

func (at *ActionTiming) getUserBalances() (ubs []*UserBalance) {
	for _, ubId := range at.UserBalanceIds {
		ub, err := storageGetter.GetUserBalance(ubId)
		if err != nil {
			log.Printf("Could not get user balances for therse id: %v. Skipping!", ubId)
		}
		ubs = append(ubs, ub)
	}
	return
}

func (at *ActionTiming) Execute() (err error) {
	aac, err := at.getActions()
	if err != nil {
		log.Print("Failed to get actions: ", err)
		return
	}
	for _, a := range aac {
		actionFunction, exists := actionTypeFuncMap[a.ActionType]
		if !exists {
			log.Printf("Function type %v not available, aborting execution!", a.ActionType)
			return
		}
		for _, ub := range at.getUserBalances() {
			AccLock.Guard(ub.Id, func() (float64, error) {
				err = actionFunction(ub, a)
				storageGetter.SetUserBalance(ub)
				return 0, nil
			})
		}
	}
	return
}

func (at *ActionTiming) CheckForASAP() bool {
	if at.Timing.StartTime == ASAP {
		oneMinute, _ := time.ParseDuration(ASAP_DELAY)
		timeTokens := strings.Split(time.Now().Add(oneMinute).Format(time.Stamp), " ")
		at.Timing.StartTime = timeTokens[len(timeTokens)-1]
		return true
	}
	return false
}

// Structure to store actions according to weight
type ActionTimingPriotityList []*ActionTiming

func (atpl ActionTimingPriotityList) Len() int {
	return len(atpl)
}

func (atpl ActionTimingPriotityList) Swap(i, j int) {
	atpl[i], atpl[j] = atpl[j], atpl[i]
}

func (atpl ActionTimingPriotityList) Less(i, j int) bool {
	if atpl[i].GetNextStartTime().Equal(atpl[j].GetNextStartTime()) {
		return atpl[i].Weight < atpl[j].Weight
	}
	return atpl[i].GetNextStartTime().Before(atpl[j].GetNextStartTime())
}

func (atpl ActionTimingPriotityList) Sort() {
	sort.Sort(atpl)
}

func (at *ActionTiming) String() string {
	return at.Tag + " " + at.GetNextStartTime().String() + ",w: " + strconv.FormatFloat(at.Weight, 'f', -1, 64)
}

/*
Serializes the action timing for the storage. Used for key-value storages.
*/
func (at *ActionTiming) store() (result string) {
	result += at.Id + "|"
	result += at.Tag + "|"
	for _, ubi := range at.UserBalanceIds {
		result += ubi + ","
	}
	result = strings.TrimRight(result, ",") + "|"
	result += at.Timing.store() + "|"
	result += strconv.FormatFloat(at.Weight, 'f', -1, 64) + "|"
	result += at.ActionsId
	return
}

/*
De-serializes the action timing for the storage. Used for key-value storages.
*/
func (at *ActionTiming) restore(input string) {
	elements := strings.Split(input, "|")
	at.Id = elements[0]
	at.Tag = elements[1]
	for _, ubi := range strings.Split(elements[2], ",") {
		at.UserBalanceIds = append(at.UserBalanceIds, ubi)
	}

	at.Timing = &Interval{}
	at.Timing.restore(elements[3])
	at.Weight, _ = strconv.ParseFloat(elements[4], 64)
	at.ActionsId = elements[5]
}

// helper function for uuid generation
func GenUUID() string {
	uuid := make([]byte, 16)
	n, err := rand.Read(uuid)
	if n != len(uuid) || err != nil {
		return strconv.FormatInt(time.Now().UnixNano(), 10)
	}
	// TODO: verify the two lines implement RFC 4122 correctly
	uuid[8] = 0x80 // variant bits see page 5
	uuid[4] = 0x40 // version 4 Pseudo Random, see page 7

	return hex.EncodeToString(uuid)
}
