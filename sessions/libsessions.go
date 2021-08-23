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

package sessions

import (
	"errors"
	"strings"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	jwt "github.com/dgrijalva/jwt-go"
)

var unratedReqs = engine.MapEvent{
	utils.MetaPostpaid:      struct{}{},
	utils.MetaPseudoPrepaid: struct{}{},
	utils.MetaRated:         struct{}{},
}

var authReqs = engine.MapEvent{
	utils.MetaPrepaid:       struct{}{},
	utils.MetaPseudoPrepaid: struct{}{},
}

// BiRPClient is the interface implemented by Agents which are able to
// communicate bidirectionally with SessionS and remote Communication Switch
type BiRPClient interface {
	V1DisconnectSession(ctx *context.Context, args utils.AttrDisconnectSession, reply *string) (err error)
	V1GetActiveSessionIDs(ctx *context.Context, ignParam string, sessionIDs *[]*SessionID) (err error)
	V1ReAuthorize(ctx *context.Context, originID string, reply *string) (err error)
	V1DisconnectPeer(ctx *context.Context, args *utils.DPRArgs, reply *string) (err error)
	V1WarnDisconnect(ctx *context.Context, args map[string]interface{}, reply *string) (err error)
}

// GetSetCGRID will populate the CGRID key if not present and return it
func GetSetCGRID(ev engine.MapEvent) (cgrID string) {
	cgrID = ev.GetStringIgnoreErrors(utils.CGRID)
	if cgrID == "" {
		cgrID = utils.Sha1(ev.GetStringIgnoreErrors(utils.OriginID),
			ev.GetStringIgnoreErrors(utils.OriginHost))
		ev[utils.CGRID] = cgrID
	}
	return
}

func getFlagIDs(flag string) []string {
	flagWithIDs := strings.Split(flag, utils.InInFieldSep)
	if len(flagWithIDs) <= 1 {
		return nil
	}
	return strings.Split(flagWithIDs[1], utils.ANDSep)
}

// ProcessedStirIdentity the structure that keeps all the header information
type ProcessedStirIdentity struct {
	Tokens     []string
	SigningStr string
	Signature  string
	Header     *utils.PASSporTHeader
	Payload    *utils.PASSporTPayload
}

// NewProcessedIdentity creates a proccessed header
func NewProcessedIdentity(identity string) (pi *ProcessedStirIdentity, err error) {
	pi = new(ProcessedStirIdentity)
	hdrtoken := strings.Split(utils.RemoveWhiteSpaces(identity), utils.InfieldSep)

	if len(hdrtoken) == 1 {
		err = errors.New("missing parts of the message header")
		return
	}
	pi.Tokens = hdrtoken[1:]
	btoken := utils.SplitPath(hdrtoken[0], utils.NestingSep[0], -1)
	if len(btoken) != 3 {
		err = errors.New("wrong header format")
		return
	}
	pi.SigningStr = btoken[0] + utils.NestingSep + btoken[1]
	pi.Signature = btoken[2]

	pi.Header = new(utils.PASSporTHeader)
	if err = utils.DecodeBase64JSON(btoken[0], pi.Header); err != nil {
		return
	}
	pi.Payload = new(utils.PASSporTPayload)
	err = utils.DecodeBase64JSON(btoken[1], pi.Payload)
	return
}

// VerifyHeader returns if the header is corectly populated
func (pi *ProcessedStirIdentity) VerifyHeader() (isValid bool) {
	var x5u string
	for _, pair := range pi.Tokens {
		ptoken := strings.Split(pair, utils.AttrValueSep)
		if len(ptoken) != 2 {
			continue
		}
		switch ptoken[0] {
		case utils.STIRAlgField:
			if ptoken[1] != utils.STIRAlg {
				return false
			}
		case utils.STIRPptField:
			if ptoken[1] != utils.STIRPpt &&
				ptoken[1] != "\""+utils.STIRPpt+"\"" {
				return false
			}
		case utils.STIRInfoField:
			lenParamInfo := len(ptoken[1])
			if lenParamInfo <= 2 {
				return false
			}
			x5u = ptoken[1]
			if x5u[0] == '<' && x5u[lenParamInfo-1] == '>' {
				x5u = x5u[1 : lenParamInfo-1]
			}
		}
	}

	return pi.Header.Alg == utils.STIRAlg &&
		pi.Header.Ppt == utils.STIRPpt &&
		pi.Header.Typ == utils.STIRTyp &&
		pi.Header.X5u == x5u
}

