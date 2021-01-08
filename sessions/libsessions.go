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

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/dgrijalva/jwt-go"
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
	Call(serviceMethod string, args interface{}, reply interface{}) error
	V1DisconnectSession(args utils.AttrDisconnectSession, reply *string) (err error)
	V1GetActiveSessionIDs(ignParam string, sessionIDs *[]*SessionID) (err error)
	V1ReAuthorize(originID string, reply *string) (err error)
	V1DisconnectPeer(args *utils.DPRArgs, reply *string) (err error)
	V1WarnDisconnect(args map[string]interface{}, reply *string) (err error)
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
	return strings.Split(flagWithIDs[1], utils.InfieldSep)
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
	btoken := strings.Split(hdrtoken[0], utils.NestingSep)
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
func (pi *ProcessedStirIdentity) VerifySignature(timeoutVal time.Duration) (err error) {
	var pubkey interface{}
	var ok bool
	if pubkey, ok = engine.Cache.Get(utils.CacheSTIR, pi.Header.X5u); !ok {
		if pubkey, err = utils.NewECDSAPubKey(pi.Header.X5u, timeoutVal); err != nil {
			if errCh := engine.Cache.Set(utils.CacheSTIR, pi.Header.X5u, nil,
				nil, false, utils.NonTransactional); errCh != nil {
				return errCh
			}
			return
		}
		if errCh := engine.Cache.Set(utils.CacheSTIR, pi.Header.X5u, pubkey,
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
func NewSTIRIdentity(header *utils.PASSporTHeader, payload *utils.PASSporTPayload, prvkeyPath string, timeout time.Duration) (identity string, err error) {
	var prvKey interface{}
	var ok bool
	if prvKey, ok = engine.Cache.Get(utils.CacheSTIR, prvkeyPath); !ok {
		if prvKey, err = utils.NewECDSAPrvKey(prvkeyPath, timeout); err != nil {
			if errCh := engine.Cache.Set(utils.CacheSTIR, prvkeyPath, nil,
				nil, false, utils.NonTransactional); errCh != nil {
				return utils.EmptyString, errCh
			}
			return
		}
		if errCh := engine.Cache.Set(utils.CacheSTIR, prvkeyPath, prvKey,
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
func AuthStirShaken(identity, originatorTn, originatorURI, destinationTn, destinationURI string,
	attest utils.StringSet, hdrMaxDur time.Duration) (err error) {
	var pi *ProcessedStirIdentity
	if pi, err = NewProcessedIdentity(identity); err != nil {
		return
	}
	if !pi.VerifyHeader() {
		return errors.New("wrong header")
	}
	if err = pi.VerifySignature(config.CgrConfig().GeneralCfg().ReplyTimeout); err != nil {
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
	Opts               map[string]interface{}
}

// V1STIRIdentityArgs are the arguments for STIRIdentity API
type V1STIRIdentityArgs struct {
	Payload        *utils.PASSporTPayload // the STIR payload
	PublicKeyPath  string                 // the path to the public key used in the header
	PrivateKeyPath string                 // the private key path
	OverwriteIAT   bool                   // if true the IAT from payload is overwrited with the present unix timestamp
	Opts           map[string]interface{}
}

// getDerivedEvents returns only the *raw event if derivedReply flag is not specified
func getDerivedEvents(events map[string]*utils.CGREventWithOpts, derivedReply bool) map[string]*utils.CGREventWithOpts {
	if derivedReply {
		return events
	}
	return map[string]*utils.CGREventWithOpts{
		utils.MetaRaw: events[utils.MetaRaw],
	}
}

// getDerivedMaxUsage returns only the *raw MaxUsage if derivedReply flag is not specified
func getDerivedMaxUsage(maxUsages map[string]time.Duration, derivedReply bool) (out map[string]time.Duration) {
	if derivedReply {
		out = maxUsages
	} else {
		out = make(map[string]time.Duration)
	}
	var maxUsage time.Duration
	var maxUsageSet bool // so we know if we have set the 0 on purpose
	for _, rplyMaxUsage := range maxUsages {
		if !maxUsageSet || rplyMaxUsage < maxUsage {
			maxUsage = rplyMaxUsage
			maxUsageSet = true
		}
	}
	out[utils.MetaRaw] = maxUsage
	return
}
