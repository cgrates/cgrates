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
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
)

var hangupEv string = `Event-Name: CHANNEL_HANGUP_COMPLETE
Core-UUID: 651a8db2-4f67-4cf8-b622-169e8a482e50
FreeSWITCH-Hostname: CgrDev1
FreeSWITCH-Switchname: CgrDev1
FreeSWITCH-IPv4: 10.0.3.15
FreeSWITCH-IPv6: %3A%3A1
Event-Date-Local: 2015-07-07%2016%3A53%3A14
Event-Date-GMT: Tue,%2007%20Jul%202015%2014%3A53%3A14%20GMT
Event-Date-Timestamp: 1436280794030635
Event-Calling-File: switch_core_state_machine.c
Event-Calling-Function: switch_core_session_reporting_state
Event-Calling-Line-Number: 834
Event-Sequence: 1035
Hangup-Cause: NORMAL_CLEARING
Channel-State: CS_REPORTING
Channel-Call-State: HANGUP
Channel-State-Number: 11
Channel-Name: sofia/cgrtest/1001%40127.0.0.1
Unique-ID: e3133bf7-dcde-4daf-9663-9a79ffcef5ad
Call-Direction: inbound
Presence-Call-Direction: inbound
Channel-HIT-Dialplan: true
Channel-Call-UUID: e3133bf7-dcde-4daf-9663-9a79ffcef5ad
Answer-State: hangup
Hangup-Cause: NORMAL_CLEARING
Channel-Read-Codec-Name: SPEEX
Channel-Read-Codec-Rate: 32000
Channel-Read-Codec-Bit-Rate: 44000
Channel-Write-Codec-Name: SPEEX
Channel-Write-Codec-Rate: 32000
Channel-Write-Codec-Bit-Rate: 44000
Caller-Direction: inbound
Caller-Logical-Direction: inbound
Caller-Username: 1001
Caller-Dialplan: XML
Caller-Caller-ID-Name: 1001
Caller-Caller-ID-Number: 1001
Caller-Orig-Caller-ID-Name: 1001
Caller-Orig-Caller-ID-Number: 1001
Caller-Callee-ID-Name: Outbound%20Call
Caller-Callee-ID-Number: 1003
Caller-Network-Addr: 127.0.0.1
Caller-ANI: 1001
Caller-Destination-Number: 1003
Caller-Unique-ID: e3133bf7-dcde-4daf-9663-9a79ffcef5ad
Caller-Source: mod_sofia
Caller-Transfer-Source: 1436280728%3Ae7c250e8-6ad7-4bd4-8962-318e0b0da728%3Abl_xfer%3A1003/default/XML
Caller-Context: default
Caller-RDNIS: 1003
Caller-Channel-Name: sofia/cgrtest/1001%40127.0.0.1
Caller-Profile-Index: 2
Caller-Profile-Created-Time: 1436280728930693
Caller-Channel-Created-Time: 1436280728471153
Caller-Channel-Answered-Time: 1436280728971147
Caller-Channel-Progress-Time: 0
Caller-Channel-Progress-Media-Time: 0
Caller-Channel-Hangup-Time: 1436280794010851
Caller-Channel-Transfer-Time: 0
Caller-Channel-Resurrect-Time: 0
Caller-Channel-Bridged-Time: 1436280728971147
Caller-Channel-Last-Hold: 0
Caller-Channel-Hold-Accum: 0
Caller-Screen-Bit: true
Caller-Privacy-Hide-Name: false
Caller-Privacy-Hide-Number: false
Other-Type: originatee
Other-Leg-Direction: outbound
Other-Leg-Logical-Direction: inbound
Other-Leg-Username: 1001
Other-Leg-Dialplan: XML
Other-Leg-Caller-ID-Name: Extension%201001
Other-Leg-Caller-ID-Number: 1001
Other-Leg-Orig-Caller-ID-Name: 1001
Other-Leg-Orig-Caller-ID-Number: 1001
Other-Leg-Callee-ID-Name: Outbound%20Call
Other-Leg-Callee-ID-Number: 1003
Other-Leg-Network-Addr: 127.0.0.1
Other-Leg-ANI: 1001
Other-Leg-Destination-Number: 1003
Other-Leg-Unique-ID: 0a30dd7c-c222-482f-a322-b1218a15f8cd
Other-Leg-Source: mod_sofia
Other-Leg-Context: default
Other-Leg-RDNIS: 1003
Other-Leg-Channel-Name: sofia/cgrtest/1003%40127.0.0.1%3A5070
Other-Leg-Profile-Created-Time: 1436280728950627
Other-Leg-Channel-Created-Time: 1436280728950627
Other-Leg-Channel-Answered-Time: 1436280728950627
Other-Leg-Channel-Progress-Time: 0
Other-Leg-Channel-Progress-Media-Time: 0
Other-Leg-Channel-Hangup-Time: 0
Other-Leg-Channel-Transfer-Time: 0
Other-Leg-Channel-Resurrect-Time: 0
Other-Leg-Channel-Bridged-Time: 0
Other-Leg-Channel-Last-Hold: 0
Other-Leg-Channel-Hold-Accum: 0
Other-Leg-Screen-Bit: true
Other-Leg-Privacy-Hide-Name: false
Other-Leg-Privacy-Hide-Number: false
variable_direction: inbound
variable_uuid: e3133bf7-dcde-4daf-9663-9a79ffcef5ad
variable_session_id: 4
variable_sip_from_user: 1001
variable_sip_from_uri: 1001%40127.0.0.1
variable_sip_from_host: 127.0.0.1
variable_channel_name: sofia/cgrtest/1001%40127.0.0.1
variable_ep_codec_string: speex%4016000h%4020i,speex%408000h%4020i,speex%4032000h%4020i,GSM%408000h%4020i%4013200b,PCMU%408000h%4020i%4064000b,PCMA%408000h%4020i%4064000b,G722%408000h%4020i%4064000b
variable_sip_local_network_addr: 127.0.0.1
variable_sip_network_ip: 127.0.0.1
variable_sip_network_port: 46615
variable_sip_received_ip: 127.0.0.1
variable_sip_received_port: 46615
variable_sip_via_protocol: tcp
variable_sip_authorized: true
variable_Event-Name: REQUEST_PARAMS
variable_Core-UUID: 651a8db2-4f67-4cf8-b622-169e8a482e50
variable_FreeSWITCH-Hostname: CgrDev1
variable_FreeSWITCH-Switchname: CgrDev1
variable_FreeSWITCH-IPv4: 10.0.3.15
variable_FreeSWITCH-IPv6: %3A%3A1
variable_Event-Date-Local: 2015-07-07%2016%3A52%3A08
variable_Event-Date-GMT: Tue,%2007%20Jul%202015%2014%3A52%3A08%20GMT
variable_Event-Date-Timestamp: 1436280728471153
variable_Event-Calling-File: sofia.c
variable_Event-Calling-Function: sofia_handle_sip_i_invite
variable_Event-Calling-Line-Number: 9056
variable_Event-Sequence: 515
variable_sip_number_alias: 1001
variable_sip_auth_username: 1001
variable_sip_auth_realm: 127.0.0.1
variable_number_alias: 1001
variable_requested_domain_name: cgrates.org
variable_record_stereo: true
variable_transfer_fallback_extension: operator
variable_toll_allow: domestic,international,local
variable_accountcode: 1001
variable_user_context: default
variable_effective_caller_id_name: Extension%201001
variable_effective_caller_id_number: 1001
variable_outbound_caller_id_name: FreeSWITCH
variable_outbound_caller_id_number: 0000000000
variable_callgroup: techsupport
variable_cgr_reqtype: *prepaid
variable_cgr_supplier: supplier1
variable_user_name: 1001
variable_domain_name: cgrates.org
variable_sip_from_user_stripped: 1001
variable_sofia_profile_name: cgrtest
variable_recovery_profile_name: cgrtest
variable_sip_full_route: %3Csip%3A127.0.0.1%3A25060%3Blr%3E
variable_sip_recover_via: SIP/2.0/TCP%20127.0.0.1%3A46615%3Brport%3D46615%3Bbranch%3Dz9hG4bKPjGj7AlihmVwAVz9McwVeI64NeBHlPmXAN%3Balias
variable_sip_req_user: 1003
variable_sip_req_uri: 1003%40127.0.0.1
variable_sip_req_host: 127.0.0.1
variable_sip_to_user: 1003
variable_sip_to_uri: 1003%40127.0.0.1
variable_sip_to_host: 127.0.0.1
variable_sip_contact_params: ob
variable_sip_contact_user: 1001
variable_sip_contact_port: 5072
variable_sip_contact_uri: 1001%40127.0.0.1%3A5072
variable_sip_contact_host: 127.0.0.1
variable_sip_via_host: 127.0.0.1
variable_sip_via_port: 46615
variable_sip_via_rport: 46615
variable_switch_r_sdp: v%3D0%0D%0Ao%3D-%203645269528%203645269528%20IN%20IP4%2010.0.3.15%0D%0As%3Dpjmedia%0D%0Ab%3DAS%3A84%0D%0At%3D0%200%0D%0Aa%3DX-nat%3A0%0D%0Am%3Daudio%204006%20RTP/AVP%2098%2097%2099%20104%203%200%208%209%2096%0D%0Ac%3DIN%20IP4%2010.0.3.15%0D%0Ab%3DAS%3A64000%0D%0Aa%3Drtpmap%3A98%20speex/16000%0D%0Aa%3Drtpmap%3A97%20speex/8000%0D%0Aa%3Drtpmap%3A99%20speex/32000%0D%0Aa%3Drtpmap%3A104%20iLBC/8000%0D%0Aa%3Dfmtp%3A104%20mode%3D30%0D%0Aa%3Drtpmap%3A3%20GSM/8000%0D%0Aa%3Drtpmap%3A0%20PCMU/8000%0D%0Aa%3Drtpmap%3A8%20PCMA/8000%0D%0Aa%3Drtpmap%3A9%20G722/8000%0D%0Aa%3Drtpmap%3A96%20telephone-event/8000%0D%0Aa%3Dfmtp%3A96%200-16%0D%0Aa%3Drtcp%3A4007%20IN%20IP4%2010.0.3.15%0D%0A
variable_rtp_remote_audio_rtcp_port: 4007%20IN%20IP4%2010.0.3.15
variable_rtp_audio_recv_pt: 99
variable_rtp_use_codec_name: SPEEX
variable_rtp_use_codec_rate: 32000
variable_rtp_use_codec_ptime: 20
variable_rtp_use_codec_channels: 1
variable_rtp_last_audio_codec_string: SPEEX%4032000h%4020i%401c
variable_read_codec: SPEEX
variable_original_read_codec: SPEEX
variable_read_rate: 32000
variable_original_read_rate: 32000
variable_write_codec: SPEEX
variable_write_rate: 32000
variable_dtmf_type: rfc2833
variable_execute_on_answer: sched_hangup%20%2B3120%20alloted_timeout
variable_cgr_notify: %2BAUTH_OK
variable_max_forwards: 69
variable_transfer_history: 1436280728%3Ae7c250e8-6ad7-4bd4-8962-318e0b0da728%3Abl_xfer%3A1003/default/XML
variable_transfer_source: 1436280728%3Ae7c250e8-6ad7-4bd4-8962-318e0b0da728%3Abl_xfer%3A1003/default/XML
variable_DP_MATCH: ARRAY%3A%3A1003%7C%3A1003
variable_call_uuid: e3133bf7-dcde-4daf-9663-9a79ffcef5ad
variable_ringback: %25(2000,4000,440,480)
variable_call_timeout: 30
variable_dialed_user: 1003
variable_dialed_domain: cgrates.org
variable_originated_legs: ARRAY%3A%3A0a30dd7c-c222-482f-a322-b1218a15f8cd%3BOutbound%20Call%3B1003%7C%3A0a30dd7c-c222-482f-a322-b1218a15f8cd%3BOutbound%20Call%3B1003
variable_switch_m_sdp: v%3D0%0D%0Ao%3D-%203645269528%203645269529%20IN%20IP4%2010.0.3.15%0D%0As%3Dpjmedia%0D%0Ab%3DAS%3A84%0D%0At%3D0%200%0D%0Aa%3DX-nat%3A0%0D%0Am%3Daudio%204018%20RTP/AVP%2099%20101%0D%0Ac%3DIN%20IP4%2010.0.3.15%0D%0Ab%3DAS%3A64000%0D%0Aa%3Drtpmap%3A99%20speex/32000%0D%0Aa%3Drtpmap%3A101%20telephone-event/8000%0D%0Aa%3Dfmtp%3A101%200-16%0D%0Aa%3Drtcp%3A4019%20IN%20IP4%2010.0.3.15%0D%0A
variable_rtp_local_sdp_str: v%3D0%0Ao%3DFreeSWITCH%201436250882%201436250883%20IN%20IP4%2010.0.3.15%0As%3DFreeSWITCH%0Ac%3DIN%20IP4%2010.0.3.15%0At%3D0%200%0Am%3Daudio%2029846%20RTP/AVP%2099%2096%0Aa%3Drtpmap%3A99%20speex/32000%0Aa%3Drtpmap%3A96%20telephone-event/8000%0Aa%3Dfmtp%3A96%200-16%0Aa%3Dptime%3A20%0Aa%3Dsendrecv%0Aa%3Drtcp%3A29847%20IN%20IP4%2010.0.3.15%0A
variable_local_media_ip: 10.0.3.15
variable_local_media_port: 29846
variable_advertised_media_ip: 10.0.3.15
variable_rtp_use_pt: 99
variable_rtp_use_ssrc: 1470667272
variable_rtp_2833_send_payload: 96
variable_rtp_2833_recv_payload: 96
variable_remote_media_ip: 10.0.3.15
variable_remote_media_port: 4006
variable_endpoint_disposition: ANSWER
variable_current_application_data: %2B3120%20alloted_timeout
variable_current_application: sched_hangup
variable_originate_causes: ARRAY%3A%3A0a30dd7c-c222-482f-a322-b1218a15f8cd%3BNONE%7C%3A0a30dd7c-c222-482f-a322-b1218a15f8cd%3BNONE
variable_originate_disposition: SUCCESS
variable_DIALSTATUS: SUCCESS
variable_last_bridge_to: 0a30dd7c-c222-482f-a322-b1218a15f8cd
variable_bridge_channel: sofia/cgrtest/1003%40127.0.0.1%3A5070
variable_bridge_uuid: 0a30dd7c-c222-482f-a322-b1218a15f8cd
variable_signal_bond: 0a30dd7c-c222-482f-a322-b1218a15f8cd
variable_sip_to_tag: 5Qt4ecvreSHZN
variable_sip_from_tag: YwuG8U3rRbqIn.xYTnU8NrI3giyxDBHJ
variable_sip_cseq: 4178
variable_sip_call_id: r3xaJ8CLpyTAIHWUZG7gtZQYgAPEGf9S
variable_sip_full_via: SIP/2.0/UDP%2010.0.3.15%3A5072%3Brport%3D5072%3Bbranch%3Dz9hG4bKPjPqma7vnLxDkBqcCH3eXLmLYZoPS.6MDc%3Breceived%3D127.0.0.1
variable_sip_full_from: sip%3A1001%40127.0.0.1%3Btag%3DYwuG8U3rRbqIn.xYTnU8NrI3giyxDBHJ
variable_sip_full_to: sip%3A1003%40127.0.0.1%3Btag%3D5Qt4ecvreSHZN
variable_last_sent_callee_id_name: Outbound%20Call
variable_last_sent_callee_id_number: 1003
variable_sip_term_status: 200
variable_proto_specific_hangup_cause: sip%3A200
variable_sip_term_cause: 16
variable_last_bridge_role: originator
variable_sip_user_agent: PJSUA%20v2.3%20Linux-3.2.0.4/x86_64/glibc-2.13
variable_sip_hangup_disposition: recv_bye
variable_bridge_hangup_cause: NORMAL_CLEARING
variable_hangup_cause: NORMAL_CLEARING
variable_hangup_cause_q850: 16
variable_digits_dialed: none
variable_start_stamp: 2015-07-07%2016%3A52%3A08
variable_profile_start_stamp: 2015-07-07%2016%3A52%3A08
variable_answer_stamp: 2015-07-07%2016%3A52%3A08
variable_bridge_stamp: 2015-07-07%2016%3A52%3A08
variable_end_stamp: 2015-07-07%2016%3A53%3A14
variable_start_epoch: 1436280728
variable_start_uepoch: 1436280728471153
variable_profile_start_epoch: 1436280728
variable_profile_start_uepoch: 1436280728930693
variable_answer_epoch: 1436280728
variable_answer_uepoch: 1436280728971147
variable_bridge_epoch: 1436280728
variable_bridge_uepoch: 1436280728971147
variable_last_hold_epoch: 0
variable_last_hold_uepoch: 0
variable_hold_accum_seconds: 0
variable_hold_accum_usec: 0
variable_hold_accum_ms: 0
variable_resurrect_epoch: 0
variable_resurrect_uepoch: 0
variable_progress_epoch: 0
variable_progress_uepoch: 0
variable_progress_media_epoch: 0
variable_progress_media_uepoch: 0
variable_end_epoch: 1436280794
variable_end_uepoch: 1436280794010851
variable_last_app: sched_hangup
variable_last_arg: %2B3120%20alloted_timeout
variable_caller_id: %221001%22%20%3C1001%3E
variable_duration: 66
variable_billsec: 66
variable_progresssec: 0
variable_answersec: 0
variable_waitsec: 0
variable_progress_mediasec: 0
variable_flow_billsec: 66
variable_mduration: 65539
variable_billmsec: 65039
variable_progressmsec: 28
variable_answermsec: 500
variable_waitmsec: 500
variable_progress_mediamsec: 28
variable_cgr_acd: 30
variable_flow_billmsec: 65539
variable_uduration: 65539698
variable_billusec: 65039704
variable_progressusec: 0
variable_answerusec: 499994
variable_waitusec: 499994
variable_progress_mediausec: 0
variable_flow_billusec: 65539698
variable_rtp_audio_in_raw_bytes: 6770
variable_rtp_audio_in_media_bytes: 6762
variable_rtp_audio_in_packet_count: 192
variable_rtp_audio_in_media_packet_count: 190
variable_rtp_audio_in_skip_packet_count: 6
variable_rtp_audio_in_jitter_packet_count: 0
variable_rtp_audio_in_dtmf_packet_count: 0
variable_rtp_audio_in_cng_packet_count: 0
variable_rtp_audio_in_flush_packet_count: 2
variable_rtp_audio_in_largest_jb_size: 0
variable_rtp_audio_in_jitter_min_variance: 26.73
variable_rtp_audio_in_jitter_max_variance: 6716.71
variable_rtp_audio_in_jitter_loss_rate: 0.00
variable_rtp_audio_in_jitter_burst_rate: 0.00
variable_rtp_audio_in_mean_interval: 36.67
variable_rtp_audio_in_flaw_total: 0
variable_rtp_audio_in_quality_percentage: 100.00
variable_rtp_audio_in_mos: 4.50
variable_rtp_audio_out_raw_bytes: 4686
variable_rtp_audio_out_media_bytes: 4686
variable_rtp_audio_out_packet_count: 108
variable_rtp_audio_out_media_packet_count: 108
variable_rtp_audio_out_skip_packet_count: 0
variable_rtp_audio_out_dtmf_packet_count: 0
variable_rtp_audio_out_cng_packet_count: 0
variable_rtp_audio_rtcp_packet_count: 1450
variable_rtp_audio_rtcp_octet_count: 45940`

