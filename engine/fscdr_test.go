/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2012-2015 ITsysCOM

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

package engine

import (
	"reflect"
	"regexp"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

var body = []byte(`{"core-uuid":"651a8db2-4f67-4cf8-b622-169e8a482e50","switchname":"CgrDev1","channel_data":{"state":"CS_REPORTING","direction":"inbound","state_number":"11","flags":"0=1;1=1;37=1;38=1;40=1;43=1;48=1;53=1;105=1;111=1;112=1;116=1;118=1","caps":"1=1;2=1;3=1;4=1;5=1;6=1"},"variables":{"direction":"inbound","uuid":"e3133bf7-dcde-4daf-9663-9a79ffcef5ad","session_id":"4","sip_from_user":"1001","sip_from_uri":"1001@127.0.0.1","sip_from_host":"127.0.0.1","channel_name":"sofia/cgrtest/1001@127.0.0.1","ep_codec_string":"speex@16000h@20i,speex@8000h@20i,speex@32000h@20i,GSM@8000h@20i@13200b,PCMU@8000h@20i@64000b,PCMA@8000h@20i@64000b,G722@8000h@20i@64000b","sip_local_network_addr":"127.0.0.1","sip_network_ip":"127.0.0.1","sip_network_port":"46615","sip_received_ip":"127.0.0.1","sip_received_port":"46615","sip_via_protocol":"tcp","sip_authorized":"true","Event-Name":"REQUEST_PARAMS","Core-UUID":"651a8db2-4f67-4cf8-b622-169e8a482e50","FreeSWITCH-Hostname":"CgrDev1","FreeSWITCH-Switchname":"CgrDev1","FreeSWITCH-IPv4":"10.0.3.15","FreeSWITCH-IPv6":"::1","Event-Date-Local":"2015-07-07 16:52:08","Event-Date-GMT":"Tue, 07 Jul 2015 14:52:08 GMT","Event-Date-Timestamp":"1436280728471153","Event-Calling-File":"sofia.c","Event-Calling-Function":"sofia_handle_sip_i_invite","Event-Calling-Line-Number":"9056","Event-Sequence":"515","sip_number_alias":"1001","sip_auth_username":"1001","sip_auth_realm":"127.0.0.1","number_alias":"1001","requested_domain_name":"cgrates.org","record_stereo":"true","transfer_fallback_extension":"operator","toll_allow":"domestic,international,local","accountcode":"1001","user_context":"default","effective_caller_id_name":"Extension 1001","effective_caller_id_number":"1001","outbound_caller_id_name":"FreeSWITCH","outbound_caller_id_number":"0000000000","callgroup":"techsupport","cgr_reqtype":"*prepaid","cgr_supplier":"supplier1","user_name":"1001","domain_name":"cgrates.org","sip_from_user_stripped":"1001","sofia_profile_name":"cgrtest","recovery_profile_name":"cgrtest","sip_full_route":"<sip:127.0.0.1:25060;lr>","sip_recover_via":"SIP/2.0/TCP 127.0.0.1:46615;rport=46615;branch=z9hG4bKPjGj7AlihmVwAVz9McwVeI64NeBHlPmXAN;alias","sip_req_user":"1003","sip_req_uri":"1003@127.0.0.1","sip_req_host":"127.0.0.1","sip_to_user":"1003","sip_to_uri":"1003@127.0.0.1","sip_to_host":"127.0.0.1","sip_contact_params":"ob","sip_contact_user":"1001","sip_contact_port":"5072","sip_contact_uri":"1001@127.0.0.1:5072","sip_contact_host":"127.0.0.1","sip_via_host":"127.0.0.1","sip_via_port":"46615","sip_via_rport":"46615","switch_r_sdp":"v=0\r\no=- 3645269528 3645269528 IN IP4 10.0.3.15\r\ns=pjmedia\r\nb=AS:84\r\nt=0 0\r\na=X-nat:0\r\nm=audio 4006 RTP/AVP 98 97 99 104 3 0 8 9 96\r\nc=IN IP4 10.0.3.15\r\nb=AS:64000\r\na=rtpmap:98 speex/16000\r\na=rtpmap:97 speex/8000\r\na=rtpmap:99 speex/32000\r\na=rtpmap:104 iLBC/8000\r\na=fmtp:104 mode=30\r\na=rtpmap:3 GSM/8000\r\na=rtpmap:0 PCMU/8000\r\na=rtpmap:8 PCMA/8000\r\na=rtpmap:9 G722/8000\r\na=rtpmap:96 telephone-event/8000\r\na=fmtp:96 0-16\r\na=rtcp:4007 IN IP4 10.0.3.15\r\n","rtp_remote_audio_rtcp_port":"4007 IN IP4 10.0.3.15","rtp_audio_recv_pt":"99","rtp_use_codec_name":"SPEEX","rtp_use_codec_rate":"32000","rtp_use_codec_ptime":"20","rtp_use_codec_channels":"1","rtp_last_audio_codec_string":"SPEEX@32000h@20i@1c","read_codec":"SPEEX","original_read_codec":"SPEEX","read_rate":"32000","original_read_rate":"32000","write_codec":"SPEEX","write_rate":"32000","dtmf_type":"rfc2833","execute_on_answer":"sched_hangup +3120 alloted_timeout","cgr_notify":"+AUTH_OK","max_forwards":"69","transfer_history":"1436280728:e7c250e8-6ad7-4bd4-8962-318e0b0da728:bl_xfer:1003/default/XML","transfer_source":"1436280728:e7c250e8-6ad7-4bd4-8962-318e0b0da728:bl_xfer:1003/default/XML","DP_MATCH":"ARRAY::1003|:1003","call_uuid":"e3133bf7-dcde-4daf-9663-9a79ffcef5ad","ringback":"%(2000,4000,440,480)","call_timeout":"30","dialed_user":"1003","dialed_domain":"cgrates.org","originated_legs":"ARRAY::0a30dd7c-c222-482f-a322-b1218a15f8cd;Outbound Call;1003|:0a30dd7c-c222-482f-a322-b1218a15f8cd;Outbound Call;1003","switch_m_sdp":"v=0\r\no=- 3645269528 3645269529 IN IP4 10.0.3.15\r\ns=pjmedia\r\nb=AS:84\r\nt=0 0\r\na=X-nat:0\r\nm=audio 4018 RTP/AVP 99 101\r\nc=IN IP4 10.0.3.15\r\nb=AS:64000\r\na=rtpmap:99 speex/32000\r\na=rtpmap:101 telephone-event/8000\r\na=fmtp:101 0-16\r\na=rtcp:4019 IN IP4 10.0.3.15\r\n","rtp_local_sdp_str":"v=0\no=FreeSWITCH 1436250882 1436250883 IN IP4 10.0.3.15\ns=FreeSWITCH\nc=IN IP4 10.0.3.15\nt=0 0\nm=audio 29846 RTP/AVP 99 96\na=rtpmap:99 speex/32000\na=rtpmap:96 telephone-event/8000\na=fmtp:96 0-16\na=ptime:20\na=sendrecv\na=rtcp:29847 IN IP4 10.0.3.15\n","local_media_ip":"10.0.3.15","local_media_port":"29846","advertised_media_ip":"10.0.3.15","rtp_use_pt":"99","rtp_use_ssrc":"1470667272","rtp_2833_send_payload":"96","rtp_2833_recv_payload":"96","remote_media_ip":"10.0.3.15","remote_media_port":"4006","endpoint_disposition":"ANSWER","current_application_data":"+3120 alloted_timeout","current_application":"sched_hangup","originate_causes":"ARRAY::0a30dd7c-c222-482f-a322-b1218a15f8cd;NONE|:0a30dd7c-c222-482f-a322-b1218a15f8cd;NONE","originate_disposition":"SUCCESS","DIALSTATUS":"SUCCESS","last_bridge_to":"0a30dd7c-c222-482f-a322-b1218a15f8cd","bridge_channel":"sofia/cgrtest/1003@127.0.0.1:5070","bridge_uuid":"0a30dd7c-c222-482f-a322-b1218a15f8cd","signal_bond":"0a30dd7c-c222-482f-a322-b1218a15f8cd","sip_to_tag":"5Qt4ecvreSHZN","sip_from_tag":"YwuG8U3rRbqIn.xYTnU8NrI3giyxDBHJ","sip_cseq":"4178","sip_call_id":"r3xaJ8CLpyTAIHWUZG7gtZQYgAPEGf9S","sip_full_via":"SIP/2.0/UDP 10.0.3.15:5072;rport=5072;branch=z9hG4bKPjPqma7vnLxDkBqcCH3eXLmLYZoPS.6MDc;received=127.0.0.1","sip_full_from":"sip:1001@127.0.0.1;tag=YwuG8U3rRbqIn.xYTnU8NrI3giyxDBHJ","sip_full_to":"sip:1003@127.0.0.1;tag=5Qt4ecvreSHZN","last_sent_callee_id_name":"Outbound Call","last_sent_callee_id_number":"1003","sip_term_status":"200","proto_specific_hangup_cause":"sip:200","sip_term_cause":"16","last_bridge_role":"originator","sip_user_agent":"PJSUA v2.3 Linux-3.2.0.4/x86_64/glibc-2.13","sip_hangup_disposition":"recv_bye","bridge_hangup_cause":"NORMAL_CLEARING","hangup_cause":"NORMAL_CLEARING","hangup_cause_q850":"16","digits_dialed":"none","start_stamp":"2015-07-07 16:52:08","profile_start_stamp":"2015-07-07 16:52:08","answer_stamp":"2015-07-07 16:52:08","bridge_stamp":"2015-07-07 16:52:08","end_stamp":"2015-07-07 16:53:14","start_epoch":"1436280728","start_uepoch":"1436280728471153","profile_start_epoch":"1436280728","profile_start_uepoch":"1436280728930693","answer_epoch":"1436280728","answer_uepoch":"1436280728971147","bridge_epoch":"1436280728","bridge_uepoch":"1436280728971147","last_hold_epoch":"0","last_hold_uepoch":"0","hold_accum_seconds":"0","hold_accum_usec":"0","hold_accum_ms":"0","resurrect_epoch":"0","resurrect_uepoch":"0","progress_epoch":"0","progress_uepoch":"0","progress_media_epoch":"0","progress_media_uepoch":"0","end_epoch":"1436280794","end_uepoch":"1436280794010851","last_app":"sched_hangup","last_arg":"+3120 alloted_timeout","caller_id":"\"1001\" <1001>","duration":"66","billsec":"66","progresssec":"0","answersec":"0","waitsec":"0","progress_mediasec":"0","flow_billsec":"66","mduration":"65539","billmsec":"65039","progressmsec":"28","answermsec":"500","waitmsec":"500","progress_mediamsec":"28","flow_billmsec":"65539","uduration":"65539698","billusec":"65039704","progressusec":"0","answerusec":"499994","waitusec":"499994","progress_mediausec":"0","flow_billusec":"65539698","rtp_audio_in_raw_bytes":"6770","rtp_audio_in_media_bytes":"6762","rtp_audio_in_packet_count":"192","rtp_audio_in_media_packet_count":"190","rtp_audio_in_skip_packet_count":"6","rtp_audio_in_jitter_packet_count":"0","rtp_audio_in_dtmf_packet_count":"0","rtp_audio_in_cng_packet_count":"0","rtp_audio_in_flush_packet_count":"2","rtp_audio_in_largest_jb_size":"0","rtp_audio_in_jitter_min_variance":"26.73","rtp_audio_in_jitter_max_variance":"6716.71","rtp_audio_in_jitter_loss_rate":"0.00","rtp_audio_in_jitter_burst_rate":"0.00","rtp_audio_in_mean_interval":"36.67","rtp_audio_in_flaw_total":"0","rtp_audio_in_quality_percentage":"100.00","rtp_audio_in_mos":"4.50","rtp_audio_out_raw_bytes":"4686","rtp_audio_out_media_bytes":"4686","rtp_audio_out_packet_count":"108","rtp_audio_out_media_packet_count":"108","rtp_audio_out_skip_packet_count":"0","rtp_audio_out_dtmf_packet_count":"0","rtp_audio_out_cng_packet_count":"0","rtp_audio_rtcp_packet_count":"1450","rtp_audio_rtcp_octet_count":"45940"},"app_log":{"applications":[{"app_name":"info","app_data":""},{"app_name":"park","app_data":""},{"app_name":"info","app_data":""},{"app_name":"set","app_data":"ringback=%(2000,4000,440,480)"},{"app_name":"set","app_data":"call_timeout=30"},{"app_name":"bridge","app_data":"user/1003@cgrates.org"},{"app_name":"sched_hangup","app_data":"+3120 alloted_timeout"}]},"callflow":{"dialplan":"XML","profile_index":"2","extension":{"name":"call_debug","number":"1003","applications":[{"app_name":"info","app_data":""},{"app_name":"set","app_data":"ringback=${us-ring}"},{"app_name":"set","app_data":"call_timeout=30"},{"app_name":"bridge","app_data":"user/${destination_number}@${domain_name}"}]},"caller_profile":{"username":"1001","dialplan":"XML","caller_id_name":"1001","ani":"1001","aniii":"","caller_id_number":"1001","network_addr":"127.0.0.1","rdnis":"1003","destination_number":"1003","uuid":"e3133bf7-dcde-4daf-9663-9a79ffcef5ad","source":"mod_sofia","context":"default","chan_name":"sofia/cgrtest/1001@127.0.0.1","originatee":{"originatee_caller_profiles":[{"username":"1001","dialplan":"XML","caller_id_name":"Extension 1001","ani":"1001","aniii":"","caller_id_number":"1001","network_addr":"127.0.0.1","rdnis":"1003","destination_number":"1003","uuid":"0a30dd7c-c222-482f-a322-b1218a15f8cd","source":"mod_sofia","context":"default","chan_name":"sofia/cgrtest/1003@127.0.0.1:5070"},{"username":"1001","dialplan":"XML","caller_id_name":"Extension 1001","ani":"1001","aniii":"","caller_id_number":"1001","network_addr":"127.0.0.1","rdnis":"1003","destination_number":"1003","uuid":"0a30dd7c-c222-482f-a322-b1218a15f8cd","source":"mod_sofia","context":"default","chan_name":"sofia/cgrtest/1003@127.0.0.1:5070"}]}},"times":{"created_time":"1436280728471153","profile_created_time":"1436280728930693","progress_time":"0","progress_media_time":"0","answered_time":"1436280728971147","bridged_time":"1436280728971147","last_hold_time":"0","hold_accum_time":"0","hangup_time":"1436280794010851","resurrect_time":"0","transfer_time":"0"}},"callflow":{"dialplan":"XML","profile_index":"1","extension":{"name":"call_debug","number":"1003","applications":[{"app_name":"info","app_data":""},{"app_name":"park","app_data":""}]},"caller_profile":{"username":"1001","dialplan":"XML","caller_id_name":"1001","ani":"1001","aniii":"","caller_id_number":"1001","network_addr":"127.0.0.1","rdnis":"","destination_number":"1003","uuid":"e3133bf7-dcde-4daf-9663-9a79ffcef5ad","source":"mod_sofia","context":"default","chan_name":"sofia/cgrtest/1001@127.0.0.1"},"times":{"created_time":"1436280728471153","profile_created_time":"1436280728471153","progress_time":"0","progress_media_time":"0","answered_time":"0","bridged_time":"0","last_hold_time":"0","hold_accum_time":"0","hangup_time":"0","resurrect_time":"0","transfer_time":"1436280728930693"}}}`)
var fsCdrCfg *config.CGRConfig

func TestFsCdrInterfaces(t *testing.T) {
	var _ RawCdr = new(FSCdr)
}

func TestFirstNonEmpty(t *testing.T) {
	fsCdrCfg, _ = config.NewDefaultCGRConfig()
	fsCdr, err := NewFSCdr(body, fsCdrCfg)
	if err != nil {
		t.Errorf("Error loading cdr: %v", err)
	}
	//fsc := fsCdr.(FSCdr)
	if _, ok := fsCdr.vars["cgr_reqtype"]; !ok {
		t.Error("Error parsing cdr: ", fsCdr)
	}
}

func TestCDRFields(t *testing.T) {
	fsCdrCfg.CDRSExtraFields = []*utils.RSRField{&utils.RSRField{Id: "sip_user_agent"}}
	fsCdr, err := NewFSCdr(body, fsCdrCfg)
	if err != nil {
		t.Errorf("Error loading cdr: %v", err)
	}
	setupTime, _ := utils.ParseTimeDetectLayout("1436280728", "")
	answerTime, _ := utils.ParseTimeDetectLayout("1436280728", "")
	expctCDR := &CDR{CGRID: "164b0422fdc6a5117031b427439482c6a4f90e41", TOR: utils.VOICE, OriginID: "e3133bf7-dcde-4daf-9663-9a79ffcef5ad",
		OriginHost: "127.0.0.1", Source: "freeswitch_json", Direction: utils.OUT, Category: "call", RequestType: utils.META_PREPAID, Tenant: "cgrates.org", Account: "1001", Subject: "1001",
		Destination: "1003", SetupTime: setupTime, PDD: time.Duration(28) * time.Millisecond, AnswerTime: answerTime, Usage: time.Duration(66) * time.Second, Supplier: "supplier1",
		DisconnectCause: "NORMAL_CLEARING", ExtraFields: map[string]string{"sip_user_agent": "PJSUA v2.3 Linux-3.2.0.4/x86_64/glibc-2.13"}, Cost: -1}
	if CDR := fsCdr.AsStoredCdr(""); !reflect.DeepEqual(expctCDR, CDR) {
		t.Errorf("Expecting: %v, received: %v", expctCDR, CDR)
	}
}

func TestSearchExtraFieldLast(t *testing.T) {
	fsCdr, _ := NewFSCdr(body, fsCdrCfg)
	value := fsCdr.searchExtraField("transfer_time", fsCdr.body)
	if value != "1436280728930693" {
		t.Error("Error finding extra field: ", value)
	}
}

func TestSearchExtraField(t *testing.T) {
	fsCdr, _ := NewFSCdr(body, fsCdrCfg)
	rsrSt1, _ := utils.NewRSRField("^injected_value")
	rsrSt2, _ := utils.NewRSRField("^injected_hdr::injected_value/")
	fsCdrCfg.CDRSExtraFields = []*utils.RSRField{&utils.RSRField{Id: "caller_id_name"}, rsrSt1, rsrSt2}
	extraFields := fsCdr.getExtraFields()
	if len(extraFields) != 3 || extraFields["caller_id_name"] != "1001" ||
		extraFields["injected_value"] != "injected_value" ||
		extraFields["injected_hdr"] != "injected_value" {
		t.Error("Error parsing extra fields: ", extraFields)
	}
}

func TestSearchExtraFieldInSlice(t *testing.T) {
	fsCdr, _ := NewFSCdr(body, fsCdrCfg)
	value := fsCdr.searchExtraField("app_data", fsCdr.body)
	if value != "ringback=%(2000,4000,440,480)" {
		t.Error("Error finding extra field: ", value)
	}
}

func TestSearchReplaceInExtraFields(t *testing.T) {
	fsCdrCfg.CDRSExtraFields = []*utils.RSRField{&utils.RSRField{Id: "read_codec"},
		&utils.RSRField{Id: "sip_user_agent", RSRules: []*utils.ReSearchReplace{&utils.ReSearchReplace{SearchRegexp: regexp.MustCompile(`([A-Za-z]*).+`), ReplaceTemplate: "$1"}}},
		&utils.RSRField{Id: "write_codec"}}
	fsCdr, _ := NewFSCdr(body, fsCdrCfg)
	extraFields := fsCdr.getExtraFields()
	if len(extraFields) != 3 {
		t.Error("Error parsing extra fields: ", extraFields)
	}
	if extraFields["sip_user_agent"] != "PJSUA" {
		t.Error("Error parsing extra fields: ", extraFields)
	}
}

func TestDDazRSRExtraFields(t *testing.T) {
	eFieldsCfg := `{"cdrs": {
	"extra_fields": ["~effective_caller_id_number:s/(\\d+)/+$1/"],
},}`
	simpleJsonCdr := []byte(`{
    "core-uuid": "feef0b51-7fdf-4c4a-878e-aff233752de2",
    "channel_data": {
        "state": "CS_REPORTING",
        "direction": "inbound",
        "state_number": "11",
        "flags": "0=1;1=1;3=1;36=1;37=1;39=1;42=1;47=1;52=1;73=1;75=1;94=1",
        "caps": "1=1;2=1;3=1;4=1;5=1;6=1"
    },
    "variables": {
        "direction": "inbound",
        "uuid": "86cfd6e2-dbda-45a3-b59d-f683ec368e8b",
        "session_id": "5",
        "accountcode": "1001",
        "user_context": "default",
        "effective_caller_id_name": "Extension 1001",
        "effective_caller_id_number": "4986517174963",
        "outbound_caller_id_name": "FreeSWITCH",
        "outbound_caller_id_number": "0000000000"
    },
    "times": {
        "created_time": "1396984221278217",
        "profile_created_time": "1396984221278217",
        "progress_time": "0",
        "progress_media_time": "0",
        "answered_time": "0",
        "hangup_time": "0",
        "resurrect_time": "0",
        "transfer_time": "1396984221377035"
    }
}`)
	var err error
	fsCdrCfg, err = config.NewCGRConfigFromJsonString(eFieldsCfg)
	if err != nil {
		t.Error("Could not parse the config", err.Error())
	} else if !reflect.DeepEqual(fsCdrCfg.CDRSExtraFields, []*utils.RSRField{&utils.RSRField{Id: "effective_caller_id_number",
		RSRules: []*utils.ReSearchReplace{&utils.ReSearchReplace{SearchRegexp: regexp.MustCompile(`(\d+)`), ReplaceTemplate: "+$1"}}}}) {
		t.Errorf("Unexpected value for config CdrsExtraFields: %v", fsCdrCfg.CDRSExtraFields)
	}
	fsCdr, err := NewFSCdr(simpleJsonCdr, fsCdrCfg)
	if err != nil {
		t.Error("Could not parse cdr", err.Error())
	}
	extraFields := fsCdr.getExtraFields()
	if extraFields["effective_caller_id_number"] != "+4986517174963" {
		t.Error("Unexpected effective_caller_id_number received", extraFields["effective_caller_id_number"])
	}
}
