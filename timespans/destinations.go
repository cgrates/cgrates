package timespans

import (
	"strings"
)

type Destination struct {
	Id       string
	Prefixes []string
}

func (d *Destination) GetKey() (result string) {
	return d.Id
}

func (d *Destination) store() (result string) {
	for _, p := range d.Prefixes {
		result += p + ","
	}
	result = strings.TrimRight(result, ",")
	return
}

func (d *Destination) restore(input string) {
	d.Prefixes = strings.Split(input, ",")
}

/*
Returns true if the bucket contains specified prefix.
*/
func (d *Destination) containsPrefix(prefix string) bool {
	for _, p := range d.Prefixes {
		if prefix == p {
			return true
		}
	}
	return false
}
