/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>
*/
package engine

import (
	"reflect"
	"testing"

	"github.com/cgrates/cgrates/utils"
)

func TestLibSuppliersSortCost(t *testing.T) {
	sSpls := &SortedSuppliers{
		SortedSuppliers: []*SortedSupplier{
			&SortedSupplier{
				SupplierID: "supplier1",
				SortingData: map[string]interface{}{
					utils.Cost:   0.1,
					utils.Weight: 10.0,
				},
				SupplierParameters: "param1",
			},
			&SortedSupplier{
				SupplierID: "supplier2",
				SortingData: map[string]interface{}{
					utils.Cost:   0.1,
					utils.Weight: 20.0,
				},
				SupplierParameters: "param2",
			},
			&SortedSupplier{
				SupplierID: "supplier3",
				SortingData: map[string]interface{}{
					utils.Cost:   0.05,
					utils.Weight: 10.0,
				},
				SupplierParameters: "param3",
			},
		},
	}
	sSpls.SortLeastCost()
	eOrderedSpls := &SortedSuppliers{
		SortedSuppliers: []*SortedSupplier{
			&SortedSupplier{
				SupplierID: "supplier3",
				SortingData: map[string]interface{}{
					utils.Cost:   0.05,
					utils.Weight: 10.0,
				},
				SupplierParameters: "param3",
			},
			&SortedSupplier{
				SupplierID: "supplier2",
				SortingData: map[string]interface{}{
					utils.Cost:   0.1,
					utils.Weight: 20.0,
				},
				SupplierParameters: "param2",
			},
			&SortedSupplier{
				SupplierID: "supplier1",
				SortingData: map[string]interface{}{
					utils.Cost:   0.1,
					utils.Weight: 10.0,
				},
				SupplierParameters: "param1",
			},
		},
	}
	if !reflect.DeepEqual(eOrderedSpls, sSpls) {
		t.Errorf("Expecting: %s, received: %s",
			utils.ToJSON(eOrderedSpls), utils.ToJSON(sSpls))
	}
}

func TestLibSuppliersSortWeight(t *testing.T) {
	sSpls := &SortedSuppliers{
		SortedSuppliers: []*SortedSupplier{
			&SortedSupplier{
				SupplierID: "supplier1",
				SortingData: map[string]interface{}{
					utils.Weight: 10.0,
				},
				SupplierParameters: "param1",
			},
			&SortedSupplier{
				SupplierID: "supplier2",
				SortingData: map[string]interface{}{
					utils.Weight: 20.0,
				},
				SupplierParameters: "param2",
			},
			&SortedSupplier{
				SupplierID: "supplier3",
				SortingData: map[string]interface{}{
					utils.Weight: 10.5,
				},
				SupplierParameters: "param3",
			},
		},
	}
	sSpls.SortWeight()
	eOrderedSpls := &SortedSuppliers{
		SortedSuppliers: []*SortedSupplier{
			&SortedSupplier{
				SupplierID: "supplier2",
				SortingData: map[string]interface{}{
					utils.Weight: 20.0,
				},
				SupplierParameters: "param2",
			},
			&SortedSupplier{
				SupplierID: "supplier3",
				SortingData: map[string]interface{}{
					utils.Weight: 10.5,
				},
				SupplierParameters: "param3",
			},
			&SortedSupplier{
				SupplierID: "supplier1",
				SortingData: map[string]interface{}{
					utils.Weight: 10.0,
				},
				SupplierParameters: "param1",
			},
		},
	}
	if !reflect.DeepEqual(eOrderedSpls, sSpls) {
		t.Errorf("Expecting: %s, received: %s",
			utils.ToJSON(eOrderedSpls), utils.ToJSON(sSpls))
	}
}

func TestSortedSuppliersDigest(t *testing.T) {
	eSpls := SortedSuppliers{
		ProfileID: "SPL_WEIGHT_1",
		Sorting:   utils.MetaWeight,
		SortedSuppliers: []*SortedSupplier{
			&SortedSupplier{
				SupplierID: "supplier2",
				SortingData: map[string]interface{}{
					"Weight": 20.0,
				},
				SupplierParameters: "param2",
			},
			&SortedSupplier{
				SupplierID: "supplier1",
				SortingData: map[string]interface{}{
					"Weight": 10.0,
				},
				SupplierParameters: "param1",
			},
		},
	}
	exp := "supplier2:param2,supplier1:param1"
	rcv := eSpls.Digest()
	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("Expecting: %+v, received: %+v", exp, rcv)
	}
}

