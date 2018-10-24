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
package engine

import (
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

var body = []byte(`{
	"core-uuid": "eb8bcdd2-d9eb-4f8f-80c3-1a4042aabe31",
	"switchname": "FSDev1",
	"channel_data": {
		"state": "CS_REPORTING",
		"direction": "inbound",
		"state_number": "11",
		"flags": "0=1;1=1;3=1;20=1;37=1;38=1;40=1;43=1;48=1;53=1;75=1;98=1;112=1;113=1;122=1;134=1",
		"caps": "1=1;2=1;3=1;4=1;5=1;6=1"
	},
	"callStats": {
		"audio": {
			"inbound": {
				"raw_bytes": 572588,
				"media_bytes": 572588,
				"packet_count": 3329,
				"media_packet_count": 3329,
				"skip_packet_count": 10,
				"jitter_packet_count": 0,
				"dtmf_packet_count": 0,
				"cng_packet_count": 0,
				"flush_packet_count": 0,
				"largest_jb_size": 0,
				"jitter_min_variance": 0,
				"jitter_max_variance": 0,
				"jitter_loss_rate": 0,
				"jitter_burst_rate": 0,
				"mean_interval": 0,
				"flaw_total": 0,
				"quality_percentage": 100,
				"mos": 4.500000
			},
			"outbound": {
				"raw_bytes": 0,
				"media_bytes": 0,
				"packet_count": 0,
				"media_packet_count": 0,
				"skip_packet_count": 0,
				"dtmf_packet_count": 0,
				"cng_packet_count": 0,
				"rtcp_packet_count": 0,
				"rtcp_octet_count": 0
			}
		}
	},
	"variables": {
		"direction": "inbound",
		"uuid": "3da8bf84-c133-4959-9e24-e72875cb33a1",
		"session_id": "7",
		"sip_from_user": "1001",
		"sip_from_uri": "1001@10.10.10.204",
		"sip_from_host": "10.10.10.204",
		"channel_name": "sofia/internal/1001@10.10.10.204",
		"ep_codec_string": "CORE_PCM_MODULE.PCMU@8000h@20i@64000b,CORE_PCM_MODULE.PCMA@8000h@20i@64000b",
		"sip_local_network_addr": "10.10.10.204",
		"sip_network_ip": "10.10.10.100",
		"sip_network_port": "5060",
		"sip_invite_stamp": "1515666344534355",
		"sip_received_ip": "10.10.10.100",
		"sip_received_port": "5060",
		"sip_via_protocol": "udp",
		"sip_from_user_stripped": "1001",
		"sofia_profile_name": "internal",
		"recovery_profile_name": "internal",
		"sip_req_user": "1002",
		"sip_req_uri": "1002@10.10.10.204",
		"sip_req_host": "10.10.10.204",
		"sip_to_user": "1002",
		"sip_to_uri": "1002@10.10.10.204",
		"sip_to_host": "10.10.10.204",
		"sip_contact_params": "transport=udp;registering_acc=10_10_10_204",
		"sip_contact_user": "1001",
		"sip_contact_port": "5060",
		"sip_contact_uri": "1001@10.10.10.100:5060",
		"sip_contact_host": "10.10.10.100",
		"sip_user_agent": "Jitsi2.10.5550Linux",
		"sip_via_host": "10.10.10.100",
		"sip_via_port": "5060",
		"presence_id": "1001@10.10.10.204",
		"cgr_notify": "AUTH_OK",
		"max_forwards": "69",
		"transfer_history": "1515666344:b4300942-e809-4393-99cb-d39a1bc3c219:bl_xfer:1002/default/XML",
		"transfer_source": "1515666344:b4300942-e809-4393-99cb-d39a1bc3c219:bl_xfer:1002/default/XML",
		"DP_MATCH": "ARRAY::1002|:1002",
		"call_uuid": "3da8bf84-c133-4959-9e24-e72875cb33a1",
		"call_timeout": "30",
		"current_application_data": "user/1002@10.10.10.204",
		"current_application": "bridge",
		"dialed_user": "1002",
		"dialed_domain": "10.10.10.204",
		"originated_legs": "ARRAY::f52c26f1-b018-4963-bf6d-a3111d1a0320;Outbound Call;1002|:f52c26f1-b018-4963-bf6d-a3111d1a0320;Outbound Call;1002",
		"switch_m_sdp": "v=0\r\no=1002-jitsi.org 0 0 IN IP4 10.10.10.100\r\ns=-\r\nc=IN IP4 10.10.10.100\r\nt=0 0\r\nm=audio 5022 RTP/AVP 0 8 101\r\na=rtpmap:0 PCMU/8000\r\na=rtpmap:8 PCMA/8000\r\na=rtpmap:101 telephone-event/8000\r\n",
		"rtp_use_codec_name": "PCMU",
		"rtp_use_codec_rate": "8000",
		"rtp_use_codec_ptime": "20",
		"rtp_use_codec_channels": "1",
		"rtp_last_audio_codec_string": "PCMU@8000h@20i@1c",
		"read_codec": "PCMU",
		"original_read_codec": "PCMU",
		"read_rate": "8000",
		"original_read_rate": "8000",
		"write_codec": "PCMU",
		"write_rate": "8000",
		"video_possible": "true",
		"video_media_flow": "sendonly",
		"local_media_ip": "10.10.10.204",
		"local_media_port": "21566",
		"advertised_media_ip": "10.10.10.204",
		"rtp_use_timer_name": "soft",
		"rtp_use_pt": "0",
		"rtp_use_ssrc": "1448966920",
		"endpoint_disposition": "ANSWER",
		"originate_causes": "ARRAY::f52c26f1-b018-4963-bf6d-a3111d1a0320;NONE|:f52c26f1-b018-4963-bf6d-a3111d1a0320;NONE",
		"originate_disposition": "SUCCESS",
		"DIALSTATUS": "SUCCESS",
		"last_bridge_to": "f52c26f1-b018-4963-bf6d-a3111d1a0320",
		"bridge_channel": "sofia/internal/1002@10.10.10.100:5060",
		"bridge_uuid": "f52c26f1-b018-4963-bf6d-a3111d1a0320",
		"signal_bond": "f52c26f1-b018-4963-bf6d-a3111d1a0320",
		"last_sent_callee_id_name": "Outbound Call",
		"last_sent_callee_id_number": "1002",
		"switch_r_sdp": "v=0\r\no=1001-jitsi.org 0 1 IN IP4 10.10.10.100\r\ns=-\r\nc=IN IP4 10.10.10.100\r\nt=0 0\r\nm=audio 5018 RTP/AVP 96 97 98 9 100 102 0 8 103 3 104 101\r\na=rtpmap:96 opus/48000/2\r\na=fmtp:96 usedtx=1\r\na=rtpmap:97 SILK/24000\r\na=rtpmap:98 SILK/16000\r\na=rtpmap:9 G722/8000\r\na=rtpmap:100 speex/32000\r\na=rtpmap:102 speex/16000\r\na=rtpmap:0 PCMU/8000\r\na=rtpmap:8 PCMA/8000\r\na=rtpmap:103 iLBC/8000\r\na=rtpmap:3 GSM/8000\r\na=rtpmap:104 speex/8000\r\na=rtpmap:101 telephone-event/8000\r\na=sendonly\r\na=ptime:20\r\na=extmap:1 urn:ietf:params:rtp-hdrext:csrc-audio-level\r\na=extmap:2 urn:ietf:params:rtp-hdrext:ssrc-audio-level\r\na=rtcp-xr:voip-metrics\r\nm=video 0 RTP/AVP 105 99\r\n",
		"rtp_use_codec_string": "PCMU,PCMA",
		"audio_media_flow": "recvonly",
		"remote_media_ip": "10.10.10.100",
		"remote_media_port": "5018",
		"rtp_audio_recv_pt": "0",
		"dtmf_type": "rfc2833",
		"rtp_2833_send_payload": "101",
		"rtp_2833_recv_payload": "101",
		"rtp_local_sdp_str": "v=0\r\no=FreeSWITCH 1515644781 1515644783 IN IP4 10.10.10.204\r\ns=FreeSWITCH\r\nc=IN IP4 10.10.10.204\r\nt=0 0\r\nm=audio 21566 RTP/AVP 0 101\r\na=rtpmap:0 PCMU/8000\r\na=rtpmap:101 telephone-event/8000\r\na=fmtp:101 0-16\r\na=ptime:20\r\na=recvonly\r\n",
		"sip_to_tag": "m3g4NZ4rXFX3p",
		"sip_from_tag": "f25afe20",
		"sip_cseq": "2",
		"sip_call_id": "818e26f805701988c1a330175d7d2629@0:0:0:0:0:0:0:0",
		"sip_full_via": "SIP/2.0/UDP 10.10.10.100:5060;branch=z9hG4bK-313838-2ee350643dea826b4a74f8049852f307",
		"sip_from_display": "1001",
		"sip_full_from": "\"1001\" <sip:1001@10.10.10.204>;tag=f25afe20",
		"sip_full_to": "<sip:1002@10.10.10.204>;tag=m3g4NZ4rXFX3p",
		"sip_hangup_phrase": "OK",
		"last_bridge_hangup_cause": "NORMAL_CLEARING",
		"last_bridge_proto_specific_hangup_cause": "sip:200",
		"bridge_hangup_cause": "NORMAL_CLEARING",
		"hangup_cause": "NORMAL_CLEARING",
		"hangup_cause_q850": "16",
		"digits_dialed": "none",
		"start_stamp": "2018-01-11 11:25:44",
		"profile_start_stamp": "2018-01-11 11:25:44",
		"answer_stamp": "2018-01-11 11:25:47",
		"bridge_stamp": "2018-01-11 11:25:47",
		"hold_stamp": "2018-01-11 11:25:48",
		"progress_stamp": "2018-01-11 11:25:44",
		"progress_media_stamp": "2018-01-11 11:25:47",
		"hold_events": "{{1515666348363496,1515666415502648}}",
		"end_stamp": "2018-01-11 11:26:55",
		"start_epoch": "1515666344",
		"start_uepoch": "1515666344534355",
		"profile_start_epoch": "1515666344",
		"profile_start_uepoch": "1515666344534355",
		"answer_epoch": "1515666347",
		"answer_uepoch": "1515666347954373",
		"bridge_epoch": "1515666347",
		"bridge_uepoch": "1515666347954373",
		"last_hold_epoch": "1515666348",
		"last_hold_uepoch": "1515666348363496",
		"hold_accum_seconds": "67",
		"hold_accum_usec": "67139151",
		"hold_accum_ms": "67139",
		"resurrect_epoch": "0",
		"resurrect_uepoch": "0",
		"progress_epoch": "1515666344",
		"progress_uepoch": "1515666344594267",
		"progress_media_epoch": "1515666347",
		"progress_media_uepoch": "1515666347954373",
		"end_epoch": "1515666415",
		"end_uepoch": "1515666415494269",
		"last_app": "bridge",
		"last_arg": "user/1002@10.10.10.204",
		"caller_id": "\"1001\" <1001>",
		"duration": "71",
		"billsec": "68",
		"progresssec": "0",
		"answersec": "3",
		"waitsec": "3",
		"progress_mediasec": "3",
		"flow_billsec": "71",
		"mduration": "70960",
		"billmsec": "67540",
		"progressmsec": "60",
		"answermsec": "3420",
		"waitmsec": "3420",
		"progress_mediamsec": "3420",
		"flow_billmsec": "70960",
		"uduration": "70959914",
		"billusec": "67539896",
		"progressusec": "59912",
		"answerusec": "3420018",
		"waitusec": "3420018",
		"progress_mediausec": "3420018",
		"flow_billusec": "70959914",
		"sip_hangup_disposition": "send_bye",
		"rtp_audio_in_raw_bytes": "572588",
		"rtp_audio_in_media_bytes": "572588",
		"rtp_audio_in_packet_count": "3329",
		"rtp_audio_in_media_packet_count": "3329",
		"rtp_audio_in_skip_packet_count": "10",
		"rtp_audio_in_jitter_packet_count": "0",
		"rtp_audio_in_dtmf_packet_count": "0",
		"rtp_audio_in_cng_packet_count": "0",
		"rtp_audio_in_flush_packet_count": "0",
		"rtp_audio_in_largest_jb_size": "0",
		"rtp_audio_in_jitter_min_variance": "0.00",
		"rtp_audio_in_jitter_max_variance": "0.00",
		"rtp_audio_in_jitter_loss_rate": "0.00",
		"rtp_audio_in_jitter_burst_rate": "0.00",
		"rtp_audio_in_mean_interval": "0.00",
		"rtp_audio_in_flaw_total": "0",
		"rtp_audio_in_quality_percentage": "100.00",
		"rtp_audio_in_mos": "4.50",
		"rtp_audio_out_raw_bytes": "0",
		"rtp_audio_out_media_bytes": "0",
		"rtp_audio_out_packet_count": "0",
		"rtp_audio_out_media_packet_count": "0",
		"rtp_audio_out_skip_packet_count": "0",
		"rtp_audio_out_dtmf_packet_count": "0",
		"rtp_audio_out_cng_packet_count": "0",
		"rtp_audio_rtcp_packet_count": "0",
		"rtp_audio_rtcp_octet_count": "0"
	},
	"app_log": {
		"applications": [{
			"app_name": "park",
			"app_data": "",
			"app_stamp": "1515666344548466"
		}, {
			"app_name": "set",
			"app_data": "ringback=",
			"app_stamp": "1515666344575066"
		}, {
			"app_name": "set",
			"app_data": "call_timeout=30",
			"app_stamp": "1515666344576009"
		}, {
			"app_name": "bridge",
			"app_data": "user/1002@10.10.10.204",
			"app_stamp": "1515666344576703"
		}]
	},
	"callflow": [{
		"dialplan": "XML",
		"profile_index": "2",
		"extension": {
			"name": "Local_Extension",
			"number": "1002",
			"applications": [{
				"app_name": "set",
				"app_data": "ringback=${us-ring}"
			}, {
				"app_name": "set",
				"app_data": "call_timeout=30"
			}, {
				"app_name": "bridge",
				"app_data": "user/${destination_number}@${domain_name}"
			}]
		},
		"caller_profile": {
			"username": "1001",
			"dialplan": "XML",
			"caller_id_name": "1001",
			"ani": "1001",
			"aniii": "",
			"caller_id_number": "1001",
			"network_addr": "10.10.10.100",
			"rdnis": "1002",
			"destination_number": "1002",
			"uuid": "3da8bf84-c133-4959-9e24-e72875cb33a1",
			"source": "mod_sofia",
			"context": "default",
			"chan_name": "sofia/internal/1001@10.10.10.204",
			"originatee": {
				"originatee_caller_profiles": [{
					"username": "1001",
					"dialplan": "XML",
					"caller_id_name": "1001",
					"ani": "1001",
					"aniii": "",
					"caller_id_number": "1001",
					"network_addr": "10.10.10.100",
					"rdnis": "1002",
					"destination_number": "1002",
					"uuid": "f52c26f1-b018-4963-bf6d-a3111d1a0320",
					"source": "mod_sofia",
					"context": "default",
					"chan_name": "sofia/internal/1002@10.10.10.100:5060"
				}, {
					"username": "1001",
					"dialplan": "XML",
					"caller_id_name": "1001",
					"ani": "1001",
					"aniii": "",
					"caller_id_number": "1001",
					"network_addr": "10.10.10.100",
					"rdnis": "1002",
					"destination_number": "1002",
					"uuid": "f52c26f1-b018-4963-bf6d-a3111d1a0320",
					"source": "mod_sofia",
					"context": "default",
					"chan_name": "sofia/internal/1002@10.10.10.100:5060"
				}]
			}
		},
		"times": {
			"created_time": "1515666344534355",
			"profile_created_time": "1515666344534355",
			"progress_time": "1515666344594267",
			"progress_media_time": "1515666347954373",
			"answered_time": "1515666347954373",
			"bridged_time": "1515666347954373",
			"last_hold_time": "1515666348363496",
			"hold_accum_time": "67139151",
			"hangup_time": "1515666415494269",
			"resurrect_time": "0",
			"transfer_time": "0"
		}
	}, {
		"dialplan": "XML",
		"profile_index": "1",
		"extension": {
			"name": "CGRateS_Auth",
			"number": "1002",
			"applications": [{
				"app_name": "park",
				"app_data": ""
			}]
		},
		"caller_profile": {
			"username": "1001",
			"dialplan": "XML",
			"caller_id_name": "1001",
			"ani": "1001",
			"aniii": "",
			"caller_id_number": "1001",
			"network_addr": "10.10.10.100",
			"rdnis": "",
			"destination_number": "1002",
			"uuid": "3da8bf84-c133-4959-9e24-e72875cb33a1",
			"source": "mod_sofia",
			"context": "default",
			"chan_name": "sofia/internal/1001@10.10.10.204"
		},
		"times": {
			"created_time": "1515666344534355",
			"profile_created_time": "1515666344534355",
			"progress_time": "0",
			"progress_media_time": "0",
			"answered_time": "0",
			"bridged_time": "0",
			"last_hold_time": "0",
			"hold_accum_time": "0",
			"hangup_time": "0",
			"resurrect_time": "0",
			"transfer_time": "1515666344534355"
		}
	}]
}`)

var fsCdrCfg *config.CGRConfig

func TestFsCdrInterfaces(t *testing.T) {
	var _ RawCdr = new(FSCdr)
}

func TestFsCdrFirstNonEmpty(t *testing.T) {
	fsCdrCfg, _ = config.NewDefaultCGRConfig()
	fsCdr, err := NewFSCdr(body, fsCdrCfg)
	if err != nil {
		t.Errorf("Error loading cdr: %v", err)
	}
	//fsc := fsCdr.(FSCdr)
	if _, ok := fsCdr.vars["cgr_notify"]; !ok {
		t.Error("Error parsing cdr: ", fsCdr)
	}
}

func TestFsCdrCDRFields(t *testing.T) {
	fsCdrCfg.CdrsCfg().CDRSExtraFields = []*utils.RSRField{{Id: "sip_user_agent"}}
	fsCdr, err := NewFSCdr(body, fsCdrCfg)
	if err != nil {
		t.Errorf("Error loading cdr: %v", err)
	}
	setupTime, _ := utils.ParseTimeDetectLayout("1515666344", "")
	answerTime, _ := utils.ParseTimeDetectLayout("1515666347", "")
	expctCDR := &CDR{
		CGRID: "24b5766be325fa751fab5a0a06373e106f33a257",
		ToR:   utils.VOICE, OriginID: "3da8bf84-c133-4959-9e24-e72875cb33a1",
		OriginHost: "", Source: "freeswitch_json", Category: "call",
		RequestType: utils.META_RATED, Tenant: "cgrates.org",
		Account: "1001", Subject: "1001",
		Destination: "1002", SetupTime: setupTime,
		AnswerTime: answerTime, Usage: time.Duration(68) * time.Second,
		Cost:        -1,
		ExtraFields: map[string]string{"sip_user_agent": "Jitsi2.10.5550Linux"}}
	if CDR := fsCdr.AsCDR(""); !reflect.DeepEqual(expctCDR, CDR) {
		t.Errorf("Expecting: %+v, received: %+v", expctCDR, CDR)
	}
}

func TestFsCdrSearchExtraFieldLast(t *testing.T) {
	fsCdr, _ := NewFSCdr(body, fsCdrCfg)
	value := fsCdr.searchExtraField("progress_media_time", fsCdr.body)
	if value != "1515666347954373" {
		t.Error("Error finding extra field: ", value)
	}
}

func TestFsCdrSearchExtraField(t *testing.T) {
	fsCdr, _ := NewFSCdr(body, fsCdrCfg)
	rsrSt1, _ := utils.NewRSRField("^injected_value")
	rsrSt2, _ := utils.NewRSRField("^injected_hdr::injected_value/")
	fsCdrCfg.CdrsCfg().CDRSExtraFields = []*utils.RSRField{{Id: "caller_id_name"}, rsrSt1, rsrSt2}
	extraFields := fsCdr.getExtraFields()
	if len(extraFields) != 3 || extraFields["caller_id_name"] != "1001" ||
		extraFields["injected_value"] != "injected_value" ||
		extraFields["injected_hdr"] != "injected_value" {
		t.Error("Error parsing extra fields: ", extraFields)
	}

}

func TestFsCdrSearchExtraFieldInSlice(t *testing.T) {
	fsCdr, _ := NewFSCdr(body, fsCdrCfg)
	if value := fsCdr.searchExtraField("floatfld1", map[string]interface{}{"floatfld1": 6.4}); value != "6.4" {
		t.Errorf("Expecting: 6.4, received: %s", value)
	}
}

func TestFsCdrSearchReplaceInExtraFields(t *testing.T) {
	fsCdrCfg.CdrsCfg().CDRSExtraFields = utils.ParseRSRFieldsMustCompile(`read_codec;~sip_user_agent:s/([A-Za-z]*).+/$1/;write_codec`, utils.INFIELD_SEP)
	fsCdr, _ := NewFSCdr(body, fsCdrCfg)
	extraFields := fsCdr.getExtraFields()
	if len(extraFields) != 3 {
		t.Error("Error parsing extra fields: ", extraFields)
	}
	if extraFields["sip_user_agent"] != "Jitsi" {
		t.Error("Error parsing extra fields: ", extraFields)
	}
}

func TestFsCdrDDazRSRExtraFields(t *testing.T) {
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
	fsCdrCfg, err = config.NewCGRConfigFromJsonStringWithDefaults(eFieldsCfg)
	expCdrExtra := utils.ParseRSRFieldsMustCompile(`~effective_caller_id_number:s/(\d+)/+$1/`, utils.INFIELD_SEP)
	if err != nil {
		t.Error("Could not parse the config", err.Error())
	} else if !reflect.DeepEqual(expCdrExtra[0], fsCdrCfg.CdrsCfg().CDRSExtraFields[0]) { // Kinda deepEqual bug since without index does not match
		t.Errorf("Expecting: %+v, received: %+v", expCdrExtra, fsCdrCfg.CdrsCfg().CDRSExtraFields)
	}
	fsCdr, err := NewFSCdr(simpleJsonCdr, fsCdrCfg)
	if err != nil {
		t.Error("Could not parse cdr", err.Error())
	}
	extraFields := fsCdr.getExtraFields()
	if extraFields["effective_caller_id_number"] != "+4986517174963" {
		t.Errorf("Unexpected effective_caller_id_number received: %+v", extraFields["effective_caller_id_number"])
	}
}

func TestFsCdrFirstDefined(t *testing.T) {
	fsCdr, _ := NewFSCdr(body, fsCdrCfg)
	value := fsCdr.firstDefined([]string{utils.CGR_SUBJECT, utils.CGR_ACCOUNT, FS_USERNAME}, FsUsername)
	if value != "1001" {
		t.Errorf("Expecting: 1001, received: %s", value)
	}
	value = fsCdr.firstDefined([]string{utils.CGR_ACCOUNT, FS_USERNAME}, FsUsername)
	if value != "1001" {
		t.Errorf("Expecting: 1001, received: %s", value)
	}
}
