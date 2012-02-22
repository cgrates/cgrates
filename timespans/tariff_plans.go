package timespans

import (
	"strconv"
	"strings"
)

/*
Structure describing a tariff plan's number of bonus items. It is uset to restore
these numbers to the user budget every month.
*/
type TariffPlan struct {
	Id            string
	SmsCredit     int
	MinuteBuckets []*MinuteBucket
}

/*
Serializes the tariff plan for the storage. Used for key-value storages.
*/
func (tp *TariffPlan) store() (result string) {
	result += strconv.Itoa(tp.SmsCredit) + ";"
	for _, mb := range tp.MinuteBuckets {
		var mbs string
		mbs += strconv.Itoa(int(mb.Seconds)) + "|"
		mbs += strconv.Itoa(int(mb.Priority)) + "|"
		mbs += strconv.FormatFloat(mb.Price, 'f', -1, 64) + "|"
		mbs += mb.DestinationId
		result += mbs + ";"
	}
	return
}

/*
De-serializes the tariff plan for the storage. Used for key-value storages.
*/
func (tp *TariffPlan) restore(input string) {
	elements := strings.Split(input, ";")
	tp.SmsCredit, _ = strconv.Atoi(elements[0])
	for _, mbs := range elements[1 : len(elements)-1] {
		mb := &MinuteBucket{}
		mbse := strings.Split(mbs, "|")
		mb.Seconds, _ = strconv.ParseFloat(mbse[0], 64)
		mb.Priority, _ = strconv.Atoi(mbse[1])
		mb.Price, _ = strconv.ParseFloat(mbse[2], 64)
		mb.DestinationId = mbse[3]

		tp.MinuteBuckets = append(tp.MinuteBuckets, mb)
	}
}
