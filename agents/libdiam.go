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
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/fiorix/go-diameter/v4/diam"
	"github.com/fiorix/go-diameter/v4/diam/avp"
	"github.com/fiorix/go-diameter/v4/diam/datatype"
	"github.com/fiorix/go-diameter/v4/diam/dict"
)

func loadDictionaries(dictsDir, componentID string) error {
	fi, err := os.Stat(dictsDir)
	if err != nil {
		if strings.HasSuffix(err.Error(), "no such file or directory") {
			return fmt.Errorf("<%s> Invalid dictionaries folder: <%s>", componentID, dictsDir)
		}
		return err
	} else if !fi.IsDir() { // If config dir defined, needs to exist
		return fmt.Errorf("<DiameterAgent> Path: <%s> is not a directory", dictsDir)
	}
	return filepath.Walk(dictsDir, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			return nil
		}
		cfgFiles, err := filepath.Glob(filepath.Join(path, "*.xml")) // Only consider .xml files
		if err != nil {
			return err
		}
		if cfgFiles == nil { // No need of processing further since there are no dictionary files in the folder
			return nil
		}
		for _, filePath := range cfgFiles {
			utils.Logger.Info(fmt.Sprintf("<%s> Loading dictionary out of file %s", componentID, filePath))
			if err := dict.Default.LoadFile(filePath); err != nil {
				return err
			}
		}
		return nil
	})
}

// diamAVPValue will extract the go primary value out of diameter type value
func diamAVPAsIface(dAVP *diam.AVP) (val any, err error) {
	if dAVP == nil {
		return nil, errors.New("nil AVP")
	}
	switch dAVP.Data.Type() {
	default:
		return nil, fmt.Errorf("unsupported AVP data type: %d", dAVP.Data.Type())
	case datatype.AddressType:
		return net.IP([]byte(dAVP.Data.(datatype.Address))), nil
	case datatype.DiameterIdentityType:
		return string(dAVP.Data.(datatype.DiameterIdentity)), nil
	case datatype.DiameterURIType:
		return string(dAVP.Data.(datatype.DiameterURI)), nil
	case datatype.EnumeratedType:
		return int32(dAVP.Data.(datatype.Enumerated)), nil
	case datatype.Float32Type:
		return float32(dAVP.Data.(datatype.Float32)), nil
	case datatype.Float64Type:
		return float64(dAVP.Data.(datatype.Float64)), nil
	case datatype.IPFilterRuleType:
		return string(dAVP.Data.(datatype.IPFilterRule)), nil
	case datatype.IPv4Type:
		return net.IP([]byte(dAVP.Data.(datatype.IPv4))), nil
	case datatype.Integer32Type:
		return int32(dAVP.Data.(datatype.Integer32)), nil
	case datatype.Integer64Type:
		return int64(dAVP.Data.(datatype.Integer64)), nil
	case datatype.OctetStringType:
		return string(dAVP.Data.(datatype.OctetString)), nil
	case datatype.QoSFilterRuleType:
		return string(dAVP.Data.(datatype.QoSFilterRule)), nil
	case datatype.TimeType:
		return time.Time(dAVP.Data.(datatype.Time)), nil
	case datatype.UTF8StringType:
		return string(dAVP.Data.(datatype.UTF8String)), nil
	case datatype.Unsigned32Type:
		return uint32(dAVP.Data.(datatype.Unsigned32)), nil
	case datatype.Unsigned64Type:
		return uint64(dAVP.Data.(datatype.Unsigned64)), nil
	}
}

func diamAVPAsString(dAVP *diam.AVP) (s string, err error) {
	var iface any
	if iface, err = diamAVPAsIface(dAVP); err != nil {
		return
	}
	return utils.IfaceAsString(iface), nil
}

