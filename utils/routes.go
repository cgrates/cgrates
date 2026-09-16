// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package utils

import (
	"maps"
	"sort"
	"strconv"
	"strings"
)

// RoutesDefaultRatio is the default ratio value for routes when none is
// explicitly specified in the profile. Defined here to avoid circular
// dependencies with the config package.
var RoutesDefaultRatio = 1

// RouteProfile represents the configuration of a Route profile.
type RouteProfile struct {
	Tenant            string
	ID                string // LCR Profile ID
	FilterIDs         []string
	Weights           DynamicWeights
	Blockers          DynamicBlockers
	Sorting           string // Sorting strategy
	SortingParameters []string
	Routes            []*Route
}

// Clone method for RouteProfile
func (rp *RouteProfile) Clone() *RouteProfile {
	if rp == nil {
		return nil
	}
	clone := &RouteProfile{
		Tenant:  rp.Tenant,
		ID:      rp.ID,
		Sorting: rp.Sorting,
	}
	if rp.FilterIDs != nil {
		clone.FilterIDs = make([]string, len(rp.FilterIDs))
		copy(clone.FilterIDs, rp.FilterIDs)
	}
	if rp.SortingParameters != nil {
		clone.SortingParameters = make([]string, len(rp.SortingParameters))
		copy(clone.SortingParameters, rp.SortingParameters)
	}
	if rp.Routes != nil {
		clone.Routes = make([]*Route, len(rp.Routes))
		for i, route := range rp.Routes {
			clone.Routes[i] = route.Clone()
		}
	}
	if rp.Weights != nil {
		clone.Weights = rp.Weights.Clone()
	}
	if rp.Blockers != nil {
		clone.Blockers = rp.Blockers.Clone()
	}
	return clone
}

// CacheClone returns a clone of RouteProfile used by ltcache CacheCloner
func (rp *RouteProfile) CacheClone() any {
	return rp.Clone()
}

// RouteProfileWithAPIOpts wraps RouteProfile with APIOpts.
type RouteProfileWithAPIOpts struct {
	*RouteProfile
	APIOpts map[string]any
}

// compileCacheParameters prepares route ratios for MetaLoad sorting by parsing the
// SortingParameters and applying appropriate ratio values to each route.
func (rp *RouteProfile) compileCacheParameters() error {
	if rp.Sorting != MetaLoad {
		return nil
	}

	// Parse route ID to ratio mappings.
	ratioMap := make(map[string]int)
	for _, param := range rp.SortingParameters {
		parts := strings.Split(param, ConcatenatedKeySep)
		ratio, err := strconv.Atoi(parts[1])
		if err != nil {
			return err
		}
		ratioMap[parts[0]] = ratio
	}

	// Get default ratio (from map or config).
	defaultRatio := RoutesDefaultRatio
	if mapDefault, exists := ratioMap[MetaDefault]; exists {
		defaultRatio = mapDefault
	}

	// Apply appropriate ratio to each route.
	for _, route := range rp.Routes {
		route.cacheRoute = make(map[string]any)
		ratio, exists := ratioMap[route.ID]
		if !exists {
			ratio = defaultRatio
		}
		route.cacheRoute[MetaRatio] = ratio
	}

	return nil
}

// Compile is a wrapper for convenience setting up the RouteProfile.
func (rp *RouteProfile) Compile() error {
	return rp.compileCacheParameters()
}

// TenantID returns unique identifier of the LCRProfile in a multi-tenant environment.
func (rp *RouteProfile) TenantID() string {
	return ConcatenatedKey(rp.Tenant, rp.ID)
}

