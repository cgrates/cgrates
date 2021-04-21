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

package config

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/cgrates/cgrates/utils"
)

var (
	spltRgxp  = regexp.MustCompile(`:s\/`)
	rulesRgxp = regexp.MustCompile(`(?:(.*[^\\])\/(.*[^\\])*\/){1,}`)
)

// NewRSRParsers creates a new RSRParsers by spliting the rule using the separator
func NewRSRParsers(parsersRules string, rsrSeparator string) (prsrs RSRParsers, err error) {
	if parsersRules == utils.EmptyString {
		return
	}
	if count := strings.Count(parsersRules, utils.RSRConstSep); count%2 != 0 { // check if we have matching `
		return nil, fmt.Errorf("Closed unspilit syntax ")
	} else if count != 0 {
		var splitedRule []string
		for idx := strings.IndexByte(parsersRules, utils.RSRConstChar); idx != -1; idx = strings.IndexByte(parsersRules, utils.RSRConstChar) {
			insideARulePrefix := !strings.HasSuffix(parsersRules[:idx], utils.InfieldSep) // if doesn't have ; we need to concatenate it with last rule
			if insideARulePrefix {
				splitedRule = append(splitedRule, strings.Split(parsersRules[:idx], utils.InfieldSep)...)
			} else {
				splitedRule = append(splitedRule, strings.Split(parsersRules[:idx-1], utils.InfieldSep)...)
			}
			parsersRules = parsersRules[idx+1:]
			idx = strings.IndexByte(parsersRules, utils.RSRConstChar)
			if insideARulePrefix {
				splitedRule[len(splitedRule)-1] += parsersRules[:idx]
			} else {
				splitedRule = append(splitedRule, parsersRules[:idx])
			}
			parsersRules = parsersRules[idx+1:]
			count -= 2 // the number of ` remaining
			if len(parsersRules) == 0 {
				continue
			}
			insideARuleSufix := !strings.HasPrefix(parsersRules, utils.InfieldSep) // if doesn't have ; we need to concatenate it with last rule
			if insideARuleSufix {
				idx = strings.IndexByte(parsersRules, utils.FallbackSep) // ';'
				if idx == -1 {
					idx = len(parsersRules)
					splitedRule[len(splitedRule)-1] += parsersRules[:idx]
					break
				}
				splitedRule[len(splitedRule)-1] += parsersRules[:idx]
			} else {
				idx = 0
			}
			parsersRules = parsersRules[idx+1:]
			if len(parsersRules) == 0 {
				break
			}
			if count == 0 { // no more ` so add the rest
				splitedRule = append(splitedRule, strings.Split(parsersRules, utils.InfieldSep)...)
				break
			}
		}
		return NewRSRParsersFromSlice(splitedRule)
	}
	return NewRSRParsersFromSlice(strings.Split(parsersRules, rsrSeparator))
}

// NewRSRParsersFromSlice creates a new RSRParsers from a slice
func NewRSRParsersFromSlice(parsersRules []string) (prsrs RSRParsers, err error) {
	prsrs = make(RSRParsers, len(parsersRules))
	for i, rlStr := range parsersRules {
		if prsrs[i], err = NewRSRParser(rlStr); err != nil {
			return nil, err
		} else if prsrs[i] == nil {
			return nil, fmt.Errorf("emtpy RSRParser in rule: <%s>", rlStr)
		}
	}
	return
}

// NewRSRParsersMustCompile creates a new RSRParsers and panic if fails
func NewRSRParsersMustCompile(parsersRules string, rsrSeparator string) (prsrs RSRParsers) {
	var err error
	if prsrs, err = NewRSRParsers(parsersRules, rsrSeparator); err != nil {
		panic(fmt.Sprintf("rule: <%s>, error: %s", parsersRules, err.Error()))
	}
	return
}

// RSRParsers is a set of RSRParser
type RSRParsers []*RSRParser

// GetRule returns the original string from which the rules were composed
func (prsrs RSRParsers) GetRule(sep string) (out string) {
	for _, prsr := range prsrs {
		out += sep + prsr.Rules
	}
	if len(out) != 0 {
		out = out[1:]
	}
	return
}

// Compile parses Rules string and repopulates other fields
func (prsrs RSRParsers) Compile() (err error) {
	for _, prsr := range prsrs {
		if err = prsr.Compile(); err != nil {
			return
		}
	}
	return
}

// ParseValue will parse the value out considering converters
func (prsrs RSRParsers) ParseValue(value interface{}) (out string, err error) {
	for _, prsr := range prsrs {
		var outPrsr string
		if outPrsr, err = prsr.ParseValue(value); err != nil {
			return utils.EmptyString, err
		}
		out += outPrsr
	}
	return
}

