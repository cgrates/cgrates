// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package engine

import (
	"testing"

	"github.com/cgrates/cgrates/utils"
)

func TestSplsWeightSortSuppliers(t *testing.T) {
	str := "test"
	ws := &WeightSorter{
		sorting: str,
		spS:     &SupplierService{},
	}
	slc := []string{str}
	suppls := []*Supplier{{
		ID:                 str,
		FilterIDs:          slc,
		AccountIDs:         slc,
		RatingPlanIDs:      slc,
		ResourceIDs:        slc,
		StatIDs:            slc,
		Weight:             1.2,
		Blocker:            true,
		SupplierParameters: str,

		cacheSupplier: map[string]any{"test": 1},
	}}
	suplEv := &utils.CGREvent{
		Tenant: str,
		ID:     str,
		Time:   &tm,
		Event:  map[string]any{"AnswerTime": fl},
	}
	extraOpts := &optsGetSuppliers{
		ignoreErrors:      false,
		maxCost:           1.2,
		sortingParameters: slc,
		sortingStragety:   str,
	}
	argDsp := &utils.ArgDispatcher{
		APIKey:  &str,
		RouteID: &str,
	}

	rcv, err := ws.SortSuppliers(str, suppls, suplEv, extraOpts, argDsp)

	if err != nil {
		if err.Error() != "MANDATORY_IE_MISSING: [Account]" {
			t.Fatal(err)
		}
	} else {
		t.Fatal("was expectng an error")
	}

	if rcv != nil {
		t.Error(rcv)
	}
}

func TestSplsWeightSortSuppliersResourceDescendentSorter(t *testing.T) {
	str := "test"
	ws := &ResourceDescendentSorter{
		sorting: str,
		spS:     &SupplierService{},
	}
	slc := []string{str}
	suppls := []*Supplier{{
		ID:                 str,
		FilterIDs:          slc,
		AccountIDs:         slc,
		RatingPlanIDs:      slc,
		ResourceIDs:        slc,
		StatIDs:            slc,
		Weight:             1.2,
		Blocker:            true,
		SupplierParameters: str,

		cacheSupplier: map[string]any{"test": 1},
	}}
	suplEv := &utils.CGREvent{
		Tenant: str,
		ID:     str,
		Time:   &tm,
		Event:  map[string]any{"AnswerTime": fl},
	}
	extraOpts := &optsGetSuppliers{
		ignoreErrors:      false,
		maxCost:           1.2,
		sortingParameters: slc,
		sortingStragety:   str,
	}
	argDsp := &utils.ArgDispatcher{
		APIKey:  &str,
		RouteID: &str,
	}

	rcv, err := ws.SortSuppliers(str, suppls, suplEv, extraOpts, argDsp)

	if err != nil {
		if err.Error() != "MANDATORY_IE_MISSING: [Account]" {
			t.Fatal(err)
		}
	} else {
		t.Fatal("was expectng an error")
	}

	if rcv != nil {
		t.Error(rcv)
	}

	slc2 := []string{}
	suppls2 := []*Supplier{{
		ID:                 str,
		FilterIDs:          slc,
		AccountIDs:         slc,
		RatingPlanIDs:      slc,
		ResourceIDs:        slc2,
		StatIDs:            slc,
		Weight:             1.2,
		Blocker:            true,
		SupplierParameters: str,

		cacheSupplier: map[string]any{"test": 1},
	}}
	rcv, err = ws.SortSuppliers(str, suppls2, suplEv, extraOpts, argDsp)

	if err != nil {
		if err.Error() != "MANDATORY_IE_MISSING: [ResourceIDs]" {
			t.Fatal(err)
		}
	} else {
		t.Fatal("was expectng an error")
	}

	if rcv != nil {
		t.Error(rcv)
	}
}