func TestEventCreation(t *testing.T) {
	body := `Event-Name: RE_SCHEDULE
Core-UUID: 792e181c-b6e6-499c-82a1-52a778e7d82d
FreeSWITCH-Hostname: h1.ip-switch.net
FreeSWITCH-Switchname: h1.ip-switch.net
FreeSWITCH-IPv4: 88.198.12.156
FreeSWITCH-IPv6: %3A%3A1
Event-Date-Local: 2012-10-05%2013%3A41%3A38
Event-Date-GMT: Fri,%2005%20Oct%202012%2011%3A41%3A38%20GMT
Event-Date-Timestamp: 1349437298012866
Event-Calling-File: switch_scheduler.c
Event-Calling-Function: switch_scheduler_execute
Event-Calling-Line-Number: 65
Event-Sequence: 34263
Task-ID: 2
Task-Desc: heartbeat
Task-Group: core
Task-Runtime: 1349437318`
	ev := NewFSEvent(body)
	if ev.GetName() != "RE_SCHEDULE" {
		t.Error("Event not parsed correctly: ", ev)
	}
	l := len(ev)
	if l != 17 {
		t.Error("Incorrect number of event fields: ", l)
	}
}

// Detects if any of the parsers do not return static values
func TestEventParseStatic(t *testing.T) {
	ev := NewFSEvent("")
	setupTime, _ := ev.GetSetupTime("^2013-12-07 08:42:24", "")
	answerTime, _ := ev.GetAnswerTime("^2013-12-07 08:42:24", "")
	dur, _ := ev.GetDuration("^60s")
	if ev.GetReqType("^test") != "test" ||
		ev.GetTenant("^test") != "test" ||
		ev.GetCategory("^test") != "test" ||
		ev.GetAccount("^test") != "test" ||
		ev.GetSubject("^test") != "test" ||
		ev.GetDestination("^test") != "test" ||
		setupTime != time.Date(2013, 12, 7, 8, 42, 24, 0, time.UTC) ||
		answerTime != time.Date(2013, 12, 7, 8, 42, 24, 0, time.UTC) ||
		dur != time.Duration(60)*time.Second {
		t.Error("Values out of static not matching",
			ev.GetReqType("^test") != "test",
			ev.GetTenant("^test") != "test",
			ev.GetCategory("^test") != "test",
			ev.GetAccount("^test") != "test",
			ev.GetSubject("^test") != "test",
			ev.GetDestination("^test") != "test",
			setupTime != time.Date(2013, 12, 7, 8, 42, 24, 0, time.UTC),
			answerTime != time.Date(2013, 12, 7, 8, 42, 24, 0, time.UTC),
			dur != time.Duration(60)*time.Second)
	}
}

