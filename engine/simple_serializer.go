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
	"strconv"
	"strings"
	"time"
)

type Serializer interface {
	Store() (string, error)
	Restore(string) error
}

func notEnoughElements(element, line string) error {
	return fmt.Errorf("Too few elements to restore (%v): %v", element, line)
}

func (ap *ActivationPeriod) Store() (result string, err error) {
	result += ap.ActivationTime.Format(time.RFC3339) + "|"
	for _, i := range ap.Intervals {
		str, err := i.Store()
		if err != nil {
			return "", err
		}
		result += str + "|"
	}
	result = strings.TrimRight(result, "|")
	return
}

func (ap *ActivationPeriod) Restore(input string) error {
	elements := strings.Split(input, "|")
	var err error
	ap.ActivationTime, err = time.Parse(time.RFC3339, elements[0])
	if err != nil {
		return err
	}
	els := elements[1:]
	if len(els) > 1 {
		els = elements[1 : len(elements)-1]
	}
	for _, is := range els {
		i := &Interval{}
		if err := i.Restore(is); err != nil {
			return err
		}
		ap.Intervals = append(ap.Intervals, i)
	}
	return nil
}

func (rp *RatingProfile) Store() (result string, err error) {
	result += rp.FallbackKey + ">"
	for k, aps := range rp.DestinationMap {
		result += k + "="
		for _, ap := range aps {
			aps, err := ap.Store()
			if err != nil {
				return result, err
			}
			result += aps + "<"
		}
		result = strings.TrimRight(result, "<")
		result += ">"
	}
	result = strings.TrimRight(result, ">")
	return
}

func (rp *RatingProfile) Restore(input string) error {
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
			err := ap.Restore(aps)
			if err != nil {
				return err
			}
			newAps = append(newAps, ap)
		}
		rp.DestinationMap[pair[0]] = newAps
	}
	return nil
}

func (a *Action) Store() (result string, err error) {
	result += a.Id + "|"
	result += a.ActionType + "|"
	result += a.BalanceId + "|"
	result += a.Direction + "|"
	result += a.ExpirationString + "|"
	result += a.ExpirationDate.Format(time.RFC3339) + "|"
	result += strconv.FormatFloat(a.Units, 'f', -1, 64) + "|"
	result += strconv.FormatFloat(a.Weight, 'f', -1, 64)
	if a.MinuteBucket != nil {
		result += "|"
		str, err := a.MinuteBucket.Store()
		if err != nil {
			return "", err
		}
		result += str
	}
	return
}

func (a *Action) Restore(input string) (err error) {
	elements := strings.Split(input, "|")
	if len(elements) < 8 {
		return notEnoughElements("Action", input)
	}
	a.Id = elements[0]
	a.ActionType = elements[1]
	a.BalanceId = elements[2]
	a.Direction = elements[3]
	a.ExpirationString = elements[4]
	if a.ExpirationDate, err = time.Parse(time.RFC3339, elements[5]); err != nil {
		return err
	}
	a.Units, _ = strconv.ParseFloat(elements[6], 64)
	a.Weight, _ = strconv.ParseFloat(elements[7], 64)
	if len(elements) == 9 {
		a.MinuteBucket = &MinuteBucket{}
		if err := a.MinuteBucket.Restore(elements[8]); err != nil {
			return err
		}
	}
	return nil
}

func (as Actions) Store() (result string, err error) {
	for _, a := range as {
		str, err := a.Store()
		if err != nil {
			return "", err
		}
		result += str + "~"
	}
	result = strings.TrimRight(result, "~")
	return
}

func (as *Actions) Restore(input string) error {
	for _, a_string := range strings.Split(input, "~") {
		if len(a_string) > 0 {
			a := &Action{}
			if err := a.Restore(a_string); err != nil {
				return err
			}
			*as = append(*as, a)
		}
	}
	return nil
}

func (at *ActionTiming) Store() (result string, err error) {
	result += at.Id + "|"
	result += at.Tag + "|"
	for _, ubi := range at.UserBalanceIds {
		result += ubi + ","
	}
	result = strings.TrimRight(result, ",") + "|"
	if at.Timing != nil {
		str, err := at.Timing.Store()
		if err != nil {
			return "", err
		}
		result += str + "|"
	} else {
		result += " |"
	}
	result += strconv.FormatFloat(at.Weight, 'f', -1, 64) + "|"
	result += at.ActionsId
	return
}

