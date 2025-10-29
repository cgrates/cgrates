/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/
package engine

import (
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

func TestLibSuppliersSortCost(t *testing.T) {
	sSpls := &SortedSuppliers{
		SortedSuppliers: []*SortedSupplier{
			{
				SupplierID: "supplier1",
				SortingData: map[string]any{
					utils.Cost:   0.1,
					utils.Weight: 10.0,
				},
				SupplierParameters: "param1",
			},
			{
				SupplierID: "supplier2",
				SortingData: map[string]any{
					utils.Cost:   0.1,
					utils.Weight: 20.0,
				},
				SupplierParameters: "param2",
			},
			{
				SupplierID: "supplier3",
				SortingData: map[string]any{
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
			{
				SupplierID: "supplier3",
				SortingData: map[string]any{
					utils.Cost:   0.05,
					utils.Weight: 10.0,
				},
				SupplierParameters: "param3",
			},
			{
				SupplierID: "supplier2",
				SortingData: map[string]any{
					utils.Cost:   0.1,
					utils.Weight: 20.0,
				},
				SupplierParameters: "param2",
			},
			{
				SupplierID: "supplier1",
				SortingData: map[string]any{
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
			{
				SupplierID: "supplier1",
				SortingData: map[string]any{
					utils.Weight: 10.0,
				},
				SupplierParameters: "param1",
			},
			{
				SupplierID: "supplier2",
				SortingData: map[string]any{
					utils.Weight: 20.0,
				},
				SupplierParameters: "param2",
			},
			{
				SupplierID: "supplier3",
				SortingData: map[string]any{
					utils.Weight: 10.5,
				},
				SupplierParameters: "param3",
			},
		},
	}
	sSpls.SortWeight()
	eOrderedSpls := &SortedSuppliers{
		SortedSuppliers: []*SortedSupplier{
			{
				SupplierID: "supplier2",
				SortingData: map[string]any{
					utils.Weight: 20.0,
				},
				SupplierParameters: "param2",
			},
			{
				SupplierID: "supplier3",
				SortingData: map[string]any{
					utils.Weight: 10.5,
				},
				SupplierParameters: "param3",
			},
			{
				SupplierID: "supplier1",
				SortingData: map[string]any{
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
			{
				SupplierID: "supplier2",
				SortingData: map[string]any{
					"Weight": 20.0,
				},
				SupplierParameters: "param2",
			},
			{
				SupplierID: "supplier1",
				SortingData: map[string]any{
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
			{
				SupplierID: "supplier1",
				SortingData: map[string]any{
					"Weight": 30.0,
				},
				SupplierParameters: "param1",
			},
			{
				SupplierID: "supplier2",
				SortingData: map[string]any{
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
			{
				SupplierID: "supplier1",
				SortingData: map[string]any{
					utils.Cost:   0.1,
					utils.Weight: 15.0,
				},
				SupplierParameters: "param1",
			},
			{
				SupplierID: "supplier2",
				SortingData: map[string]any{
					utils.Cost:   0.2,
					utils.Weight: 20.0,
				},
				SupplierParameters: "param2",
			},
			{
				SupplierID: "supplier3",
				SortingData: map[string]any{
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
			{
				SupplierID: "supplier2",
				SortingData: map[string]any{
					utils.Cost:   0.2,
					utils.Weight: 20.0,
				},
				SupplierParameters: "param2",
			},
			{
				SupplierID: "supplier1",
				SortingData: map[string]any{
					utils.Cost:   0.1,
					utils.Weight: 15.0,
				},
				SupplierParameters: "param1",
			},
			{
				SupplierID: "supplier3",
				SortingData: map[string]any{
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

// sort based on *acd and *tcd
func TestLibSuppliersSortQOS(t *testing.T) {
	sSpls := &SortedSuppliers{
		SortedSuppliers: []*SortedSupplier{
			{
				//the average value for supplier1 for *acd is 0.5 , *tcd  1.1
				SupplierID: "supplier1",
				SortingData: map[string]any{
					utils.Cost:    0.5,
					utils.Weight:  10.0,
					utils.MetaACD: 0.5,
					utils.MetaTCD: 1.1,
				},
			},
			{
				//the average value for supplier2 for *acd is 0.5 , *tcd 4.1
				SupplierID: "supplier2",
				SortingData: map[string]any{
					utils.Cost:    0.1,
					utils.Weight:  15.0,
					utils.MetaACD: 0.5,
					utils.MetaTCD: 4.1,
				},
			},
			{
				//the average value for supplier3 for *acd is 0.4 , *tcd 5.1
				SupplierID: "supplier3",
				SortingData: map[string]any{
					utils.Cost:    1.1,
					utils.Weight:  17.8,
					utils.MetaACD: 0.4,
					utils.MetaTCD: 5.1,
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

// sort based on *acd and *tcd
func TestLibSuppliersSortQOS2(t *testing.T) {
	sSpls := &SortedSuppliers{
		SortedSuppliers: []*SortedSupplier{
			{
				//the average value for supplier1 for *acd is 0.5 , *tcd  1.1
				SupplierID: "supplier1",
				SortingData: map[string]any{
					utils.Weight:  10.0,
					utils.MetaACD: 0.5,
					utils.MetaTCD: 1.1,
				},
			},
			{
				//the worst value for supplier1 for *acd is 0.5 , *tcd  1.1
				//supplier1 and supplier2 have the same value for *acd and *tcd
				//will be sorted based on weight
				SupplierID: "supplier2",
				SortingData: map[string]any{
					utils.Weight:  17.0,
					utils.MetaACD: 0.5,
					utils.MetaTCD: 1.1,
				},
			},
			{

				SupplierID: "supplier3",
				SortingData: map[string]any{
					utils.Cost:    0.5,
					utils.Weight:  10.0,
					utils.MetaACD: 0.7,
					utils.MetaTCD: 1.1,
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

// sort based on *pdd
func TestLibSuppliersSortQOS3(t *testing.T) {
	sSpls := &SortedSuppliers{
		SortedSuppliers: []*SortedSupplier{
			{
				//the worst value for supplier1 for *pdd is 0.7 , *tcd  1.1
				//supplier1 and supplier3 have the same value for *pdd
				//will be sorted based on weight
				SupplierID: "supplier1",
				SortingData: map[string]any{
					utils.Weight:  15.0,
					utils.MetaPDD: 0.7,
					utils.MetaTCD: 1.1,
				},
			},
			{
				//the worst value for supplier2 for *pdd is 1.2, *tcd  1.1
				SupplierID: "supplier2",
				SortingData: map[string]any{
					utils.Weight:  10.0,
					utils.MetaPDD: 1.2,
					utils.MetaTCD: 1.1,
				},
			},
			{
				//the worst value for supplier3 for *pdd is 0.7, *tcd  10.1
				SupplierID: "supplier3",
				SortingData: map[string]any{
					utils.Weight:  10.0,
					utils.MetaPDD: 0.7,
					utils.MetaTCD: 10.1,
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
			{
				SupplierID: "supplier1",
				SortingData: map[string]any{
					utils.MetaACD: 0.2,
					utils.MetaTCD: 15.0,
					utils.MetaASR: 1.2,
				},
			},
			{
				SupplierID: "supplier2",
				SortingData: map[string]any{
					utils.MetaACD: 0.2,
					utils.MetaTCD: 20.0,
					utils.MetaASR: -1.0,
				},
			},
			{
				SupplierID: "supplier3",
				SortingData: map[string]any{
					utils.MetaACD: 0.1,
					utils.MetaTCD: 10.0,
					utils.MetaASR: 1.2,
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
			{
				SupplierID: "supplier1",
				SortingData: map[string]any{
					utils.MetaACD: 0.2,
					utils.MetaTCD: 15.0,
					utils.MetaASR: -1.0,
					utils.MetaTCC: 10.1,
				},
			},
			{
				SupplierID: "supplier2",
				SortingData: map[string]any{
					utils.MetaACD: 0.2,
					utils.MetaTCD: 20.0,
					utils.MetaASR: 1.2,
					utils.MetaTCC: 10.1,
				},
			},
			{
				SupplierID: "supplier3",
				SortingData: map[string]any{
					utils.MetaACD: 0.1,
					utils.MetaTCD: 10.0,
					utils.MetaASR: 1.2,
					utils.MetaTCC: 10.1,
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
			{
				SupplierID: "supplier1",
				SortingData: map[string]any{
					utils.Weight:  15.0,
					utils.MetaACD: 0.2,
				},
			},
			{
				SupplierID: "supplier2",
				SortingData: map[string]any{
					utils.Weight:  25.0,
					utils.MetaACD: 0.2,
				},
			},
			{
				SupplierID: "supplier3",
				SortingData: map[string]any{
					utils.Weight:  20.0,
					utils.MetaACD: 0.1,
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
			{
				SupplierID: "supplier1",
				SortingData: map[string]any{
					utils.Weight:  15.0,
					utils.MetaACD: -1.0,
				},
			},
			{
				SupplierID: "supplier2",
				SortingData: map[string]any{
					utils.Weight:  25.0,
					utils.MetaACD: -1.0,
				},
			},
			{
				SupplierID: "supplier3",
				SortingData: map[string]any{
					utils.Weight:  20.0,
					utils.MetaACD: -1.0,
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
			{
				SupplierID: "supplier1",
				SortingData: map[string]any{
					utils.Weight:  15.0,
					utils.MetaACD: -1.0,
				},
			},
			{
				SupplierID: "supplier2",
				SortingData: map[string]any{
					utils.Weight:  25.0,
					utils.MetaACD: -1.0,
				},
			},
			{
				SupplierID: "supplier3",
				SortingData: map[string]any{
					utils.Weight:  20.0,
					utils.MetaACD: 10.0,
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

func TestLibSuppliersSortLoadDistribution(t *testing.T) {
	sSpls := &SortedSuppliers{
		SortedSuppliers: []*SortedSupplier{
			{
				SupplierID: "supplier1",
				SortingData: map[string]any{
					utils.Weight: 25.0,
					utils.Ratio:  4.0,
					utils.Load:   3.0,
				},
			},
			{
				SupplierID: "supplier2",
				SortingData: map[string]any{
					utils.Weight: 15.0,
					utils.Ratio:  10.0,
					utils.Load:   5.0,
				},
			},
			{
				SupplierID: "supplier3",
				SortingData: map[string]any{
					utils.Weight: 25.0,
					utils.Ratio:  1.0,
					utils.Load:   1.0,
				},
			},
		},
	}
	sSpls.SortLoadDistribution()
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

func TestSortSuppliersSortResourceAscendent(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	cfg.SupplierSCfg().ResourceSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResources)}
	cfg.SupplierSCfg().RALsConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaRALs)}
	clientConn := make(chan birpc.ClientConnector, 1)
	clientConn <- &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.ResourceSv1GetResource: func(ctx *context.Context, args, reply any) error {
				res := Resource{
					ID:     "ResourceSupplier2",
					Tenant: "cgrates.org",
					tUsage: utils.Float64Pointer(46.2),
				}
				*reply.(*Resource) = res
				return nil
			},
			utils.ResponderGetCostOnRatingPlans: func(ctx *context.Context, args, reply any) error {
				rpl := map[string]any{
					utils.Cost: 23.1,
				}
				*reply.(*map[string]any) = rpl
				return nil
			},
		},
	}
	connMgr := NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResources): clientConn,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaRALs):      clientConn,
	})
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	spS, err := NewSupplierService(dm, nil, cfg, connMgr)
	if err != nil {
		t.Error(err)
	}
	rscAscSort := NewResourceAscendetSorter(spS)
	rscDscSort := NewResourceDescendentSorter(spS)
	qosSuppSort := NewQOSSupplierSorter(spS)
	lDistSort := NewLoadDistributionSorter(spS)
	lCostSort := NewLeastCostSorter(spS)
	hCostSort := NewHighestCostSorter(spS)
	suppls := []*Supplier{
		{
			ID:            "supplier2",
			ResourceIDs:   []string{"ResourceSupplier2"},
			StatIDs:       []string{"Stat1"},
			RatingPlanIDs: []string{"RPL_1"},
			Weight:        20,
			Blocker:       false,
			cacheSupplier: map[string]any{
				utils.MetaRatio: 3.3,
			},
		},
		{
			ID:            "supplier3",
			ResourceIDs:   []string{"ResourceSupplier3"},
			StatIDs:       []string{"Stat2"},
			RatingPlanIDs: []string{"RPL_2"},
			Weight:        35,
			Blocker:       false,
			cacheSupplier: map[string]any{
				utils.MetaRatio: 2.4,
			},
		},
	}
	suplEv := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "utils.CGREvent1",
		Event: map[string]any{
			"Account":        "1001",
			"Destination":    "1002",
			"Supplier":       "SupplierProfile2",
			utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
			"UsageInterval":  "1s",
			"PddInterval":    "1s",
			utils.SetupTime:  time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
			"Weight":         "20.0",
		},
	}
	spl := &optsGetSuppliers{
		maxCost: 25.0,
		sortingParameters: []string{
			"ResourceUsage",
		},
		sortingStragety: utils.MetaLoad,
	}
	if _, err := rscAscSort.SortSuppliers("ASC_SUPP", suppls, suplEv, spl, nil); err != nil {
		t.Error(err)
	}
	if _, err := rscDscSort.SortSuppliers("DSC_SUPP", suppls, suplEv, spl, nil); err != nil {
		t.Error(err)
	}
	if _, err := qosSuppSort.SortSuppliers("QOS_SUPP", suppls, suplEv, spl, nil); err != nil {
		t.Error(err)
	}
	if _, err := lDistSort.SortSuppliers("LDS_SUPP", suppls, suplEv, spl, nil); err != nil {
		t.Error(err)
	}
	if _, err := lCostSort.SortSuppliers("LST_SUPP", suppls, suplEv, spl, nil); err != nil {
		t.Error(err)
	}
	if _, err := hCostSort.SortSuppliers("HCS_SUPP", suppls, suplEv, spl, nil); err != nil {
		t.Error(err)
	}
}

func TestPopulateSortingDataStatMetrics(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	cfg.SupplierSCfg().StatSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStatS)}
	clientConn := make(chan birpc.ClientConnector, 1)
	clientConn <- &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.StatSv1GetQueueFloatMetrics: func(ctx *context.Context, args, reply any) error {
				rpl := map[string]float64{
					"metric1": 22.1,
				}
				*reply.(*map[string]float64) = rpl
				return nil
			},
		},
	}
	connMgr := NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStatS): clientConn,
	})
	spS, err := NewSupplierService(dm, nil, cfg, connMgr)
	if err != nil {
		t.Error(err)
	}

	suppls := &Supplier{
		ID:          "supplier2",
		ResourceIDs: []string{"ResourceSupplier2"},
		StatIDs:     []string{"Stat1"},
		Weight:      20,
		Blocker:     false,
		cacheSupplier: map[string]any{
			utils.MetaRatio: 3.3,
		},
	}

	suplEv := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "utils.CGREvent1",
		Event: map[string]any{
			"Account":        "1001",
			"Destination":    "1002",
			"Supplier":       "SupplierProfile2",
			utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
			"UsageInterval":  "1s",
			"PddInterval":    "1s",
			utils.SetupTime:  time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
			"Weight":         "20.0",
		},
	}
	spl := &optsGetSuppliers{
		maxCost: 25.0,
		sortingParameters: []string{
			"ResourceUsage",
		},
		sortingStragety: utils.MetaReload,
	}

	if _, pass, err := spS.populateSortingData(suplEv, suppls, spl, nil); err != nil || !pass {
		t.Error(err)
	}
	spl.sortingStragety = utils.MetaLoad
	if _, pass, err := spS.populateSortingData(suplEv, suppls, spl, nil); err != nil || !pass {
		t.Error(err)
	}
}

func TestLibSuppliersSupplierIDs(t *testing.T) {
	sSpls := &SortedSuppliers{
		ProfileID: str,
		Sorting:   str,
		Count:     nm,
		SortedSuppliers: []*SortedSupplier{
			{
				SupplierID:         str,
				SupplierParameters: str,
				SortingData:        map[string]any{str: str},
			},
		},
	}

	rcv := sSpls.SupplierIDs()
	exp := []string{str}

	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("expected %s, received %s", utils.ToJSON(exp), utils.ToJSON(rcv))
	}
}

func TestLibSupplierAsNavigableMap(t *testing.T) {
	ss := &SortedSupplier{
		SupplierID:         str,
		SupplierParameters: str,
		SortingData:        map[string]any{str: str},
	}

	rcv := ss.AsNavigableMap()
	exp := utils.NavigableMap2{
		"SupplierID":         utils.NewNMData(ss.SupplierID),
		"SupplierParameters": utils.NewNMData(ss.SupplierParameters),
		"SortingData":        utils.NavigableMap2{str: utils.NewNMData(str)},
	}

	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("expected %s, received %s", utils.ToJSON(exp), utils.ToJSON(rcv))
	}
}

func TestLibSuppliersAsNavigableMap(t *testing.T) {
	sSpls := &SortedSuppliers{
		ProfileID: str,
		Sorting:   str,
		Count:     nm,
		SortedSuppliers: []*SortedSupplier{
			{
				SupplierID:         str,
				SupplierParameters: str,
				SortingData:        map[string]any{str: str},
			},
		},
	}

	rcv := sSpls.AsNavigableMap()
	exp := utils.NavigableMap2{
		"ProfileID":       utils.NewNMData(sSpls.ProfileID),
		"Sorting":         utils.NewNMData(sSpls.Sorting),
		"Count":           utils.NewNMData(sSpls.Count),
		"SortedSuppliers": &utils.NMSlice{sSpls.SortedSuppliers[0].AsNavigableMap()},
	}

	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("expected %s, received %s", utils.ToJSON(exp), utils.ToJSON(rcv))
	}
}

func TestLibSuppliersSortSuppliers(t *testing.T) {
	ssd := SupplierSortDispatcher{}

	_, err := ssd.SortSuppliers(str, str, []*Supplier{}, &utils.CGREvent{}, &optsGetSuppliers{}, nil)

	if err != nil {
		if err.Error() != "unsupported sorting strategy: test" {
			t.Error(err)
		}
	}
}