func TestSplsWeightSortSuppliersResourceAscendentSorter(t *testing.T) {
	str := "test"
	ws := &ResourceAscendentSorter{
		sorting: str,
		spS:     &SupplierService{},
	}
	slc := []string{str}
	suppls := []*Supplier{{
		ID:                 str,
		FilterIDs:          slc,
		AccountIDs:         slc,
		RatingPlanIDs:      slc,
		ResourceIDs:        slc,
		StatIDs:            slc,
		Weight:             1.2,
		Blocker:            true,
		SupplierParameters: str,

		cacheSupplier: map[string]any{"test": 1},
	}}
	suplEv := &utils.CGREvent{
		Tenant: str,
		ID:     str,
		Time:   &tm,
		Event:  map[string]any{"AnswerTime": fl},
	}
	extraOpts := &optsGetSuppliers{
		ignoreErrors:      false,
		maxCost:           1.2,
		sortingParameters: slc,
		sortingStragety:   str,
	}
	argDsp := &utils.ArgDispatcher{
		APIKey:  &str,
		RouteID: &str,
	}

	rcv, err := ws.SortSuppliers(str, suppls, suplEv, extraOpts, argDsp)

	if err != nil {
		if err.Error() != "MANDATORY_IE_MISSING: [Account]" {
			t.Fatal(err)
		}
	} else {
		t.Fatal("was expectng an error")
	}

	if rcv != nil {
		t.Error(rcv)
	}

	slc2 := []string{}
	suppls2 := []*Supplier{{
		ID:                 str,
		FilterIDs:          slc,
		AccountIDs:         slc,
		RatingPlanIDs:      slc,
		ResourceIDs:        slc2,
		StatIDs:            slc,
		Weight:             1.2,
		Blocker:            true,
		SupplierParameters: str,

		cacheSupplier: map[string]any{"test": 1},
	}}
	rcv, err = ws.SortSuppliers(str, suppls2, suplEv, extraOpts, argDsp)

	if err != nil {
		if err.Error() != "MANDATORY_IE_MISSING: [ResourceIDs]" {
			t.Fatal(err)
		}
	} else {
		t.Fatal("was expectng an error")
	}

	if rcv != nil {
		t.Error(rcv)
	}
}

func TestSplsWeightSortSuppliersQOSSupplierSorter(t *testing.T) {
	str := "test"
	ws := &QOSSupplierSorter{
		sorting: str,
		spS:     &SupplierService{},
	}
	slc := []string{str}
	suppls := []*Supplier{{
		ID:                 str,
		FilterIDs:          slc,
		AccountIDs:         slc,
		RatingPlanIDs:      slc,
		ResourceIDs:        slc,
		StatIDs:            slc,
		Weight:             1.2,
		Blocker:            true,
		SupplierParameters: str,

		cacheSupplier: map[string]any{"test": 1},
	}}
	suplEv := &utils.CGREvent{
		Tenant: str,
		ID:     str,
		Time:   &tm,
		Event:  map[string]any{"AnswerTime": fl},
	}
	extraOpts := &optsGetSuppliers{
		ignoreErrors:      false,
		maxCost:           1.2,
		sortingParameters: slc,
		sortingStragety:   str,
	}
	argDsp := &utils.ArgDispatcher{
		APIKey:  &str,
		RouteID: &str,
	}

	rcv, err := ws.SortSuppliers(str, suppls, suplEv, extraOpts, argDsp)

	if err != nil {
		if err.Error() != "MANDATORY_IE_MISSING: [Account]" {
			t.Fatal(err)
		}
	} else {
		t.Fatal("was expectng an error")
	}

	if rcv != nil {
		t.Error(rcv)
	}

	extraOpts2 := &optsGetSuppliers{
		ignoreErrors:      true,
		maxCost:           1.2,
		sortingParameters: slc,
		sortingStragety:   str,
	}
	rcv, err = ws.SortSuppliers(str, suppls, suplEv, extraOpts2, argDsp)

	if err != nil {
		if err.Error() != "NOT_FOUND" {
			t.Fatal(err)
		}
	} else {
		t.Fatal("was expectng an error")
	}

	if rcv != nil {
		t.Error(rcv)
	}
}