func (at *ActionTiming) Restore(input string) error {
	elements := strings.Split(input, "|")
	at.Id = elements[0]
	at.Tag = elements[1]
	for _, ubi := range strings.Split(elements[2], ",") {
		if strings.TrimSpace(ubi) != "" {
			at.UserBalanceIds = append(at.UserBalanceIds, ubi)
		}
	}

	at.Timing = &Interval{}
	if err := at.Timing.Restore(elements[3]); err != nil {
		return err
	}
	at.Weight, _ = strconv.ParseFloat(elements[4], 64)
	at.ActionsId = elements[5]
	return nil
}

func (ats ActionTimings) Store() (result string, err error) {
	for _, at := range ats {
		str, err := at.Store()
		if err != nil {
			return "", err
		}
		result += str + "~"
	}
	result = strings.TrimRight(result, "~")
	return
}

func (ats *ActionTimings) Restore(input string) error {
	for _, at_string := range strings.Split(input, "~") {
		if len(at_string) > 0 {
			at := &ActionTiming{}
			if err := at.Restore(at_string); err != nil {
				return err
			}
			*ats = append(*ats, at)
		}
	}
	return nil
}

func (at *ActionTrigger) Store() (result string, err error) {
	result += at.Id + ";"
	result += at.BalanceId + ";"
	result += at.Direction + ";"
	result += at.DestinationId + ";"
	result += at.ActionsId + ";"
	result += strconv.FormatFloat(at.ThresholdValue, 'f', -1, 64) + ";"
	result += at.ThresholdType + ";"
	result += strconv.FormatFloat(at.Weight, 'f', -1, 64) + ";"
	result += strconv.FormatBool(at.Executed)
	return
}

func (at *ActionTrigger) Restore(input string) error {
	elements := strings.Split(input, ";")
	if len(elements) != 9 {
		return notEnoughElements("ActionTrigger", input)
	}
	at.Id = elements[0]
	at.BalanceId = elements[1]
	at.Direction = elements[2]
	at.DestinationId = elements[3]
	at.ActionsId = elements[4]
	at.ThresholdValue, _ = strconv.ParseFloat(elements[5], 64)
	at.ThresholdType = elements[6]
	at.Weight, _ = strconv.ParseFloat(elements[7], 64)
	at.Executed, _ = strconv.ParseBool(elements[8])
	return nil
}

func (b *Balance) Store() (result string, err error) {
	result += b.Id + ")"
	result += strconv.FormatFloat(b.Value, 'f', -1, 64) + ")"
	result += strconv.FormatFloat(b.Weight, 'f', -1, 64) + ")"
	result += b.ExpirationDate.Format(time.RFC3339)
	return result, nil
}

func (b *Balance) Restore(input string) (err error) {
	elements := strings.Split(input, ")")
	b.Id = elements[0]
	b.Value, _ = strconv.ParseFloat(elements[1], 64)
	b.Weight, _ = strconv.ParseFloat(elements[2], 64)
	b.ExpirationDate, err = time.Parse(time.RFC3339, elements[3])
	return nil
}

func (bc BalanceChain) Store() (result string, err error) {
	for _, b := range bc {
		str, err := b.Store()
		if err != nil {
			return "", err
		}
		result += str + "^"
	}
	result = strings.TrimRight(result, "^")
	return result, nil
}

func (bc *BalanceChain) Restore(input string) error {
	elements := strings.Split(input, "^")
	for _, element := range elements {
		b := &Balance{}
		err := b.Restore(element)
		if err != nil {
			return err
		}
		*bc = append(*bc, b)
	}
	return nil
}

func (ub *UserBalance) Store() (result string, err error) {
	result += ub.Id + "|"
	result += ub.Type + "|"
	for k, v := range ub.BalanceMap {
		bc, err := v.Store()
		if err != nil {
			return "", err
		}
		result += k + "=" + bc + "#"
	}
	result = strings.TrimRight(result, "#") + "|"
	for _, mb := range ub.MinuteBuckets {
		str, err := mb.Store()
		if err != nil {
			return "", err
		}
		result += str + "#"
	}
	result = strings.TrimRight(result, "#") + "|"
	for _, uc := range ub.UnitCounters {
		str, err := uc.Store()
		if err != nil {
			return "", err
		}
		result += str + "#"
	}
	result = strings.TrimRight(result, "#") + "|"
	for _, at := range ub.ActionTriggers {
		res, err := at.Store()
		if err != nil {
			return "", err
		}
		result += res + "#"
	}
	result = strings.TrimRight(result, "#")
	return
}