// Test here if the answer is selected out of headers we specify, even if not default defined
func TestEventSelectiveHeaders(t *testing.T) {
	body := `Event-Name: RE_SCHEDULE
Core-UUID: 792e181c-b6e6-499c-82a1-52a778e7d82d
FreeSWITCH-Hostname: h1.ip-switch.net
FreeSWITCH-Switchname: h1.ip-switch.net
FreeSWITCH-IPv4: 88.198.12.156
FreeSWITCH-IPv6: %3A%3A1
Event-Date-Local: 2012-10-05%2013%3A41%3A38
Event-Date-GMT: Fri,%2005%20Oct%202012%2011%3A41%3A38%20GMT
Event-Date-Timestamp: 1349437298012866
Event-Calling-File: switch_scheduler.c
Event-Calling-Function: switch_scheduler_execute
Event-Calling-Line-Number: 65
Event-Sequence: 34263
Task-ID: 2
Task-Desc: heartbeat
Task-Group: core
Task-Runtime: 1349437318`
	cfg, _ := config.NewDefaultCGRConfig()
	config.SetCgrConfig(cfg)
	ev := NewFSEvent(body)
	setupTime, _ := ev.GetSetupTime("Event-Date-Local", "")
	answerTime, _ := ev.GetAnswerTime("Event-Date-Local", "")
	dur, _ := ev.GetDuration("Event-Calling-Line-Number")
	if ev.GetReqType("FreeSWITCH-Hostname") != "h1.ip-switch.net" ||
		ev.GetTenant("FreeSWITCH-Hostname") != "h1.ip-switch.net" ||
		ev.GetCategory("FreeSWITCH-Hostname") != "h1.ip-switch.net" ||
		ev.GetAccount("FreeSWITCH-Hostname") != "h1.ip-switch.net" ||
		ev.GetSubject("FreeSWITCH-Hostname") != "h1.ip-switch.net" ||
		ev.GetDestination("FreeSWITCH-Hostname") != "h1.ip-switch.net" ||
		setupTime != time.Date(2012, 10, 5, 13, 41, 38, 0, time.UTC) ||
		answerTime != time.Date(2012, 10, 5, 13, 41, 38, 0, time.UTC) ||
		dur != time.Duration(65)*time.Second {
		t.Error("Values out of static not matching",
			ev.GetReqType("FreeSWITCH-Hostname") != "h1.ip-switch.net",
			ev.GetTenant("FreeSWITCH-Hostname") != "h1.ip-switch.net",
			ev.GetCategory("FreeSWITCH-Hostname") != "h1.ip-switch.net",
			ev.GetAccount("FreeSWITCH-Hostname") != "h1.ip-switch.net",
			ev.GetSubject("FreeSWITCH-Hostname") != "h1.ip-switch.net",
			ev.GetDestination("FreeSWITCH-Hostname") != "h1.ip-switch.net",
			setupTime != time.Date(2012, 10, 5, 13, 41, 38, 0, time.UTC),
			answerTime != time.Date(2012, 10, 5, 13, 41, 38, 0, time.UTC),
			dur != time.Duration(65)*time.Second)
	}
}

func TestDDazEmptyTime(t *testing.T) {
	body := `Event-Name: RE_SCHEDULE
Core-UUID: 792e181c-b6e6-499c-82a1-52a778e7d82d
FreeSWITCH-Hostname: h1.ip-switch.net
FreeSWITCH-Switchname: h1.ip-switch.net
FreeSWITCH-IPv4: 88.198.12.156
Caller-Channel-Created-Time: 0
Caller-Channel-Answered-Time
Task-Runtime: 1349437318`
	var nilTime time.Time
	ev := NewFSEvent(body)
	if setupTime, err := ev.GetSetupTime("", ""); err != nil {
		t.Error("Error when parsing empty setupTime")
	} else if setupTime != nilTime {
		t.Error("Expecting nil time, got: ", setupTime)
	}
	if answerTime, err := ev.GetAnswerTime("", ""); err != nil {
		t.Error("Error when parsing empty setupTime")
	} else if answerTime != nilTime {
		t.Error("Expecting nil time, got: ", answerTime)
	}
}

func TestParseFsHangup(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	config.SetCgrConfig(cfg)
	ev := NewFSEvent(hangupEv)
	setupTime, _ := ev.GetSetupTime(utils.META_DEFAULT, "")
	answerTime, _ := ev.GetAnswerTime(utils.META_DEFAULT, "")
	dur, _ := ev.GetDuration(utils.META_DEFAULT)
	if ev.GetReqType(utils.META_DEFAULT) != utils.META_PREPAID ||
		ev.GetTenant(utils.META_DEFAULT) != "cgrates.org" ||
		ev.GetCategory(utils.META_DEFAULT) != "call" ||
		ev.GetAccount(utils.META_DEFAULT) != "1001" ||
		ev.GetSubject(utils.META_DEFAULT) != "1001" ||
		ev.GetDestination(utils.META_DEFAULT) != "1003" ||
		setupTime.UTC() != time.Date(2015, 7, 7, 14, 52, 8, 0, time.UTC) ||
		answerTime.UTC() != time.Date(2015, 7, 7, 14, 52, 8, 0, time.UTC) ||
		dur != time.Duration(66)*time.Second {
		t.Error("Default values not matching",
			ev.GetReqType(utils.META_DEFAULT) != utils.META_PREPAID,
			ev.GetTenant(utils.META_DEFAULT) != "cgrates.org",
			ev.GetCategory(utils.META_DEFAULT) != "call",
			ev.GetAccount(utils.META_DEFAULT) != "1001",
			ev.GetSubject(utils.META_DEFAULT) != "1001",
			ev.GetDestination(utils.META_DEFAULT) != "1003",
			setupTime.UTC() != time.Date(2015, 7, 7, 14, 52, 8, 0, time.UTC),
			answerTime.UTC() != time.Date(2015, 7, 7, 14, 52, 8, 0, time.UTC),
			dur != time.Duration(66)*time.Second)
	}
}

func TestParseEventValue(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	config.SetCgrConfig(cfg)
	ev := NewFSEvent(hangupEv)
	if tor := ev.ParseEventValue(&utils.RSRField{Id: utils.TOR}, ""); tor != utils.VOICE {
		t.Error("Unexpected tor parsed", tor)
	}
	if accid := ev.ParseEventValue(&utils.RSRField{Id: utils.OriginID}, ""); accid != "e3133bf7-dcde-4daf-9663-9a79ffcef5ad" {
		t.Error("Unexpected result parsed", accid)
	}
	if parsed := ev.ParseEventValue(&utils.RSRField{Id: utils.OriginHost}, ""); parsed != "10.0.3.15" {
		t.Error("Unexpected result parsed", parsed)
	}
	if parsed := ev.ParseEventValue(&utils.RSRField{Id: utils.Source}, ""); parsed != "FS_EVENT" {
		t.Error("Unexpected result parsed", parsed)
	}
	if parsed := ev.ParseEventValue(&utils.RSRField{Id: utils.RequestType}, ""); parsed != utils.META_PREPAID {
		t.Error("Unexpected result parsed", parsed)
	}
	if parsed := ev.ParseEventValue(&utils.RSRField{Id: utils.Direction}, ""); parsed != utils.OUT {
		t.Error("Unexpected result parsed", parsed)
	}
	if parsed := ev.ParseEventValue(&utils.RSRField{Id: utils.Tenant}, ""); parsed != "cgrates.org" {
		t.Error("Unexpected result parsed", parsed)
	}
	if parsed := ev.ParseEventValue(&utils.RSRField{Id: utils.Category}, ""); parsed != "call" {
		t.Error("Unexpected result parsed", parsed)
	}
	if parsed := ev.ParseEventValue(&utils.RSRField{Id: utils.Account}, ""); parsed != "1001" {
		t.Error("Unexpected result parsed", parsed)
	}
	if parsed := ev.ParseEventValue(&utils.RSRField{Id: utils.Subject}, ""); parsed != "1001" {
		t.Error("Unexpected result parsed", parsed)
	}
	if parsed := ev.ParseEventValue(&utils.RSRField{Id: utils.Destination}, ""); parsed != "1003" {
		t.Error("Unexpected result parsed", parsed)
	}
	sTime, _ := utils.ParseTimeDetectLayout("1436280728471153"[:len("1436280728471153")-6], "") // We discard nanoseconds information so we can correlate csv
	if parsed := ev.ParseEventValue(&utils.RSRField{Id: utils.SetupTime}, ""); parsed != sTime.String() {
		t.Errorf("Expecting: %s, parsed: %s", sTime.String(), parsed)
	}
	aTime, _ := utils.ParseTimeDetectLayout("1436280728971147"[:len("1436280728971147")-6], "")
	if parsed := ev.ParseEventValue(&utils.RSRField{Id: utils.AnswerTime}, ""); parsed != aTime.String() {
		t.Errorf("Expecting: %s, parsed: %s", aTime.String(), parsed)
	}
	if parsed := ev.ParseEventValue(&utils.RSRField{Id: utils.Usage}, ""); parsed != "66000000000" {
		t.Error("Unexpected result parsed", parsed)
	}
	if parsed := ev.ParseEventValue(&utils.RSRField{Id: utils.PDD}, ""); parsed != "0.028" {
		t.Error("Unexpected result parsed", parsed)
	}
	if parsed := ev.ParseEventValue(&utils.RSRField{Id: utils.SUPPLIER}, ""); parsed != "supplier1" {
		t.Error("Unexpected result parsed", parsed)
	}
	if parsed := ev.ParseEventValue(&utils.RSRField{Id: utils.MEDI_RUNID}, ""); parsed != utils.DEFAULT_RUNID {
		t.Error("Unexpected result parsed", parsed)
	}
	if parsed := ev.ParseEventValue(&utils.RSRField{Id: utils.COST}, ""); parsed != "-1" {
		t.Error("Unexpected result parsed", parsed)
	}
	if parsed := ev.ParseEventValue(&utils.RSRField{Id: "Hangup-Cause"}, ""); parsed != "NORMAL_CLEARING" {
		t.Error("Unexpected result parsed", parsed)
	}
}

