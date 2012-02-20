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
	SmsCredit          int
	MinuteBuckets []*MinuteBucket
}

/*
Serializes the activation periods for the storage. Used for key-value storages.
*/
func (tp *TariffPlan) store() (result string) {
	result += strconv.Itoa(tp.SmsCredit) + ";"
	for _, mb := range tp.MinuteBuckets {
		var mbs string
		mbs += strconv.Itoa(int(mb.seconds)) + "|"
		mbs += strconv.Itoa(int(mb.priority)) + "|"
		mbs += strconv.FormatFloat(mb.price, 'f', -1, 64) + "|"
		mbs += mb.destinationId
		result += mbs + ";"
	}
	return
}

/*
De-serializes the activation periods for the storage. Used for key-value storages.
*/
func (tp *TariffPlan) restore(input string) {
	elements := strings.Split(input, ";")
	tp.SmsCredit,_ = strconv.Atoi(elements[0])
	for _, mbs := range elements[1 : len(elements)-1] {
		mb := &MinuteBucket{}
		mbse := strings.Split(mbs, "|")
		mb.seconds,_ = strconv.Atoi(mbse[0])
		mb.priority,_ = strconv.Atoi(mbse[1])
		mb.price,_ = strconv.ParseFloat(mbse[2], 64)
		mb.destinationId = mbse[3]

		tp.MinuteBuckets = append(tp.MinuteBuckets, mb)
	}
}
