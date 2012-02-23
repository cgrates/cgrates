package timespans

import (
	"log"
	"sort"
	"strconv"
	"strings"
	"sync"
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
	mux                sync.RWMutex
}

type AmountTooBig byte

func (a AmountTooBig) Error() string {
	return "Amount excedes budget!"
}

/*
Structure to store minute buckets according to priority, precision or price.
*/
type bucketsorter []*MinuteBucket

func (bs bucketsorter) Len() int {
	return len(bs)
}

func (bs bucketsorter) Swap(i, j int) {
	bs[i], bs[j] = bs[j], bs[i]
}

func (bs bucketsorter) Less(j, i int) bool {
	return bs[i].Priority < bs[j].Priority ||
		bs[i].precision < bs[j].precision ||
		bs[i].Price > bs[j].Price
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
		mb.Seconds, _ = strconv.ParseFloat(mbse[0], 64)
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
func (ub *UserBudget) getSecondsForPrefix(sg StorageGetter, prefix string) (seconds float64, bucketList bucketsorter) {
	if len(ub.MinuteBuckets) == 0 {
		log.Print("There are no minute buckets to check for user", ub.Id)
		return
	}

	for _, mb := range ub.MinuteBuckets {
		d := mb.getDestination(sg)
		if d == nil {
			continue
		}
		contains, precision := d.containsPrefix(prefix)
		if contains {
			mb.precision = precision
			if mb.Seconds > 0 {
				bucketList = append(bucketList, mb)
			}
		}
	}
	sort.Sort(bucketList) // sorts the buckets according to priority, precision or price
	credit := ub.Credit
	for _, mb := range bucketList {
		s := mb.GetSecondsForCredit(credit)
		credit -= s * mb.Price
		seconds += s
	}
	return
}

/*
Debits some amount of user's money credit. Returns the remaining credit in user's budget.
*/
func (ub *UserBudget) debitMoneyBudget(sg StorageGetter, amount float64) float64 {
	ub.mux.Lock()
	defer ub.mux.Unlock()
	ub.Credit -= amount
	sg.SetUserBudget(ub)
	return ub.Credit
}

/*
Debits the recived amount of seconds from user's minute buckets.
All the appropriate buckets will be debited until all amount of minutes is consumed.
If the amount is bigger than the sum of all seconds in the minute buckets than nothing will be
debited and an error will be returned.
*/
func (ub *UserBudget) debitMinutesBudget(sg StorageGetter, amount float64, prefix string) error {
	ub.mux.Lock()
	defer ub.mux.Unlock()
	avaliableNbSeconds, bucketList := ub.getSecondsForPrefix(sg, prefix)
	if avaliableNbSeconds < amount {
		return new(AmountTooBig)
	}
	for _, mb := range bucketList {
		if mb.Seconds < amount {
			amount -= mb.Seconds
			mb.Seconds = 0
		} else {
			mb.Seconds -= amount
			break
		}
	}
	sg.SetUserBudget(ub)
	return nil
}

/*
Debits some amount of user's SMS budget. Returns the remaining SMS in user's budget.
If the amount is bigger than the budget than nothing wil be debited and an error will be returned
*/
func (ub *UserBudget) debitSMSBuget(sg StorageGetter, amount int) (int, error) {
	ub.mux.Lock()
	defer ub.mux.Unlock()
	if ub.SmsCredit < amount {
		return ub.SmsCredit, new(AmountTooBig)
	}
	ub.SmsCredit -= amount
	sg.SetUserBudget(ub)
	return ub.SmsCredit, nil
}