func TestSortedSuppliersDigest2(t *testing.T) {
	eSpls := SortedSuppliers{
		ProfileID: "SPL_WEIGHT_1",
		Sorting:   utils.MetaWeight,
		SortedSuppliers: []*SortedSupplier{
			&SortedSupplier{
				SupplierID: "supplier1",
				SortingData: map[string]interface{}{
					"Weight": 30.0,
				},
				SupplierParameters: "param1",
			},
			&SortedSupplier{
				SupplierID: "supplier2",
				SortingData: map[string]interface{}{
					"Weight": 20.0,
				},
				SupplierParameters: "param2",
			},
		},
	}
	exp := "supplier1:param1,supplier2:param2"
	rcv := eSpls.Digest()
	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("Expecting: %+v, received: %+v", exp, rcv)
	}
}

func TestSortedSuppliersDigest3(t *testing.T) {
	eSpls := SortedSuppliers{
		ProfileID:       "SPL_WEIGHT_1",
		Sorting:         utils.MetaWeight,
		SortedSuppliers: []*SortedSupplier{},
	}
	exp := ""
	rcv := eSpls.Digest()
	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("Expecting: %+v, received: %+v", exp, rcv)
	}
}

func TestLibSuppliersSortHighestCost(t *testing.T) {
	sSpls := &SortedSuppliers{
		SortedSuppliers: []*SortedSupplier{
			&SortedSupplier{
				SupplierID: "supplier1",
				SortingData: map[string]interface{}{
					utils.Cost:   0.1,
					utils.Weight: 15.0,
				},
				SupplierParameters: "param1",
			},
			&SortedSupplier{
				SupplierID: "supplier2",
				SortingData: map[string]interface{}{
					utils.Cost:   0.2,
					utils.Weight: 20.0,
				},
				SupplierParameters: "param2",
			},
			&SortedSupplier{
				SupplierID: "supplier3",
				SortingData: map[string]interface{}{
					utils.Cost:   0.05,
					utils.Weight: 10.0,
				},
				SupplierParameters: "param3",
			},
		},
	}
	sSpls.SortHighestCost()
	eOrderedSpls := &SortedSuppliers{
		SortedSuppliers: []*SortedSupplier{
			&SortedSupplier{
				SupplierID: "supplier2",
				SortingData: map[string]interface{}{
					utils.Cost:   0.2,
					utils.Weight: 20.0,
				},
				SupplierParameters: "param2",
			},
			&SortedSupplier{
				SupplierID: "supplier1",
				SortingData: map[string]interface{}{
					utils.Cost:   0.1,
					utils.Weight: 15.0,
				},
				SupplierParameters: "param1",
			},
			&SortedSupplier{
				SupplierID: "supplier3",
				SortingData: map[string]interface{}{
					utils.Cost:   0.05,
					utils.Weight: 10.0,
				},
				SupplierParameters: "param3",
			},
		},
	}
	if !reflect.DeepEqual(eOrderedSpls, sSpls) {
		t.Errorf("Expecting: %s, received: %s",
			utils.ToJSON(eOrderedSpls), utils.ToJSON(sSpls))
	}
}

func TestLibSuppliersStatMetricSort(t *testing.T) {
	sm := SplStatMetrics{
		&SplStatMetric{StatID: "SampleStat",
			metricType:  utils.MetaACD,
			MetricValue: 10.1},
		&SplStatMetric{StatID: "SampleStat2",
			metricType:  utils.MetaACD,
			MetricValue: 23.1},
		&SplStatMetric{StatID: "SampleStat3",
			metricType:  utils.MetaACD,
			MetricValue: 10.0},
	}
	sm.Sort()
	exp := SplStatMetrics{
		&SplStatMetric{StatID: "SampleStat3",
			metricType:  utils.MetaACD,
			MetricValue: 10.0},
		&SplStatMetric{StatID: "SampleStat",
			metricType:  utils.MetaACD,
			MetricValue: 10.1},
		&SplStatMetric{StatID: "SampleStat2",
			metricType:  utils.MetaACD,
			MetricValue: 23.1},
	}
	if !reflect.DeepEqual(exp, sm) {
		t.Errorf("Expecting: %s, received: %s",
			utils.ToJSON(exp), utils.ToJSON(sm))
	}
}