// VerifySignature returns if the signature is valid
func (pi *ProcessedStirIdentity) VerifySignature(ctx *context.Context, timeoutVal time.Duration) (err error) {
	var pubkey interface{}
	var ok bool
	if pubkey, ok = engine.Cache.Get(utils.CacheSTIR, pi.Header.X5u); !ok {
		if pubkey, err = utils.NewECDSAPubKey(pi.Header.X5u, timeoutVal); err != nil {
			if errCh := engine.Cache.Set(ctx, utils.CacheSTIR, pi.Header.X5u, nil,
				nil, false, utils.NonTransactional); errCh != nil {
				return errCh
			}
			return
		}
		if errCh := engine.Cache.Set(ctx, utils.CacheSTIR, pi.Header.X5u, pubkey,
			nil, false, utils.NonTransactional); errCh != nil {
			return errCh
		}
	}

	sigMethod := jwt.GetSigningMethod(pi.Header.Alg)
	return sigMethod.Verify(pi.SigningStr, pi.Signature, pubkey)

}

// VerifyPayload returns if the payload is corectly populated
func (pi *ProcessedStirIdentity) VerifyPayload(originatorTn, originatorURI, destinationTn, destinationURI string,
	hdrMaxDur time.Duration, attest utils.StringSet) (err error) {
	if !attest.Has(utils.MetaAny) && !attest.Has(pi.Payload.ATTest) {
		return errors.New("wrong attest level")
	}
	if hdrMaxDur >= 0 && time.Now().After(time.Unix(pi.Payload.IAT, 0).Add(hdrMaxDur)) {
		return errors.New("expired payload")
	}
	if originatorURI != utils.EmptyString {
		if originatorURI != pi.Payload.Orig.URI {
			return errors.New("wrong originatorURI")
		}
	} else if originatorTn != pi.Payload.Orig.Tn {
		return errors.New("wrong originatorTn")
	}
	if destinationURI != utils.EmptyString {
		if !utils.SliceHasMember(pi.Payload.Dest.URI, destinationURI) {
			return errors.New("wrong destinationURI")
		}
	} else if !utils.SliceHasMember(pi.Payload.Dest.Tn, destinationTn) {
		return errors.New("wrong destinationTn")
	}
	return
}

// NewSTIRIdentity returns the identiy for stir header
func NewSTIRIdentity(ctx *context.Context, header *utils.PASSporTHeader, payload *utils.PASSporTPayload, prvkeyPath string, timeout time.Duration) (identity string, err error) {
	var prvKey interface{}
	var ok bool
	if prvKey, ok = engine.Cache.Get(utils.CacheSTIR, prvkeyPath); !ok {
		if prvKey, err = utils.NewECDSAPrvKey(prvkeyPath, timeout); err != nil {
			if errCh := engine.Cache.Set(ctx, utils.CacheSTIR, prvkeyPath, nil,
				nil, false, utils.NonTransactional); errCh != nil {
				return utils.EmptyString, errCh
			}
			return
		}
		if errCh := engine.Cache.Set(ctx, utils.CacheSTIR, prvkeyPath, prvKey,
			nil, false, utils.NonTransactional); errCh != nil {
			return utils.EmptyString, errCh
		}
	}
	var headerStr, payloadStr string
	if headerStr, err = utils.EncodeBase64JSON(header); err != nil {
		return
	}
	if payloadStr, err = utils.EncodeBase64JSON(payload); err != nil {
		return
	}
	identity = headerStr + utils.NestingSep + payloadStr

	sigMethod := jwt.GetSigningMethod(header.Alg)
	var signature string
	if signature, err = sigMethod.Sign(identity, prvKey); err != nil {
		return
	}
	identity += utils.NestingSep + signature
	identity += utils.STIRExtraInfoPrefix + header.X5u + utils.STIRExtraInfoSuffix
	return
}

// AuthStirShaken autentificates the given identity using STIR/SHAKEN
func AuthStirShaken(ctx *context.Context, identity, originatorTn, originatorURI, destinationTn, destinationURI string,
	attest utils.StringSet, hdrMaxDur time.Duration) (err error) {
	var pi *ProcessedStirIdentity
	if pi, err = NewProcessedIdentity(identity); err != nil {
		return
	}
	if !pi.VerifyHeader() {
		return errors.New("wrong header")
	}
	if err = pi.VerifySignature(ctx, config.CgrConfig().GeneralCfg().ReplyTimeout); err != nil {
		return
	}
	return pi.VerifyPayload(originatorTn, originatorURI, destinationTn, destinationURI, hdrMaxDur, attest)
}

// V1STIRAuthenticateArgs are the arguments for STIRAuthenticate API
type V1STIRAuthenticateArgs struct {
	Attest             []string // what attest levels are allowed
	DestinationTn      string   // the expected destination telephone number
	DestinationURI     string   // the expected destination URI; if this is populated the DestinationTn is ignored
	Identity           string   // the identity header
	OriginatorTn       string   // the expected originator telephone number
	OriginatorURI      string   // the expected originator URI; if this is populated the OriginatorTn is ignored
	PayloadMaxDuration string   // the duration the payload is valid after it's creation
	APIOpts            map[string]interface{}
}

