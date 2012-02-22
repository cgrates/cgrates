package timespans

import "math"

type MinuteBucket struct {
	Seconds       float64
	Priority      int
	Price         float64
	DestinationId string
	destination   *Destination
	precision     int
}

/*
Returns the destination loading it from the storage if necessary.
*/
func (mb *MinuteBucket) getDestination(storage StorageGetter) (dest *Destination) {
	if mb.destination == nil {
		mb.destination, _ = storage.GetDestination(mb.DestinationId)
	}
	return mb.destination
}

func (mb *MinuteBucket) GetSecondsForCredit(credit float64) (seconds float64) {
	seconds = mb.Seconds
	if mb.Price > 0 {
		seconds = math.Min(credit/mb.Price, mb.Seconds)
	}
	return
}