// Set implements the profile interface, setting values in RouteProfile based on path.
func (rp *RouteProfile) Set(path []string, val any, newBranch bool) (err error) {
	switch len(path) {
	default:
		return ErrWrongPath
	case 1:
		switch path[0] {
		default:
			return ErrWrongPath
		case Tenant:
			rp.Tenant = IfaceAsString(val)
		case ID:
			rp.ID = IfaceAsString(val)
		case FilterIDs:
			var valA []string
			valA, err = IfaceAsStringSlice(val)
			rp.FilterIDs = append(rp.FilterIDs, valA...)
		case SortingParameters:
			var valA []string
			valA, err = IfaceAsStringSlice(val)
			rp.SortingParameters = append(rp.SortingParameters, valA...)
		case Sorting:
			if valStr := IfaceAsString(val); len(valStr) != 0 {
				rp.Sorting = valStr
			}
		case Weights:
			if val != EmptyString {
				rp.Weights, err = NewDynamicWeightsFromString(IfaceAsString(val), InfieldSep, ANDSep)
			}
		case Blockers:
			if val != EmptyString {
				rp.Blockers, err = NewDynamicBlockersFromString(IfaceAsString(val), InfieldSep, ANDSep)
			}
		}
	case 2:
		if val == EmptyString {
			return
		}
		if path[0] != Routes {
			return ErrWrongPath
		}
		if len(rp.Routes) == 0 || newBranch {
			rp.Routes = append(rp.Routes, new(Route))
		}
		rt := rp.Routes[len(rp.Routes)-1]
		switch path[1] {
		case ID:
			rt.ID = IfaceAsString(val)
		case FilterIDs:
			var valA []string
			valA, err = IfaceAsStringSlice(val)
			rt.FilterIDs = append(rt.FilterIDs, valA...)
		case AccountIDs:
			var valA []string
			valA, err = IfaceAsStringSlice(val)
			rt.AccountIDs = append(rt.AccountIDs, valA...)
		case RateProfileIDs:
			var valA []string
			valA, err = IfaceAsStringSlice(val)
			rt.RateProfileIDs = append(rt.RateProfileIDs, valA...)
		case ResourceIDs:
			var valA []string
			valA, err = IfaceAsStringSlice(val)
			rt.ResourceIDs = append(rt.ResourceIDs, valA...)
		case StatIDs:
			var valA []string
			valA, err = IfaceAsStringSlice(val)
			rt.StatIDs = append(rt.StatIDs, valA...)
		case Weights:
			if val != EmptyString {
				rt.Weights, err = NewDynamicWeightsFromString(IfaceAsString(val), InfieldSep, ANDSep)
			}
		case Blockers:
			if val != EmptyString {
				rt.Blockers, err = NewDynamicBlockersFromString(IfaceAsString(val), InfieldSep, ANDSep)
			}
		case RouteParameters:
			rt.RouteParameters = IfaceAsString(val)
		default:
			return ErrWrongPath
		}
	}
	return
}

// Merge implements the profile interface, merging values from another RouteProfile.
func (rp *RouteProfile) Merge(v2 any) {
	vi := v2.(*RouteProfile)
	if len(vi.Tenant) != 0 {
		rp.Tenant = vi.Tenant
	}
	if len(vi.ID) != 0 {
		rp.ID = vi.ID
	}
	rp.FilterIDs = append(rp.FilterIDs, vi.FilterIDs...)
	rp.SortingParameters = append(rp.SortingParameters, vi.SortingParameters...)
	var equal bool
	for _, routeV2 := range vi.Routes {
		for _, route := range rp.Routes {
			if route.ID == routeV2.ID {
				route.Merge(routeV2)
				equal = true
				break
			}
		}
		if !equal {
			rp.Routes = append(rp.Routes, routeV2)
		}
		equal = false
	}
	rp.Weights = append(rp.Weights, vi.Weights...)
	rp.Blockers = append(rp.Blockers, vi.Blockers...)
	if len(vi.Sorting) != 0 {
		rp.Sorting = vi.Sorting
	}
}

// String implements the DataProvider interface, returning the RouteProfile in JSON format.
func (rp *RouteProfile) String() string { return ToJSON(rp) }

// FieldAsString implements the DataProvider interface, retrieving field value as string.
func (rp *RouteProfile) FieldAsString(fldPath []string) (_ string, err error) {
	var val any
	if val, err = rp.FieldAsInterface(fldPath); err != nil {
		return
	}
	return IfaceAsString(val), nil
}

// FieldAsInterface implements the DataProvider interface, retrieving field value as interface.
func (rp *RouteProfile) FieldAsInterface(fldPath []string) (_ any, err error) {
	if len(fldPath) == 1 {
		switch fldPath[0] {
		default:
			fld, idx := GetPathIndex(fldPath[0])
			if idx != nil {
				switch fld {
				case SortingParameters:
					if *idx < len(rp.SortingParameters) {
						return rp.SortingParameters[*idx], nil
					}
				case FilterIDs:
					if *idx < len(rp.FilterIDs) {
						return rp.FilterIDs[*idx], nil
					}
				case Routes:
					if *idx < len(rp.Routes) {
						return rp.Routes[*idx], nil
					}
				}
			}
			return nil, ErrNotFound
		case Tenant:
			return rp.Tenant, nil
		case ID:
			return rp.ID, nil
		case FilterIDs:
			return rp.FilterIDs, nil
		case Weights:
			return rp.Weights.String(InfieldSep, ANDSep), nil
		case SortingParameters:
			return rp.SortingParameters, nil
		case Sorting:
			return rp.Sorting, nil
		case Blockers:
			return rp.Blockers.String(InfieldSep, ANDSep), nil
		case Routes:
			return rp.Routes, nil
		}
	}
	if len(fldPath) == 0 {
		return nil, ErrNotFound
	}
	fld, idx := GetPathIndex(fldPath[0])
	if fld != Routes ||
		idx == nil {
		return nil, ErrNotFound
	}
	if *idx >= len(rp.Routes) {
		return nil, ErrNotFound
	}
	return rp.Routes[*idx].FieldAsInterface(fldPath[1:])
}