// V1STIRIdentityArgs are the arguments for STIRIdentity API
type V1STIRIdentityArgs struct {
	Payload        *utils.PASSporTPayload // the STIR payload
	PublicKeyPath  string                 // the path to the public key used in the header
	PrivateKeyPath string                 // the private key path
	OverwriteIAT   bool                   // if true the IAT from payload is overwrited with the present unix timestamp
	APIOpts        map[string]interface{}
}

// getDerivedEvents returns only the *raw event if derivedReply flag is not specified
func getDerivedEvents(events map[string]*utils.CGREvent, derivedReply bool) map[string]*utils.CGREvent {
	if derivedReply {
		return events
	}
	return map[string]*utils.CGREvent{
		utils.MetaRaw: events[utils.MetaRaw],
	}
}

// V1ProcessEventArgs are the options passed to ProcessEvent API
type V1ProcessEventArgs struct {
	*utils.CGREvent
	utils.Paginator // for routes
}

// V1ProcessEventReply is the reply for the ProcessEvent API
type V1ProcessEventReply struct {
	MaxUsage           map[string]time.Duration                  `json:",omitempty"`
	Cost               map[string]float64                        `json:",omitempty"` // Cost is the cost received from Rater, ignoring accounting part
	ResourceAllocation map[string]string                         `json:",omitempty"`
	Attributes         map[string]*engine.AttrSProcessEventReply `json:",omitempty"`
	RouteProfiles      map[string]engine.SortedRoutesList        `json:",omitempty"`
	ThresholdIDs       map[string][]string                       `json:",omitempty"`
	StatQueueIDs       map[string][]string                       `json:",omitempty"`
	STIRIdentity       map[string]string                         `json:",omitempty"`
}

// AsNavigableMap is part of engine.NavigableMapper interface
func (v1Rply *V1ProcessEventReply) AsNavigableMap() map[string]*utils.DataNode {
	cgrReply := make(map[string]*utils.DataNode)
	if v1Rply.MaxUsage != nil {
		usage := &utils.DataNode{Type: utils.NMMapType, Map: make(map[string]*utils.DataNode)}
		for k, v := range v1Rply.MaxUsage {
			usage.Map[k] = utils.NewLeafNode(v)
		}
		cgrReply[utils.CapMaxUsage] = usage
	}
	if v1Rply.ResourceAllocation != nil {
		res := &utils.DataNode{Type: utils.NMMapType, Map: make(map[string]*utils.DataNode)}
		for k, v := range v1Rply.ResourceAllocation {
			res.Map[k] = utils.NewLeafNode(v)
		}
		cgrReply[utils.CapResourceAllocation] = res
	}
	if v1Rply.Attributes != nil {
		atts := &utils.DataNode{Type: utils.NMMapType, Map: make(map[string]*utils.DataNode)}
		for k, att := range v1Rply.Attributes {
			attrs := &utils.DataNode{Type: utils.NMMapType, Map: make(map[string]*utils.DataNode)}
			for _, fldName := range att.AlteredFields {
				fldName = strings.TrimPrefix(fldName, utils.MetaReq+utils.NestingSep)
				if att.CGREvent.HasField(fldName) {
					attrs.Map[fldName] = utils.NewLeafNode(att.CGREvent.Event[fldName])
				}
			}
			atts.Map[k] = attrs
		}
		cgrReply[utils.CapAttributes] = atts
	}
	if v1Rply.RouteProfiles != nil {
		routes := &utils.DataNode{Type: utils.NMMapType, Map: make(map[string]*utils.DataNode)}
		for k, route := range v1Rply.RouteProfiles {
			routes.Map[k] = route.AsNavigableMap()
		}
		cgrReply[utils.CapRouteProfiles] = routes
	}
	if v1Rply.ThresholdIDs != nil {
		th := &utils.DataNode{Type: utils.NMMapType, Map: make(map[string]*utils.DataNode)}
		for k, thr := range v1Rply.ThresholdIDs {
			thIDs := &utils.DataNode{Type: utils.NMSliceType, Slice: make([]*utils.DataNode, len(thr))}
			for i, v := range thr {
				thIDs.Slice[i] = utils.NewLeafNode(v)
			}
			th.Map[k] = thIDs
		}
		cgrReply[utils.CapThresholds] = th
	}
	if v1Rply.StatQueueIDs != nil {
		st := &utils.DataNode{Type: utils.NMMapType, Map: make(map[string]*utils.DataNode)}
		for k, sts := range v1Rply.StatQueueIDs {
			stIDs := &utils.DataNode{Type: utils.NMSliceType, Slice: make([]*utils.DataNode, len(sts))}
			for i, v := range sts {
				stIDs.Slice[i] = utils.NewLeafNode(v)
			}
			st.Map[k] = stIDs
		}
		cgrReply[utils.CapStatQueues] = st
	}
	if v1Rply.Cost != nil {
		costs := &utils.DataNode{Type: utils.NMMapType, Map: make(map[string]*utils.DataNode)}
		for k, cost := range v1Rply.Cost {
			costs.Map[k] = utils.NewLeafNode(cost)
		}
	}
	if v1Rply.STIRIdentity != nil {
		stir := &utils.DataNode{Type: utils.NMMapType, Map: make(map[string]*utils.DataNode)}
		for k, v := range v1Rply.STIRIdentity {
			stir.Map[k] = utils.NewLeafNode(v)
		}
		cgrReply[utils.OptsStirIdentity] = stir
	}
	return cgrReply
}

