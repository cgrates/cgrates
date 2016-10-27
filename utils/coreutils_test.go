package utils

import (
	"testing"
	"time"
)

func TestTimeIs0h(t *testing.T) {
	t1, err := time.Parse(time.RFC3339, "2012-11-01T22:08:41+00:00")
	if err != nil {
		t.Log(err)
		t.Errorf("time parsing error")
	}
	result := TimeIs0h(t1)
	if result != false {
		t.Errorf("setting time")
	}

}