func TestFsEvAsCGREvent(t *testing.T) {
	timezone := config.CgrConfig().DefaultTimezone
	ev := NewFSEvent(hangupEv)
	sTime, err := ev.GetSetupTime(utils.META_DEFAULT, timezone)
	if err != nil {
		t.Error(err)
	}
	expected := &utils.CGREvent{
		Tenant: ev.GetTenant(utils.META_DEFAULT),
		ID:     utils.UUIDSha1Prefix(),
		Time:   &sTime,
		Event:  ev.AsMapStringInterface(timezone),
	}
	if rcv, err := ev.AsCGREvent(timezone); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected.Tenant, rcv.Tenant) {
		t.Errorf("Expecting: %+v, received: %+v", expected.Tenant, rcv.Tenant)
	} else if !reflect.DeepEqual(expected.Time, rcv.Time) {
		t.Errorf("Expecting: %+v, received: %+v", expected.Time, rcv.Time)
	} else if !reflect.DeepEqual(expected.Event, rcv.Event) {
		t.Errorf("Expecting: %+v, received: %+v", expected.Event, rcv.Event)
	}
}

func TestFsEvAsMapStringInterface(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	config.SetCgrConfig(cfg)
	ev := NewFSEvent(hangupEv)
	setupTime, _ := utils.ParseTimeDetectLayout("1436280728", "")
	aTime, _ := utils.ParseTimeDetectLayout("1436280728", "")
	expectedMap := make(map[string]interface{})
	expectedMap[utils.TOR] = utils.VOICE
	expectedMap[utils.OriginID] = "e3133bf7-dcde-4daf-9663-9a79ffcef5ad"
	expectedMap[utils.OriginHost] = "10.0.3.15"
	expectedMap[utils.Source] = "FS_CHANNEL_HANGUP_COMPLETE"
	expectedMap[utils.Category] = "call"
	expectedMap[utils.SetupTime] = setupTime
	expectedMap[utils.AnswerTime] = aTime
	expectedMap[utils.RequestType] = utils.META_PREPAID
	expectedMap[utils.Direction] = "*out"
	expectedMap[utils.Destination] = "1003"
	expectedMap[utils.Usage] = time.Duration(66) * time.Second
	expectedMap[utils.Tenant] = "cgrates.org"
	expectedMap[utils.Account] = "1001"
	expectedMap[utils.Subject] = "1001"
	expectedMap[utils.Cost] = -1
	expectedMap[utils.PDD] = time.Duration(28) * time.Millisecond
	expectedMap[utils.ACD] = time.Duration(30) * time.Second
	expectedMap[utils.DISCONNECT_CAUSE] = "NORMAL_CLEARING"
	expectedMap[utils.SUPPLIER] = "supplier1"
	if storedMap := ev.AsMapStringInterface(""); !reflect.DeepEqual(expectedMap, storedMap) {
		t.Errorf("Expecting: %+v, received: %+v", expectedMap, storedMap)
	}
}

func TestFsEvGetExtraFields(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	cfg.FsAgentCfg().ExtraFields = []*utils.RSRField{
		&utils.RSRField{Id: "Channel-Read-Codec-Name"},
		&utils.RSRField{Id: "Channel-Write-Codec-Name"},
		&utils.RSRField{Id: "NonExistingHeader"}}
	config.SetCgrConfig(cfg)
	ev := NewFSEvent(hangupEv)
	expectedExtraFields := map[string]string{
		"Channel-Read-Codec-Name":  "SPEEX",
		"Channel-Write-Codec-Name": "SPEEX", "NonExistingHeader": ""}
	if extraFields := ev.GetExtraFields(); !reflect.DeepEqual(expectedExtraFields, extraFields) {
		t.Errorf("Expecting: %+v, received: %+v", expectedExtraFields, extraFields)
	}
}

func TestSliceAsFsArray(t *testing.T) {
	items := []string{}
	if fsArray := SliceAsFsArray(items); fsArray != "" {
		t.Error(fsArray)
	}
	items = []string{"item1", "item2", "item3"}
	if fsArray := SliceAsFsArray(items); fsArray != "ARRAY::3|:item1|:item2|:item3" {
		t.Error(fsArray)
	}
}