// newDiamDataType constructs dataType from valStr
func newDiamDataType(typ datatype.TypeID, valStr,
	tmz string) (dt datatype.Type, err error) {
	switch typ {
	default:
		return nil, fmt.Errorf("unsupported AVP data type: %d", typ)
	case datatype.AddressType:
		return datatype.Address(net.ParseIP(valStr)), nil
	case datatype.DiameterIdentityType:
		return datatype.DiameterIdentity(valStr), nil
	case datatype.DiameterURIType:
		return datatype.DiameterURI(valStr), nil
	case datatype.EnumeratedType:
		i, err := strconv.ParseInt(valStr, 10, 32)
		if err != nil {
			return nil, err
		}
		return datatype.Enumerated(int32(i)), nil
	case datatype.Float32Type:
		f, err := strconv.ParseFloat(valStr, 32)
		if err != nil {
			return nil, err
		}
		return datatype.Float32(float32(f)), nil
	case datatype.Float64Type:
		f, err := strconv.ParseFloat(valStr, 64)
		if err != nil {
			return nil, err
		}
		return datatype.Float64(f), nil
	case datatype.IPFilterRuleType:
		return datatype.IPFilterRule(valStr), nil
	case datatype.IPv4Type:
		return datatype.IPv4(net.ParseIP(valStr)), nil
	case datatype.Integer32Type:
		i, err := strconv.ParseInt(valStr, 10, 32)
		if err != nil {
			return nil, err
		}
		return datatype.Integer32(int32(i)), nil
	case datatype.Integer64Type:
		i, err := strconv.ParseInt(valStr, 10, 64)
		if err != nil {
			return nil, err
		}
		return datatype.Integer64(i), nil
	case datatype.OctetStringType:
		return datatype.OctetString(valStr), nil
	case datatype.QoSFilterRuleType:
		return datatype.QoSFilterRule(valStr), nil
	case datatype.TimeType:
		t, err := utils.ParseTimeDetectLayout(valStr, tmz)
		if err != nil {
			return nil, err
		}
		return datatype.Time(t), nil
	case datatype.UTF8StringType:
		return datatype.UTF8String(valStr), nil
	case datatype.Unsigned32Type:
		i, err := strconv.ParseUint(valStr, 10, 32)
		if err != nil {
			return nil, err
		}
		return datatype.Unsigned32(uint32(i)), nil
	case datatype.Unsigned64Type:
		i, err := strconv.ParseUint(valStr, 10, 64)
		if err != nil {
			return nil, err
		}
		return datatype.Unsigned64(i), nil
	}
}

func headerLen(a *diam.AVP) int {
	if a.Flags&avp.Vbit == avp.Vbit {
		return 12
	}
	return 8
}

func updateAVPLength(avps []*diam.AVP) (l int) {
	for _, avp := range avps {
		if v, ok := (avp.Data).(*diam.GroupedAVP); ok {
			avp.Length = headerLen(avp) + updateAVPLength(v.AVP)
		}
		l += avp.Length
	}
	return
}

