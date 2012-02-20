package timespans

type TariffPlan struct {
	Id            string
	MinuteBuckets []*MinuteBucket
}