func TestLibSuppliersStatMetricSort2(t *testing.T) {
	sm := SplStatMetrics{
		&SplStatMetric{StatID: "SampleStat",
			metricType:  utils.MetaPDD,
			MetricValue: 10.1},
		&SplStatMetric{StatID: "SampleStat2",
			metricType:  utils.MetaPDD,
			MetricValue: 23.1},
		&SplStatMetric{StatID: "SampleStat3",
			metricType:  utils.MetaPDD,
			MetricValue: 10.0},
	}
	sm.Sort()
	exp := SplStatMetrics{
		&SplStatMetric{StatID: "SampleStat2",
			metricType:  utils.MetaPDD,
			MetricValue: 23.1},
		&SplStatMetric{StatID: "SampleStat",
			metricType:  utils.MetaPDD,
			MetricValue: 10.1},
		&SplStatMetric{StatID: "SampleStat3",
			metricType:  utils.MetaPDD,
			MetricValue: 10.0},
	}
	if !reflect.DeepEqual(exp, sm) {
		t.Errorf("Expecting: %s, received: %s",
			utils.ToJSON(exp), utils.ToJSON(sm))
	}
}

func TestLibSuppliersSplDataProvide(t *testing.T) {
	//simulatedData simulate sortingData
	simulatedData := map[string]interface{}{
		utils.Cost: 12.45,
		utils.MetaACD: SplStatMetrics{
			&SplStatMetric{
				StatID:      utils.META_NONE,
				metricType:  utils.MetaACD,
				MetricValue: 9.0},
			&SplStatMetric{
				StatID:      utils.META_NONE,
				metricType:  utils.MetaACD,
				MetricValue: 10.0},
		},
		utils.MetaPDD: SplStatMetrics{
			&SplStatMetric{
				StatID:      utils.META_NONE,
				metricType:  utils.MetaPDD,
				MetricValue: 12.0},
			&SplStatMetric{
				StatID:      utils.META_NONE,
				metricType:  utils.MetaPDD,
				MetricValue: 5.0},
		},
	}
	ev := map[string]interface{}{
		utils.Account: "1001",
	}
	sDP := newSplDataProvider(ev, simulatedData)
	exp := "1001"
	if rcv, err := sDP.FieldAsInterface([]string{utils.MetaReq, utils.Account}); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("Expecting: %+v, received: %+v", exp, rcv)
	}
	exp2 := 12.45
	if rcv, err := sDP.FieldAsInterface([]string{utils.MetaVars, utils.Cost}); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, exp2) {
		t.Errorf("Expecting: %+v, received: %+v", exp2, rcv)
	}
	exp3 := 9.0
	if rcv, err := sDP.FieldAsInterface([]string{utils.MetaVars, utils.MetaACD}); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, exp3) {
		t.Errorf("Expecting: %+v, received: %+v", exp3, rcv)
	}
	exp4 := 12.0
	if rcv, err := sDP.FieldAsInterface([]string{utils.MetaVars, utils.MetaPDD}); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, exp4) {
		t.Errorf("Expecting: %+v, received: %+v", exp4, rcv)
	}
}