// Make sure processing of the hangup event produces the same output as FS-JSON CDR
func TestSyncFsEventWithJsonCdr(t *testing.T) {
	body := []byte(`
		{"core-uuid":"61b38010-fb8e-4022-89b0-3c0be6e500b9","switchname":"teo","channel_data":{"state":"CS_REPORTING","direction":"inbound","state_number":"11","flags":"0=1;1=1;3=1;20=1;37=1;38=1;40=1;43=1;48=1;53=1;75=1;77=1;112=1;113=1;122=1;134=1","caps":"1=1;2=1;3=1;4=1;5=1;6=1"},"callStats":{"audio":{"inbound":{"raw_bytes":505772,"media_bytes":501988,"packet_count":2961,"media_packet_count":2939,"skip_packet_count":8,"jitter_packet_count":0,"dtmf_packet_count":0,"cng_packet_count":0,"flush_packet_count":22,"largest_jb_size":0,"jitter_min_variance":18,"jitter_max_variance":338,"jitter_loss_rate":90.827031,"jitter_burst_rate":89.827031,"mean_interval":20.016053,"flaw_total":1,"quality_percentage":99,"mos":4.492027,"errorLog":[{"start":1520944107460545,"stop":1520944114478926,"flaws":242610,"consecutiveFlaws":0,"durationMS":7018}]},"outbound":{"raw_bytes":488136,"media_bytes":488136,"packet_count":2838,"media_packet_count":2838,"skip_packet_count":0,"dtmf_packet_count":0,"cng_packet_count":0,"rtcp_packet_count":0,"rtcp_octet_count":0}}},"variables":{"direction":"inbound","uuid":"5282f87c-368c-43a2-babd-bc63ff12dc79","session_id":"5","sip_from_user":"1001","sip_from_uri":"1001@192.168.56.202","sip_from_host":"192.168.56.202","channel_name":"sofia/internal/1001@192.168.56.202","ep_codec_string":"mod_spandsp.G722@8000h@20i@64000b,CORE_PCM_MODULE.PCMU@8000h@20i@64000b,CORE_PCM_MODULE.PCMA@8000h@20i@64000b,mod_spandsp.GSM@8000h@20i@13200b","sip_local_network_addr":"192.168.56.202","sip_network_ip":"192.168.56.1","sip_network_port":"5060","sip_invite_stamp":"1520944103557418","sip_received_ip":"192.168.56.1","sip_received_port":"5060","sip_via_protocol":"udp","sip_authorized":"true","Event-Name":"REQUEST_PARAMS","Core-UUID":"61b38010-fb8e-4022-89b0-3c0be6e500b9","FreeSWITCH-Hostname":"teo","FreeSWITCH-Switchname":"teo","FreeSWITCH-IPv4":"10.0.2.15","FreeSWITCH-IPv6":"::1","Event-Date-Local":"2018-03-13 08:28:23","Event-Date-GMT":"Tue, 13 Mar 2018 12:28:23 GMT","Event-Date-Timestamp":"1520944103557418","Event-Calling-File":"sofia.c","Event-Calling-Function":"sofia_handle_sip_i_invite","Event-Calling-Line-Number":"10096","Event-Sequence":"742","sip_number_alias":"1001","sip_auth_username":"1001","sip_auth_realm":"192.168.56.202","number_alias":"1001","requested_user_name":"1001","requested_domain_name":"192.168.56.202","record_stereo":"true","transfer_fallback_extension":"operator","toll_allow":"domestic,international,local","accountcode":"1001","user_context":"default","effective_caller_id_name":"Extension 1001","effective_caller_id_number":"1001","callgroup":"techsupport","cgr_reqtype":"*prepaid","cgr_subsystems":"*resources;*attributes;*sessions;*suppliers;*thresholds;*stats;*accounts","cgr_acd":"30","user_name":"1001","domain_name":"192.168.56.202","sip_from_user_stripped":"1001","sofia_profile_name":"internal","recovery_profile_name":"internal","sip_req_user":"1002","sip_req_uri":"1002@192.168.56.202","sip_req_host":"192.168.56.202","sip_to_user":"1002","sip_to_uri":"1002@192.168.56.202","sip_to_host":"192.168.56.202","sip_contact_params":"transport=udp;registering_acc=192_168_56_202","sip_contact_user":"1001","sip_contact_port":"5060","sip_contact_uri":"1001@192.168.56.1:5060","sip_contact_host":"192.168.56.1","sip_user_agent":"Jitsi2.10.5550Windows 10","sip_via_host":"192.168.56.1","sip_via_port":"5060","presence_id":"1001@192.168.56.202","execute_on_answer":"sched_hangup +3048 alloted_timeout","cgr_resource_allocation":"ResGroup1","cgr_suppliers":"ARRAY::3|:supplier3|:supplier1|:supplier2","cgr_notify":"AUTH_OK","max_forwards":"69","transfer_history":"1520944103:ec99d6a9-72b7-4184-88aa-87b240f5465c:bl_xfer:1002/default/XML","transfer_source":"1520944103:ec99d6a9-72b7-4184-88aa-87b240f5465c:bl_xfer:1002/default/XML","DP_MATCH":"ARRAY::1002|:1002","call_uuid":"5282f87c-368c-43a2-babd-bc63ff12dc79","call_timeout":"30","dialed_user":"1002","dialed_domain":"192.168.56.202","originated_legs":"ARRAY::fa3173aa-3c1a-41e5-ab82-60623396fa08;Outbound Call;1002|:fa3173aa-3c1a-41e5-ab82-60623396fa08;Outbound Call;1002","switch_m_sdp":"v=0\r\no=1002-jitsi.org 0 0 IN IP4 192.168.56.1\r\ns=-\r\nc=IN IP4 192.168.56.1\r\nt=0 0\r\nm=audio 5106 RTP/AVP 9 0 8 3 101\r\na=rtpmap:9 G722/8000\r\na=rtpmap:0 PCMU/8000\r\na=rtpmap:8 PCMA/8000\r\na=rtpmap:3 GSM/8000\r\na=rtpmap:101 telephone-event/8000\r\n","rtp_use_codec_name":"G722","rtp_use_codec_rate":"8000","rtp_use_codec_ptime":"20","rtp_use_codec_channels":"1","rtp_last_audio_codec_string":"G722@8000h@20i@1c","read_codec":"G722","original_read_codec":"G722","read_rate":"16000","original_read_rate":"16000","write_codec":"G722","write_rate":"16000","local_media_ip":"192.168.56.202","local_media_port":"22992","advertised_media_ip":"192.168.56.202","rtp_use_timer_name":"soft","rtp_use_pt":"9","rtp_use_ssrc":"850142423","endpoint_disposition":"ANSWER","current_application_data":"+3048 alloted_timeout","current_application":"sched_hangup","originate_causes":"ARRAY::fa3173aa-3c1a-41e5-ab82-60623396fa08;NONE|:fa3173aa-3c1a-41e5-ab82-60623396fa08;NONE","originate_disposition":"SUCCESS","DIALSTATUS":"SUCCESS","last_bridge_to":"fa3173aa-3c1a-41e5-ab82-60623396fa08","bridge_channel":"sofia/internal/1002@192.168.56.1:5060","bridge_uuid":"fa3173aa-3c1a-41e5-ab82-60623396fa08","signal_bond":"fa3173aa-3c1a-41e5-ab82-60623396fa08","last_sent_callee_id_name":"Outbound Call","last_sent_callee_id_number":"1002","switch_r_sdp":"v=0\r\no=1001-jitsi.org 0 2 IN IP4 192.168.56.1\r\ns=-\r\nc=IN IP4 192.168.56.1\r\nt=0 0\r\nm=audio 5102 RTP/AVP 96 97 98 9 100 102 0 8 103 3 104 4 101\r\na=rtpmap:96 opus/48000/2\r\na=fmtp:96 usedtx=1\r\na=rtpmap:97 SILK/24000\r\na=rtpmap:98 SILK/16000\r\na=rtpmap:9 G722/8000\r\na=rtpmap:100 speex/32000\r\na=rtpmap:102 speex/16000\r\na=rtpmap:0 PCMU/8000\r\na=rtpmap:8 PCMA/8000\r\na=rtpmap:103 iLBC/8000\r\na=rtpmap:3 GSM/8000\r\na=rtpmap:104 speex/8000\r\na=rtpmap:4 G723/8000\r\na=fmtp:4 annexa=no;bitrate=6.3\r\na=rtpmap:101 telephone-event/8000\r\na=ptime:20\r\na=extmap:1 urn:ietf:params:rtp-hdrext:csrc-audio-level\r\na=extmap:2 urn:ietf:params:rtp-hdrext:ssrc-audio-level\r\na=rtcp-xr:voip-metrics\r\na=zrtp-hash:1.10 dba9c36338c885a1bd67b120bf2f1aaff977a5f496ce54803911e465daa33eb2\r\nm=video 5108 RTP/AVP 105 99\r\na=rtpmap:105 H264/90000\r\na=fmtp:105 profile-level-id=4DE01f;packetization-mode=1\r\na=rtpmap:99 H264/90000\r\na=fmtp:99 profile-level-id=4DE01f\r\na=recvonly\r\na=imageattr:105 send * recv [x=[1:1920],y=[1:1080]]\r\na=imageattr:99 send * recv [x=[1:1920],y=[1:1080]]\r\n","rtp_use_codec_string":"G722,PCMU,PCMA,GSM","r_sdp_audio_zrtp_hash":"1.10 dba9c36338c885a1bd67b120bf2f1aaff977a5f496ce54803911e465daa33eb2","audio_media_flow":"sendrecv","remote_media_ip":"192.168.56.1","remote_media_port":"5102","rtp_audio_recv_pt":"9","dtmf_type":"rfc2833","rtp_2833_send_payload":"101","rtp_2833_recv_payload":"101","video_possible":"true","video_media_flow":"sendonly","rtp_local_sdp_str":"v=0\r\no=FreeSWITCH 1520921113 1520921116 IN IP4 192.168.56.202\r\ns=FreeSWITCH\r\nc=IN IP4 192.168.56.202\r\nt=0 0\r\nm=audio 22992 RTP/AVP 9 101\r\na=rtpmap:9 G722/8000\r\na=rtpmap:101 telephone-event/8000\r\na=fmtp:101 0-16\r\na=ptime:20\r\na=sendrecv\r\nm=video 0 RTP/AVP 19\r\n","sip_to_tag":"KjBaScg1vBF8r","sip_from_tag":"a27a6086","sip_cseq":"4","sip_call_id":"b1acaf2198f12cc429ab8a615cac083d@0:0:0:0:0:0:0:0","sip_full_via":"SIP/2.0/UDP 192.168.56.1:5060;branch=z9hG4bK-333939-4e9c820926945f1a2c6d0bf5ac2da12e","sip_from_display":"1001","sip_full_from":"\"1001\" <sip:1001@192.168.56.202>;tag=a27a6086","sip_full_to":"<sip:1002@192.168.56.202>;tag=KjBaScg1vBF8r","sip_hangup_phrase":"OK","last_bridge_hangup_cause":"NORMAL_CLEARING","last_bridge_proto_specific_hangup_cause":"sip:200","bridge_hangup_cause":"NORMAL_CLEARING","hangup_cause":"NORMAL_CLEARING","hangup_cause_q850":"16","digits_dialed":"none","start_stamp":"2018-03-13 08:28:23","profile_start_stamp":"2018-03-13 08:28:23","answer_stamp":"2018-03-13 08:28:25","bridge_stamp":"2018-03-13 08:28:25","hold_stamp":"2018-03-13 08:28:25","progress_stamp":"2018-03-13 08:28:23","progress_media_stamp":"2018-03-13 08:28:25","hold_events":"{{1520944105098164,1520944107306456}}","end_stamp":"2018-03-13 08:29:24","start_epoch":"1520944103","start_uepoch":"1520944103557418","profile_start_epoch":"1520944103","profile_start_uepoch":"1520944103617274","answer_epoch":"1520944105","answer_uepoch":"1520944105037574","bridge_epoch":"1520944105","bridge_uepoch":"1520944105037574","last_hold_epoch":"1520944105","last_hold_uepoch":"1520944105098164","hold_accum_seconds":"2","hold_accum_usec":"2208292","hold_accum_ms":"2208","resurrect_epoch":"0","resurrect_uepoch":"0","progress_epoch":"1520944103","progress_uepoch":"1520944103637959","progress_media_epoch":"1520944105","progress_media_uepoch":"1520944105037574","end_epoch":"1520944164","end_uepoch":"1520944164057232","last_app":"sched_hangup","last_arg":"+3048 alloted_timeout","caller_id":"\"1001\" <1001>","duration":"61","billsec":"59","progresssec":"0","answersec":"2","waitsec":"2","progress_mediasec":"2","flow_billsec":"61","mduration":"60500","billmsec":"59020","progressmsec":"80","answermsec":"1480","waitmsec":"1480","progress_mediamsec":"1480","flow_billmsec":"60500","uduration":"60499814","billusec":"59019658","progressusec":"80541","answerusec":"1480156","waitusec":"1480156","progress_mediausec":"1480156","flow_billusec":"60499814","sip_hangup_disposition":"send_bye","rtp_audio_in_raw_bytes":"505772","rtp_audio_in_media_bytes":"501988","rtp_audio_in_packet_count":"2961","rtp_audio_in_media_packet_count":"2939","rtp_audio_in_skip_packet_count":"8","rtp_audio_in_jitter_packet_count":"0","rtp_audio_in_dtmf_packet_count":"0","rtp_audio_in_cng_packet_count":"0","rtp_audio_in_flush_packet_count":"22","rtp_audio_in_largest_jb_size":"0","rtp_audio_in_jitter_min_variance":"18.00","rtp_audio_in_jitter_max_variance":"338.00","rtp_audio_in_jitter_loss_rate":"90.83","rtp_audio_in_jitter_burst_rate":"89.83","rtp_audio_in_mean_interval":"20.02","rtp_audio_in_flaw_total":"1","rtp_audio_in_quality_percentage":"99.00","rtp_audio_in_mos":"4.49","rtp_audio_out_raw_bytes":"488136","rtp_audio_out_media_bytes":"488136","rtp_audio_out_packet_count":"2838","rtp_audio_out_media_packet_count":"2838","rtp_audio_out_skip_packet_count":"0","rtp_audio_out_dtmf_packet_count":"0","rtp_audio_out_cng_packet_count":"0","rtp_audio_rtcp_packet_count":"0","rtp_audio_rtcp_octet_count":"0"},"app_log":{"applications":[{"app_name":"info","app_data":"","app_stamp":"1520944103580845"},{"app_name":"park","app_data":"","app_stamp":"1520944103581317"},{"app_name":"set","app_data":"ringback=","app_stamp":"1520944103622870"},{"app_name":"set","app_data":"call_timeout=30","app_stamp":"1520944103623092"},{"app_name":"bridge","app_data":"user/1002@192.168.56.202","app_stamp":"1520944103623254"},{"app_name":"sched_hangup","app_data":"+3048 alloted_timeout","app_stamp":"1520944105040579"}]},"callflow":[{"dialplan":"XML","profile_index":"2","extension":{"name":"Local_Extension","number":"1002","applications":[{"app_name":"set","app_data":"ringback=${us-ring}"},{"app_name":"set","app_data":"call_timeout=30"},{"app_name":"bridge","app_data":"user/${destination_number}@${domain_name}"}]},"caller_profile":{"username":"1001","dialplan":"XML","caller_id_name":"1001","ani":"1001","aniii":"","caller_id_number":"1001","network_addr":"192.168.56.1","rdnis":"1002","destination_number":"1002","uuid":"5282f87c-368c-43a2-babd-bc63ff12dc79","source":"mod_sofia","context":"default","chan_name":"sofia/internal/1001@192.168.56.202","originatee":{"originatee_caller_profiles":[{"username":"1001","dialplan":"XML","caller_id_name":"Extension 1001","ani":"1001","aniii":"","caller_id_number":"1001","network_addr":"192.168.56.1","rdnis":"1002","destination_number":"1002","uuid":"fa3173aa-3c1a-41e5-ab82-60623396fa08","source":"mod_sofia","context":"default","chan_name":"sofia/internal/1002@192.168.56.1:5060"},{"username":"1001","dialplan":"XML","caller_id_name":"Extension 1001","ani":"1001","aniii":"","caller_id_number":"1001","network_addr":"192.168.56.1","rdnis":"1002","destination_number":"1002","uuid":"fa3173aa-3c1a-41e5-ab82-60623396fa08","source":"mod_sofia","context":"default","chan_name":"sofia/internal/1002@192.168.56.1:5060"}]}},"times":{"created_time":"1520944103557418","profile_created_time":"1520944103617274","progress_time":"1520944103637959","progress_media_time":"1520944105037574","answered_time":"1520944105037574","bridged_time":"1520944105037574","last_hold_time":"1520944105098164","hold_accum_time":"2208292","hangup_time":"1520944164057232","resurrect_time":"0","transfer_time":"0"}},{"dialplan":"XML","profile_index":"1","extension":{"name":"CGRateS_Auth","number":"1002","applications":[{"app_name":"info","app_data":""},{"app_name":"park","app_data":""}]},"caller_profile":{"username":"1001","dialplan":"XML","caller_id_name":"1001","ani":"1001","aniii":"","caller_id_number":"1001","network_addr":"192.168.56.1","rdnis":"","destination_number":"1002","uuid":"5282f87c-368c-43a2-babd-bc63ff12dc79","source":"mod_sofia","context":"default","chan_name":"sofia/internal/1001@192.168.56.202"},"times":{"created_time":"1520944103557418","profile_created_time":"1520944103557418","progress_time":"0","progress_media_time":"0","answered_time":"0","bridged_time":"0","last_hold_time":"0","hold_accum_time":"0","hangup_time":"0","resurrect_time":"0","transfer_time":"1520944103617274"}}]}
		`)
	hangUp := `Event-Name: CHANNEL_HANGUP_COMPLETE
Core-UUID: 61b38010-fb8e-4022-89b0-3c0be6e500b9
FreeSWITCH-Hostname: teo
FreeSWITCH-Switchname: teo
FreeSWITCH-IPv4: 10.0.2.15
FreeSWITCH-IPv6: %3A%3A1
Event-Date-Local: 2018-03-13%2011%3A17%3A44
Event-Date-GMT: Tue,%2013%20Mar%202018%2015%3A17%3A44%20GMT
Event-Date-Timestamp: 1520954264757374
Event-Calling-File: switch_core_state_machine.c
Event-Calling-Function: switch_core_session_reporting_state
Event-Calling-Line-Number: 949
Event-Sequence: 2920
Hangup-Cause: MANAGER_REQUEST
Channel-State: CS_REPORTING
Channel-Call-State: HANGUP
Channel-State-Number: 11
Channel-Name: sofia/internal/1001%40192.168.56.202
Unique-ID: a511b320-18a2-425f-b574-81a088fcd70a
Call-Direction: inbound
Presence-Call-Direction: inbound
Channel-HIT-Dialplan: true
Channel-Presence-ID: 1001%40192.168.56.202
Channel-Call-UUID: a511b320-18a2-425f-b574-81a088fcd70a
Answer-State: hangup
Hangup-Cause: MANAGER_REQUEST
Channel-Read-Codec-Name: G722
Channel-Read-Codec-Rate: 16000
Channel-Read-Codec-Bit-Rate: 64000
Channel-Write-Codec-Name: G722
Channel-Write-Codec-Rate: 16000
Channel-Write-Codec-Bit-Rate: 64000
Caller-Direction: inbound
Caller-Logical-Direction: inbound
Caller-Username: 1001
Caller-Dialplan: XML
Caller-Caller-ID-Name: 1001
Caller-Caller-ID-Number: 1001
Caller-Orig-Caller-ID-Name: 1001
Caller-Orig-Caller-ID-Number: 1001
Caller-Callee-ID-Name: Outbound%20Call
Caller-Callee-ID-Number: 1002
Caller-Network-Addr: 192.168.56.1
Caller-ANI: 1001
Caller-Destination-Number: 1002
Caller-Unique-ID: a511b320-18a2-425f-b574-81a088fcd70a
Caller-Source: mod_sofia
Caller-Transfer-Source: 1520954259%3Af6338829-f2e3-4874-8841-6baf5c7962b2%3Abl_xfer%3A1002/default/XML
Caller-Context: default
Caller-RDNIS: 1002
Caller-Channel-Name: sofia/internal/1001%40192.168.56.202
Caller-Profile-Index: 2
Caller-Profile-Created-Time: 1520954259237372
Caller-Channel-Created-Time: 1520954259207931
Caller-Channel-Answered-Time: 1520954264358500
Caller-Channel-Progress-Time: 1520954259257731
Caller-Channel-Progress-Media-Time: 1520954264358500
Caller-Channel-Hangup-Time: 1520954264737400
Caller-Channel-Transfer-Time: 0
Caller-Channel-Resurrect-Time: 0
Caller-Channel-Bridged-Time: 1520954264358500
Caller-Channel-Last-Hold: 0
Caller-Channel-Hold-Accum: 0
Caller-Screen-Bit: true
Caller-Privacy-Hide-Name: false
Caller-Privacy-Hide-Number: false
Other-Type: originatee
Other-Leg-Direction: outbound
Other-Leg-Logical-Direction: inbound
Other-Leg-Username: 1001
Other-Leg-Dialplan: XML
Other-Leg-Caller-ID-Name: Extension%201001
Other-Leg-Caller-ID-Number: 1001
Other-Leg-Orig-Caller-ID-Name: 1001
Other-Leg-Orig-Caller-ID-Number: 1001
Other-Leg-Callee-ID-Name: Outbound%20Call
Other-Leg-Callee-ID-Number: 1002
Other-Leg-Network-Addr: 192.168.56.1
Other-Leg-ANI: 1001
Other-Leg-Destination-Number: 1002
Other-Leg-Unique-ID: b64a5ab7-3572-4c82-88ea-1c39f8157023
Other-Leg-Source: mod_sofia
Other-Leg-Context: default
Other-Leg-RDNIS: 1002
Other-Leg-Channel-Name: sofia/internal/1002%40192.168.56.1%3A5060
Other-Leg-Profile-Created-Time: 1520954259237372
Other-Leg-Channel-Created-Time: 1520954259237372
Other-Leg-Channel-Answered-Time: 1520954264345966
Other-Leg-Channel-Progress-Time: 1520954259257731
Other-Leg-Channel-Progress-Media-Time: 0
Other-Leg-Channel-Hangup-Time: 0
Other-Leg-Channel-Transfer-Time: 0
Other-Leg-Channel-Resurrect-Time: 0
Other-Leg-Channel-Bridged-Time: 0
Other-Leg-Channel-Last-Hold: 0
Other-Leg-Channel-Hold-Accum: 0
Other-Leg-Screen-Bit: true
Other-Leg-Privacy-Hide-Name: false
Other-Leg-Privacy-Hide-Number: false
variable_direction: inbound
variable_uuid: a511b320-18a2-425f-b574-81a088fcd70a
variable_session_id: 9
variable_sip_from_user: 1001
variable_sip_from_uri: 1001%40192.168.56.202
variable_sip_from_host: 192.168.56.202
variable_channel_name: sofia/internal/1001%40192.168.56.202
variable_ep_codec_string: mod_spandsp.G722%408000h%4020i%4064000b,CORE_PCM_MODULE.PCMU%408000h%4020i%4064000b,CORE_PCM_MODULE.PCMA%408000h%4020i%4064000b,mod_spandsp.GSM%408000h%4020i%4013200b
variable_sip_local_network_addr: 192.168.56.202
variable_sip_network_ip: 192.168.56.1
variable_sip_network_port: 5060
variable_sip_invite_stamp: 1520954259207931
variable_sip_received_ip: 192.168.56.1
variable_sip_received_port: 5060
variable_sip_via_protocol: udp
variable_sip_authorized: true
variable_Event-Name: REQUEST_PARAMS
variable_Core-UUID: 61b38010-fb8e-4022-89b0-3c0be6e500b9
variable_FreeSWITCH-Hostname: teo
variable_FreeSWITCH-Switchname: teo
variable_FreeSWITCH-IPv4: 10.0.2.15
variable_FreeSWITCH-IPv6: %3A%3A1
variable_Event-Date-Local: 2018-03-13%2011%3A17%3A39
variable_Event-Date-GMT: Tue,%2013%20Mar%202018%2015%3A17%3A39%20GMT
variable_Event-Date-Timestamp: 1520954259207931
variable_Event-Calling-File: sofia.c
variable_Event-Calling-Function: sofia_handle_sip_i_invite
variable_Event-Calling-Line-Number: 10096
variable_Event-Sequence: 2835
variable_sip_number_alias: 1001
variable_sip_auth_username: 1001
variable_sip_auth_realm: 192.168.56.202
variable_number_alias: 1001
variable_requested_user_name: 1001
variable_requested_domain_name: 192.168.56.202
variable_record_stereo: true
variable_transfer_fallback_extension: operator
variable_toll_allow: domestic,international,local
variable_accountcode: 1001
variable_user_context: default
variable_effective_caller_id_name: Extension%201001
variable_effective_caller_id_number: 1001
variable_callgroup: techsupport
variable_cgr_reqtype: *prepaid
variable_cgr_subsystems: *resources%3B*attributes%3B*sessions%3B*suppliers%3B*thresholds%3B*stats%3B*accounts
variable_cgr_acd: 30
variable_user_name: 1001
variable_domain_name: 192.168.56.202
variable_sip_from_user_stripped: 1001
variable_sofia_profile_name: internal
variable_recovery_profile_name: internal
variable_sip_req_user: 1002
variable_sip_req_uri: 1002%40192.168.56.202
variable_sip_req_host: 192.168.56.202
variable_sip_to_user: 1002
variable_sip_to_uri: 1002%40192.168.56.202
variable_sip_to_host: 192.168.56.202
variable_sip_contact_params: transport%3Dudp%3Bregistering_acc%3D192_168_56_202
variable_sip_contact_user: 1001
variable_sip_contact_port: 5060
variable_sip_contact_uri: 1001%40192.168.56.1%3A5060
variable_sip_contact_host: 192.168.56.1
variable_sip_user_agent: Jitsi2.10.5550Windows%2010
variable_sip_via_host: 192.168.56.1
variable_sip_via_port: 5060
variable_presence_id: 1001%40192.168.56.202
variable_switch_r_sdp: v%3D0%0D%0Ao%3D1001-jitsi.org%200%200%20IN%20IP4%20192.168.56.1%0D%0As%3D-%0D%0Ac%3DIN%20IP4%20192.168.56.1%0D%0At%3D0%200%0D%0Am%3Daudio%205116%20RTP/AVP%2096%2097%2098%209%20100%20102%200%208%20103%203%20104%204%20101%0D%0Aa%3Drtpmap%3A96%20opus/48000/2%0D%0Aa%3Dfmtp%3A96%20usedtx%3D1%0D%0Aa%3Drtpmap%3A97%20SILK/24000%0D%0Aa%3Drtpmap%3A98%20SILK/16000%0D%0Aa%3Drtpmap%3A9%20G722/8000%0D%0Aa%3Drtpmap%3A100%20speex/32000%0D%0Aa%3Drtpmap%3A102%20speex/16000%0D%0Aa%3Drtpmap%3A0%20PCMU/8000%0D%0Aa%3Drtpmap%3A8%20PCMA/8000%0D%0Aa%3Drtpmap%3A103%20iLBC/8000%0D%0Aa%3Drtpmap%3A3%20GSM/8000%0D%0Aa%3Drtpmap%3A104%20speex/8000%0D%0Aa%3Drtpmap%3A4%20G723/8000%0D%0Aa%3Dfmtp%3A4%20annexa%3Dno%3Bbitrate%3D6.3%0D%0Aa%3Drtpmap%3A101%20telephone-event/8000%0D%0Aa%3Dptime%3A20%0D%0Aa%3Dextmap%3A1%20urn%3Aietf%3Aparams%3Artp-hdrext%3Acsrc-audio-level%0D%0Aa%3Dextmap%3A2%20urn%3Aietf%3Aparams%3Artp-hdrext%3Assrc-audio-level%0D%0Aa%3Drtcp-xr%3Avoip-metrics%0D%0Am%3Dvideo%205118%20RTP/AVP%20105%2099%0D%0Aa%3Drtpmap%3A105%20H264/90000%0D%0Aa%3Dfmtp%3A105%20profile-level-id%3D4DE01f%3Bpacketization-mode%3D1%0D%0Aa%3Drtpmap%3A99%20H264/90000%0D%0Aa%3Dfmtp%3A99%20profile-level-id%3D4DE01f%0D%0Aa%3Drecvonly%0D%0Aa%3Dimageattr%3A105%20send%20*%20recv%20%5Bx%3D%5B1%3A1920%5D,y%3D%5B1%3A1080%5D%5D%0D%0Aa%3Dimageattr%3A99%20send%20*%20recv%20%5Bx%3D%5B1%3A1920%5D,y%3D%5B1%3A1080%5D%5D%0D%0A
variable_max_forwards: 69
variable_transfer_history: 1520954259%3Af6338829-f2e3-4874-8841-6baf5c7962b2%3Abl_xfer%3A1002/default/XML
variable_transfer_source: 1520954259%3Af6338829-f2e3-4874-8841-6baf5c7962b2%3Abl_xfer%3A1002/default/XML
variable_DP_MATCH: ARRAY%3A%3A1002%7C%3A1002
variable_call_uuid: a511b320-18a2-425f-b574-81a088fcd70a
variable_call_timeout: 30
variable_dialed_user: 1002
variable_dialed_domain: 192.168.56.202
variable_originated_legs: ARRAY%3A%3Ab64a5ab7-3572-4c82-88ea-1c39f8157023%3BOutbound%20Call%3B1002%7C%3Ab64a5ab7-3572-4c82-88ea-1c39f8157023%3BOutbound%20Call%3B1002
variable_rtp_use_codec_string: G722,PCMU,PCMA,GSM
variable_audio_media_flow: sendrecv
variable_rtp_audio_recv_pt: 9
variable_rtp_use_codec_name: G722
variable_rtp_use_codec_rate: 8000
variable_rtp_use_codec_ptime: 20
variable_rtp_use_codec_channels: 1
variable_rtp_last_audio_codec_string: G722%408000h%4020i%401c
variable_read_codec: G722
variable_original_read_codec: G722
variable_read_rate: 16000
variable_original_read_rate: 16000
variable_write_codec: G722
variable_write_rate: 16000
variable_dtmf_type: rfc2833
variable_video_possible: true
variable_video_media_flow: sendonly
variable_local_media_ip: 192.168.56.202
variable_local_media_port: 17842
variable_advertised_media_ip: 192.168.56.202
variable_rtp_use_timer_name: soft
variable_rtp_use_pt: 9
variable_rtp_use_ssrc: 850152579
variable_rtp_2833_send_payload: 101
variable_rtp_2833_recv_payload: 101
variable_remote_media_ip: 192.168.56.1
variable_remote_media_port: 5116
variable_rtp_local_sdp_str: v%3D0%0D%0Ao%3DFreeSWITCH%201520936422%201520936423%20IN%20IP4%20192.168.56.202%0D%0As%3DFreeSWITCH%0D%0Ac%3DIN%20IP4%20192.168.56.202%0D%0At%3D0%200%0D%0Am%3Daudio%2017842%20RTP/AVP%209%20101%0D%0Aa%3Drtpmap%3A9%20G722/8000%0D%0Aa%3Drtpmap%3A101%20telephone-event/8000%0D%0Aa%3Dfmtp%3A101%200-16%0D%0Aa%3Dptime%3A20%0D%0Aa%3Dsendrecv%0D%0Am%3Dvideo%200%20RTP/AVP%2019%0D%0A
variable_endpoint_disposition: ANSWER
variable_originate_causes: ARRAY%3A%3Ab64a5ab7-3572-4c82-88ea-1c39f8157023%3BNONE%7C%3Ab64a5ab7-3572-4c82-88ea-1c39f8157023%3BNONE
variable_originate_disposition: SUCCESS
variable_DIALSTATUS: SUCCESS
variable_last_bridge_to: b64a5ab7-3572-4c82-88ea-1c39f8157023
variable_bridge_channel: sofia/internal/1002%40192.168.56.1%3A5060
variable_bridge_uuid: b64a5ab7-3572-4c82-88ea-1c39f8157023
variable_signal_bond: b64a5ab7-3572-4c82-88ea-1c39f8157023
variable_sip_to_tag: K61NU1g84rUQB
variable_sip_from_tag: e806c8da
variable_sip_cseq: 2
variable_sip_call_id: 6cd120fa2b430439e6bceb81be52f3ef%400%3A0%3A0%3A0%3A0%3A0%3A0%3A0
variable_sip_full_via: SIP/2.0/UDP%20192.168.56.1%3A5060%3Bbranch%3Dz9hG4bK-333939-6be85d43d3e8db68310b028366c28d4a
variable_sip_from_display: 1001
variable_sip_full_from: %221001%22%20%3Csip%3A1001%40192.168.56.202%3E%3Btag%3De806c8da
variable_sip_full_to: %3Csip%3A1002%40192.168.56.202%3E%3Btag%3DK61NU1g84rUQB
variable_cgr_notify: SERVER_ERROR
variable_last_sent_callee_id_name: Outbound%20Call
variable_last_sent_callee_id_number: 1002
variable_switch_m_sdp: v%3D0%0D%0Ao%3D1002-jitsi.org%200%201%20IN%20IP4%20192.168.56.1%0D%0As%3D-%0D%0Ac%3DIN%20IP4%20192.168.56.1%0D%0At%3D0%200%0D%0Am%3Daudio%205120%20RTP/AVP%2096%2097%2098%209%20100%20102%200%208%20103%203%20104%204%20101%0D%0Aa%3Drtpmap%3A96%20opus/48000/2%0D%0Aa%3Dfmtp%3A96%20usedtx%3D1%0D%0Aa%3Drtpmap%3A97%20SILK/24000%0D%0Aa%3Drtpmap%3A98%20SILK/16000%0D%0Aa%3Drtpmap%3A9%20G722/8000%0D%0Aa%3Drtpmap%3A100%20speex/32000%0D%0Aa%3Drtpmap%3A102%20speex/16000%0D%0Aa%3Drtpmap%3A0%20PCMU/8000%0D%0Aa%3Drtpmap%3A8%20PCMA/8000%0D%0Aa%3Drtpmap%3A103%20iLBC/8000%0D%0Aa%3Drtpmap%3A3%20GSM/8000%0D%0Aa%3Drtpmap%3A104%20speex/8000%0D%0Aa%3Drtpmap%3A4%20G723/8000%0D%0Aa%3Dfmtp%3A4%20annexa%3Dno%3Bbitrate%3D6.3%0D%0Aa%3Drtpmap%3A101%20telephone-event/8000%0D%0Aa%3Dsendonly%0D%0Aa%3Dptime%3A20%0D%0Aa%3Dextmap%3A1%20urn%3Aietf%3Aparams%3Artp-hdrext%3Acsrc-audio-level%0D%0Aa%3Dextmap%3A2%20urn%3Aietf%3Aparams%3Artp-hdrext%3Assrc-audio-level%0D%0Aa%3Drtcp-xr%3Avoip-metrics%0D%0Aa%3Dzrtp-hash%3A1.10%200128e1ef0c736175faa6fd8e329b6b04e1ba19c0de8701869b24cc29ee22533a%0D%0A
variable_l_sdp_audio_zrtp_hash: 1.10%200128e1ef0c736175faa6fd8e329b6b04e1ba19c0de8701869b24cc29ee22533a
variable_current_application_data: local_stream%3A//moh
variable_current_application: playback
variable_current_application_response: FILE%20NOT%20FOUND
variable_last_bridge_role: originator
variable_bridge_hangup_cause: NORMAL_CLEARING
variable_hangup_cause: MANAGER_REQUEST
variable_hangup_cause_q850: 16
variable_digits_dialed: none
variable_start_stamp: 2018-03-13%2011%3A17%3A39
variable_profile_start_stamp: 2018-03-13%2011%3A17%3A39
variable_answer_stamp: 2018-03-13%2011%3A17%3A44
variable_bridge_stamp: 2018-03-13%2011%3A17%3A44
variable_progress_stamp: 2018-03-13%2011%3A17%3A39
variable_progress_media_stamp: 2018-03-13%2011%3A17%3A44
variable_end_stamp: 2018-03-13%2011%3A17%3A44
variable_start_epoch: 1520954259
variable_start_uepoch: 1520954259207931
variable_profile_start_epoch: 1520954259
variable_profile_start_uepoch: 1520954259237372
variable_answer_epoch: 1520954264
variable_answer_uepoch: 1520954264358500
variable_bridge_epoch: 1520954264
variable_bridge_uepoch: 1520954264358500
variable_last_hold_epoch: 0
variable_last_hold_uepoch: 0
variable_hold_accum_seconds: 0
variable_hold_accum_usec: 0
variable_hold_accum_ms: 0
variable_resurrect_epoch: 0
variable_resurrect_uepoch: 0
variable_progress_epoch: 1520954259
variable_progress_uepoch: 1520954259257731
variable_progress_media_epoch: 1520954264
variable_progress_media_uepoch: 1520954264358500
variable_end_epoch: 1520954264
variable_end_uepoch: 1520954264737400
variable_last_app: playback
variable_last_arg: local_stream%3A//moh
variable_caller_id: %221001%22%20%3C1001%3E
variable_duration: 5
variable_billsec: 0
variable_progresssec: 0
variable_answersec: 5
variable_waitsec: 5
variable_progress_mediasec: 5
variable_flow_billsec: 5
variable_mduration: 5530
variable_billmsec: 379
variable_progressmsec: 50
variable_answermsec: 5151
variable_waitmsec: 5151
variable_progress_mediamsec: 5151
variable_flow_billmsec: 5530
variable_uduration: 5529469
variable_billusec: 378900
variable_progressusec: 49800
variable_answerusec: 5150569
variable_waitusec: 5150569
variable_progress_mediausec: 5150569
variable_flow_billusec: 5529469
variable_sip_hangup_disposition: send_bye
variable_rtp_audio_in_raw_bytes: 1376
variable_rtp_audio_in_media_bytes: 1376
variable_rtp_audio_in_packet_count: 8
variable_rtp_audio_in_media_packet_count: 8
variable_rtp_audio_in_skip_packet_count: 13
variable_rtp_audio_in_jitter_packet_count: 0
variable_rtp_audio_in_dtmf_packet_count: 0
variable_rtp_audio_in_cng_packet_count: 0
variable_rtp_audio_in_flush_packet_count: 0
variable_rtp_audio_in_largest_jb_size: 0
variable_rtp_audio_in_jitter_min_variance: 0.00
variable_rtp_audio_in_jitter_max_variance: 0.00
variable_rtp_audio_in_jitter_loss_rate: 0.00
variable_rtp_audio_in_jitter_burst_rate: 0.00
variable_rtp_audio_in_mean_interval: 0.00
variable_rtp_audio_in_flaw_total: 0
variable_rtp_audio_in_quality_percentage: 100.00
variable_rtp_audio_in_mos: 4.50
variable_rtp_audio_out_raw_bytes: 0
variable_rtp_audio_out_media_bytes: 0
variable_rtp_audio_out_packet_count: 0
variable_rtp_audio_out_media_packet_count: 0
variable_rtp_audio_out_skip_packet_count: 0
variable_rtp_audio_out_dtmf_packet_count: 0
variable_rtp_audio_out_cng_packet_count: 0
variable_rtp_audio_rtcp_packet_count: 0
variable_rtp_audio_rtcp_octet_count: 0`

	var fsCdrCfg *config.CGRConfig
	timezone := config.CgrConfig().DefaultTimezone
	fsCdrCfg, _ = config.NewDefaultCGRConfig()
	fsCdr, _ := engine.NewFSCdr(body, fsCdrCfg)
	smGev := sessions.SMGenericEvent(NewFSEvent(hangUp).AsMapStringInterface(timezone))
	valCGRID := smGev.AsCDR(fsCdrCfg, timezone)
	// convertit in SMGEvent
	// pe urma comapart fsCDR.CGRID cu
	value := fsCdr.AsCDR(timezone)
	if value.CGRID != valCGRID.CGRID {
		t.Errorf("Expecting: %s, received: %s", valCGRID.CGRID, value.CGRID)
	}
}