func (ub *UserBalance) Restore(input string) error {
	elements := strings.Split(input, "|")
	if len(elements) < 2 {
		return notEnoughElements("UserBalance", input)
	}
	ub.Id = elements[0]
	ub.Type = elements[1]
	if ub.BalanceMap == nil {
		ub.BalanceMap = make(map[string]BalanceChain, 0)
	}
	for _, maps := range strings.Split(elements[2], "#") {
		kv := strings.Split(maps, "=")
		if len(kv) != 2 {
			continue
		}
		bc := BalanceChain{}
		if err := (&bc).Restore(kv[1]); err != nil {
			return err
		}
		ub.BalanceMap[kv[0]] = bc
	}
	for _, mbs := range strings.Split(elements[3], "#") {
		if mbs == "" {
			continue
		}
		mb := &MinuteBucket{}
		if err := mb.Restore(mbs); err != nil {
			return err
		}
		ub.MinuteBuckets = append(ub.MinuteBuckets, mb)
	}
	for _, ucs := range strings.Split(elements[4], "#") {
		if ucs == "" {
			continue
		}
		uc := &UnitsCounter{}
		if err := uc.Restore(ucs); err != nil {
			return err
		}
		ub.UnitCounters = append(ub.UnitCounters, uc)
	}
	for _, ats := range strings.Split(elements[5], "#") {
		if ats == "" {
			continue
		}
		at := &ActionTrigger{}
		if err := at.Restore(ats); err != nil {
			return err
		}
		ub.ActionTriggers = append(ub.ActionTriggers, at)
	}
	return nil
}

/*
Serializes the unit counter for the storage. Used for key-value storages.
*/
func (uc *UnitsCounter) Store() (result string, err error) {
	result += uc.Direction + "/"
	result += uc.BalanceId + "/"
	result += strconv.FormatFloat(uc.Units, 'f', -1, 64) + "/"
	for _, mb := range uc.MinuteBuckets {
		str, err := mb.Store()
		if err != nil {
			return "", err
		}
		result += str + ","
	}
	result = strings.TrimRight(result, ",")
	return
}

/*
De-serializes the unit counter for the storage. Used for key-value storages.
*/
func (uc *UnitsCounter) Restore(input string) error {
	elements := strings.Split(input, "/")
	if len(elements) != 4 {
		return notEnoughElements("UnitsCounter", input)
	}
	uc.Direction = elements[0]
	uc.BalanceId = elements[1]
	uc.Units, _ = strconv.ParseFloat(elements[2], 64)
	for _, mbs := range strings.Split(elements[3], ",") {
		mb := &MinuteBucket{}
		if err := mb.Restore(mbs); err != nil {
			return err
		}
		uc.MinuteBuckets = append(uc.MinuteBuckets, mb)
	}
	return nil
}

func (d *Destination) Store() (result string, err error) {
	for _, p := range d.Prefixes {
		result += p + ","
	}
	result = strings.TrimRight(result, ",")
	return
}

func (d *Destination) Restore(input string) error {
	d.Prefixes = strings.Split(input, ",")
	return nil
}

func (pg PriceGroups) Store() (result string, err error) {
	for _, p := range pg {
		result += p.GroupIntervalStart.String() +
			":" + strconv.FormatFloat(p.Value, 'f', -1, 64) +
			":" + p.RateIncrement.String() +
			":" + p.RateUnit.String() +
			","
	}
	result = strings.TrimRight(result, ",")
	return
}

func (pg *PriceGroups) Restore(input string) error {
	elements := strings.Split(input, ",")
	for _, element := range elements {
		priceElements := strings.Split(element, ":")
		if len(priceElements) != 4 {
			continue
		}
		ss, err := time.ParseDuration(priceElements[0])
		if err != nil {
			return err
		}
		v, err := strconv.ParseFloat(priceElements[1], 64)
		if err != nil {
			return err
		}
		ri, err := time.ParseDuration(priceElements[2])
		if err != nil {
			return err
		}
		ru, err := time.ParseDuration(priceElements[3])
		if err != nil {
			return err
		}
		price := &Price{
			GroupIntervalStart: ss,
			Value:              v,
			RateIncrement:      ri,
			RateUnit:           ru,
		}
		*pg = append(*pg, price)
	}
	return nil
}

