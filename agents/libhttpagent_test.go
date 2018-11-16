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
	"bufio"
	"bytes"
	"net/http"
	"strings"
	"testing"
	//"github.com/cgrates/cgrates/utils"
)

func TestHttpUrlDPFieldAsInterface(t *testing.T) {
	br := bufio.NewReader(strings.NewReader(`GET /cdr?request_type=MOSMS_CDR&timestamp=2008-08-15%2017:49:21&message_date=2008-08-15%2017:49:21&transactionid=100744&CDR_ID=123456&carrierid=1&mcc=222&mnc=10&imsi=235180000000000&msisdn=%2B4977000000000&destination=%2B497700000001&message_status=0&IOT=0&service_id=1 HTTP/1.1
Host: api.cgrates.org

`))
	req, err := http.ReadRequest(br)
	if err != nil {
		t.Error(err)
	}
	hU, _ := newHTTPUrlDP(req)
	if data, err := hU.FieldAsString([]string{"request_type"}); err != nil {
		t.Error(err)
	} else if data != "MOSMS_CDR" {
		t.Errorf("expecting: MOSMS_CDR, received: <%s>", data)
	}
	if data, err := hU.FieldAsString([]string{"transactionid"}); err != nil {
		t.Error(err)
	} else if data != "100744" {
		t.Errorf("expecting: MOSMS_CDR, received: <%s>", data)
	}
	if data, err := hU.FieldAsString([]string{"nonexistent"}); err != nil {
		t.Error(err)
	} else if data != "" {
		t.Errorf("received: <%s>", data)
	}
}

/*
<?xml version="1.0"?>
<response status="success">
<api_call>SampleAPIMethod</api_call>
<SIM>
<PublicNumber>497924804904</PublicNumber>
</SIM>
</response>
*/

func TestHttpXmlDPFieldAsInterface(t *testing.T) {
	body := `<complete-success-notification callid="109870">
	<createtime>2005-08-26T14:16:42</createtime>
	<connecttime>2005-08-26T14:16:56</connecttime>
	<endtime>2005-08-26T14:17:34</endtime>
	<reference>My Call Reference</reference>
	<userid>386</userid>
	<username>sampleusername</username>
	<customerid>1</customerid>
	<companyname>Conecto LLC</companyname>
	<totalcost amount="0.21" currency="USD">US$0.21</totalcost>
	<hasrecording>yes</hasrecording>
	<hasvoicemail>no</hasvoicemail>
	<agenttotalcost amount="0.13" currency="USD">US$0.13</agenttotalcost>
	<agentid>44</agentid>
	<callleg calllegid="222146">
		<number>+441624828505</number>
		<description>Isle of Man</description>
		<seconds>38</seconds>
		<perminuterate amount="0.0200" currency="USD">US$0.0200</perminuterate>
		<cost amount="0.0140" currency="USD">US$0.0140</cost>
		<agentperminuterate amount="0.0130" currency="USD">US$0.0130</agentperminuterate>
		<agentcost amount="0.0082" currency="USD">US$0.0082</agentcost>
	</callleg>
	<callleg calllegid="222147">
		<number>+44 7624 494075</number>
		<description>Isle of Man</description>
		<seconds>37</seconds>
		<perminuterate amount="0.2700" currency="USD">US$0.2700</perminuterate>
		<cost amount="0.1890" currency="USD">US$0.1890</cost>
		<agentperminuterate amount="0.1880" currency="USD">US$0.1880</agentperminuterate>
		<agentcost amount="0.1159" currency="USD">US$0.1159</agentcost>
	</callleg>
</complete-success-notification>
`

	req, err := http.NewRequest("POST", "http://localhost:8080/", bytes.NewBuffer([]byte(body)))
	if err != nil {
		t.Error(err)
	}
	dP, _ := newHTTPXmlDP(req)
	if data, err := dP.FieldAsString([]string{"complete-success-notification", "userid"}); err != nil {
		t.Error(err)
	} else if data != "386" {
		t.Errorf("expecting: 386, received: <%s>", data)
	}
	if data, err := dP.FieldAsString([]string{"complete-success-notification", "username"}); err != nil {
		t.Error(err)
	} else if data != "sampleusername" {
		t.Errorf("expecting: sampleusername, received: <%s>", data)
	}
	if data, err := dP.FieldAsString([]string{"complete-success-notification", "callleg", "seconds"}); err != nil {
		t.Error(err)
	} else if data != "38" {
		t.Errorf("expecting: 38, received: <%s>", data)
	}
	if data, err := dP.FieldAsString([]string{"complete-success-notification", "callleg[1]", "seconds"}); err != nil {
		t.Error(err)
	} else if data != "37" {
		t.Errorf("expecting: 37, received: <%s>", data)
	}
	if data, err := dP.FieldAsString([]string{"complete-success-notification", "callleg[@calllegid='222147']", "seconds"}); err != nil {
		t.Error(err)
	} else if data != "37" {
		t.Errorf("expecting: 37, received: <%s>", data)
	}
	if data, err := dP.FieldAsString([]string{"complete-success-notification", "callleg", "@calllegid"}); err != nil {
		t.Error(err)
	} else if data != "222146" {
		t.Errorf("expecting: 222146, received: <%s>", data)
	}
	if data, err := dP.FieldAsString([]string{"complete-success-notification", "callleg[1]", "@calllegid"}); err != nil {
		t.Error(err)
	} else if data != "222147" {
		t.Errorf("expecting: 222147, received: <%s>", data)
	}
}

