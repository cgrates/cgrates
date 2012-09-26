package mediator

import (
	"testing"
)

func TestIndexLoad(t *testing.T) {
	idxs := make(mediatorFieldIdxs, 0)
	err := idxs.Load("1,2,3")
	if err != nil && len(idxs) != 3 {
		t.Error("Error parsing indexses ", err)
	}
}

func TestIndexLoadError(t *testing.T) {
	idxs := make(mediatorFieldIdxs, 0)
	err := idxs.Load("1,a,2")
	if err == nil {
		t.Error("Error parsing indexses ", err)
	}
}

func TestIndexLoadEmpty(t *testing.T) {
	idxs := make(mediatorFieldIdxs, 0)
	err := idxs.Load("")
	if err == nil {
		t.Error("Error parsing indexses ", err)
	}
}

func TestIndexLengthSame(t *testing.T) {
	m := new(Mediator)
	objs := []mediatorFieldIdxs{m.directionIndexs, m.torIndexs, m.tenantIndexs, m.subjectIndexs,
		m.accountIndexs, m.destinationIndexs, m.timeStartIndexs, m.durationIndexs, m.uuidIndexs}
	for _, o := range objs {
		o.Load("1,2,3")
	}
	if !m.validateIndexses() {
		t.Error("Error checking length")
	}
}

func TestIndexLengthDifferent(t *testing.T) {
	m := new(Mediator)
	objs := []mediatorFieldIdxs{m.directionIndexs, m.torIndexs, m.tenantIndexs, m.subjectIndexs,
		m.accountIndexs, m.timeStartIndexs, m.durationIndexs, m.uuidIndexs}
	for _, o := range objs {
		o.Load("1,2,3")
	}
	m.destinationIndexs.Load("4,5")
	if m.validateIndexses() {
		t.Error("Error checking length")
	}
}