// ParseDataProvider will parse the dataprovider using DPDynamicString
func (prsrs RSRParsers) ParseDataProvider(dP utils.DataProvider) (out string, err error) {
	for _, prsr := range prsrs {
		var outPrsr string
		if outPrsr, err = prsr.ParseDataProvider(dP); err != nil {
			return utils.EmptyString, err
		}
		out += outPrsr
	}
	return
}

// ParseDataProviderWithInterfaces will parse the dataprovider using DPDynamicInterface
func (prsrs RSRParsers) ParseDataProviderWithInterfaces(dP utils.DataProvider) (out string, err error) {
	for _, prsr := range prsrs {
		var outPrsr string
		if outPrsr, err = prsr.ParseDataProviderWithInterfaces(dP); err != nil {
			return utils.EmptyString, err
		}
		out += outPrsr
	}
	return
}

// GetIfaceFromValues returns an interface for each RSRParser
func (prsrs RSRParsers) GetIfaceFromValues(evNm utils.DataProvider) (iFaceVals []interface{}, err error) {
	iFaceVals = make([]interface{}, len(prsrs))
	for i, val := range prsrs {
		var strVal string
		if strVal, err = val.ParseDataProvider(evNm); err != nil {
			return
		}
		iFaceVals[i] = utils.StringToInterface(strVal)
	}
	return
}

// Clone returns a deep copy of RSRParsers
func (prsrs RSRParsers) Clone() (cln RSRParsers) {
	if prsrs == nil {
		return nil
	}
	cln = make(RSRParsers, len(prsrs))
	for i, prsr := range prsrs {
		cln[i] = prsr.Clone()
	}
	return
}

func (prsrs RSRParsers) AsStringSlice() (v []string) {
	v = make([]string, len(prsrs))
	for i, val := range prsrs {
		v[i] = val.Rules
	}
	return
}

// NewRSRParser builds one RSRParser
func NewRSRParser(parserRules string) (rsrParser *RSRParser, err error) {
	if len(parserRules) == 0 {
		return
	}
	rsrParser = &RSRParser{Rules: parserRules}
	if err = rsrParser.Compile(); err != nil {
		rsrParser = nil
	}
	return
}

// NewRSRParserMustCompile creates a new RSRParser and panic if fails
func NewRSRParserMustCompile(parserRules string) (rsrPrsr *RSRParser) {
	var err error
	if rsrPrsr, err = NewRSRParser(parserRules); err != nil {
		panic(fmt.Sprintf("compiling rules: <%s>, error: %s", parserRules, err.Error()))
	}
	return
}

// RSRParser is a parser for data coming from various sources
type RSRParser struct {
	Rules string // Rules container holding the string rules, public so it can be stored

	path       string                   // instruct extracting info out of header in event
	rsrRules   []*utils.ReSearchReplace // rules to use when parsing value
	converters utils.DataConverters     // set of converters to apply on output

	dynIdxStart int
	dynIdxEnd   int
	dynRules    RSRParsers
}

// AttrName exports the attribute name of the RSRParser
func (prsr *RSRParser) AttrName() string {
	return strings.TrimPrefix(prsr.path, utils.DynamicDataPrefix)
}

// Compile parses Rules string and repopulates other fields
func (prsr *RSRParser) Compile() (err error) {
	parserRules := prsr.Rules

	if dynIdxStart := strings.IndexByte(parserRules, utils.RSRDynStartChar); dynIdxStart != -1 {
		if dynIdxEnd := strings.IndexByte(parserRules[dynIdxStart:], utils.RSRDynEndChar); dynIdxEnd != -1 {
			var dynrules RSRParsers
			if dynrules, err = NewRSRParsers(parserRules[dynIdxStart+1:dynIdxStart+dynIdxEnd],
				CgrConfig().GeneralCfg().RSRSep); err != nil {
				return
			}
			prsr.dynRules = dynrules
			prsr.dynIdxStart = dynIdxStart
			prsr.dynIdxEnd = dynIdxStart + dynIdxEnd + 1
			return
		}
	}

	if idxConverters := strings.Index(parserRules, utils.RSRDataConverterPrefix); idxConverters != -1 { // converters in the string
		if !strings.HasSuffix(parserRules, utils.RSRDataConverterSufix) {
			return fmt.Errorf("invalid converter terminator in rule: <%s>",
				parserRules)
		}
		convertersStr := parserRules[idxConverters+1 : len(parserRules)-1] // strip also {}
		convsSplt := strings.Split(convertersStr, utils.ANDSep)
		prsr.converters = make(utils.DataConverters, len(convsSplt))
		for i, convStr := range convsSplt {
			var conv utils.DataConverter
			if conv, err = utils.NewDataConverter(convStr); err != nil {
				return fmt.Errorf("invalid converter value in string: <%s>, err: %s",
					convStr, err.Error())
			}
			prsr.converters[i] = conv
		}
		parserRules = parserRules[:idxConverters]
	}
	if !strings.HasPrefix(parserRules, utils.DynamicDataPrefix) ||
		len(parserRules) == 1 { // special case when RSR is defined as static attribute
		prsr.path = parserRules
		return
	}
	// dynamic content via attributeNames
	spltRules := spltRgxp.Split(parserRules, -1)
	prsr.path = spltRules[0] // in form ~hdr_name
	prsr.rsrRules = make([]*utils.ReSearchReplace, 0, len(spltRules[1:]))
	if len(spltRules) > 1 {
		for _, ruleStr := range spltRules[1:] { // :s/ already removed through split
			allMatches := rulesRgxp.FindStringSubmatch(ruleStr)
			if len(allMatches) != 3 {
				return fmt.Errorf("not enough members in Search&Replace, ruleStr: <%s>, matches: %v, ", ruleStr, allMatches)
			}
			var srRegexp *regexp.Regexp
			if srRegexp, err = regexp.Compile(allMatches[1]); err != nil {
				return fmt.Errorf("invalid Search&Replace subfield rule: <%s>", allMatches[1])
			}
			prsr.rsrRules = append(prsr.rsrRules, &utils.ReSearchReplace{
				SearchRegexp:    srRegexp,
				ReplaceTemplate: allMatches[2],
			})
		}
	}
	return
}

