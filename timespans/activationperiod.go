package timespans

import (
	"strconv"
	"strings"
	"time"
)

/*
The struture that is saved to storage.
*/
type ActivationPeriod struct {
	ActivationTime time.Time
	Intervals      []*Interval
}

/*
Adds one ore more intervals to the internal interval list.
*/
func (ap *ActivationPeriod) AddInterval(is ...*Interval) {
	for _, i := range is {
		ap.Intervals = append(ap.Intervals, i)
	}
}

/*
Serializes the objects for the storage.
*/
func (ap *ActivationPeriod) store() (result string) {
	result += strconv.FormatInt(ap.ActivationTime.UnixNano(), 10) + ";"
	var is string
	for _, i := range ap.Intervals {
		is = strconv.Itoa(int(i.Month)) + "|"
		is += strconv.Itoa(i.MonthDay) + "|"
		for _, wd := range i.WeekDays {
			is += strconv.Itoa(int(wd)) + ","
		}
		is = strings.TrimRight(is, ",") + "|"
		is += i.StartTime + "|"
		is += i.EndTime + "|"
		is += strconv.FormatFloat(i.Ponder, 'f', -1, 64) + "|"
		is += strconv.FormatFloat(i.ConnectFee, 'f', -1, 64) + "|"
		is += strconv.FormatFloat(i.Price, 'f', -1, 64) + "|"
		is += strconv.FormatFloat(i.BillingUnit, 'f', -1, 64)
		result += is + ";"
	}
	return
}

/*
De-serializes the objects for the storage.
*/
func (ap *ActivationPeriod) restore(input string) {
	elements := strings.Split(input, ";")
	unixNano, _ := strconv.ParseInt(elements[0], 10, 64)
	ap.ActivationTime = time.Unix(0, unixNano).In(time.UTC)
	for _, is := range elements[1 : len(elements)-1] {
		i := &Interval{}
		ise := strings.Split(is, "|")
		month, _ := strconv.Atoi(ise[0])
		i.Month = time.Month(month)
		i.MonthDay, _ = strconv.Atoi(ise[1])
		for _, d := range strings.Split(ise[2], ",") {
			wd, _ := strconv.Atoi(d)
			i.WeekDays = append(i.WeekDays, time.Weekday(wd))
		}
		i.StartTime = ise[3]
		i.EndTime = ise[4]
		i.Ponder, _ = strconv.ParseFloat(ise[5], 64)
		i.ConnectFee, _ = strconv.ParseFloat(ise[6], 64)
		i.Price, _ = strconv.ParseFloat(ise[7], 64)
		i.BillingUnit, _ = strconv.ParseFloat(ise[8], 64)

		ap.Intervals = append(ap.Intervals, i)
	}
}