// AsMapStringInterface converts RouteProfile struct to map[string]any
func (rp *RouteProfile) AsMapStringInterface() map[string]any {
	if rp == nil {
		return nil
	}
	return map[string]any{
		Tenant:            rp.Tenant,
		ID:                rp.ID,
		FilterIDs:         rp.FilterIDs,
		Weights:           rp.Weights,
		Blockers:          rp.Blockers,
		Sorting:           rp.Sorting,
		SortingParameters: rp.SortingParameters,
		Routes:            rp.Routes,
	}
}

// MapStringInterfaceToRouteProfile converts map[string]any to RouteProfile struct
func MapStringInterfaceToRouteProfile(m map[string]any) *RouteProfile {
	rp := &RouteProfile{}

	if v, ok := m[Tenant].(string); ok {
		rp.Tenant = v
	}
	if v, ok := m[ID].(string); ok {
		rp.ID = v
	}
	rp.FilterIDs = InterfaceToStringSlice(m[FilterIDs])
	rp.Weights = InterfaceToDynamicWeights(m[Weights])
	rp.Blockers = InterfaceToDynamicBlockers(m[Blockers])
	if v, ok := m[Sorting].(string); ok {
		rp.Sorting = v
	}
	rp.SortingParameters = InterfaceToStringSlice(m[SortingParameters])
	rp.Routes = InterfaceToRoutes(m[Routes])
	return rp
}

// InterfaceToRoutes converts interface to []*Route
func InterfaceToRoutes(v any) []*Route {
	if v == nil {
		return nil
	}
	switch val := v.(type) {
	case []*Route:
		return val
	case []any:
		var result []*Route
		for _, item := range val {
			if m, ok := item.(map[string]any); ok {
				route := MapStringInterfaceToRoute(m)
				result = append(result, route)
			} else if r, ok := item.(*Route); ok {
				result = append(result, r)
			}
		}
		return result
	default:
		return nil
	}
}

// MapStringInterfaceToRoute converts map[string]any to Route struct
func MapStringInterfaceToRoute(m map[string]any) *Route {
	r := &Route{}
	if v, ok := m[ID].(string); ok {
		r.ID = v
	}
	r.FilterIDs = InterfaceToStringSlice(m[FilterIDs])
	r.AccountIDs = InterfaceToStringSlice(m[AccountIDs])
	r.RateProfileIDs = InterfaceToStringSlice(m[RateProfileIDs])
	r.ResourceIDs = InterfaceToStringSlice(m[ResourceIDs])
	r.StatIDs = InterfaceToStringSlice(m[StatIDs])
	r.Weights = InterfaceToDynamicWeights(m[Weights])
	r.Blockers = InterfaceToDynamicBlockers(m[Blockers])
	if v, ok := m[RouteParameters].(string); ok {
		r.RouteParameters = v
	}
	return r
}

// Route defines a single route within a RouteProfile.
type Route struct {
	ID              string // RouteID
	FilterIDs       []string
	AccountIDs      []string
	RateProfileIDs  []string // used when computing price
	ResourceIDs     []string // queried in some strategies
	StatIDs         []string // queried in some strategies
	Weights         DynamicWeights
	Blockers        DynamicBlockers // if true, stops processing further routes
	RouteParameters string

	// Internal cache for route properties
	// Example: cacheRoute["*ratio"] contains the route's ratio value
	cacheRoute map[string]any
}