//sort based on *acd and *tcd
func TestLibSuppliersSortQOS(t *testing.T) {
	sSpls := &SortedSuppliers{
		SortedSuppliers: []*SortedSupplier{
			&SortedSupplier{
				//the worst value for supplier1 for *acd is 0.5 , *tcd  1.1
				SupplierID: "supplier1",
				SortingData: map[string]interface{}{
					utils.Cost:   0.5,
					utils.Weight: 10.0,
					utils.MetaACD: SplStatMetrics{
						&SplStatMetric{StatID: "SampleStat",
							metricType:  utils.MetaACD,
							MetricValue: 0.5},
						&SplStatMetric{StatID: "SampleStat2",
							metricType:  utils.MetaACD,
							MetricValue: 23.1},
						&SplStatMetric{StatID: "SampleStat3",
							metricType:  utils.MetaACD,
							MetricValue: 10.0},
					},
					utils.MetaTCD: SplStatMetrics{
						&SplStatMetric{StatID: "SampleStat",
							metricType:  utils.MetaTCD,
							MetricValue: 1.1},
						&SplStatMetric{StatID: "SampleStat2",
							metricType:  utils.MetaTCD,
							MetricValue: 12.1},
					},
				},
			},
			&SortedSupplier{
				//the worst value for supplier2 for *acd is 0.5 , *tcd 4.1
				SupplierID: "supplier2",
				SortingData: map[string]interface{}{
					utils.Cost:   0.1,
					utils.Weight: 15.0,
					utils.MetaACD: SplStatMetrics{
						&SplStatMetric{StatID: "SampleStat",
							metricType:  utils.MetaACD,
							MetricValue: 1.1},
						&SplStatMetric{StatID: "SampleStat2",
							metricType:  utils.MetaACD,
							MetricValue: 12.1},
						&SplStatMetric{StatID: "SampleStat3",
							metricType:  utils.MetaACD,
							MetricValue: 0.5},
					},
					utils.MetaTCD: SplStatMetrics{
						&SplStatMetric{StatID: "SampleStat",
							metricType:  utils.MetaTCD,
							MetricValue: 12.1},
						&SplStatMetric{StatID: "SampleStat2",
							metricType:  utils.MetaTCD,
							MetricValue: 4.1},
					},
				},
			},
			&SortedSupplier{
				//the worst value for supplier3 for *acd is 0.4 , *tcd 5.1
				SupplierID: "supplier3",
				SortingData: map[string]interface{}{
					utils.Cost:   1.1,
					utils.Weight: 17.8,
					utils.MetaACD: SplStatMetrics{
						&SplStatMetric{StatID: "SampleStat",
							metricType:  utils.MetaACD,
							MetricValue: 0.4},
						&SplStatMetric{StatID: "SampleStat2",
							metricType:  utils.MetaACD,
							MetricValue: 2.1},
						&SplStatMetric{StatID: "SampleStat3",
							metricType:  utils.MetaACD,
							MetricValue: 0.5},
					},
					utils.MetaTCD: SplStatMetrics{
						&SplStatMetric{StatID: "SampleStat",
							metricType:  utils.MetaTCD,
							MetricValue: 6.1},
						&SplStatMetric{StatID: "SampleStat2",
							metricType:  utils.MetaTCD,
							MetricValue: 5.1},
					},
				},
			},
		},
	}

	//sort base on *acd and *tcd
	sSpls.SortQOS([]string{utils.MetaACD, utils.MetaTCD})
	rcv := make([]string, len(sSpls.SortedSuppliers))
	eIds := []string{"supplier2", "supplier1", "supplier3"}
	for i, spl := range sSpls.SortedSuppliers {
		rcv[i] = spl.SupplierID

	}
	if !reflect.DeepEqual(eIds, rcv) {
		t.Errorf("Expecting: %+v, \n received: %+v",
			eIds, rcv)
	}
}

