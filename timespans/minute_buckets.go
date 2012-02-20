package timespans

type MinuteBucket struct {
	seconds     int
	priority    int
	price       float64
	destinationId string
	destination *Destination
}

func (mb *MinuteBucket) getDestination(storage StorageGetter) (dest *Destination) {
	if mb.destination == nil {
		mb.destination,_ = storage.GetDestination(mb.destinationId)
	}
	return mb.destination
}