// Clone method for Route
func (r *Route) Clone() *Route {
	if r == nil {
		return nil
	}
	clone := &Route{
		ID:              r.ID,
		RouteParameters: r.RouteParameters,
	}
	if r.FilterIDs != nil {
		clone.FilterIDs = make([]string, len(r.FilterIDs))
		copy(clone.FilterIDs, r.FilterIDs)
	}
	if r.Weights != nil {
		clone.Weights = r.Weights.Clone()
	}
	if r.Blockers != nil {
		clone.Blockers = r.Blockers.Clone()
	}
	if r.AccountIDs != nil {
		clone.AccountIDs = make([]string, len(r.AccountIDs))
		copy(clone.AccountIDs, r.AccountIDs)
	}
	if r.RateProfileIDs != nil {
		clone.RateProfileIDs = make([]string, len(r.RateProfileIDs))
		copy(clone.RateProfileIDs, r.RateProfileIDs)
	}
	if r.ResourceIDs != nil {
		clone.ResourceIDs = make([]string, len(r.ResourceIDs))
		copy(clone.ResourceIDs, r.ResourceIDs)
	}
	if r.StatIDs != nil {
		clone.StatIDs = make([]string, len(r.StatIDs))
		copy(clone.StatIDs, r.StatIDs)
	}
	if r.cacheRoute != nil {
		clone.cacheRoute = make(map[string]any)
		maps.Copy(clone.cacheRoute, r.cacheRoute)
	}
	return clone
}

// Ratio returns the cached ratio value for this route.
func (r *Route) Ratio() any {
	return r.cacheRoute[MetaRatio]
}

// Merge implements the merge interface, merging values from another Route.
func (r *Route) Merge(v2 *Route) {
	if len(v2.ID) != 0 {
		r.ID = v2.ID
	}
	if len(v2.RouteParameters) != 0 {
		r.RouteParameters = v2.RouteParameters
	}
	r.Weights = append(r.Weights, v2.Weights...)
	r.Blockers = append(r.Blockers, v2.Blockers...)
	r.FilterIDs = append(r.FilterIDs, v2.FilterIDs...)
	r.AccountIDs = append(r.AccountIDs, v2.AccountIDs...)
	r.RateProfileIDs = append(r.RateProfileIDs, v2.RateProfileIDs...)
	r.ResourceIDs = append(r.ResourceIDs, v2.ResourceIDs...)
	r.StatIDs = append(r.StatIDs, v2.StatIDs...)
}

// String implements the DataProvider interface, returning the Route in JSON format.
func (r *Route) String() string { return ToJSON(r) }

// FieldAsString implements the DataProvider interface, retrieving field value as string.
func (r *Route) FieldAsString(fldPath []string) (_ string, err error) {
	var val any
	if val, err = r.FieldAsInterface(fldPath); err != nil {
		return
	}
	return IfaceAsString(val), nil
}

// FieldAsInterface implements the DataProvider interface, retrieving field value as interface.
func (r *Route) FieldAsInterface(fldPath []string) (_ any, err error) {
	if len(fldPath) != 1 {
		return nil, ErrNotFound
	}
	switch fldPath[0] {
	default:
		fld, idx := GetPathIndex(fldPath[0])
		if idx != nil {
			switch fld {
			case AccountIDs:
				if *idx < len(r.AccountIDs) {
					return r.AccountIDs[*idx], nil
				}
			case FilterIDs:
				if *idx < len(r.FilterIDs) {
					return r.FilterIDs[*idx], nil
				}
			case RateProfileIDs:
				if *idx < len(r.RateProfileIDs) {
					return r.RateProfileIDs[*idx], nil
				}
			case ResourceIDs:
				if *idx < len(r.ResourceIDs) {
					return r.ResourceIDs[*idx], nil
				}
			case StatIDs:
				if *idx < len(r.StatIDs) {
					return r.StatIDs[*idx], nil
				}
			}
		}
		return nil, ErrNotFound
	case ID:
		return r.ID, nil
	case FilterIDs:
		return r.FilterIDs, nil
	case AccountIDs:
		return r.AccountIDs, nil
	case RateProfileIDs:
		return r.RateProfileIDs, nil
	case ResourceIDs:
		return r.ResourceIDs, nil
	case StatIDs:
		return r.StatIDs, nil
	case Weights:
		return r.Weights.String(InfieldSep, ANDSep), nil
	case Blockers:
		return r.Blockers.String(InfieldSep, ANDSep), nil
	case RouteParameters:
		return r.RouteParameters, nil
	}
}

// SortedRoute represents one route in SortedRoutes
type SortedRoute struct {
	RouteID            string
	RouteParameters    string
	SortingData        map[string]any      // store here extra info like cost or stats (can contain the data that we do not use to sort after)
	sortingDataDecimal map[string]*Decimal // only the data we sort after
}