//sort based on *acd and *tcd
func TestLibSuppliersSortQOS2(t *testing.T) {
	sSpls := &SortedSuppliers{
		SortedSuppliers: []*SortedSupplier{
			&SortedSupplier{
				//the worst value for supplier1 for *acd is 0.5 , *tcd  1.1
				SupplierID: "supplier1",
				SortingData: map[string]interface{}{
					utils.Weight: 10.0,
					utils.MetaACD: SplStatMetrics{
						&SplStatMetric{StatID: "SampleStat",
							metricType:  utils.MetaACD,
							MetricValue: 0.5},
						&SplStatMetric{StatID: "SampleStat2",
							metricType:  utils.MetaACD,
							MetricValue: 0.7},
						&SplStatMetric{StatID: "SampleStat3",
							metricType:  utils.MetaACD,
							MetricValue: 0.6},
					},
					utils.MetaTCD: SplStatMetrics{
						&SplStatMetric{StatID: "SampleStat",
							metricType:  utils.MetaTCD,
							MetricValue: 1.1},
						&SplStatMetric{StatID: "SampleStat2",
							metricType:  utils.MetaTCD,
							MetricValue: 12.1},
					},
				},
			},
			&SortedSupplier{
				//the worst value for supplier1 for *acd is 0.5 , *tcd  1.1
				//supplier1 and supplier2 have the same value for *acd and *tcd
				//will be sorted based on weight
				SupplierID: "supplier2",
				SortingData: map[string]interface{}{
					utils.Weight: 17.0,
					utils.MetaACD: SplStatMetrics{
						&SplStatMetric{StatID: "SampleStat",
							metricType:  utils.MetaACD,
							MetricValue: 0.5},
						&SplStatMetric{StatID: "SampleStat2",
							metricType:  utils.MetaACD,
							MetricValue: 0.5},
						&SplStatMetric{StatID: "SampleStat3",
							metricType:  utils.MetaACD,
							MetricValue: 0.5},
					},
					utils.MetaTCD: SplStatMetrics{
						&SplStatMetric{StatID: "SampleStat",
							metricType:  utils.MetaTCD,
							MetricValue: 1.1},
						&SplStatMetric{StatID: "SampleStat2",
							metricType:  utils.MetaTCD,
							MetricValue: 12.1},
					},
				},
			},
			&SortedSupplier{
				//the worst value for supplier1 for *acd is 0.5 , *tcd  1.1
				SupplierID: "supplier3",
				SortingData: map[string]interface{}{
					utils.Cost:   0.5,
					utils.Weight: 10.0,
					utils.MetaACD: SplStatMetrics{
						&SplStatMetric{StatID: "SampleStat",
							metricType:  utils.MetaACD,
							MetricValue: 0.5},
						&SplStatMetric{StatID: "SampleStat2",
							metricType:  utils.MetaACD,
							MetricValue: 0.5},
						&SplStatMetric{StatID: "SampleStat3",
							metricType:  utils.MetaACD,
							MetricValue: 0.5},
					},
					utils.MetaTCD: SplStatMetrics{
						&SplStatMetric{StatID: "SampleStat",
							metricType:  utils.MetaTCD,
							MetricValue: 1.2},
					},
				},
			},
		},
	}
	sSpls.SortQOS([]string{utils.MetaACD, utils.MetaTCD})
	rcv := make([]string, len(sSpls.SortedSuppliers))
	eIds := []string{"supplier3", "supplier2", "supplier1"}
	for i, spl := range sSpls.SortedSuppliers {
		rcv[i] = spl.SupplierID

	}
	if !reflect.DeepEqual(eIds, rcv) {
		t.Errorf("Expecting: %+v, \n received: %+v",
			eIds, rcv)
	}
}

//sort based on *pdd
func TestLibSuppliersSortQOS3(t *testing.T) {
	sSpls := &SortedSuppliers{
		SortedSuppliers: []*SortedSupplier{
			&SortedSupplier{
				//the worst value for supplier1 for *pdd is 0.7 , *tcd  1.1
				//supplier1 and supplier3 have the same value for *pdd
				//will be sorted based on weight
				SupplierID: "supplier1",
				SortingData: map[string]interface{}{
					utils.Weight: 15.0,
					utils.MetaPDD: SplStatMetrics{
						&SplStatMetric{StatID: "SampleStat",
							metricType:  utils.MetaPDD,
							MetricValue: 0.5},
						&SplStatMetric{StatID: "SampleStat2",
							metricType:  utils.MetaACD,
							MetricValue: 0.7},
					},
					utils.MetaTCD: SplStatMetrics{
						&SplStatMetric{StatID: "SampleStat",
							metricType:  utils.MetaTCD,
							MetricValue: 1.1},
						&SplStatMetric{StatID: "SampleStat2",
							metricType:  utils.MetaTCD,
							MetricValue: 12.1},
					},
				},
			},
			&SortedSupplier{
				//the worst value for supplier2 for *pdd is 1.2, *tcd  1.1
				SupplierID: "supplier2",
				SortingData: map[string]interface{}{
					utils.Weight: 10.0,
					utils.MetaPDD: SplStatMetrics{
						&SplStatMetric{StatID: "SampleStat",
							metricType:  utils.MetaPDD,
							MetricValue: 0.9},
						&SplStatMetric{StatID: "SampleStat2",
							metricType:  utils.MetaACD,
							MetricValue: 1.2},
					},
					utils.MetaTCD: SplStatMetrics{
						&SplStatMetric{StatID: "SampleStat",
							metricType:  utils.MetaTCD,
							MetricValue: 1.1},
					},
				},
			},
			&SortedSupplier{
				//the worst value for supplier3 for *pdd is 0.7, *tcd  10.1
				SupplierID: "supplier3",
				SortingData: map[string]interface{}{
					utils.Weight: 10.0,
					utils.MetaPDD: SplStatMetrics{
						&SplStatMetric{StatID: "SampleStat",
							metricType:  utils.MetaPDD,
							MetricValue: 0.7},
					},
					utils.MetaTCD: SplStatMetrics{
						&SplStatMetric{StatID: "SampleStat",
							metricType:  utils.MetaTCD,
							MetricValue: 10.1},
					},
				},
			},
		},
	}
	sSpls.SortQOS([]string{utils.MetaPDD})
	rcv := make([]string, len(sSpls.SortedSuppliers))
	eIds := []string{"supplier1", "supplier3", "supplier2"}
	for i, spl := range sSpls.SortedSuppliers {
		rcv[i] = spl.SupplierID

	}
	if !reflect.DeepEqual(eIds, rcv) {
		t.Errorf("Expecting: %+v, \n received: %+v",
			eIds, rcv)
	}
}