func TestSplsWeightSortSuppliersLoadDistributionSorter(t *testing.T) {
	str := "test"
	ws := &LoadDistributionSorter{
		sorting: str,
		spS:     &SupplierService{},
	}
	slc := []string{str}
	suppls := []*Supplier{{
		ID:                 str,
		FilterIDs:          slc,
		AccountIDs:         slc,
		RatingPlanIDs:      slc,
		ResourceIDs:        slc,
		StatIDs:            slc,
		Weight:             1.2,
		Blocker:            true,
		SupplierParameters: str,

		cacheSupplier: map[string]any{"test": 1},
	}}
	suplEv := &utils.CGREvent{
		Tenant: str,
		ID:     str,
		Time:   &tm,
		Event:  map[string]any{"AnswerTime": fl},
	}
	extraOpts := &optsGetSuppliers{
		ignoreErrors:      false,
		maxCost:           1.2,
		sortingParameters: slc,
		sortingStragety:   str,
	}
	argDsp := &utils.ArgDispatcher{
		APIKey:  &str,
		RouteID: &str,
	}

	rcv, err := ws.SortSuppliers(str, suppls, suplEv, extraOpts, argDsp)

	if err != nil {
		if err.Error() != "MANDATORY_IE_MISSING: [Account]" {
			t.Fatal(err)
		}
	} else {
		t.Fatal("was expectng an error")
	}

	if rcv != nil {
		t.Error(rcv)
	}

	slc2 := []string{}
	suppls2 := []*Supplier{{
		ID:                 str,
		FilterIDs:          slc,
		AccountIDs:         slc,
		RatingPlanIDs:      slc,
		ResourceIDs:        slc2,
		StatIDs:            slc2,
		Weight:             1.2,
		Blocker:            true,
		SupplierParameters: str,

		cacheSupplier: map[string]any{"test": 1},
	}}
	rcv, err = ws.SortSuppliers(str, suppls2, suplEv, extraOpts, argDsp)

	if err != nil {
		if err.Error() != "MANDATORY_IE_MISSING: [StatIDs]" {
			t.Fatal(err)
		}
	} else {
		t.Fatal("was expectng an error")
	}

	if rcv != nil {
		t.Error(rcv)
	}
}

func TestSplsWeightSortSuppliersLeastCostSorter(t *testing.T) {
	str := "test"
	ws := &LeastCostSorter{
		sorting: str,
		spS:     &SupplierService{},
	}
	slc := []string{str}
	suppls := []*Supplier{{
		ID:                 str,
		FilterIDs:          slc,
		AccountIDs:         slc,
		RatingPlanIDs:      slc,
		ResourceIDs:        slc,
		StatIDs:            slc,
		Weight:             1.2,
		Blocker:            true,
		SupplierParameters: str,

		cacheSupplier: map[string]any{"test": 1},
	}}
	suplEv := &utils.CGREvent{
		Tenant: str,
		ID:     str,
		Time:   &tm,
		Event:  map[string]any{"AnswerTime": fl},
	}
	extraOpts := &optsGetSuppliers{
		ignoreErrors:      false,
		maxCost:           1.2,
		sortingParameters: slc,
		sortingStragety:   str,
	}
	argDsp := &utils.ArgDispatcher{
		APIKey:  &str,
		RouteID: &str,
	}

	rcv, err := ws.SortSuppliers(str, suppls, suplEv, extraOpts, argDsp)

	if err != nil {
		if err.Error() != "MANDATORY_IE_MISSING: [Account]" {
			t.Fatal(err)
		}
	} else {
		t.Fatal("was expectng an error")
	}

	if rcv != nil {
		t.Error(rcv)
	}

	slc2 := []string{}
	suppls2 := []*Supplier{{
		ID:                 str,
		FilterIDs:          slc,
		AccountIDs:         slc,
		RatingPlanIDs:      slc2,
		ResourceIDs:        slc2,
		StatIDs:            slc2,
		Weight:             1.2,
		Blocker:            true,
		SupplierParameters: str,

		cacheSupplier: map[string]any{"test": 1},
	}}
	rcv, err = ws.SortSuppliers(str, suppls2, suplEv, extraOpts, argDsp)

	if err != nil {
		if err.Error() != "MANDATORY_IE_MISSING: [RatingPlanIDs]" {
			t.Fatal(err)
		}
	} else {
		t.Fatal("was expectng an error")
	}

	if rcv != nil {
		t.Error(rcv)
	}

	extraOpts2 := &optsGetSuppliers{
		ignoreErrors:      true,
		maxCost:           1.2,
		sortingParameters: slc,
		sortingStragety:   str,
	}
	rcv, err = ws.SortSuppliers(str, suppls, suplEv, extraOpts2, argDsp)

	if err != nil {
		if err.Error() != "NOT_FOUND" {
			t.Fatal(err)
		}
	} else {
		t.Fatal("was expectng an error")
	}

	if rcv != nil {
		t.Error(rcv)
	}
}

