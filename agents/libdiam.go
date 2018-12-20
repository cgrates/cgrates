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
	"github.com/fiorix/go-diameter/diam"
	"github.com/fiorix/go-diameter/diam/avp"
	"github.com/fiorix/go-diameter/diam/datatype"
	"github.com/fiorix/go-diameter/diam/dict"
)

func loadDictionaries(dictsDir, componentId string) error {
	fi, err := os.Stat(dictsDir)
	if err != nil {
		if strings.HasSuffix(err.Error(), "no such file or directory") {
			return fmt.Errorf("<%s> Invalid dictionaries folder: <%s>", componentId, dictsDir)
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
			utils.Logger.Info(fmt.Sprintf("<%s> Loading dictionary out of file %s", componentId, filePath))
			if err := dict.Default.LoadFile(filePath); err != nil {
				return err
			}
		}
		return nil
	})
}

// diamAVPValue will extract the go primary value out of diameter type value
func diamAVPAsIface(dAVP *diam.AVP) (val interface{}, err error) {
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
	var iface interface{}
	if iface, err = diamAVPAsIface(dAVP); err != nil {
		return
	}
	return utils.IfaceAsString(iface)
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
		if err == nil {
			return nil, err
		}
		return datatype.Float32(float32(f)), nil
	case datatype.Float64Type:
		f, err := strconv.ParseFloat(valStr, 64)
		if err == nil {
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

// messageAddAVPsWithPath will dynamically add AVPs into the message
// 	append:	append to the message, on false overwrite if AVP is single or add to group if AVP is Grouped
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
		newMsgAVP := diam.NewAVP(dictAVPs[i].Code, avp.Mbit, dictAVPs[i].VendorID, typeVal) // FixMe: maybe Mbit with dictionary one
		if i == lastAVPIdx-1 && !newBranch {
			for idx := i + 1; idx > 0; idx-- { //check if we can append to the last AVP
				avps, err := m.FindAVPsWithPath(path[:idx], dict.UndefinedVendorID)
				if err != nil {
					return err
				}
				if len(avps) != 0 { // Group AVP already in the message
					prevGrpData := avps[len(avps)-1].Data.(*diam.GroupedAVP)               // Take the last avp found to append there
					if newMsgAVP.Data.Type() == diam.GroupedAVPType && idx != lastAVPIdx { // check if we need to add a group to last avp
						prevGrpData.AVP = append(prevGrpData.AVP, newMsgAVP)
						m.Header.MessageLength += uint32(newMsgAVP.Len())
					} else {
						prevGrpData.AVP = append(prevGrpData.AVP, msgAVP)
						m.Header.MessageLength += uint32(msgAVP.Len())
					}
					return nil
				}
			}
		}
		msgAVP = newMsgAVP
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
			return nil
		}
	}
	m.AVP = append(m.AVP, msgAVP)
	m.Header.MessageLength += uint32(msgAVP.Len())
	return nil
}

// writeOnConn writes the message on connection, logs failures
func writeOnConn(c diam.Conn, m *diam.Message) {
	if _, err := m.WriteTo(c); err != nil {
		utils.Logger.Warning(fmt.Sprintf("<%s> failed writing message to %s, err: %s, msg: %s",
			utils.DiameterAgent, c.RemoteAddr(), err.Error(), m))
	}
}

// newDADataProvider constructs a DataProvider for a diameter message
func newDADataProvider(c diam.Conn, m *diam.Message) config.DataProvider {
	return &diameterDP{c: c, m: m, cache: config.NewNavigableMap(nil)}

}

// diameterDP implements engine.DataProvider, serving as diam.Message data decoder
// decoded data is only searched once and cached
type diameterDP struct {
	c     diam.Conn
	m     *diam.Message
	cache *config.NavigableMap
}

// String is part of engine.DataProvider interface
// when called, it will display the already parsed values out of cache
func (dP *diameterDP) String() string {
	return dP.m.String()
}

// AsNavigableMap is part of engine.DataProvider interface
func (dP *diameterDP) AsNavigableMap([]*config.FCTemplate) (
	nm *config.NavigableMap, err error) {
	return nil, utils.ErrNotImplemented
}

// FieldAsString is part of engine.DataProvider interface
func (dP *diameterDP) FieldAsString(fldPath []string) (data string, err error) {
	var valIface interface{}
	valIface, err = dP.FieldAsInterface(fldPath)
	if err != nil {
		return
	}
	return utils.IfaceAsString(valIface)
}

// RemoteHost is part of engine.DataProvider interface
func (dP *diameterDP) RemoteHost() net.Addr {
	return utils.NewNetAddr(dP.c.RemoteAddr().Network(), dP.c.RemoteAddr().String())
}