// NewV1ProcessMessageArgs is a constructor for MessageArgs used by ProcessMessage
func NewV1ProcessMessageArgs(attrs bool, attributeIDs []string,
	thds bool, thresholdIDs []string, stats bool, statIDs []string, resrc, acnts,
	routes, routesIgnoreErrs, routesEventCost bool, cgrEv *utils.CGREvent,
	routePaginator utils.Paginator, forceDuration bool, routesMaxCost string) (args *V1ProcessMessageArgs) {
	args = &V1ProcessMessageArgs{
		AllocateResources:  resrc,
		Debit:              acnts,
		GetAttributes:      attrs,
		ProcessThresholds:  thds,
		ProcessStats:       stats,
		RoutesIgnoreErrors: routesIgnoreErrs,
		GetRoutes:          routes,
		CGREvent:           cgrEv,
		ForceDuration:      forceDuration,
	}
	if routesEventCost {
		args.RoutesMaxCost = utils.MetaEventCost
	} else {
		args.RoutesMaxCost = routesMaxCost
	}
	args.Paginator = routePaginator
	if len(attributeIDs) != 0 {
		args.AttributeIDs = attributeIDs
	}
	if len(thresholdIDs) != 0 {
		args.ThresholdIDs = thresholdIDs
	}
	if len(statIDs) != 0 {
		args.StatIDs = statIDs
	}
	return
}

// V1ProcessMessageArgs are the options passed to ProcessMessage API
type V1ProcessMessageArgs struct {
	GetAttributes      bool
	AllocateResources  bool
	Debit              bool
	ForceDuration      bool
	ProcessThresholds  bool
	ProcessStats       bool
	GetRoutes          bool
	RoutesMaxCost      string
	RoutesIgnoreErrors bool
	AttributeIDs       []string
	ThresholdIDs       []string
	StatIDs            []string
	*utils.CGREvent
	utils.Paginator
}

// ParseFlags will populate the V1ProcessMessageArgs flags
func (args *V1ProcessMessageArgs) ParseFlags(flags, sep string) {
	for _, subsystem := range strings.Split(flags, sep) {
		switch {
		case subsystem == utils.MetaAccounts:
			args.Debit = true
		case subsystem == utils.MetaResources:
			args.AllocateResources = true
		case subsystem == utils.MetaRoutes:
			args.GetRoutes = true
		case subsystem == utils.MetaRoutesIgnoreErrors:
			args.RoutesIgnoreErrors = true
		case subsystem == utils.MetaRoutesEventCost:
			args.RoutesMaxCost = utils.MetaEventCost
		case strings.HasPrefix(subsystem, utils.MetaRoutesMaxCost):
			args.RoutesMaxCost = strings.TrimPrefix(subsystem, utils.MetaRoutesMaxCost+utils.InInFieldSep)
		case strings.Index(subsystem, utils.MetaAttributes) != -1:
			args.GetAttributes = true
			args.AttributeIDs = getFlagIDs(subsystem)
		case strings.Index(subsystem, utils.MetaThresholds) != -1:
			args.ProcessThresholds = true
			args.ThresholdIDs = getFlagIDs(subsystem)
		case strings.Index(subsystem, utils.MetaStats) != -1:
			args.ProcessStats = true
			args.StatIDs = getFlagIDs(subsystem)
		case subsystem == utils.MetaFD:
			args.ForceDuration = true
		}
	}
	args.Paginator, _ = utils.GetRoutePaginatorFromOpts(args.APIOpts)
}

// V1ProcessMessageReply is the reply for the ProcessMessage API
type V1ProcessMessageReply struct {
	MaxUsage           *time.Duration                 `json:",omitempty"`
	ResourceAllocation *string                        `json:",omitempty"`
	Attributes         *engine.AttrSProcessEventReply `json:",omitempty"`
	RouteProfiles      engine.SortedRoutesList        `json:",omitempty"`
	ThresholdIDs       *[]string                      `json:",omitempty"`
	StatQueueIDs       *[]string                      `json:",omitempty"`

	needsMaxUsage bool // for gob encoding only
}

// SetMaxUsageNeeded used by agent that use the reply as NavigableMapper
// only used for gob encoding
func (v1Rply *V1ProcessMessageReply) SetMaxUsageNeeded(getMaxUsage bool) {
	if v1Rply == nil {
		return
	}
	v1Rply.needsMaxUsage = getMaxUsage
}

