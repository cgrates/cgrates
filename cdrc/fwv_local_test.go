/*
Real-time Charging System for Telecom & ISP environments
Copyright (C) 2012-2015 ITsysCOM GmbH

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

package cdrc

import ()

/*
var fwvCfgPath string
var fwvCfg *config.CGRConfig
var fwvRpc *rpc.Client
var fwvCdrcCfg *config.CdrcConfig
*/
var FW_CDR_FILE1 = `HDR0001DDB     ABC                                     Some Connect A.B.                       DDB-Some-10022-20120711-309.CDR         00030920120711100255                                                                    
CDR0000010  0 20120708181506000123451234         0040123123120                  004                                            000018009980010001ISDN  ABC   10Buiten uw regio                         EHV 00000009190000000009
CDR0000020  0 20120708190945000123451234         0040123123120                  004                                            000016009980010001ISDN  ABC   10Buiten uw regio                         EHV 00000009190000000009
CDR0000030  0 20120708191009000123451234         0040123123120                  004                                            000020009980010001ISDN  ABC   10Buiten uw regio                         EHV 00000009190000000009
CDR0000040  0 20120708231043000123451234         0040123123120                  004                                            000011009980010001ISDN  ABC   10Buiten uw regio                         EHV 00000009190000000009
CDR0000050  0 20120709122216000123451235         004212                         004                                            000217009980010001ISDN  ABC   10Buiten uw regio                         HMR 00000000190000000000
CDR0000060  0 20120709130542000123451236         0012323453                     004                                            000019009980010001ISDN  ABC   35Sterdiensten                            AP  00000000190000000000
CDR0000070  0 20120709140032000123451237         0040012323453100               001                                            000050009980030001ISDN  ABD   20Internationaal                          NLB 00000000190000000000
CDR0000080  0 20120709140142000123451237         0040012323453100               001                                            000050009980030001ISDN  ABD   20Internationaal                          NLB 00000000190000000000
CDR0000090  0 20120709150305000123451237         0040012323453100               001                                            000050009980030001ISDN  ABD   20Internationaal                          NLB 00000000190000000000
CDR0000100  0 20120709150414000123451237         0040012323453100               001                                            000057009980030001ISDN  ABD   20Internationaal                          NLB 00000000190000000000
CDR0000110  0 20120709150531000123451237         0040012323453100               001                                            000059009980030001ISDN  ABD   20Internationaal                          NLB 00000000190000000000
CDR0000120  0 20120709150635000123451237         0040012323453100               001                                            000050009980030001ISDN  ABD   20Internationaal                          NLB 00000000190000000000
CDR0000130  0 20120709151756000123451237         0040012323453100               001                                            000050009980030001ISDN  ABD   20Internationaal                          NLB 00000000190000000000
CDR0000140  0 20120709154549000123451237         0040012323453100               001                                            000052009980030001ISDN  ABD   20Internationaal                          NLB 00000000190000000000
CDR0000150  0 20120709154701000123451237         0040012323453100               001                                            000121009980030001ISDN  ABD   20Internationaal                          NLB 00000000190000000000
CDR0000160  0 20120709154842000123451237         0040012323453100               001                                            000055009980030001ISDN  ABD   20Internationaal                          NLB 00000000190000000000
CDR0000170  0 20120709154956000123451237         0040012323453100               001                                            000115009980030001ISDN  ABD   20Internationaal                          NLB 00000000190000000000
CDR0000180  0 20120709155131000123451237         0040012323453100               001                                            000059009980030001ISDN  ABD   20Internationaal                          NLB 00000000190000000000
CDR0000190  0 20120709155236000123451237         0040012323453100               001                                            000050009980030001ISDN  ABD   20Internationaal                          NLB 00000000190000000000
CDR0000200  0 20120709160309000123451237         0040012323453100               001                                            000100009980030001ISDN  ABD   20Internationaal                          NLB 00000000190000000000
CDR0000210  0 20120709160415000123451237         0040012323453100               001                                            000050009980030001ISDN  ABD   20Internationaal                          NLB 00000000190000000000
CDR0000220  0 20120709161739000123451237         0040012323453100               001                                            000058009980030001ISDN  ABD   20Internationaal                          NLB 00000000190000000000
CDR0000230  0 20120709170356000123123459         0040123234531                  004                                            000012002760010001ISDN  276   10Buiten uw regio                         TB  00000009190000000009
CDR0000240  0 20120709181036000123123450         0012323453                     004                                            000042009980010001ISDN  ABC   05Binnen uw regio                         AP  00000010190000000010
CDR0000250  0 20120709191245000123123458         0040123232350                  004                                            000012002760000001PSTN  276   10Buiten uw regio                         TB  00000009190000000009
CDR0000260  0 20120709202324000123123459         0040123234531                  004                                            000011002760010001ISDN  276   10Buiten uw regio                         TB  00000009190000000009
CDR0000270  0 20120709211756000123451237         0040012323453100               001                                            000051009980030001ISDN  ABD   20Internationaal                          NLB 00000000190000000000
CDR0000280  0 20120709211852000123451237         0040012323453100               001                                            000050009980030001ISDN  ABD   20Internationaal                          NLB 00000000190000000000
CDR0000290  0 20120709212904000123123458         0040123232350                  004                                            000012002760000001PSTN  276   10Buiten uw regio                         TB  00000009190000000009
CDR0000300  0 20120709073707000123123459         0040123234531                  004                                            000012002760010001ISDN  276   10Buiten uw regio                         TB  00000009190000000009
CDR0000310  0 20120709085451000123451237         0040012323453100               001                                            000744009980030001ISDN  ABD   20Internationaal                          NLB 00000000190000000000
CDR0000320  0 20120709091756000123451237         0040012323453100               001                                            000050009980030001ISDN  ABD   20Internationaal                          NLB 00000000190000000000
CDR0000330  0 20120710070434000123123458         0040123232350                  004                                            000012002760000001PSTN  276   10Buiten uw regio                         TB  00000009190000000009
TRL0001DDB     ABC                                     Some Connect A.B.                       DDB-Some-10022-20120711-309.CDR         0003090000003300000030550000000001000000000100Y                                         `
