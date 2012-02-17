package timespans

import (
	"strings"
	"strconv"
	"time"
	"github.com/simonz05/godis"
)

type RedisStorage struct {
	db *godis.Client
}

func NewRedisStorage(address string, db int) (*RedisStorage, error) {
	ndb := godis.New(address, db, "")
	return &RedisStorage{db: ndb}, nil
}

func (rs *RedisStorage) Close() {
	rs.db.Quit()
}

func (rs *RedisStorage) GetActivationPeriods(key string) (aps []*ActivationPeriod, err error) {
	elem, err := rs.db.Get(key)
	values:= elem.String()
	if err == nil {
		for _, ap_string := range strings.Split(values, "\n") {
				if len(ap_string) > 0 {
					ap := rs.restore(ap_string)
					aps = append(aps, ap)
				}
		}
	}
	return aps, err
}

func (rs *RedisStorage) SetActivationPeriods(key string, aps []*ActivationPeriod){
	result := ""
	for _, ap := range aps {
		result += rs.store(ap) + "\n"
	}
	rs.db.Set(key, result)
}

/*
Serializes the activation periods for the storage.
*/
func (rs *RedisStorage) store(ap *ActivationPeriod) (result string) {
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
De-serializes the activation periods for the storage.
*/
func (rs *RedisStorage) restore(input string) (ap *ActivationPeriod) {
	elements := strings.Split(input, ";")
	unixNano, _ := strconv.ParseInt(elements[0], 10, 64)
	ap = &ActivationPeriod{}
	ap.ActivationTime = time.Unix(0, unixNano).In(time.UTC)
	for _, is := range elements[1 : len(elements)-1] {
		i := &Interval{}
		ise := strings.Split(is, "|")
		month, _ := strconv.Atoi(ise[0])
		i.Month = time.Month(month)
		i.MonthDay, _ = strconv.Atoi(ise[1])
		for _, d := range strings.Split(ise[2], ",") {
			if d != "" {
				wd, _ := strconv.Atoi(d)
				i.WeekDays = append(i.WeekDays, time.Weekday(wd))
			}
		}
		i.StartTime = ise[3]
		i.EndTime = ise[4]
		i.Ponder, _ = strconv.ParseFloat(ise[5], 64)
		i.ConnectFee, _ = strconv.ParseFloat(ise[6], 64)
		i.Price, _ = strconv.ParseFloat(ise[7], 64)
		i.BillingUnit, _ = strconv.ParseFloat(ise[8], 64)

		ap.Intervals = append(ap.Intervals, i)
	}
	return
}
