package timespans

type MinuteBucket struct {
	Seconds       int
	Priority      int
	Price         float64
	DestinationId string
	destination   *Destination
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