func TestFsEvV1AuthorizeArgs(t *testing.T) {
	timezone := config.CgrConfig().DefaultTimezone
	ev := NewFSEvent(hangupEv)
	sTime, err := ev.GetSetupTime(utils.META_DEFAULT, timezone)
	if err != nil {
		t.Error(err)
	}
	expected := &sessions.V1AuthorizeArgs{
		GetMaxUsage: true,
		CGREvent: utils.CGREvent{
			Tenant: ev.GetTenant(utils.META_DEFAULT),
			ID:     utils.UUIDSha1Prefix(),
			Time:   &sTime,
			Event:  ev.AsMapStringInterface(timezone),
		},
	}
	rcv := ev.V1AuthorizeArgs()
	if !reflect.DeepEqual(expected.CGREvent.Tenant, rcv.CGREvent.Tenant) {
		t.Errorf("Expecting: %+v, received: %+v", expected.CGREvent.Tenant, rcv.CGREvent.Tenant)
	} else if !reflect.DeepEqual(expected.CGREvent.Time, rcv.CGREvent.Time) {
		t.Errorf("Expecting: %+v, received: %+v", expected.CGREvent.Time, rcv.CGREvent.Time)
	} else if !reflect.DeepEqual(expected.CGREvent.Event, rcv.CGREvent.Event) {
		t.Errorf("Expecting: %+v, received: %+v", expected.CGREvent.Event, rcv.CGREvent.Event)
	} else if !reflect.DeepEqual(expected.CGREvent.Event, rcv.CGREvent.Event) {
		t.Errorf("Expecting: %+v, received: %+v", expected.CGREvent.Event, rcv.CGREvent.Event)
	} else if !reflect.DeepEqual(expected.GetMaxUsage, rcv.GetMaxUsage) {
		t.Errorf("Expecting: %+v, received: %+v", expected.GetMaxUsage, rcv.GetMaxUsage)
	} else if !reflect.DeepEqual(expected.GetSuppliers, rcv.GetSuppliers) {
		t.Errorf("Expecting: %+v, received: %+v", expected.GetSuppliers, rcv.GetSuppliers)
	} else if !reflect.DeepEqual(expected.GetAttributes, rcv.GetAttributes) {
		t.Errorf("Expecting: %+v, received: %+v", expected.GetAttributes, rcv.GetAttributes)
	}
}

