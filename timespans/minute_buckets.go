package timespans

type MinuteBucket struct {
	seconds     int
	priority    int
	price       float64
	destination *Destination
}

/*
Returns true if the bucket contains specified prefix.
*/
func (mb *MinuteBucket) containsPrefix(prefix string) bool {
	for _, p := range mb.destination.Prefixes {
		if prefix == p {
			return true
		}
	}
	return false
}