func TestLibSuppliersSortQOS4(t *testing.T) {
	sSpls := &SortedSuppliers{
		SortedSuppliers: []*SortedSupplier{
			&SortedSupplier{
				SupplierID: "supplier1",
				SortingData: map[string]interface{}{
					utils.MetaACD: SplStatMetrics{
						&SplStatMetric{StatID: "SampleStat",
							metricType:  utils.MetaACD,
							MetricValue: 0.2},
					},
					utils.MetaTCD: SplStatMetrics{
						&SplStatMetric{StatID: "SampleStat",
							metricType:  utils.MetaTCD,
							MetricValue: 15.0},
					},
					utils.MetaASR: SplStatMetrics{
						&SplStatMetric{StatID: "SampleStat",
							metricType:  utils.MetaASR,
							MetricValue: 1.2},
					},
				},
			},
			&SortedSupplier{
				SupplierID: "supplier2",
				SortingData: map[string]interface{}{
					utils.MetaACD: SplStatMetrics{
						&SplStatMetric{StatID: "SampleStat",
							metricType:  utils.MetaACD,
							MetricValue: 0.2},
					},
					utils.MetaTCD: SplStatMetrics{
						&SplStatMetric{StatID: "SampleStat",
							metricType:  utils.MetaTCD,
							MetricValue: 20.0},
					},
					utils.MetaASR: SplStatMetrics{
						&SplStatMetric{StatID: "SampleStat",
							metricType:  utils.MetaASR,
							MetricValue: -1.0},
					},
				},
			},
			&SortedSupplier{
				SupplierID: "supplier3",
				SortingData: map[string]interface{}{
					utils.MetaACD: SplStatMetrics{
						&SplStatMetric{StatID: "SampleStat",
							metricType:  utils.MetaACD,
							MetricValue: 0.1},
					},
					utils.MetaTCD: SplStatMetrics{
						&SplStatMetric{StatID: "SampleStat",
							metricType:  utils.MetaTCD,
							MetricValue: 10.0},
					},
					utils.MetaASR: SplStatMetrics{
						&SplStatMetric{StatID: "SampleStat",
							metricType:  utils.MetaASR,
							MetricValue: 1.2},
					},
				},
			},
		},
	}
	sSpls.SortQOS([]string{utils.MetaASR, utils.MetaACD, utils.MetaTCD})
	rcv := make([]string, len(sSpls.SortedSuppliers))
	eIds := []string{"supplier1", "supplier3", "supplier2"}
	for i, spl := range sSpls.SortedSuppliers {
		rcv[i] = spl.SupplierID

	}
	if !reflect.DeepEqual(eIds, rcv) {
		t.Errorf("Expecting: %+v, \n received: %+v",
			eIds, rcv)
	}
}