// AsNavigableMap is part of engine.NavigableMapper interface
func (v1Rply *V1ProcessMessageReply) AsNavigableMap() map[string]*utils.DataNode {
	cgrReply := make(map[string]*utils.DataNode)
	if v1Rply.MaxUsage != nil {
		cgrReply[utils.CapMaxUsage] = utils.NewLeafNode(*v1Rply.MaxUsage)
	} else if v1Rply.needsMaxUsage {
		cgrReply[utils.CapMaxUsage] = utils.NewLeafNode(0)
	}
	if v1Rply.ResourceAllocation != nil {
		cgrReply[utils.CapResourceAllocation] = utils.NewLeafNode(*v1Rply.ResourceAllocation)
	}
	if v1Rply.Attributes != nil {
		attrs := &utils.DataNode{Type: utils.NMMapType, Map: make(map[string]*utils.DataNode)}
		for _, fldName := range v1Rply.Attributes.AlteredFields {
			fldName = strings.TrimPrefix(fldName, utils.MetaReq+utils.NestingSep)
			if v1Rply.Attributes.CGREvent.HasField(fldName) {
				attrs.Map[fldName] = utils.NewLeafNode(v1Rply.Attributes.CGREvent.Event[fldName])
			}
		}
		cgrReply[utils.CapAttributes] = attrs
	}
	if v1Rply.RouteProfiles != nil {
		cgrReply[utils.CapRouteProfiles] = v1Rply.RouteProfiles.AsNavigableMap()
	}
	if v1Rply.ThresholdIDs != nil {
		thIDs := &utils.DataNode{Type: utils.NMSliceType, Slice: make([]*utils.DataNode, len(*v1Rply.ThresholdIDs))}
		for i, v := range *v1Rply.ThresholdIDs {
			thIDs.Slice[i] = utils.NewLeafNode(v)
		}
		cgrReply[utils.CapThresholds] = thIDs
	}
	if v1Rply.StatQueueIDs != nil {
		stIDs := &utils.DataNode{Type: utils.NMSliceType, Slice: make([]*utils.DataNode, len(*v1Rply.StatQueueIDs))}
		for i, v := range *v1Rply.StatQueueIDs {
			stIDs.Slice[i] = utils.NewLeafNode(v)
		}
		cgrReply[utils.CapStatQueues] = stIDs
	}
	return cgrReply
}

// NewV1AuthorizeArgs is a constructor for V1AuthorizeArgs
func NewV1AuthorizeArgs(attrs bool, attributeIDs []string,
	thrslds bool, thresholdIDs []string, statQueues bool, statIDs []string,
	res, maxUsage, routes, routesIgnoreErrs, routesEventCost bool,
	cgrEv *utils.CGREvent, routePaginator utils.Paginator,
	forceDuration bool, routesMaxCost string) (args *V1AuthorizeArgs) {
	args = &V1AuthorizeArgs{
		GetAttributes:      attrs,
		AuthorizeResources: res,
		GetMaxUsage:        maxUsage,
		ProcessThresholds:  thrslds,
		ProcessStats:       statQueues,
		RoutesIgnoreErrors: routesIgnoreErrs,
		GetRoutes:          routes,
		CGREvent:           cgrEv,
		ForceDuration:      forceDuration,
	}
	if routesEventCost {
		args.RoutesMaxCost = utils.MetaEventCost
	} else {
		args.RoutesMaxCost = routesMaxCost
	}
	args.Paginator = routePaginator
	if len(attributeIDs) != 0 {
		args.AttributeIDs = attributeIDs
	}
	if len(thresholdIDs) != 0 {
		args.ThresholdIDs = thresholdIDs
	}
	if len(statIDs) != 0 {
		args.StatIDs = statIDs
	}

	return
}

// V1AuthorizeArgs are options available in auth request
type V1AuthorizeArgs struct {
	GetAttributes      bool
	AuthorizeResources bool
	GetMaxUsage        bool
	ForceDuration      bool
	ProcessThresholds  bool
	ProcessStats       bool
	GetRoutes          bool
	RoutesMaxCost      string
	RoutesIgnoreErrors bool
	AttributeIDs       []string
	ThresholdIDs       []string
	StatIDs            []string
	*utils.CGREvent
	utils.Paginator
}

