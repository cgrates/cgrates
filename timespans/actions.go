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
	"time"
	"log"
	"fmt"
	"sort"
)

const (
	FORMAT = "2006-1-2 15:04:05 MST"
)

// Amount of a trafic of a certain type (TOR)
type UnitsCounter struct {
	Direction     string
	TOR           string
	Units         float64
	Weight        float64
	DestinationId string
	destination   *Destination
}

// Structure to store actions according to weight
type countersorter []*UnitsCounter

func (s countersorter) Len() int {
	return len(s)
}

func (s countersorter) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s countersorter) Less(j, i int) bool {
	return s[i].Weight < s[j].Weight
}

/*
Returns the destination loading it from the storage if necessary.
*/
func (uc *UnitsCounter) getDestination() (dest *Destination) {
	if uc.destination == nil {
		uc.destination, _ = storageGetter.GetDestination(uc.DestinationId)
	}
	return uc.destination
}

/*
Structure to be filled for each tariff plan with the bonus value for received calls minutes.
*/
type Action struct {
	ActionType   string
	BalanceId    string
	Units        float64
	MinuteBucket *MinuteBucket
}

type actionTypeFunc func(a *Action) error

var (
	actionTypeFuncMap = map[string]actionTypeFunc{
		"LOG": logAction,
	}
)

func logAction(a *Action) (err error) {
	log.Printf("%v %v %v", a.BalanceId, a.Units, a.MinuteBucket)
	return
}

// Structure to store actions according to weight
type actionsorter []*Action

func (s actionsorter) Len() int {
	return len(s)
}

func (s actionsorter) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s actionsorter) Less(j, i int) bool {
	return s[i].MinuteBucket.Weight < s[j].MinuteBucket.Weight
}

type ActionTrigger struct {
	BalanceId      string
	ThresholdValue float64
	DestinationId  string
	destination    *Destination
	ActionsId      string
	actions        []*Action
}

type ActionTiming struct {
	UserBalanceIds []string
	Timing         *Interval
	ActionsId      string
	actions        []*Action
}

func (at *ActionTiming) getActions() (a []*Action, err error) {
	if at.actions == nil {
		a, err = storageGetter.GetActions(at.ActionsId)
		at.actions = a
	}
	return
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

	if i.WeekDays != nil && len(i.WeekDays) > 0 {
		sort.Sort(i.WeekDays)
		if t.IsZero() {
			t = time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second(), 0, now.Location())
		}
		for _, j := range []int{0, 1, 2, 3, 4, 5, 6} {
			t = time.Date(t.Year(), t.Month(), t.Day()+j, t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), t.Location())
			for _, wd := range i.WeekDays {
				if t.Weekday() == wd {
					return
				}
			}
		}
	}

	if i.MonthDays != nil && len(i.MonthDays) > 0 {
		sort.Sort(i.MonthDays)
		now := time.Now()
		x := sort.SearchInts(i.MonthDays, now.Day())
		d = i.MonthDays[0]
		if x < len(i.MonthDays) {
			if i.MonthDays[x] == now.Day() {
				if now.Before(t) {
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
		sort.Sort(i.Months)
		now := time.Now()
		x := sort.Search(len(i.Months), func(x int) bool { return i.Months[x] >= now.Month() })
		m = i.Months[0]
		if x < len(i.Months) {
			if i.Months[x] == now.Month() {
				if now.Before(t) {
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

func (at *ActionTiming) Execute() (err error) {
	aac, err := at.getActions()
	if err != nil {
		return
	}
	for _, a := range aac {
		actionFunction, exists := actionTypeFuncMap[a.ActionType]
		if !exists {
			log.Printf("Function type %v not available, aborting execution!", a.ActionType)
			return
		}
		err = actionFunction(a)
	}
	return
}
