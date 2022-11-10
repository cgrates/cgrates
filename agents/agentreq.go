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

package agents

import (
	"fmt"
	"strings"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// NewAgentRequest returns a new AgentRequest
func NewAgentRequest(req utils.DataProvider,
	vars, cgrRply *utils.DataNode,
	rply *utils.OrderedNavigableMap,
	opts utils.MapStorage,
	tntTpl config.RSRParsers,
	dfltTenant, timezone string,
	filterS *engine.FilterS,
	extraDP map[string]utils.DataProvider) (ar *AgentRequest) {
	if cgrRply == nil {
		cgrRply = &utils.DataNode{Type: utils.NMMapType, Map: make(map[string]*utils.DataNode)}
	}
	if vars == nil {
		vars = &utils.DataNode{Type: utils.NMMapType, Map: make(map[string]*utils.DataNode)}
	}
	if rply == nil {
		rply = utils.NewOrderedNavigableMap()
	}
	if opts == nil {
		opts = make(utils.MapStorage)
	}
	if extraDP == nil {
		extraDP = make(map[string]utils.DataProvider)
	}
	ar = &AgentRequest{
		Request:    req,
		Tenant:     dfltTenant,
		Vars:       vars,
		CGRRequest: utils.NewOrderedNavigableMap(),
		diamreq:    utils.NewOrderedNavigableMap(), // special case when CGRateS is building the request
		CGRReply:   cgrRply,
		Reply:      rply,
		Timezone:   timezone,
		filterS:    filterS,
		Opts:       opts,
		Cfg:        config.CgrConfig().GetDataProvider(),
		ExtraDP:    extraDP,
	}
	if tnt, err := tntTpl.ParseDataProvider(ar); err == nil && tnt != utils.EmptyString {
		ar.Tenant = tnt
	}
	ar.Vars.Set([]string{utils.NodeID}, config.CgrConfig().GeneralCfg().NodeID)
	return
}

// AgentRequest represents data related to one request towards agent
// implements utils.DataProvider so we can pass it to filters
type AgentRequest struct {
	Request    utils.DataProvider         // request
	Vars       *utils.DataNode            // shared data
	CGRRequest *utils.OrderedNavigableMap // Used in reply to access the request that was send
	CGRReply   *utils.DataNode
	Reply      *utils.OrderedNavigableMap
	Tenant     string
	Timezone   string
	filterS    *engine.FilterS
	diamreq    *utils.OrderedNavigableMap // used in case of building requests (ie. DisconnectSession)
	tmp        *utils.DataNode            // used in case you want to store temporary items and access them later
	Opts       utils.MapStorage
	Cfg        utils.DataProvider
	ExtraDP    map[string]utils.DataProvider
}

// String implements utils.DataProvider
func (ar *AgentRequest) String() string {
	return utils.ToIJSON(ar)
}

// FieldAsInterface implements utils.DataProvider
func (ar *AgentRequest) FieldAsInterface(fldPath []string) (val interface{}, err error) {
	switch fldPath[0] {
	default:
		dp, has := ar.ExtraDP[fldPath[0]]
		if !has {
			return nil, fmt.Errorf("unsupported field prefix: <%s>", fldPath[0])
		}
		val, err = dp.FieldAsInterface(fldPath[1:])
	case utils.MetaReq:
		if len(fldPath) != 1 {
			val, err = ar.Request.FieldAsInterface(fldPath[1:])
		} else {
			val = ar.Request
		}
	case utils.MetaVars:
		if len(fldPath) != 1 {
			val, err = ar.Vars.FieldAsInterface(fldPath[1:])
		} else {
			val = ar.Vars
		}
	case utils.MetaCgreq:
		if len(fldPath) != 1 {
			val, err = ar.CGRRequest.FieldAsInterface(fldPath[1:])
		} else {
			val = ar.CGRRequest
		}
	case utils.MetaCgrep:
		if len(fldPath) != 1 {
			val, err = ar.CGRReply.FieldAsInterface(fldPath[1:])
		} else {
			val = ar.CGRReply
		}
	case utils.MetaDiamreq:
		if len(fldPath) != 1 {
			val, err = ar.diamreq.FieldAsInterface(fldPath[1:])
		} else {
			val = ar.diamreq
		}
	case utils.MetaRep:
		if len(fldPath) != 1 {
			val, err = ar.Reply.FieldAsInterface(fldPath[1:])
		} else {
			val = ar.Reply
		}
	case utils.MetaTmp:
		if len(fldPath) != 1 {
			val, err = ar.tmp.FieldAsInterface(fldPath[1:])
		} else {
			val = ar.tmp
		}
	case utils.MetaUCH:
		if cacheVal, ok := engine.Cache.Get(utils.CacheUCH, strings.Join(fldPath[1:], utils.NestingSep)); !ok {
			err = utils.ErrNotFound
		} else {
			val = cacheVal
		}
	case utils.MetaOpts:
		if len(fldPath) != 1 {
			val, err = ar.Opts.FieldAsInterface(fldPath[1:])
		} else {
			val = ar.Opts
		}
	case utils.MetaCfg:
		if len(fldPath) != 1 {
			val, err = ar.Cfg.FieldAsInterface(fldPath[1:])
		} else {
			val = ar.Cfg
		}
	case utils.MetaTenant:
		return ar.Tenant, nil
	}
	if err != nil {
		return
	}
	if nmItems, isNMItems := val.([]*utils.DataNode); isNMItems { // special handling of NMItems, take the last value out of it
		el := nmItems[len(nmItems)-1]
		if el.Type == utils.NMDataType {
			val = el.Value.Data
		}
	}
	return
}

// FieldAsString implements utils.DataProvider
func (ar *AgentRequest) FieldAsString(fldPath []string) (val string, err error) {
	var iface interface{}
	if iface, err = ar.FieldAsInterface(fldPath); err != nil {
		return
	}
	return utils.IfaceAsString(iface), nil
}

// SetFields will populate fields of AgentRequest out of templates
func (ar *AgentRequest) SetFields(tplFlds []*config.FCTemplate) (err error) {
	ar.tmp = &utils.DataNode{Type: utils.NMMapType, Map: make(map[string]*utils.DataNode)}
	for _, tplFld := range tplFlds {
		if pass, err := ar.filterS.Pass(ar.Tenant,
			tplFld.Filters, ar); err != nil {
			return err
		} else if !pass {
			continue
		}
		switch tplFld.Type {
		case utils.MetaNone:
		case utils.MetaRemove:
			if err = ar.Remove(&utils.FullPath{
				PathSlice: tplFld.GetPathSlice(),
				Path:      tplFld.Path,
			}); err != nil {
				return
			}
		case utils.MetaRemoveAll:
			if err = ar.RemoveAll(tplFld.GetPathSlice()[0]); err != nil {
				return
			}
		default:
			var out interface{}
			out, err = ar.ParseField(tplFld)
			if err != nil {
				if err == utils.ErrNotFound {
					if !tplFld.Mandatory {
						err = nil
						continue
					}
					err = utils.ErrPrefixNotFound(tplFld.Tag)
				}
				return
			}
			var fullPath *utils.FullPath
			if fullPath, err = utils.GetFullFieldPath(tplFld.Path, ar); err != nil {
				return
			} else if fullPath == nil { // no dynamic path
				fullPath = &utils.FullPath{
					PathSlice: utils.CloneStringSlice(tplFld.GetPathSlice()), // need to clone so me do not modify the template
					Path:      tplFld.Path,
				}
			}

			nMItm := &utils.DataLeaf{Data: out, NewBranch: tplFld.NewBranch, AttributeID: tplFld.AttributeID}
			switch tplFld.Type {
			case utils.MetaComposed:
				err = ar.Compose(fullPath, nMItm)
			case utils.MetaGroup: // in case of *group type simply append to valSet
				err = ar.Append(fullPath, nMItm)
			default:
				err = ar.SetAsSlice(fullPath, nMItm)
			}
			if err != nil {
				return
			}
		}
		if tplFld.Blocker { // useful in case of processing errors first
			break
		}
	}
	return
}

// Set implements utils.NMInterface
func (ar *AgentRequest) SetAsSlice(fullPath *utils.FullPath, nm *utils.DataLeaf) (err error) {
	switch fullPath.PathSlice[0] {
	default:
		return fmt.Errorf("unsupported field prefix: <%s> when set field", fullPath.PathSlice[0])
	case utils.MetaVars:
		_, err = ar.Vars.Set(fullPath.PathSlice[1:], []*utils.DataNode{{Type: utils.NMDataType, Value: nm}})
		return
	case utils.MetaCgreq:
		return ar.CGRRequest.SetAsSlice(&utils.FullPath{
			PathSlice: fullPath.PathSlice[1:],
			Path:      fullPath.Path[7:],
		}, []*utils.DataNode{{Type: utils.NMDataType, Value: nm}})
	case utils.MetaCgrep:
		_, err = ar.CGRReply.Set(fullPath.PathSlice[1:], []*utils.DataNode{{Type: utils.NMDataType, Value: nm}})
		return
	case utils.MetaRep:
		return ar.Reply.SetAsSlice(&utils.FullPath{
			PathSlice: fullPath.PathSlice[1:],
			Path:      fullPath.Path[5:],
		}, []*utils.DataNode{{Type: utils.NMDataType, Value: nm}})
	case utils.MetaDiamreq:
		return ar.diamreq.SetAsSlice(&utils.FullPath{
			PathSlice: fullPath.PathSlice[1:],
			Path:      fullPath.Path[9:],
		}, []*utils.DataNode{{Type: utils.NMDataType, Value: nm}})
	case utils.MetaTmp:
		_, err = ar.tmp.Set(fullPath.PathSlice[1:], []*utils.DataNode{{Type: utils.NMDataType, Value: nm}})
		return
	case utils.MetaOpts:
		return ar.Opts.Set(fullPath.PathSlice[1:], nm.Data)
	case utils.MetaUCH:
		return engine.Cache.Set(utils.CacheUCH, fullPath.Path[5:], nm.Data, nil, true, utils.NonTransactional)
	}
}

// RemoveAll deletes all fields at given prefix
func (ar *AgentRequest) RemoveAll(prefix string) error {
	switch prefix {
	default:
		return fmt.Errorf("unsupported field prefix: <%s> when set fields", prefix)
	case utils.MetaVars:
		ar.Vars = &utils.DataNode{Type: utils.NMMapType, Map: make(map[string]*utils.DataNode)}
	case utils.MetaCgreq:
		ar.CGRRequest.RemoveAll()
	case utils.MetaCgrep:
		ar.CGRReply = &utils.DataNode{Type: utils.NMMapType, Map: make(map[string]*utils.DataNode)}
	case utils.MetaRep:
		ar.Reply.RemoveAll()
	case utils.MetaDiamreq:
		ar.diamreq.RemoveAll()
	case utils.MetaTmp:
		ar.tmp = &utils.DataNode{Type: utils.NMMapType, Map: make(map[string]*utils.DataNode)}
	case utils.MetaUCH:
		engine.Cache.Clear([]string{utils.CacheUCH})
	case utils.MetaOpts:
		ar.Opts = make(utils.MapStorage)
	}
	return nil
}

// Remove deletes the fields found at path with the given prefix
func (ar *AgentRequest) Remove(fullPath *utils.FullPath) error {
	switch fullPath.PathSlice[0] {
	default:
		return fmt.Errorf("unsupported field prefix: <%s> when set fields", fullPath.PathSlice[0])
	case utils.MetaVars:
		return ar.Vars.Remove(utils.CloneStringSlice(fullPath.PathSlice[1:]))
	case utils.MetaCgreq:
		return ar.CGRRequest.Remove(&utils.FullPath{
			PathSlice: fullPath.PathSlice[1:],
			Path:      fullPath.Path[7:],
		})
	case utils.MetaCgrep:
		return ar.CGRReply.Remove(utils.CloneStringSlice(fullPath.PathSlice[1:]))
	case utils.MetaRep:
		return ar.Reply.Remove(&utils.FullPath{
			PathSlice: fullPath.PathSlice[1:],
			Path:      fullPath.Path[5:],
		})
	case utils.MetaDiamreq:
		return ar.diamreq.Remove(&utils.FullPath{
			PathSlice: fullPath.PathSlice[1:],
			Path:      fullPath.Path[9:],
		})
	case utils.MetaTmp:
		return ar.tmp.Remove(utils.CloneStringSlice(fullPath.PathSlice[1:]))
	case utils.MetaOpts:
		return ar.Opts.Remove(fullPath.PathSlice[1:])
	case utils.MetaUCH:
		return engine.Cache.Remove(utils.CacheUCH, fullPath.Path[5:], true, utils.NonTransactional)
	}
}

// ParseField outputs the value based on the template item
func (ar *AgentRequest) ParseField(
	cfgFld *config.FCTemplate) (out interface{}, err error) {
	tmpType := cfgFld.Type
	switch tmpType {
	case utils.MetaFiller:
		cfgFld.Padding = utils.MetaRight
		tmpType = utils.MetaConstant
	case utils.MetaGroup:
		tmpType = utils.MetaVariable
	}
	out, err = engine.ParseAttribute(ar, tmpType, cfgFld.Path, cfgFld.Value, config.CgrConfig().GeneralCfg().RoundingDecimals, utils.FirstNonEmpty(cfgFld.Timezone, config.CgrConfig().GeneralCfg().DefaultTimezone), cfgFld.Layout, config.CgrConfig().GeneralCfg().RSRSep)

	if err != nil &&
		!strings.HasPrefix(err.Error(), "Could not find") {
		return
	}
	if utils.StringTmplType.Has(tmpType) { // format the string additionally with fmtFieldWidth
		out, err = utils.FmtFieldWidth(cfgFld.Tag, out.(string), cfgFld.Width,
			cfgFld.Strip, cfgFld.Padding, cfgFld.Mandatory)
	}
	return
}

// setCGRReply will set the aReq.cgrReply based on reply coming from upstream or error
// returns error in case of reply not converting to NavigableMap
func (ar *AgentRequest) setCGRReply(rply utils.NavigableMapper, err error) {
	ar.CGRReply.Map = make(map[string]*utils.DataNode)
	var errMsg string
	if err != nil {
		errMsg = err.Error()
	} else if rply != nil {
		ar.CGRReply.Map = rply.AsNavigableMap()
	}
	ar.CGRReply.Map[utils.Error] = utils.NewLeafNode(errMsg)
}

func needsMaxUsage(ralsFlags utils.FlagParams) bool {
	return len(ralsFlags) != 0 &&
		(ralsFlags.Has(utils.MetaAuthorize) ||
			ralsFlags.Has(utils.MetaInitiate) ||
			ralsFlags.Has(utils.MetaUpdate))
}

// Append sets the value at the given path
// this used with full path and the processed path to not calculate them for every set
func (ar *AgentRequest) Append(fullPath *utils.FullPath, val *utils.DataLeaf) (err error) {
	switch fullPath.PathSlice[0] {
	default:
		return fmt.Errorf("unsupported field prefix: <%s> when set field", fullPath.PathSlice[0])
	case utils.MetaVars:
		_, err = ar.Vars.Append(fullPath.PathSlice[1:], val)
		return
	case utils.MetaCgreq:
		return ar.CGRRequest.Append(&utils.FullPath{
			PathSlice: fullPath.PathSlice[1:],
			Path:      fullPath.Path[7:],
		}, val)
	case utils.MetaCgrep:
		_, err = ar.CGRReply.Append(fullPath.PathSlice[1:], val)
		return
	case utils.MetaRep:
		return ar.Reply.Append(&utils.FullPath{
			PathSlice: fullPath.PathSlice[1:],
			Path:      fullPath.Path[5:],
		}, val)
	case utils.MetaDiamreq:
		return ar.diamreq.Append(&utils.FullPath{
			PathSlice: fullPath.PathSlice[1:],
			Path:      fullPath.Path[9:],
		}, val)
	case utils.MetaTmp:
		_, err = ar.tmp.Append(fullPath.PathSlice[1:], val)
		return
	case utils.MetaOpts:
		return ar.Opts.Set(fullPath.PathSlice[1:], val.Data)
	case utils.MetaUCH:
		return engine.Cache.Set(utils.CacheUCH, fullPath.Path[5:], val.Data, nil, true, utils.NonTransactional)
	}
}

// Set sets the value at the given path
// this used with full path and the processed path to not calculate them for every set
func (ar *AgentRequest) Compose(fullPath *utils.FullPath, val *utils.DataLeaf) (err error) {
	switch fullPath.PathSlice[0] {
	default:
		return fmt.Errorf("unsupported field prefix: <%s> when set field", fullPath.PathSlice[0])
	case utils.MetaVars:
		return ar.Vars.Compose(fullPath.PathSlice[1:], val)
	case utils.MetaCgreq:
		return ar.CGRRequest.Compose(&utils.FullPath{
			PathSlice: fullPath.PathSlice[1:],
			Path:      fullPath.Path[7:],
		}, val)
	case utils.MetaCgrep:
		return ar.CGRReply.Compose(fullPath.PathSlice[1:], val)
	case utils.MetaRep:
		return ar.Reply.Compose(&utils.FullPath{
			PathSlice: fullPath.PathSlice[1:],
			Path:      fullPath.Path[5:],
		}, val)
	case utils.MetaDiamreq:
		return ar.diamreq.Compose(&utils.FullPath{
			PathSlice: fullPath.PathSlice[1:],
			Path:      fullPath.Path[9:],
		}, val)
	case utils.MetaTmp:
		return ar.tmp.Compose(fullPath.PathSlice[1:], val)
	case utils.MetaOpts:
		var prv interface{}
		if prv, err = ar.Opts.FieldAsInterface(fullPath.PathSlice[1:]); err != nil {
			if err != utils.ErrNotFound {
				return
			}
			prv = val.Data
		} else {
			prv = utils.IfaceAsString(prv) + utils.IfaceAsString(val.Data)
		}
		return ar.Opts.Set(fullPath.PathSlice[1:], prv)

	case utils.MetaUCH:
		path := fullPath.Path[5:]
		var prv interface{}
		if prvI, ok := engine.Cache.Get(utils.CacheUCH, path); !ok {
			prv = val.Data
		} else {
			prv = utils.IfaceAsString(prvI) + utils.IfaceAsString(val.Data)
		}
		return engine.Cache.Set(utils.CacheUCH, path, prv, nil, true, utils.NonTransactional)
	}
}