// ParseFlags will populate the V1AuthorizeArgs flags
func (args *V1AuthorizeArgs) ParseFlags(flags, sep string) {
	for _, subsystem := range strings.Split(flags, sep) {
		switch {
		case subsystem == utils.MetaAccounts:
			args.GetMaxUsage = true
		case subsystem == utils.MetaResources:
			args.AuthorizeResources = true
		case subsystem == utils.MetaRoutes:
			args.GetRoutes = true
		case subsystem == utils.MetaRoutesIgnoreErrors:
			args.RoutesIgnoreErrors = true
		case subsystem == utils.MetaRoutesEventCost:
			args.RoutesMaxCost = utils.MetaEventCost
		case strings.HasPrefix(subsystem, utils.MetaRoutesMaxCost):
			args.RoutesMaxCost = strings.TrimPrefix(subsystem, utils.MetaRoutesMaxCost+utils.InInFieldSep)
		case strings.HasPrefix(subsystem, utils.MetaAttributes):
			args.GetAttributes = true
			args.AttributeIDs = getFlagIDs(subsystem)
		case strings.HasPrefix(subsystem, utils.MetaThresholds):
			args.ProcessThresholds = true
			args.ThresholdIDs = getFlagIDs(subsystem)
		case strings.HasPrefix(subsystem, utils.MetaStats):
			args.ProcessStats = true
			args.StatIDs = getFlagIDs(subsystem)
		case subsystem == utils.MetaFD:
			args.ForceDuration = true
		}
	}
	args.Paginator, _ = utils.GetRoutePaginatorFromOpts(args.APIOpts)
}

// V1AuthorizeReply are options available in auth reply
type V1AuthorizeReply struct {
	Attributes         *engine.AttrSProcessEventReply `json:",omitempty"`
	ResourceAllocation *string                        `json:",omitempty"`
	MaxUsage           *time.Duration                 `json:",omitempty"`
	RouteProfiles      engine.SortedRoutesList        `json:",omitempty"`
	ThresholdIDs       *[]string                      `json:",omitempty"`
	StatQueueIDs       *[]string                      `json:",omitempty"`

	needsMaxUsage bool // for gob encoding only
}

// SetMaxUsageNeeded used by agent that use the reply as NavigableMapper
// only used for gob encoding
func (v1AuthReply *V1AuthorizeReply) SetMaxUsageNeeded(getMaxUsage bool) {
	if v1AuthReply == nil {
		return
	}
	v1AuthReply.needsMaxUsage = getMaxUsage
}

// AsNavigableMap is part of engine.NavigableMapper interface
func (v1AuthReply *V1AuthorizeReply) AsNavigableMap() map[string]*utils.DataNode {
	cgrReply := make(map[string]*utils.DataNode)
	if v1AuthReply.Attributes != nil {
		attrs := &utils.DataNode{Type: utils.NMMapType, Map: make(map[string]*utils.DataNode)}
		for _, fldName := range v1AuthReply.Attributes.AlteredFields {
			fldName = strings.TrimPrefix(fldName, utils.MetaReq+utils.NestingSep)
			if v1AuthReply.Attributes.CGREvent.HasField(fldName) {
				attrs.Map[fldName] = utils.NewLeafNode(v1AuthReply.Attributes.CGREvent.Event[fldName])
			}
		}
		cgrReply[utils.CapAttributes] = attrs
	}
	if v1AuthReply.ResourceAllocation != nil {
		cgrReply[utils.CapResourceAllocation] = utils.NewLeafNode(*v1AuthReply.ResourceAllocation)
	}
	if v1AuthReply.MaxUsage != nil {
		cgrReply[utils.CapMaxUsage] = utils.NewLeafNode(*v1AuthReply.MaxUsage)
	} else if v1AuthReply.needsMaxUsage {
		cgrReply[utils.CapMaxUsage] = utils.NewLeafNode(0)
	}

	if v1AuthReply.RouteProfiles != nil {
		nm := v1AuthReply.RouteProfiles.AsNavigableMap()
		cgrReply[utils.CapRouteProfiles] = nm
	}
	if v1AuthReply.ThresholdIDs != nil {
		thIDs := &utils.DataNode{Type: utils.NMSliceType, Slice: make([]*utils.DataNode, len(*v1AuthReply.ThresholdIDs))}
		for i, v := range *v1AuthReply.ThresholdIDs {
			thIDs.Slice[i] = utils.NewLeafNode(v)
		}
		cgrReply[utils.CapThresholds] = thIDs
	}
	if v1AuthReply.StatQueueIDs != nil {
		stIDs := &utils.DataNode{Type: utils.NMSliceType, Slice: make([]*utils.DataNode, len(*v1AuthReply.StatQueueIDs))}
		for i, v := range *v1AuthReply.StatQueueIDs {
			stIDs.Slice[i] = utils.NewLeafNode(v)
		}
		cgrReply[utils.CapStatQueues] = stIDs
	}
	return cgrReply
}

// V1AuthorizeReplyWithDigest contains return options for auth with digest
type V1AuthorizeReplyWithDigest struct {
	AttributesDigest   *string
	ResourceAllocation *string
	MaxUsage           float64 // special treat returning time.Duration.Seconds()
	RoutesDigest       *string
	Thresholds         *string
	StatQueues         *string
}