// messageAddAVPsWithPath will dynamically add AVPs into the message
//
//	append:	append to the message, on false overwrite if AVP is single or add to group if AVP is Grouped
func messageSetAVPsWithPath(m *diam.Message, pathStr []string,
	avpValStr string, newBranch bool, tmz string) (err error) {
	if len(pathStr) == 0 {
		return errors.New("empty path as AVP filter")
	}
	path := utils.SliceStringToIface(pathStr)
	dictAVPs := make([]*dict.AVP, len(path)) // for each subpath, one dictionary AVP
	for i, subpath := range path {
		if dictAVP, err := m.Dictionary().FindAVP(m.Header.ApplicationID, subpath); err != nil {
			return err
		} else if dictAVP == nil {
			return fmt.Errorf("cannot find AVP with id: %s", path[len(path)-1])
		} else {
			dictAVPs[i] = dictAVP
		}
	}
	lastAVPIdx := len(path) - 1
	if dictAVPs[lastAVPIdx].Data.Type == diam.GroupedAVPType {
		return errors.New("last AVP in path cannot be GroupedAVP")
	}
	var msgAVP *diam.AVP // Keep a reference here towards last AVP
	for i := lastAVPIdx; i >= 0; i-- {
		var typeVal datatype.Type
		if i == lastAVPIdx {
			if typeVal, err = newDiamDataType(dictAVPs[i].Data.Type, avpValStr, tmz); err != nil {
				return err
			}
		} else {
			typeVal = &diam.GroupedAVP{
				AVP: []*diam.AVP{msgAVP}}
		}
		msgAVP = diam.NewAVP(dictAVPs[i].Code, avp.Mbit, dictAVPs[i].VendorID, typeVal) // FixMe: maybe Mbit with dictionary one
		if i > 0 && !newBranch {
			avps, err := m.FindAVPsWithPath(path[:i], dict.UndefinedVendorID)
			if err != nil {
				return err
			}
			if len(avps) != 0 { // Group AVP already in the message
				prevGrpData, ok := avps[len(avps)-1].Data.(*diam.GroupedAVP) // Take the last avp found to append there
				if ok {
					prevGrpData.AVP = append(prevGrpData.AVP, msgAVP)
					m.Header.MessageLength += uint32(msgAVP.Len())
					// updateAVPLenght(m.AVP)
					return nil
				}
			}
		}
	}
	if !newBranch { // Not group AVP, replace the previous set one with this one
		avps, err := m.FindAVPsWithPath(path, dict.UndefinedVendorID)
		if err != nil {
			return err
		}
		if len(avps) != 0 { // Group AVP already in the message
			m.Header.MessageLength -= uint32(avps[len(avps)-1].Len()) // decrease message length since we overwrite
			*avps[len(avps)-1] = *msgAVP
			m.Header.MessageLength += uint32(msgAVP.Len())
			// updateAVPLenght(m.AVP)
			return nil
		}
	}
	m.AVP = append(m.AVP, msgAVP)
	m.Header.MessageLength += uint32(msgAVP.Len())
	// updateAVPLenght(m.AVP)
	return nil
}

// writeOnConn writes the message on connection, logs failures
func writeOnConn(c diam.Conn, m *diam.Message) (err error) {
	if _, err = m.WriteTo(c); err != nil {
		utils.Logger.Warning(fmt.Sprintf("<%s> failed writing message to %s, err: %s, msg: %s",
			utils.DiameterAgent, c.RemoteAddr(), err.Error(), m))
	}
	return
}

// newDADataProvider constructs a DataProvider for a diameter message
func newDADataProvider(c diam.Conn, m *diam.Message) utils.DataProvider {
	return &diameterDP{c: c, m: m, cache: utils.MapStorage{}}

}

// diameterDP implements utils.DataProvider, serving as diam.Message data decoder
// decoded data is only searched once and cached
type diameterDP struct {
	c     diam.Conn
	m     *diam.Message
	cache utils.MapStorage
}

// String is part of utils.DataProvider interface
// when called, it will display the already parsed values out of cache
func (dP *diameterDP) String() string {
	return dP.m.String()
}

// FieldAsString is part of utils.DataProvider interface
func (dP *diameterDP) FieldAsString(fldPath []string) (data string, err error) {
	var valIface any
	valIface, err = dP.FieldAsInterface(fldPath)
	if err != nil {
		return
	}
	return utils.IfaceAsString(valIface), nil
}