func (sr *SortedRoute) SetSortingDataDecimal(sortingDataDecimal map[string]*Decimal) {
	sr.sortingDataDecimal = sortingDataDecimal
}

func (sr *SortedRoute) AddToSortingDataDecimalMap(key string, val *Decimal) {
	if sr.sortingDataDecimal == nil {
		sr.sortingDataDecimal = make(map[string]*Decimal)
	}
	sr.sortingDataDecimal[key] = val
}

// getSortedData returns the data from map or a 0( in case it was not populated)
func (s *SortedRoute) getSortedData(field string) *Decimal {
	if v, has := s.sortingDataDecimal[field]; has {
		return v
	}
	return NewDecimalFromFloat64(0)
}

// SortedRoutes represents all viable routes inside one routing profile
type SortedRoutes struct {
	ProfileID string         // Profile matched
	Sorting   string         // Sorting algorithm
	Routes    []*SortedRoute // list of route IDs and SortingData data
}

// RouteIDs returns a list of route IDs
func (sRoutes *SortedRoutes) RouteIDs() (rIDs []string) {
	rIDs = make([]string, len(sRoutes.Routes))
	for i, sRoute := range sRoutes.Routes {
		rIDs[i] = sRoute.RouteID
	}
	return
}

// RoutesWithParams returns a list of routes IDs with Parameters
func (sRoutes *SortedRoutes) RoutesWithParams() (sPs []string) {
	sPs = make([]string, len(sRoutes.Routes))
	for i, spl := range sRoutes.Routes {
		sPs[i] = spl.RouteID
		if spl.RouteParameters != "" {
			sPs[i] += InInFieldSep + spl.RouteParameters
		}
	}
	return
}

func (sRoutes *SortedRoutes) compareWeight(i, j int) bool {
	cmp := sRoutes.Routes[i].getSortedData(Weight).Compare(sRoutes.Routes[j].getSortedData(Weight))
	if cmp == 0 {
		return BoolGenerator().RandomBool()
	}
	return cmp > 0
}

// SortWeight is part of sort interface, sort based on Weight
func (sRoutes *SortedRoutes) SortWeight() {
	sort.Slice(sRoutes.Routes, sRoutes.compareWeight)
}

// SortLeastCost is part of sort interface,
// sort ascendent based on Cost with fallback on Weight
func (sRoutes *SortedRoutes) SortLeastCost() {
	sort.Slice(sRoutes.Routes, func(i, j int) bool {
		cmp := sRoutes.Routes[i].getSortedData(Cost).Compare(sRoutes.Routes[j].getSortedData(Cost))
		if cmp == 0 {
			return sRoutes.compareWeight(i, j)
		}
		return cmp < 0
	})
}

// SortHighestCost is part of sort interface,
// sort descendent based on Cost with fallback on Weight
func (sRoutes *SortedRoutes) SortHighestCost() {
	sort.Slice(sRoutes.Routes, func(i, j int) bool {
		cmp := sRoutes.Routes[i].getSortedData(Cost).Compare(sRoutes.Routes[j].getSortedData(Cost))
		if cmp == 0 {
			return sRoutes.compareWeight(i, j)
		}
		return cmp > 0
	})
}

// SortQOS is part of sort interface,
// sort based on Stats
func (sRoutes *SortedRoutes) SortQOS(params []string) {
	//sort routes
	sort.Slice(sRoutes.Routes, func(i, j int) bool {
		for _, param := range params {
			//in case we have the same value for the current param we skip to the next one
			cmp := sRoutes.Routes[i].getSortedData(param).Compare(sRoutes.Routes[j].getSortedData(param))
			if cmp == 0 {
				continue
			}
			if param == MetaPDD {
				return cmp < 0
			}
			return cmp > 0
		}
		//in case that we have the same value for all params we sort base on weight
		return sRoutes.compareWeight(i, j)
	})
}

// SortResourceAscendent is part of sort interface,
// sort ascendent based on ResourceUsage with fallback on Weight
func (sRoutes *SortedRoutes) SortResourceAscendent() {
	sort.Slice(sRoutes.Routes, func(i, j int) bool {
		cmp := sRoutes.Routes[i].getSortedData(ResourceUsageStr).Compare(sRoutes.Routes[j].getSortedData(ResourceUsageStr))
		if cmp == 0 {
			return sRoutes.compareWeight(i, j)
		}
		return cmp < 0
	})
}

