package timespans

import (
	"github.com/fsouza/gokabinet/kc"
	"time"
	"strconv"
	"strings"
)

type KyotoStorage struct {
	db *kc.DB
}

func NewKyotoStorage(filaName string) (*KyotoStorage, error) {
	ndb, err := kc.Open(filaName, kc.READ)
	return &KyotoStorage{db: ndb}, err
}

func (ks *KyotoStorage) Close() {
	ks.db.Close()
}

func (ks *KyotoStorage) GetActivationPeriods(key string) (aps []*ActivationPeriod, err error) {
	values, err := ks.db.Get(key)

	if err == nil {
		for _, ap_string := range strings.Split(values, "\n") {
				if len(ap_string) > 0 {
					ap := ks.restore(ap_string)
					aps = append(aps, ap)
				}
		}
	}
	return aps, err
}

func (ks *KyotoStorage) SetActivationPeriods(key string, aps []*ActivationPeriod){
	result := ""
	for _, ap := range aps {
		result += ks.store(ap) + "\n"
	}
	ks.db.Set(key, result)
}

/*
Serializes the activation periods for the storage.
*/
func (ks *KyotoStorage) store(ap *ActivationPeriod) (result string) {
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
func (ks *KyotoStorage) restore(input string) (ap *ActivationPeriod) {
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