// FieldAsInterface is part of utils.DataProvider interface
func (dP *diameterDP) FieldAsInterface(fldPath []string) (data any, err error) {
	if data, err = dP.cache.FieldAsInterface(fldPath); err != nil {
		if err != utils.ErrNotFound { // item found in cache
			return nil, err
		}
		err = nil // cancel previous err
	} else {
		return // data was found in cache
	}
	// lastPath can contain selector inside
	lastPath := fldPath[len(fldPath)-1]
	var slctrStr string
	if splt := strings.Split(lastPath, utils.IdxStart); len(splt) != 1 {
		lastPath = splt[0]
		if splt[1][len(splt[1])-1:] != utils.IdxEnd {
			return nil, fmt.Errorf("filter rule <%s> needs to end in ]", splt[1])
		}
		slctrStr = splt[1][:len(splt[1])-1] // also strip the last ]
	}
	pathIface := utils.SliceStringToIface(fldPath)
	if slctrStr != "" { // last path was having selector inside before
		pathIface[len(pathIface)-1] = lastPath // last path was changed
	}
	var avps []*diam.AVP
	if avps, err = dP.m.FindAVPsWithPath(
		pathIface, dict.UndefinedVendorID); err != nil {
		return nil, err
	} else if len(avps) == 0 {
		return nil, utils.ErrNotFound
	}
	slectedIdx := 0 // by default we select AVP[0]
	if slctrStr != "" {
		if slectedIdx, err = strconv.Atoi(slctrStr); err != nil { // not int, compile it as RSRParser
			selIndxs := make(map[int]int) // use it to find intersection of all matched filters
			slctrStrs := strings.Split(slctrStr, utils.PipeSep)
			for _, slctrStr := range slctrStrs {
				var fltrs utils.RSRFilters
				if strings.HasSuffix(slctrStr, utils.FilterValEnd) { // Has filter, populate the var
					fltrStart := strings.Index(slctrStr, utils.FilterValStart)
					if fltrStart < 1 {
						return nil, fmt.Errorf("invalid RSRFilter start rule in string: <%s>", slctrStr)
					}
					fltrVal := slctrStr[fltrStart+1 : len(slctrStr)-1]
					if fltrs, err = utils.ParseRSRFilters(fltrVal, utils.ANDSep); err != nil {
						return nil, fmt.Errorf("Invalid FilterValue in string: %s, err: %s", fltrVal, err.Error())
					}
					slctrStr = slctrStr[:fltrStart] // Take the filter part out before compiling further
				}

				slctr, err := config.NewRSRParser(slctrStr)
				if err != nil {
					return nil, err
				}
				var fltrAVPs []*diam.AVP
				for i := len(pathIface) - 1; i > 0; i-- {
					pathIface[i] = slctr.AttrName() // search for AVPs which are having common path but different end element
					pathIface = pathIface[:i+1]
					if fltrAVPs, err = dP.m.FindAVPsWithPath(pathIface, dict.UndefinedVendorID); err != nil || len(fltrAVPs) != 0 {
						break // found AVPs or got error, go and compare
					}
				}
				if err != nil {
					return nil, err
				} else if len(fltrAVPs) == 0 || len(fltrAVPs) != len(avps) {
					return nil, utils.ErrNotFound
				}
				for k, fAVP := range fltrAVPs {
					if dataAVP, err := diamAVPAsIface(fAVP); err != nil {
						return nil, err
					} else if fld, err := slctr.ParseValue(dataAVP); err != nil {
						if err == utils.ErrNotFound && fltrs.FilterRules() == "^$" {
							selIndxs[k+1]++ // filter passing, index it with one higher to cover 0
							continue        // filter not passing, not really error
						}
						return nil, err
					} else if fltrs.Pass(fld, true) {
						selIndxs[k+1]++ // filter passing, index it with one higher to cover 0
					}
				}
			}
			var oneMatches bool
			for k, matches := range selIndxs {
				if matches == len(slctrStrs) { // all filters in selection matching
					oneMatches = true
					slectedIdx = k - 1 // decrease it to reflect real index
					break
				}
			}
			if !oneMatches {
				return nil, utils.ErrNotFound // return NotFound and let the other subsystem handle it (e.g. FilterS )
			}
		}
	}
	if slectedIdx >= len(avps) {
		return nil, errors.New("avp index higher than number of AVPs")
	}
	if data, err = diamAVPAsIface(avps[slectedIdx]); err != nil {
		return nil, err
	}
	dP.cache.Set(fldPath, data)
	return
}

