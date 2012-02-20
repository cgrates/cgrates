package timespans

type MinuteBucket struct {
	Seconds     int
	Priority    int
	Price       float64
	DestinationId string
	destination *Destination
}

func (mb *MinuteBucket) getDestination(storage StorageGetter) (dest *Destination) {
	if mb.destination == nil {
		mb.destination,_ = storage.GetDestination(mb.DestinationId)
	}
	return mb.destination
}