func (i *Interval) Store() (result string, err error) {
	str, err := i.Years.Store()
	if err != nil {
		return "", err
	}
	result += str + ";"
	str, err = i.Months.Store()
	if err != nil {
		return "", err
	}
	result += str + ";"
	str, err = i.MonthDays.Store()
	if err != nil {
		return "", err
	}
	result += str + ";"
	str, err = i.WeekDays.Store()
	if err != nil {
		return "", err
	}
	result += str + ";"
	result += i.StartTime + ";"
	result += i.EndTime + ";"
	result += strconv.FormatFloat(i.Weight, 'f', -1, 64) + ";"
	result += strconv.FormatFloat(i.ConnectFee, 'f', -1, 64) + ";"
	ps, err := i.Prices.Store()
	if err != nil {
		return "", err
	}
	result += ps + ";"
	result += i.RoundingMethod + ";"
	result += strconv.Itoa(i.RoundingDecimals)
	return
}

func (i *Interval) Restore(input string) error {
	is := strings.Split(input, ";")
	if len(is) != 11 {
		return notEnoughElements("Interval", input)
	}
	if err := i.Years.Restore(is[0]); err != nil {
		return err
	}
	if err := i.Months.Restore(is[1]); err != nil {
		return err
	}
	if err := i.MonthDays.Restore(is[2]); err != nil {
		return err
	}
	if err := i.WeekDays.Restore(is[3]); err != nil {
		return err
	}
	i.StartTime = is[4]
	i.EndTime = is[5]
	i.Weight, _ = strconv.ParseFloat(is[6], 64)
	i.ConnectFee, _ = strconv.ParseFloat(is[7], 64)
	err := (&i.Prices).Restore(is[8])
	if err != nil {
		return err
	}
	i.RoundingMethod = is[9]
	i.RoundingDecimals, _ = strconv.Atoi(is[10])
	return nil
}

func (mb *MinuteBucket) Store() (result string, err error) {
	result += strconv.FormatFloat(mb.Seconds, 'f', -1, 64) + ";"
	result += strconv.FormatFloat(mb.Weight, 'f', -1, 64) + ";"
	result += strconv.FormatFloat(mb.Price, 'f', -1, 64) + ";"
	result += mb.PriceType + ";"
	result += mb.DestinationId
	return
}

func (mb *MinuteBucket) Restore(input string) error {
	elements := strings.Split(input, ";")
	if len(elements) > 0 && len(elements) != 5 {
		return notEnoughElements("MinuteBucket", input)
	}
	mb.Seconds, _ = strconv.ParseFloat(elements[0], 64)
	mb.Weight, _ = strconv.ParseFloat(elements[1], 64)
	mb.Price, _ = strconv.ParseFloat(elements[2], 64)
	mb.PriceType = elements[3]
	mb.DestinationId = elements[4]
	return nil
}

func (wds WeekDays) Store() (result string, err error) {
	for _, wd := range wds {
		result += strconv.Itoa(int(wd)) + ","
	}
	result = strings.TrimRight(result, ",")
	return
}

func (wds *WeekDays) Restore(input string) error {
	for _, wd := range strings.Split(input, ",") {
		if wd != "" {
			mm, _ := strconv.Atoi(wd)
			*wds = append(*wds, time.Weekday(mm))
		}
	}
	return nil
}

func (mds MonthDays) Store() (result string, err error) {
	for _, md := range mds {
		result += strconv.Itoa(int(md)) + ","
	}
	result = strings.TrimRight(result, ",")
	return
}

func (mds *MonthDays) Restore(input string) error {
	for _, md := range strings.Split(input, ",") {
		if md != "" {
			mm, _ := strconv.Atoi(md)
			*mds = append(*mds, mm)
		}
	}
	return nil
}

func (ms Months) Store() (result string, err error) {
	for _, m := range ms {
		result += strconv.Itoa(int(m)) + ","
	}
	result = strings.TrimRight(result, ",")
	return
}

func (ms *Months) Restore(input string) error {
	for _, m := range strings.Split(input, ",") {
		if m != "" {
			mm, _ := strconv.Atoi(m)
			*ms = append(*ms, time.Month(mm))
		}
	}
	return nil
}

func (yss Years) Store() (result string, err error) {
	for _, ys := range yss {
		result += strconv.Itoa(int(ys)) + ","
	}
	result = strings.TrimRight(result, ",")
	return
}

func (yss *Years) Restore(input string) error {
	for _, ys := range strings.Split(input, ",") {
		if ys != "" {
			mm, _ := strconv.Atoi(ys)
			*yss = append(*yss, mm)
		}
	}
	return nil
}