// FieldAsInterface is part of engine.DataProvider interface
func (dP *diameterDP) FieldAsInterface(fldPath []string) (data interface{}, err error) {
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
	if splt := strings.Split(lastPath, "["); len(splt) != 1 {
		lastPath = splt[0]
		if splt[1][len(splt[1])-1:] != "]" {
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
				slctr, err := config.NewRSRParser(slctrStr, true)
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
					} else if _, err := slctr.ParseValue(dataAVP); err != nil {
						if err != utils.ErrFilterNotPassingNoCaps {
							return nil, err
						}
						continue // filter not passing, not really error
					} else {
						selIndxs[k+1] += 1 // filter passing, index it with one higher to cover 0
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
				return nil, utils.ErrFilterNotPassingNoCaps
			}
		}
	}
	if slectedIdx >= len(avps) {
		return nil, errors.New("avp index higher than number of AVPs")
	}
	if data, err = diamAVPAsIface(avps[slectedIdx]); err != nil {
		return nil, err
	}
	dP.cache.Set(fldPath, data, false, false)
	return
}

// diamAnswer builds up the answer to be sent back to the client
func diamAnswer(m *diam.Message, resCode uint32, errFlag bool,
	rply *config.NavigableMap, tmz string) (a *diam.Message, err error) {
	a = newDiamAnswer(m, resCode)
	if errFlag {
		a.Header.CommandFlags = diam.ErrorFlag
	}
	// write reply into message
	pathIdx := make(map[string]int) // group items for same path
	for _, val := range rply.Values() {

		nmItms, isNMItems := val.([]*config.NMItem)
		if !isNMItems {
			return nil, fmt.Errorf("cannot encode reply value: %s, err: not NMItems", utils.ToJSON(val))
		}
		// find out the first itm which is not an attribute
		var itm *config.NMItem
		if len(nmItms) == 1 {
			itm = nmItms[0]
		} else { // only for groups
			for i, cfgItm := range nmItms {
				itmPath := strings.Join(cfgItm.Path, utils.NestingSep)
				if i == 0 { // path is common, increase it only once
					pathIdx[itmPath] += 1
				}
				if i == pathIdx[itmPath]-1 { // revert from multiple items to only one per config path
					itm = cfgItm
					break
				}
			}
		}

		if itm == nil {
			continue // all attributes, not writable to diameter packet
		}
		itmStr, err := utils.IfaceAsString(itm.Data)
		if err != nil {
			return nil, fmt.Errorf("cannot convert data: %+v to string, err: %s", itm.Data, err)
		}
		var newBranch bool
		if itm.Config != nil && itm.Config.NewBranch {
			newBranch = true
		}
		if err = messageSetAVPsWithPath(a, itm.Path,
			itmStr, newBranch, tmz); err != nil {
			return nil, fmt.Errorf("setting item with path: %+v got err: %s", itm.Path, err.Error())
		}
	}
	return
}

// negDiamAnswer is used to return the negative answer we need previous to
func diamErr(m *diam.Message, resCode uint32,
	reqVars map[string]interface{},
	tpl []*config.FCTemplate, tnt, tmz string,
	filterS *engine.FilterS) (a *diam.Message, err error) {
	aReq := newAgentRequest(
		newDADataProvider(nil, m), reqVars,
		config.NewNavigableMap(nil),
		nil, tnt, tmz, filterS)
	var rplyData *config.NavigableMap
	if rplyData, err = aReq.AsNavigableMap(tpl); err != nil {
		return
	}
	return diamAnswer(m, resCode, true, rplyData, tmz)
}

func diamBareErr(m *diam.Message, resCode uint32) (a *diam.Message) {
	a = m.Answer(resCode)
	a.Header.CommandFlags = diam.ErrorFlag
	return
}

func disectDiamListen(addrs string) (ipAddrs []string) {
	ipPort := strings.Split(addrs, utils.InInFieldSep)
	if ipPort[0] == "" {
		return
	}
	ips := strings.Split(ipPort[0], utils.HDR_VAL_SEP)
	ipAddrs = make([]string, len(ips))
	for i, ip := range ips {
		ipAddrs[i] = ip
	}
	return
}

// newDiamAnswer temporary until fiorix will fix the issue
func newDiamAnswer(m *diam.Message, resCode uint32) *diam.Message {
	nm := diam.NewMessage(
		m.Header.CommandCode,
		m.Header.CommandFlags&^diam.RequestFlag, // Reset the Request bit.
		m.Header.ApplicationID,
		m.Header.HopByHopID,
		m.Header.EndToEndID,
		m.Dictionary(),
	)
	if resCode != 0 {
		nm.NewAVP(avp.ResultCode, avp.Mbit, 0, datatype.Unsigned32(resCode))
	}
	return nm
}
