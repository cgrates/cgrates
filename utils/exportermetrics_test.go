// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package utils

import (
	"reflect"
	"sort"
	"testing"
)

func TestExporterMetricsString(t *testing.T) {
	ms := &ExporterMetrics{
		MapStorage: MapStorage{
			"field1": 2,
		},
	}
	expected := "{\"field1\":2}"
	if reply := ms.String(); reply != expected {
		t.Errorf("Expected %s \n but received \n %s", expected, reply)
	}
}

func TestExporterMetricsFieldAsInterface(t *testing.T) {
	ms := &ExporterMetrics{
		MapStorage: MapStorage{
			"field1": 2,
		},
	}

	input := []string{"field1"}
	expected := 2
	if reply, err := ms.FieldAsInterface(input); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, reply) {
		t.Errorf("Expected %d \n but received \n %d", expected, reply)
	}
}

func TestExporterMetricsFieldAsString(t *testing.T) {
	ms := &ExporterMetrics{
		MapStorage: MapStorage{
			"field1": 2,
		},
	}

	input := []string{"field1"}
	expected := "2"
	if reply, err := ms.FieldAsString(input); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, reply) {
		t.Errorf("Expected %s \n but received \n %s", expected, reply)
	}
}

func TestExporterMetricsSet(t *testing.T) {
	ms := &ExporterMetrics{
		MapStorage: MapStorage{
			"field1": 2,
		},
	}

	expected := &ExporterMetrics{
		MapStorage: MapStorage{
			"field1": 2,
			"field2": 3,
		},
	}

	if err := ms.Set([]string{"field2"}, 3); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ms, expected) {
		t.Errorf("Expected %v \n but received \n %v", expected, ms)
	}
}

func TestExporterMetricsGetKeys(t *testing.T) {
	ms := &ExporterMetrics{
		MapStorage: MapStorage{
			"field1": 2,
			"field2": 3,
		},
	}

	expected := []string{"*req.field1", "*req.field2"}
	reply := ms.GetKeys(false, 0, MetaReq)
	sort.Strings(reply)

	if !reflect.DeepEqual(reply, expected) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", expected, reply)
	}
}

func TestExporterMetricsRemove(t *testing.T) {
	ms := &ExporterMetrics{
		MapStorage: MapStorage{
			"field1": 2,
			"field2": 3,
		},
	}

	expected := &ExporterMetrics{
		MapStorage: MapStorage{
			"field1": 2,
		},
	}

	if err := ms.Remove([]string{"field2"}); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ms, expected) {
		t.Errorf("Expected %v \n but received \n %v", expected, ms)
	}
}

func TestExporterMetricsClonedMapStorage(t *testing.T) {
	ms := &ExporterMetrics{
		MapStorage: MapStorage{
			"field1": []string{"v1", "v2"},
		},
	}
	exp := MapStorage{
		"field1": []string{"v1", "v2"},
	}
	msC := ms.ClonedMapStorage()
	// fmt.Println(msC)
	if !reflect.DeepEqual(exp, msC) {
		t.Errorf("Expected %v \n but received \n %v", exp, msC)
	}
}