// SortResourceDescendent is part of sort interface,
// sort descendent based on ResourceUsage with fallback on Weight
func (sRoutes *SortedRoutes) SortResourceDescendent() {
	sort.Slice(sRoutes.Routes, func(i, j int) bool {
		cmp := sRoutes.Routes[i].getSortedData(ResourceUsageStr).Compare(sRoutes.Routes[j].getSortedData(ResourceUsageStr))
		if cmp == 0 {
			return sRoutes.compareWeight(i, j)
		}
		return cmp > 0
	})
}

// SortLoadDistribution is part of sort interface,
// sort based on the following formula float64(metricVal/ratio) with fallback on Weight
func (sRoutes *SortedRoutes) SortLoadDistribution() {
	sort.Slice(sRoutes.Routes, func(i, j int) bool {
		// ((ratio + metricVal) / (ratio)) -1 = ((ratio+metricVal)/ratio) - (ratio/ratio) = (ratio+metricVal-ratio)/ratio = metricVal/ratio
		cmp := DivideDecimal(sRoutes.Routes[i].getSortedData(Load), sRoutes.Routes[i].getSortedData(Ratio)).Compare(
			DivideDecimal(sRoutes.Routes[j].getSortedData(Load), sRoutes.Routes[j].getSortedData(Ratio)))
		if cmp == 0 {
			return sRoutes.compareWeight(i, j)
		}
		return cmp < 0
	})
}

// Digest returns list of routeIDs + parameters for easier outside access
// format route1:route1params,route2:route2params
func (sRoutes *SortedRoutes) Digest() string {
	return strings.Join(sRoutes.RoutesWithParams(), FieldsSep)
}

func (ss *SortedRoute) AsNavigableMap() (nm *DataNode) {
	nm = &DataNode{
		Type: NMMapType,
		Map: map[string]*DataNode{
			RouteID:         NewLeafNode(ss.RouteID),
			RouteParameters: NewLeafNode(ss.RouteParameters),
		},
	}
	sd := &DataNode{Type: NMMapType, Map: map[string]*DataNode{}}
	for k, d := range ss.SortingData {
		sd.Map[k] = NewLeafNode(d)
	}
	nm.Map[SortingData] = sd
	return
}

func (sRoutes *SortedRoutes) AsNavigableMap() (nm *DataNode) {
	nm = &DataNode{
		Type: NMMapType,
		Map: map[string]*DataNode{
			ProfileID: NewLeafNode(sRoutes.ProfileID),
			Sorting:   NewLeafNode(sRoutes.Sorting),
		},
	}
	sr := make([]*DataNode, len(sRoutes.Routes))
	for i, ss := range sRoutes.Routes {
		sr[i] = ss.AsNavigableMap()
	}
	nm.Map[CapRoutes] = &DataNode{Type: NMSliceType, Slice: sr}
	return
}

// SortedRoutesList represents the list of matched routes grouped based of profile
type SortedRoutesList []*SortedRoutes

// RouteIDs returns a list of route IDs
func (sRs SortedRoutesList) RouteIDs() (rIDs []string) {
	for _, sR := range sRs {
		for _, r := range sR.Routes {
			rIDs = append(rIDs, r.RouteID)
		}
	}
	return
}

// RoutesWithParams returns a list of routes IDs with Parameters
func (sRs SortedRoutesList) RoutesWithParams() (sPs []string) {
	routeIDs := make(StringSet)
	for _, sR := range sRs {
		for _, spl := range sR.Routes {
			route := spl.RouteID
			if spl.RouteParameters != EmptyString {
				route += InInFieldSep + spl.RouteParameters
			}
			if !routeIDs.Has(route) {
				routeIDs.Add(route)
				sPs = append(sPs, route)
			}
		}
	}
	return
}

// Digest returns list of routeIDs + parameters for easier outside access
// format route1:route1params,route2:route2params
func (sRs SortedRoutesList) Digest() string {
	return strings.Join(sRs.RoutesWithParams(), FieldsSep)
}

// AsNavigableMap returns the SortedRoutesSet as NMInterface object
func (sRs SortedRoutesList) AsNavigableMap() (nm *DataNode) {
	nm = &DataNode{Type: NMSliceType, Slice: make([]*DataNode, len(sRs))}
	for i, ss := range sRs {
		nm.Slice[i] = ss.AsNavigableMap()
	}
	return
}
