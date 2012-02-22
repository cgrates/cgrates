package timespans

import (
	"log"
	"math"
	"strconv"
	"sort"
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
Structure to store minute buckets according to priority, precision or price.
*/
type BucketSorter []*MinuteBucket

func (bs BucketSorter) Len() int {
	return len(bs)
}

func (bs BucketSorter) Swap(i, j int) {
	bs[i], bs[j] = bs[j], bs[i]
}

func (bs BucketSorter) Less(j, i int) bool {
	return bs[i].Priority < bs[j].Priority ||
		bs[i].precision < bs[j].precision ||
		bs[i].Price < bs[j].Price
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
func (ub *UserBudget) getSecondsForPrefix(storage StorageGetter, prefix string) (seconds float64) {
	if len(ub.MinuteBuckets) == 0 {
		log.Print("There are no minute buckets to check for user", ub.Id)
		return
	}
	var bucketList BucketSorter
	for _, mb := range ub.MinuteBuckets {
		d := mb.getDestination(storage)
		if d == nil {
			continue
		}
		contains, precision := d.containsPrefix(prefix)
		if contains {
			mb.precision = precision
			bucketList = append(bucketList, mb)
		}
	}
	sort.Sort(bucketList)
	credit := ub.Credit
	for _, mb := range bucketList {
		s := float64(mb.Seconds)
		if mb.Price > 0 {
			s = math.Min(credit/mb.Price, s)
			credit -= s
		}
		seconds += s
	}
	return
}