// updateDiamMsgFromNavMap will update the diameter message with items from navigable map
func updateDiamMsgFromNavMap(m *diam.Message, navMp *utils.OrderedNavigableMap, tmz string) (err error) {
	// write reply into message
	for el := navMp.GetFirstElement(); el != nil; el = el.Next() {
		path := el.Value
		nmIt, _ := navMp.Field(path)
		if nmIt == nil {
			continue // all attributes, not writable to diameter packet
		}
		path = path[:len(path)-1] // remove the last index
		if err = messageSetAVPsWithPath(m,
			path, nmIt.String(),
			nmIt.NewBranch, tmz); err != nil {
			return fmt.Errorf("setting item with path: %+v got err: %s", path, err.Error())
		}
	}
	return
}

// diamAnswer builds up the answer to be sent back to the client
func diamAnswer(m *diam.Message, resCode uint32, errFlag bool,
	rply *utils.OrderedNavigableMap, tmz string) (a *diam.Message, err error) {
	a = m.Answer(resCode)
	if errFlag {
		a.Header.CommandFlags = diam.ErrorFlag
	}
	if err = updateDiamMsgFromNavMap(a, rply, tmz); err != nil {
		return nil, err
	}
	return
}

// diamErr handles Diameter error scenarios by attempting to build a customized error answer
// based on the *err template.
func diamErr(c diam.Conn, m *diam.Message, resCode uint32, reqVars *utils.DataNode, cfg *config.CGRConfig,
	filterS *engine.FilterS) {
	tnt := cfg.GeneralCfg().DefaultTenant
	tmz := cfg.GeneralCfg().DefaultTimezone
	aReq := NewAgentRequest(newDADataProvider(nil, m), reqVars,
		nil, nil, nil, nil, tnt, tmz, filterS, nil)
	if err := aReq.SetFields(cfg.TemplatesCfg()[utils.MetaErr]); err != nil {
		utils.Logger.Warning(fmt.Sprintf(
			"<%s> message: %s - failed to parse *err template: %v",
			utils.DiameterAgent, m, err))
		writeOnConn(c, diamErrMsg(m, diam.UnableToComply,
			fmt.Sprintf("failed to parse *err template: %v", err)))
		return
	}
	diamAns, err := diamAnswer(m, resCode, true, aReq.Reply, tmz)
	if err != nil {
		utils.Logger.Warning(fmt.Sprintf(
			"<%s> message: %s - failed to build error answer: %v",
			utils.DiameterAgent, m, err))
		writeOnConn(c, diamErrMsg(m, diam.UnableToComply,
			fmt.Sprintf("failed to build error answer: %v", err)))
		return
	}
	writeOnConn(c, diamAns)
}

// diamErrMsg creates a Diameter error answer with the given result code and optional error message.
func diamErrMsg(m *diam.Message, resCode uint32, msg string) *diam.Message {
	ans := m.Answer(resCode)
	ans.Header.CommandFlags = diam.ErrorFlag
	if msg != "" {
		ans.NewAVP(avp.ErrorMessage, 0, 0, datatype.UTF8String(msg))
	}
	return ans
}

func disectDiamListen(addrs string) (ipAddrs []net.IP) {
	ipPort := strings.Split(addrs, utils.InInFieldSep)
	if ipPort[0] == "" {
		return
	}
	ips := strings.Split(ipPort[0], utils.HDRValSep)
	ipAddrs = make([]net.IP, len(ips))
	for i, ip := range ips {
		ipAddrs[i] = net.ParseIP(ip)
	}
	return
}

// diamMessageData is cached when data is needed (ie. )
type diamMsgData struct {
	c    diam.Conn
	m    *diam.Message
	vars *utils.DataNode
}
