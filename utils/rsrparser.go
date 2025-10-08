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

package utils

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	spltRgxp  = regexp.MustCompile(`:s\/`)
	rulesRgxp = regexp.MustCompile(`(?:(.*[^\\])\/(.*[^\\])*\/){1,}`)
)

// NewRSRParsers creates a new RSRParsers by spliting the rule using the separator
func NewRSRParsers(parsersRules, sep string) (prsrs RSRParsers, err error) {
	if parsersRules == EmptyString {
		return
	}
	if count := strings.Count(parsersRules, RSRConstSep); count%2 != 0 { // check if we have matching `
		return nil, fmt.Errorf("Closed unspilit syntax")
	} else if count != 0 {
		var splitedRule []string
		for idx := strings.IndexByte(parsersRules, RSRConstChar); idx != -1; idx = strings.IndexByte(parsersRules, RSRConstChar) {
			insideARulePrefix := !strings.HasSuffix(parsersRules[:idx], InfieldSep) // if doesn't have ; we need to concatenate it with last rule
			if insideARulePrefix {
				splitedRule = append(splitedRule, strings.Split(parsersRules[:idx], InfieldSep)...)
			} else {
				splitedRule = append(splitedRule, strings.Split(parsersRules[:idx-1], InfieldSep)...)
			}
			parsersRules = parsersRules[idx+1:]
			idx = strings.IndexByte(parsersRules, RSRConstChar)
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
			insideARuleSufix := !strings.HasPrefix(parsersRules, InfieldSep) // if doesn't have ; we need to concatenate it with last rule
			if insideARuleSufix {
				idx = strings.IndexByte(parsersRules, FallbackSep) // ';'
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
				splitedRule = append(splitedRule, strings.Split(parsersRules, InfieldSep)...)
				break
			}
		}
		return NewRSRParsersFromSlice(splitedRule)
	}
	return NewRSRParsersFromSlice(strings.Split(parsersRules, sep))
}