func TestLibSuppliersSortQOS5(t *testing.T) {
	sSpls := &SortedSuppliers{
		SortedSuppliers: []*SortedSupplier{
			&SortedSupplier{
				SupplierID: "supplier1",
				SortingData: map[string]interface{}{
					utils.MetaACD: SplStatMetrics{
						&SplStatMetric{StatID: "SampleStat",
							metricType:  utils.MetaACD,
							MetricValue: 0.2},
					},
					utils.MetaTCD: SplStatMetrics{
						&SplStatMetric{StatID: "SampleStat",
							metricType:  utils.MetaTCD,
							MetricValue: 15.0},
					},
					utils.MetaASR: SplStatMetrics{
						&SplStatMetric{StatID: "SampleStat",
							metricType:  utils.MetaASR,
							MetricValue: -1.0},
					},
					utils.MetaTCC: SplStatMetrics{
						&SplStatMetric{StatID: "SampleStat",
							metricType:  utils.MetaTCC,
							MetricValue: 10.1},
					},
				},
			},
			&SortedSupplier{
				SupplierID: "supplier2",
				SortingData: map[string]interface{}{
					utils.MetaACD: SplStatMetrics{
						&SplStatMetric{StatID: "SampleStat",
							metricType:  utils.MetaACD,
							MetricValue: 0.2},
					},
					utils.MetaTCD: SplStatMetrics{
						&SplStatMetric{StatID: "SampleStat",
							metricType:  utils.MetaTCD,
							MetricValue: 20.0},
					},
					utils.MetaASR: SplStatMetrics{
						&SplStatMetric{StatID: "SampleStat",
							metricType:  utils.MetaASR,
							MetricValue: 1.2},
					},
					utils.MetaTCC: SplStatMetrics{
						&SplStatMetric{StatID: "SampleStat",
							metricType:  utils.MetaTCC,
							MetricValue: 10.1},
					},
				},
			},
			&SortedSupplier{
				SupplierID: "supplier3",
				SortingData: map[string]interface{}{
					utils.MetaACD: SplStatMetrics{
						&SplStatMetric{StatID: "SampleStat",
							metricType:  utils.MetaACD,
							MetricValue: 0.1},
					},
					utils.MetaTCD: SplStatMetrics{
						&SplStatMetric{StatID: "SampleStat",
							metricType:  utils.MetaTCD,
							MetricValue: 10.0},
					},
					utils.MetaASR: SplStatMetrics{
						&SplStatMetric{StatID: "SampleStat",
							metricType:  utils.MetaASR,
							MetricValue: 1.2},
					},
					utils.MetaTCC: SplStatMetrics{
						&SplStatMetric{StatID: "SampleStat",
							metricType:  utils.MetaTCC,
							MetricValue: 10.1},
					},
				},
			},
		},
	}
	sSpls.SortQOS([]string{utils.MetaTCC, utils.MetaASR, utils.MetaACD, utils.MetaTCD})
	rcv := make([]string, len(sSpls.SortedSuppliers))
	eIds := []string{"supplier2", "supplier3", "supplier1"}
	for i, spl := range sSpls.SortedSuppliers {
		rcv[i] = spl.SupplierID

	}
	if !reflect.DeepEqual(eIds, rcv) {
		t.Errorf("Expecting: %+v, \n received: %+v",
			eIds, rcv)
	}
}

func TestLibSuppliersSortQOS6(t *testing.T) {
	sSpls := &SortedSuppliers{
		SortedSuppliers: []*SortedSupplier{
			&SortedSupplier{
				SupplierID: "supplier1",
				SortingData: map[string]interface{}{
					utils.Weight: 15.0,
					utils.MetaACD: SplStatMetrics{
						&SplStatMetric{StatID: "SampleStat",
							metricType:  utils.MetaACD,
							MetricValue: 0.2},
					},
				},
			},
			&SortedSupplier{
				SupplierID: "supplier2",
				SortingData: map[string]interface{}{
					utils.Weight: 25.0,
					utils.MetaACD: SplStatMetrics{
						&SplStatMetric{StatID: "SampleStat",
							metricType:  utils.MetaACD,
							MetricValue: 0.2},
					},
				},
			},
			&SortedSupplier{
				SupplierID: "supplier3",
				SortingData: map[string]interface{}{
					utils.Weight: 20.0,
					utils.MetaACD: SplStatMetrics{
						&SplStatMetric{StatID: "SampleStat",
							metricType:  utils.MetaACD,
							MetricValue: 0.1},
					},
				},
			},
		},
	}
	sSpls.SortQOS([]string{utils.MetaACD})
	rcv := make([]string, len(sSpls.SortedSuppliers))
	eIds := []string{"supplier2", "supplier1", "supplier3"}
	for i, spl := range sSpls.SortedSuppliers {
		rcv[i] = spl.SupplierID

	}
	if !reflect.DeepEqual(eIds, rcv) {
		t.Errorf("Expecting: %+v, \n received: %+v",
			eIds, rcv)
	}
}