func TestSplsWeightSortSuppliersHightCostSorter(t *testing.T) {
	str := "test"
	ws := &HightCostSorter{
		sorting: str,
		spS:     &SupplierService{},
	}
	slc := []string{str}
	suppls := []*Supplier{{
		ID:                 str,
		FilterIDs:          slc,
		AccountIDs:         slc,
		RatingPlanIDs:      slc,
		ResourceIDs:        slc,
		StatIDs:            slc,
		Weight:             1.2,
		Blocker:            true,
		SupplierParameters: str,

		cacheSupplier: map[string]any{"test": 1},
	}}
	suplEv := &utils.CGREvent{
		Tenant: str,
		ID:     str,
		Time:   &tm,
		Event:  map[string]any{"AnswerTime": fl},
	}
	extraOpts := &optsGetSuppliers{
		ignoreErrors:      false,
		maxCost:           1.2,
		sortingParameters: slc,
		sortingStragety:   str,
	}
	argDsp := &utils.ArgDispatcher{
		APIKey:  &str,
		RouteID: &str,
	}

	rcv, err := ws.SortSuppliers(str, suppls, suplEv, extraOpts, argDsp)

	if err != nil {
		if err.Error() != "MANDATORY_IE_MISSING: [Account]" {
			t.Fatal(err)
		}
	} else {
		t.Fatal("was expectng an error")
	}

	if rcv != nil {
		t.Error(rcv)
	}

	slc2 := []string{}
	suppls2 := []*Supplier{{
		ID:                 str,
		FilterIDs:          slc,
		AccountIDs:         slc,
		RatingPlanIDs:      slc2,
		ResourceIDs:        slc2,
		StatIDs:            slc2,
		Weight:             1.2,
		Blocker:            true,
		SupplierParameters: str,

		cacheSupplier: map[string]any{"test": 1},
	}}
	rcv, err = ws.SortSuppliers(str, suppls2, suplEv, extraOpts, argDsp)

	if err != nil {
		if err.Error() != "MANDATORY_IE_MISSING: [RatingPlanIDs]" {
			t.Fatal(err)
		}
	} else {
		t.Fatal("was expectng an error")
	}

	if rcv != nil {
		t.Error(rcv)
	}

	extraOpts2 := &optsGetSuppliers{
		ignoreErrors:      true,
		maxCost:           1.2,
		sortingParameters: slc,
		sortingStragety:   str,
	}
	rcv, err = ws.SortSuppliers(str, suppls, suplEv, extraOpts2, argDsp)

	if err != nil {
		if err.Error() != "NOT_FOUND" {
			t.Fatal(err)
		}
	} else {
		t.Fatal("was expectng an error")
	}

	if rcv != nil {
		t.Error(rcv)
	}
}
