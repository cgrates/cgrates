package timespans

import (
	"log"
	"math"
	"strconv"
	"strings"
)

/*
Structure conatining information about user's credit (minutes, cents, sms...).'
*/
type UserBudget struct {
	Id                 string
	Credit             float64
	SmsCredit          int
	ResetDayOfTheMonth int
	TariffPlanId       string
	tariffPlan         *TariffPlan
	MinuteBuckets      []*MinuteBucket
}

/*
Serializes the user budget for the storage. Used for key-value storages.
*/
func (ub *UserBudget) store() (result string) {
	result += strconv.FormatFloat(ub.Credit, 'f', -1, 64) + ";"
	result += strconv.Itoa(ub.SmsCredit) + ";"
	result += strconv.Itoa(ub.ResetDayOfTheMonth) + ";"
	result += ub.TariffPlanId + ";"
	for _, mb := range ub.MinuteBuckets {
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
De-serializes the user budget for the storage. Used for key-value storages.
*/
func (ub *UserBudget) restore(input string) {
	elements := strings.Split(input, ";")
	ub.Credit, _ = strconv.ParseFloat(elements[0], 64)
	ub.SmsCredit, _ = strconv.Atoi(elements[1])
	ub.ResetDayOfTheMonth, _ = strconv.Atoi(elements[2])
	ub.TariffPlanId = elements[3]
	for _, mbs := range elements[4 : len(elements)-1] {
		mb := &MinuteBucket{}
		mbse := strings.Split(mbs, "|")
		mb.Seconds, _ = strconv.Atoi(mbse[0])
		mb.Priority, _ = strconv.Atoi(mbse[1])
		mb.Price, _ = strconv.ParseFloat(mbse[2], 64)
		mb.DestinationId = mbse[3]

		ub.MinuteBuckets = append(ub.MinuteBuckets, mb)
	}
}


/*
Returns the tariff plan loading it from the storage if necessary.
*/
func (ub *UserBudget) getTariffPlan(storage StorageGetter) (tp *TariffPlan) {
	if ub.tariffPlan == nil {
		ub.tariffPlan, _ = storage.GetTariffPlan(ub.TariffPlanId)
	}
	return ub.tariffPlan
}

/*
Returns user's avaliable minutes for the specified destination
*/
func (ub *UserBudget) GetSecondsForPrefix(storage StorageGetter, prefix string) (seconds int) {
	if len(ub.MinuteBuckets) == 0 {
		log.Print("There are no minute buckets to check for user", ub.Id)
		return
	}
	bestBucket := ub.MinuteBuckets[0]

	for _, mb := range ub.MinuteBuckets {
		d := mb.getDestination(storage)
		if d.containsPrefix(prefix) && mb.Priority > bestBucket.Priority {
			bestBucket = mb
		}
	}
	seconds = bestBucket.Seconds
	if bestBucket.Price > 0 {
		seconds = int(math.Min(ub.Credit/bestBucket.Price, float64(seconds)))
	}
	return
}