func TestFsEvV1InitSessionArgs(t *testing.T) {
	timezone := config.CgrConfig().DefaultTimezone
	ev := NewFSEvent(hangupEv)
	sTime, err := ev.GetSetupTime(utils.META_DEFAULT, timezone)
	if err != nil {
		t.Error(err)
	}
	expected := &sessions.V1InitSessionArgs{
		InitSession: true,
		CGREvent: utils.CGREvent{
			Tenant: ev.GetTenant(utils.META_DEFAULT),
			ID:     utils.UUIDSha1Prefix(),
			Time:   &sTime,
			Event:  ev.AsMapStringInterface(timezone),
		},
	}
	rcv := ev.V1InitSessionArgs()
	if !reflect.DeepEqual(expected.CGREvent.Tenant, rcv.CGREvent.Tenant) {
		t.Errorf("Expecting: %+v, received: %+v", expected.CGREvent.Tenant, rcv.CGREvent.Tenant)
	} else if !reflect.DeepEqual(expected.CGREvent.Time, rcv.CGREvent.Time) {
		t.Errorf("Expecting: %+v, received: %+v", expected.CGREvent.Time, rcv.CGREvent.Time)
	} else if !reflect.DeepEqual(expected.CGREvent.Event, rcv.CGREvent.Event) {
		t.Errorf("Expecting: %+v, received: %+v", expected.CGREvent.Event, rcv.CGREvent.Event)
	} else if !reflect.DeepEqual(expected.CGREvent.Event, rcv.CGREvent.Event) {
		t.Errorf("Expecting: %+v, received: %+v", expected.CGREvent.Event, rcv.CGREvent.Event)
	} else if !reflect.DeepEqual(expected.InitSession, rcv.InitSession) {
		t.Errorf("Expecting: %+v, received: %+v", expected.InitSession, rcv.InitSession)
	}
}