// NewRSRParsersFromSlice creates a new RSRParsers from a slice
func NewRSRParsersFromSlice(parsersRules []string) (prsrs RSRParsers, err error) {
	prsrs = make(RSRParsers, len(parsersRules))
	for i, rlStr := range parsersRules {
		if prsrs[i], err = NewRSRParser(rlStr); err != nil {
			return nil, err
		} else if prsrs[i] == nil {
			return nil, fmt.Errorf("empty RSRParser in rule: <%s>", rlStr)
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
func (prsrs RSRParsers) GetRule() (out string) {
	for _, prsr := range prsrs {
		out += RSRSep + prsr.Rules
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
func (prsrs RSRParsers) ParseValue(value any) (out string, err error) {
	for _, prsr := range prsrs {
		var outPrsr string
		if outPrsr, err = prsr.ParseValue(value); err != nil {
			return EmptyString, err
		}
		out += outPrsr
	}
	return
}

// ParseDataProvider will parse the dataprovider using DPDynamicString
func (prsrs RSRParsers) ParseDataProvider(dP DataProvider) (out string, err error) {
	for _, prsr := range prsrs {
		var outPrsr string
		if outPrsr, err = prsr.ParseDataProvider(dP); err != nil {
			return EmptyString, err
		}
		out += outPrsr
	}
	return
}

// ParseDataProviderWithInterfaces will parse the dataprovider using DPDynamicInterface
func (prsrs RSRParsers) ParseDataProviderWithInterfaces(dP DataProvider) (out string, err error) {
	for _, prsr := range prsrs {
		var outPrsr string
		if outPrsr, err = prsr.ParseDataProviderWithInterfaces(dP); err != nil {
			return EmptyString, err
		}
		out += outPrsr
	}
	return
}

// ParseDataProviderWithInterfaces will parse the dataprovider using DPDynamicInterface
func (prsrs RSRParsers) ParseDataProviderWithInterfaces2(dP DataProvider) (out any, err error) {
	for i, prsr := range prsrs {
		outPrsr, err := prsr.ParseDataProviderWithInterfaces2(dP)
		if err != nil {
			return nil, err
		}
		if i == 0 {
			out = outPrsr
		} else {
			out = IfaceAsString(out) + IfaceAsString(outPrsr)
		}
	}
	return
}

// GetIfaceFromValues returns an interface for each RSRParser
func (prsrs RSRParsers) GetIfaceFromValues(evNm DataProvider) (iFaceVals []any, err error) {
	iFaceVals = make([]any, len(prsrs))
	for i, val := range prsrs {
		var strVal string
		if strVal, err = val.ParseDataProvider(evNm); err != nil {
			return
		}
		iFaceVals[i] = StringToInterface(strVal)
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

	Path       string             // instruct extracting info out of header in event
	rsrRules   []*ReSearchReplace // rules to use when parsing value
	converters DataConverters     // set of converters to apply on output

	dynIdxStart int
	dynIdxEnd   int
	dynRules    RSRParsers
}

// AttrName exports the attribute name of the RSRParser
func (prsr *RSRParser) AttrName() string {
	return strings.TrimPrefix(prsr.Path, DynamicDataPrefix)
}

// Compile parses Rules string and repopulates other fields
func (prsr *RSRParser) Compile() (err error) {
	parserRules := prsr.Rules

	if dynIdxStart := strings.IndexByte(parserRules, RSRDynStartChar); dynIdxStart != -1 {
		if dynIdxEnd := strings.IndexByte(parserRules[dynIdxStart:], RSRDynEndChar); dynIdxEnd != -1 {
			var dynrules RSRParsers
			if dynrules, err = NewRSRParsers(parserRules[dynIdxStart+1:dynIdxStart+dynIdxEnd],
				RSRSep); err != nil {
				return
			}
			prsr.dynRules = dynrules
			prsr.dynIdxStart = dynIdxStart
			prsr.dynIdxEnd = dynIdxStart + dynIdxEnd + 1
			return
		}
	}

	if idxConverters := strings.Index(parserRules, RSRDataConverterPrefix); idxConverters != -1 { // converters in the string
		if !strings.HasSuffix(parserRules, RSRDataConverterSufix) {
			return fmt.Errorf("invalid converter terminator in rule: <%s>",
				parserRules)
		}
		convertersStr := parserRules[idxConverters+1 : len(parserRules)-1] // strip also {}
		convsSplt := strings.Split(convertersStr, ANDSep)
		prsr.converters = make(DataConverters, len(convsSplt))
		for i, convStr := range convsSplt {
			var conv DataConverter
			if conv, err = NewDataConverter(convStr); err != nil {
				return fmt.Errorf("invalid converter value in string: <%s>, err: %s",
					convStr, err.Error())
			}
			prsr.converters[i] = conv
		}
		parserRules = parserRules[:idxConverters]
	}
	if !strings.HasPrefix(parserRules, DynamicDataPrefix) ||
		len(parserRules) == 1 { // special case when RSR is defined as static attribute
		prsr.Path = parserRules
		return
	}
	// dynamic content via attributeNames
	spltRules := spltRgxp.Split(parserRules, -1)
	prsr.Path = spltRules[0] // in form ~hdr_name
	prsr.rsrRules = make([]*ReSearchReplace, 0, len(spltRules[1:]))
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
			prsr.rsrRules = append(prsr.rsrRules, &ReSearchReplace{
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

// parseValue the field value from a string
func (prsr *RSRParser) parseValueInterface(value any) (out any, err error) {
	for _, rsRule := range prsr.rsrRules {
		value = rsRule.Process(IfaceAsString(value))
	}
	return prsr.converters.ConvertInterface(value)
}

// ParseValue will parse the value out considering converters
func (prsr *RSRParser) ParseValue(value any) (out string, err error) {
	out = prsr.Path
	if out != DynamicDataPrefix &&
		strings.HasPrefix(out, DynamicDataPrefix) { // Enforce parsing of static values
		out = IfaceAsString(value)
	}
	return prsr.parseValue(out)
}

// ParseDataProvider will parse the dataprovider using DPDynamicString
func (prsr *RSRParser) ParseDataProvider(dP DataProvider) (out string, err error) {
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
	if outStr, err = DPDynamicString(prsr.Path, dP); err != nil {
		return
	}
	return prsr.parseValue(outStr)
}

// ParseDataProviderWithInterfaces will parse the dataprovider using DPDynamicInterface
func (prsr *RSRParser) ParseDataProviderWithInterfaces(dP DataProvider) (out string, err error) {
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
	var outIface any
	if outIface, err = DPDynamicInterface(prsr.Path, dP); err != nil {
		return
	}
	return prsr.parseValue(IfaceAsString(outIface))
}

// ParseDataProviderWithInterfaces will parse the dataprovider using DPDynamicInterface
func (prsr *RSRParser) ParseDataProviderWithInterfaces2(dP DataProvider) (out any, err error) {
	if prsr.dynRules != nil {
		var dynPath string
		if dynPath, err = prsr.dynRules.ParseDataProvider(dP); err != nil {
			return
		}
		var dynRSR *RSRParser
		if dynRSR, err = NewRSRParser(prsr.Rules[:prsr.dynIdxStart] + dynPath + prsr.Rules[prsr.dynIdxEnd:]); err != nil {
			return
		}
		return dynRSR.ParseDataProviderWithInterfaces2(dP)
	}
	var outIface any
	if outIface, err = DPDynamicInterface(prsr.Path, dP); err != nil {
		return
	}
	return prsr.parseValueInterface(outIface)
}

// CompileDynRule will return the compiled dynamic rule
func (prsr *RSRParser) CompileDynRule(dP DataProvider) (p string, err error) {
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
		Path:        prsr.Path,
		dynIdxStart: prsr.dynIdxStart,
		dynIdxEnd:   prsr.dynIdxEnd,
		dynRules:    prsr.dynRules.Clone(),
	}
	if prsr.rsrRules != nil {
		cln.rsrRules = make([]*ReSearchReplace, len(prsr.rsrRules))
		for i, rsr := range prsr.rsrRules {
			cln.rsrRules[i] = rsr.Clone()
		}
	}
	if prsr.converters != nil {
		cln.converters = make(DataConverters, len(prsr.converters))
		// we can't modify the convertor only overwrite it
		// safe to copy it's value
		copy(cln.converters, prsr.converters)
	}
	return
}