// NewV1InitSessionArgs is a constructor for V1InitSessionArgs
func NewV1InitSessionArgs(attrs bool, attributeIDs []string,
	thrslds bool, thresholdIDs []string, stats bool, statIDs []string,
	resrc, acnt bool, cgrEv *utils.CGREvent, forceDuration bool) (args *V1InitSessionArgs) {
	args = &V1InitSessionArgs{
		GetAttributes:     attrs,
		AllocateResources: resrc,
		InitSession:       acnt,
		ProcessThresholds: thrslds,
		ProcessStats:      stats,
		CGREvent:          cgrEv,
		ForceDuration:     forceDuration,
	}
	if len(attributeIDs) != 0 {
		args.AttributeIDs = attributeIDs
	}
	if len(thresholdIDs) != 0 {
		args.ThresholdIDs = thresholdIDs
	}
	if len(statIDs) != 0 {
		args.StatIDs = statIDs
	}
	return
}

// V1InitSessionArgs are options for session initialization request
type V1InitSessionArgs struct {
	GetAttributes     bool
	AllocateResources bool
	InitSession       bool
	ForceDuration     bool
	ProcessThresholds bool
	ProcessStats      bool
	AttributeIDs      []string
	ThresholdIDs      []string
	StatIDs           []string
	*utils.CGREvent
}

// ParseFlags will populate the V1InitSessionArgs flags
func (args *V1InitSessionArgs) ParseFlags(flags, sep string) {
	for _, subsystem := range strings.Split(flags, sep) {
		switch {
		case subsystem == utils.MetaAccounts:
			args.InitSession = true
		case subsystem == utils.MetaResources:
			args.AllocateResources = true
		case strings.HasPrefix(subsystem, utils.MetaAttributes):
			args.GetAttributes = true
			args.AttributeIDs = getFlagIDs(subsystem)
		case strings.HasPrefix(subsystem, utils.MetaThresholds):
			args.ProcessThresholds = true
			args.ThresholdIDs = getFlagIDs(subsystem)
		case strings.HasPrefix(subsystem, utils.MetaStats):
			args.ProcessStats = true
			args.StatIDs = getFlagIDs(subsystem)
		case subsystem == utils.MetaFD:
			args.ForceDuration = true
		}
	}
}

// V1InitSessionReply are options for initialization reply
type V1InitSessionReply struct {
	Attributes         *engine.AttrSProcessEventReply `json:",omitempty"`
	ResourceAllocation *string                        `json:",omitempty"`
	MaxUsage           *time.Duration                 `json:",omitempty"`
	ThresholdIDs       *[]string                      `json:",omitempty"`
	StatQueueIDs       *[]string                      `json:",omitempty"`

	needsMaxUsage bool // for gob encoding only
}

// SetMaxUsageNeeded used by agent that use the reply as NavigableMapper
// only used for gob encoding
func (v1Rply *V1InitSessionReply) SetMaxUsageNeeded(getMaxUsage bool) {
	if v1Rply == nil {
		return
	}
	v1Rply.needsMaxUsage = getMaxUsage
}

// AsNavigableMap is part of engine.NavigableMapper interface
func (v1Rply *V1InitSessionReply) AsNavigableMap() map[string]*utils.DataNode {
	cgrReply := make(map[string]*utils.DataNode)
	if v1Rply.Attributes != nil {
		attrs := &utils.DataNode{Type: utils.NMMapType, Map: make(map[string]*utils.DataNode)}
		for _, fldName := range v1Rply.Attributes.AlteredFields {
			fldName = strings.TrimPrefix(fldName, utils.MetaReq+utils.NestingSep)
			if v1Rply.Attributes.CGREvent.HasField(fldName) {
				attrs.Map[fldName] = utils.NewLeafNode(v1Rply.Attributes.CGREvent.Event[fldName])
			}
		}
		cgrReply[utils.CapAttributes] = attrs
	}
	if v1Rply.ResourceAllocation != nil {
		cgrReply[utils.CapResourceAllocation] = utils.NewLeafNode(*v1Rply.ResourceAllocation)
	}
	if v1Rply.MaxUsage != nil {
		cgrReply[utils.CapMaxUsage] = utils.NewLeafNode(*v1Rply.MaxUsage)
	} else if v1Rply.needsMaxUsage {
		cgrReply[utils.CapMaxUsage] = utils.NewLeafNode(0)
	}

	if v1Rply.ThresholdIDs != nil {
		thIDs := &utils.DataNode{Type: utils.NMSliceType, Slice: make([]*utils.DataNode, len(*v1Rply.ThresholdIDs))}
		for i, v := range *v1Rply.ThresholdIDs {
			thIDs.Slice[i] = utils.NewLeafNode(v)
		}
		cgrReply[utils.CapThresholds] = thIDs
	}
	if v1Rply.StatQueueIDs != nil {
		stIDs := &utils.DataNode{Type: utils.NMSliceType, Slice: make([]*utils.DataNode, len(*v1Rply.StatQueueIDs))}
		for i, v := range *v1Rply.StatQueueIDs {
			stIDs.Slice[i] = utils.NewLeafNode(v)
		}
		cgrReply[utils.CapStatQueues] = stIDs
	}
	return cgrReply
}