func TestFsEvV1TerminateSessionArgs(t *testing.T) {
	timezone := config.CgrConfig().DefaultTimezone
	ev := NewFSEvent(hangupEv)
	sTime, err := ev.GetSetupTime(utils.META_DEFAULT, timezone)
	if err != nil {
		t.Error(err)
	}
	expected := &sessions.V1TerminateSessionArgs{
		TerminateSession: true,
		CGREvent: utils.CGREvent{
			Tenant: ev.GetTenant(utils.META_DEFAULT),
			ID:     utils.UUIDSha1Prefix(),
			Time:   &sTime,
			Event:  ev.AsMapStringInterface(timezone),
		},
	}
	rcv := ev.V1TerminateSessionArgs()
	if !reflect.DeepEqual(expected.CGREvent.Tenant, rcv.CGREvent.Tenant) {
		t.Errorf("Expecting: %+v, received: %+v", expected.CGREvent.Tenant, rcv.CGREvent.Tenant)
	} else if !reflect.DeepEqual(expected.CGREvent.Time, rcv.CGREvent.Time) {
		t.Errorf("Expecting: %+v, received: %+v", expected.CGREvent.Time, rcv.CGREvent.Time)
	} else if !reflect.DeepEqual(expected.CGREvent.Event, rcv.CGREvent.Event) {
		t.Errorf("Expecting: %+v, received: %+v", expected.CGREvent.Event, rcv.CGREvent.Event)
	} else if !reflect.DeepEqual(expected.CGREvent.Event, rcv.CGREvent.Event) {
		t.Errorf("Expecting: %+v, received: %+v", expected.CGREvent.Event, rcv.CGREvent.Event)
	} else if !reflect.DeepEqual(expected.TerminateSession, rcv.TerminateSession) {
		t.Errorf("Expecting: %+v, received: %+v", expected.TerminateSession, rcv.TerminateSession)
	}
}
