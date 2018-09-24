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
	"github.com/cgrates/cgrates/utils"
	"github.com/fiorix/go-diameter/diam"
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

// writeOnConn writes the message on connection, logs failures
func writeOnConn(c diam.Conn, m *diam.Message) {
	if _, err := m.WriteTo(c); err != nil {
		utils.Logger.Warning(fmt.Sprintf("<%s> failed writing message to %s, err: %s, msg: %s",
			utils.DiameterAgent, c.RemoteAddr(), err.Error(), m))
	}
}

// newDADataProvider constructs a DataProvider for a diameter message
func newDADataProvider(m *diam.Message) config.DataProvider {
	return &diameterDP{m: m, cache: config.NewNavigableMap(nil)}

}

// diameterDP implements engine.DataProvider, serving as diam.Message data decoder
// decoded data is only searched once and cached
type diameterDP struct {
	m     *diam.Message
	cache *config.NavigableMap
}

// String is part of engine.DataProvider interface
// when called, it will display the already parsed values out of cache
func (dP *diameterDP) String() string {
	return utils.ToJSON(dP.cache)
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
	data, _ = utils.CastFieldIfToString(valIface)
	return
}

// FieldAsInterface is part of engine.DataProvider interface
func (dP *diameterDP) FieldAsInterface(fldPath []string) (data interface{}, err error) {
	if data, err = dP.cache.FieldAsInterface(fldPath); err == nil ||
		err != utils.ErrNotFound { // item found in cache
		return nil, err
	}
	err = nil // cancel previous err
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
				pathIface[len(pathIface)-1] = slctr.AttrName() // search for AVPs which are having common path but different end element
				fltrAVPs, err := dP.m.FindAVPsWithPath(pathIface, dict.UndefinedVendorID)
				if err != nil {
					return nil, err
				} else if len(fltrAVPs) == 0 || len(fltrAVPs) != len(avps) {
					return nil, utils.ErrFilterNotPassingNoCaps
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
	dP.cache.Set(fldPath, data, false)
	return
}
