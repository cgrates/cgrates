package timespans

import (
	"log"
	"math"
)

/*
Structure conatining information about user's credit (minutes, cents, sms...).'
*/
type UserBudget struct {
	id                 string
	minuteBuckets      []*MinuteBucket
	credit             float64
	smsCredit          int
	tariffPlan         *TariffPlan
	resetDayOfTheMonth int
}

/*
Returns user's avaliable minutes for the specified destination
*/
func (ub *UserBudget) GetSecondsForPrefix(storage StorageGetter, prefix string) (seconds int) {
	if len(ub.minuteBuckets) == 0 {
		log.Print("There are no minute buckets to check for user", ub.id)
		return
	}
	bestBucket := ub.minuteBuckets[0]

	for _, mb := range ub.minuteBuckets {
		d := mb.getDestination(storage)
		if d.containsPrefix(prefix) && mb.priority > bestBucket.priority {
			bestBucket = mb
		}
	}
	seconds = bestBucket.seconds
	if bestBucket.price > 0 {
		seconds = int(math.Min(ub.credit/bestBucket.price, float64(seconds)))
	}
	return
}
