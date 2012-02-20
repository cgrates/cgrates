package timespans

import (
	"strings"
)

/*
Structure that gathers multiple destination prefixes under a common id.
*/
type Destination struct {
	Id       string
	Prefixes []string
}

/*
Serializes the destination for the storage. Used for key-value storages.
*/
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
De-serializes the destination for the storage. Used for key-value storages.
*/
func (d *Destination) containsPrefix(prefix string) bool {
	for _, p := range d.Prefixes {
		if prefix == p {
			return true
		}
	}
	return false
}