func TestLibSuppliersSortQOS7(t *testing.T) {
	sSpls := &SortedSuppliers{
		SortedSuppliers: []*SortedSupplier{
			&SortedSupplier{
				SupplierID: "supplier1",
				SortingData: map[string]interface{}{
					utils.Weight: 15.0,
					utils.MetaACD: SplStatMetrics{
						&SplStatMetric{StatID: "SampleStat",
							metricType:  utils.MetaACD,
							MetricValue: -1.0},
					},
				},
			},
			&SortedSupplier{
				SupplierID: "supplier2",
				SortingData: map[string]interface{}{
					utils.Weight: 25.0,
					utils.MetaACD: SplStatMetrics{
						&SplStatMetric{StatID: "SampleStat",
							metricType:  utils.MetaACD,
							MetricValue: -1.0},
					},
				},
			},
			&SortedSupplier{
				SupplierID: "supplier3",
				SortingData: map[string]interface{}{
					utils.Weight: 20.0,
					utils.MetaACD: SplStatMetrics{
						&SplStatMetric{StatID: "SampleStat",
							metricType:  utils.MetaACD,
							MetricValue: -1.0},
					},
				},
			},
		},
	}
	sSpls.SortQOS([]string{utils.MetaACD})
	rcv := make([]string, len(sSpls.SortedSuppliers))
	eIds := []string{"supplier2", "supplier3", "supplier1"}
	for i, spl := range sSpls.SortedSuppliers {
		rcv[i] = spl.SupplierID

	}
	if !reflect.DeepEqual(eIds, rcv) {
		t.Errorf("Expecting: %+v, \n received: %+v",
			eIds, rcv)
	}
}

func TestLibSuppliersSortQOS8(t *testing.T) {
	sSpls := &SortedSuppliers{
		SortedSuppliers: []*SortedSupplier{
			&SortedSupplier{
				SupplierID: "supplier1",
				SortingData: map[string]interface{}{
					utils.Weight: 15.0,
					utils.MetaACD: SplStatMetrics{
						&SplStatMetric{StatID: "SampleStat",
							metricType:  utils.MetaACD,
							MetricValue: -1.0},
					},
				},
			},
			&SortedSupplier{
				SupplierID: "supplier2",
				SortingData: map[string]interface{}{
					utils.Weight: 25.0,
					utils.MetaACD: SplStatMetrics{
						&SplStatMetric{StatID: "SampleStat",
							metricType:  utils.MetaACD,
							MetricValue: -1.0},
					},
				},
				SupplierParameters: "param2",
			},
			&SortedSupplier{
				SupplierID: "supplier3",
				SortingData: map[string]interface{}{
					utils.Weight: 20.0,
					utils.MetaACD: SplStatMetrics{
						&SplStatMetric{StatID: "SampleStat",
							metricType:  utils.MetaACD,
							MetricValue: 10.0},
					},
				},
			},
		},
	}
	sSpls.SortQOS([]string{utils.MetaACD})
	rcv := make([]string, len(sSpls.SortedSuppliers))
	eIds := []string{"supplier3", "supplier2", "supplier1"}
	for i, spl := range sSpls.SortedSuppliers {
		rcv[i] = spl.SupplierID

	}
	if !reflect.DeepEqual(eIds, rcv) {
		t.Errorf("Expecting: %+v, \n received: %+v",
			eIds, rcv)
	}
}