func TestHttpXmlDPFieldAsInterface2(t *testing.T) {
	body := `<?xml version="1.0" encoding="UTF-8"?>
   <sms-notification callid="145566709">
   <createtime>2018-11-15T15:11:26</createtime>
   <reference>SMS</reference>
   <calltype calltypeid="8">smsrelay</calltype>
   <userid>1636488</userid>
   <username>447440935378</username>
   <customerid>1632715</customerid>
   <companyname>447440935378</companyname>
   <totalcost amount="0.0000" currency="USD">0.0000</totalcost>
   <agenttotalcost amount="0.0360" currency="USD">0.0360</agenttotalcost>
   <agentid>2774</agentid>
   <callleg calllegid="219816629" calllegtype="mo">
      <number>447440935378</number>
      <ratedforuseras><![CDATA[UK Mobile - O2 [GBRCN] [MSRN]]]></ratedforuseras>
      <cost amount="0.0000" currency="USD">0.0000</cost>
      <agentcost amount="0.0135" currency="USD">0.0135</agentcost>
   </callleg>
   <callleg calllegid="219816630" calllegtype="mt">
      <number>447930323266</number>
      <ratedforuseras><![CDATA[UK Mobile - T-Mobile [GBRME]]]></ratedforuseras>
      <cost amount="0.0000" currency="USD">0.0000</cost>
      <agentcost amount="0.0225" currency="USD">0.0225</agentcost>
   </callleg>
</sms-notification>
`

	req, err := http.NewRequest("POST", "http://localhost:8080/", bytes.NewBuffer([]byte(body)))
	if err != nil {
		t.Error(err)
	}
	dP, _ := newHTTPXmlDP(req)
	if data, err := dP.FieldAsString([]string{"sms-notification", "callleg", "agentcost", "@amount"}); err != nil {
		t.Error(err)
	} else if data != "0.0135" {
		t.Errorf("expecting: 0.0135, received: <%s>", data)
	}
	if data, err := dP.FieldAsString([]string{"sms-notification", "callleg[0]", "agentcost"}); err != nil {
		t.Error(err)
	} else if data != "0.0135" {
		t.Errorf("expecting: 0.0135, received: <%s>", data)
	}
	if data, err := dP.FieldAsString([]string{"sms-notification", "callleg[1]", "agentcost"}); err != nil {
		t.Error(err)
	} else if data != "0.0225" {
		t.Errorf("expecting: 0.0225, received: <%s>", data)
	}
}
