package timespans

import (
	"testing"
)

func TestGetDestination(t *testing.T) {
	getter, _ := NewRedisStorage("tcp:127.0.0.1:6379", 10)
	defer getter.Close()
	mb := &MinuteBucket{destinationId: "nationale"}
	d := mb.getDestination(getter)
	if d.Id != "nationale" || len(d.Prefixes) != 4 {
		t.Error("Got wrong destination: ", d)
	}
}