// parseValue the field value from a string
func (prsr *RSRParser) parseValue(value string) (out string, err error) {
	for _, rsRule := range prsr.rsrRules {
		value = rsRule.Process(value)
	}
	return prsr.converters.ConvertString(value)
}

// ParseValue will parse the value out considering converters
func (prsr *RSRParser) ParseValue(value interface{}) (out string, err error) {
	out = prsr.path
	if out != utils.DynamicDataPrefix &&
		strings.HasPrefix(out, utils.DynamicDataPrefix) { // Enforce parsing of static values
		out = utils.IfaceAsString(value)
	}
	return prsr.parseValue(out)
}

// ParseDataProvider will parse the dataprovider using DPDynamicString
func (prsr *RSRParser) ParseDataProvider(dP utils.DataProvider) (out string, err error) {
	if prsr.dynRules != nil {
		var dynPath string
		if dynPath, err = prsr.dynRules.ParseDataProvider(dP); err != nil {
			return
		}
		var dynRSR *RSRParser
		if dynRSR, err = NewRSRParser(prsr.Rules[:prsr.dynIdxStart] + dynPath + prsr.Rules[prsr.dynIdxEnd:]); err != nil {
			return
		}
		return dynRSR.ParseDataProvider(dP)
	}
	var outStr string
	if outStr, err = utils.DPDynamicString(prsr.path, dP); err != nil {
		return
	}
	return prsr.parseValue(outStr)
}

// ParseDataProviderWithInterfaces will parse the dataprovider using DPDynamicInterface
func (prsr *RSRParser) ParseDataProviderWithInterfaces(dP utils.DataProvider) (out string, err error) {
	if prsr.dynRules != nil {
		var dynPath string
		if dynPath, err = prsr.dynRules.ParseDataProvider(dP); err != nil {
			return
		}
		var dynRSR *RSRParser
		if dynRSR, err = NewRSRParser(prsr.Rules[:prsr.dynIdxStart] + dynPath + prsr.Rules[prsr.dynIdxEnd:]); err != nil {
			return
		}
		return dynRSR.ParseDataProviderWithInterfaces(dP)
	}
	var outIface interface{}
	if outIface, err = utils.DPDynamicInterface(prsr.path, dP); err != nil {
		return
	}
	return prsr.parseValue(utils.IfaceAsString(outIface))
}

// CompileDynRule will return the compiled dynamic rule
func (prsr *RSRParser) CompileDynRule(dP utils.DataProvider) (p string, err error) {
	if prsr.dynRules == nil {
		return prsr.Rules, nil
	}
	var dynPath string
	if dynPath, err = prsr.dynRules.ParseDataProvider(dP); err != nil {
		return
	}
	return prsr.Rules[:prsr.dynIdxStart] + dynPath + prsr.Rules[prsr.dynIdxEnd:], nil
}

// Clone returns a deep copy of RSRParser
func (prsr RSRParser) Clone() (cln *RSRParser) {
	cln = &RSRParser{
		Rules:       prsr.Rules,
		path:        prsr.path,
		dynIdxStart: prsr.dynIdxStart,
		dynIdxEnd:   prsr.dynIdxEnd,
		dynRules:    prsr.dynRules.Clone(),
	}
	if prsr.rsrRules != nil {
		cln.rsrRules = make([]*utils.ReSearchReplace, len(prsr.rsrRules))
		for i, rsr := range prsr.rsrRules {
			cln.rsrRules[i] = rsr.Clone()
		}
	}
	if prsr.converters != nil {
		cln.converters = make(utils.DataConverters, len(prsr.converters))
		for i, cnv := range prsr.converters {
			// we can't modify the convertor only overwirte it
			// safe to coppy it's value
			cln.converters[i] = cnv
		}
	}
	return
}