// V1InitReplyWithDigest is the formated reply
type V1InitReplyWithDigest struct {
	AttributesDigest   *string
	ResourceAllocation *string
	MaxUsage           float64
	Thresholds         *string
	StatQueues         *string
}

// NewV1UpdateSessionArgs is a constructor for update session arguments
func NewV1UpdateSessionArgs(attrs bool, attributeIDs []string,
	acnts bool, cgrEv *utils.CGREvent, forceDuration bool) (args *V1UpdateSessionArgs) {
	args = &V1UpdateSessionArgs{
		GetAttributes: attrs,
		UpdateSession: acnts,
		CGREvent:      cgrEv,
		ForceDuration: forceDuration,
	}
	if len(attributeIDs) != 0 {
		args.AttributeIDs = attributeIDs
	}
	return
}

// V1UpdateSessionArgs contains options for session update
type V1UpdateSessionArgs struct {
	GetAttributes bool
	UpdateSession bool
	ForceDuration bool
	AttributeIDs  []string
	*utils.CGREvent
}

// V1UpdateSessionReply contains options for session update reply
type V1UpdateSessionReply struct {
	Attributes *engine.AttrSProcessEventReply `json:",omitempty"`
	MaxUsage   *time.Duration                 `json:",omitempty"`

	needsMaxUsage bool // for gob encoding only
}

// SetMaxUsageNeeded used by agent that use the reply as NavigableMapper
// only used for gob encoding
func (v1Rply *V1UpdateSessionReply) SetMaxUsageNeeded(getMaxUsage bool) {
	if v1Rply == nil {
		return
	}
	v1Rply.needsMaxUsage = getMaxUsage
}

// AsNavigableMap is part of engine.NavigableMapper interface
func (v1Rply *V1UpdateSessionReply) AsNavigableMap() map[string]*utils.DataNode {
	cgrReply := make(map[string]*utils.DataNode)
	if v1Rply.Attributes != nil {
		attrs := &utils.DataNode{Type: utils.NMMapType, Map: make(map[string]*utils.DataNode)}
		for _, fldName := range v1Rply.Attributes.AlteredFields {
			fldName = strings.TrimPrefix(fldName, utils.MetaReq+utils.NestingSep)
			if v1Rply.Attributes.CGREvent.HasField(fldName) {
				attrs.Map[fldName] = utils.NewLeafNode(v1Rply.Attributes.CGREvent.Event[fldName])
			}
		}
		cgrReply[utils.CapAttributes] = attrs
	}
	if v1Rply.MaxUsage != nil {
		cgrReply[utils.CapMaxUsage] = utils.NewLeafNode(*v1Rply.MaxUsage)
	} else if v1Rply.needsMaxUsage {
		cgrReply[utils.CapMaxUsage] = utils.NewLeafNode(0)
	}
	return cgrReply
}

// NewV1TerminateSessionArgs creates a new V1TerminateSessionArgs using the given arguments
func NewV1TerminateSessionArgs(acnts, resrc,
	thrds bool, thresholdIDs []string, stats bool,
	statIDs []string, cgrEv *utils.CGREvent, forceDuration bool) (args *V1TerminateSessionArgs) {
	args = &V1TerminateSessionArgs{
		TerminateSession:  acnts,
		ReleaseResources:  resrc,
		ProcessThresholds: thrds,
		ProcessStats:      stats,
		CGREvent:          cgrEv,
		ForceDuration:     forceDuration,
	}
	if len(thresholdIDs) != 0 {
		args.ThresholdIDs = thresholdIDs
	}
	if len(statIDs) != 0 {
		args.StatIDs = statIDs
	}
	return
}

// V1TerminateSessionArgs is used as argumen for TerminateSession
type V1TerminateSessionArgs struct {
	TerminateSession  bool
	ForceDuration     bool
	ReleaseResources  bool
	ProcessThresholds bool
	ProcessStats      bool
	ThresholdIDs      []string
	StatIDs           []string
	*utils.CGREvent
}

// ParseFlags will populate the V1TerminateSessionArgs flags
func (args *V1TerminateSessionArgs) ParseFlags(flags, sep string) {
	for _, subsystem := range strings.Split(flags, sep) {
		switch {
		case subsystem == utils.MetaAccounts:
			args.TerminateSession = true
		case subsystem == utils.MetaResources:
			args.ReleaseResources = true
		case strings.Index(subsystem, utils.MetaThresholds) != -1:
			args.ProcessThresholds = true
			args.ThresholdIDs = getFlagIDs(subsystem)
		case strings.Index(subsystem, utils.MetaStats) != -1:
			args.ProcessStats = true
			args.StatIDs = getFlagIDs(subsystem)
		case subsystem == utils.MetaFD:
			args.ForceDuration = true
		}
	}
}

// ArgsReplicateSessions used to specify wich Session to replicate over the given connections
type ArgsReplicateSessions struct {
	CGRID   string
	Passive bool
	ConnIDs []string
}
