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
variable_rtp_audio_rtcp_octet_count: 45940
variable_cgr_subsystems: *resources%3B*attributes%3B*sessions%3B*suppliers_event_cost%3B*suppliers_ignore_errors%3B*accounts`

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
	if tor, _ := ev.ParseEventValue(config.NewRSRParserMustCompile(utils.REGEXP_PREFIX+utils.ToR, true), ""); tor != utils.VOICE {
		t.Error("Unexpected tor parsed", tor)
	}
	if accid, _ := ev.ParseEventValue(config.NewRSRParserMustCompile(utils.REGEXP_PREFIX+utils.OriginID, true), ""); accid != "e3133bf7-dcde-4daf-9663-9a79ffcef5ad" {
		t.Error("Unexpected result parsed", accid)
	}
	if parsed, _ := ev.ParseEventValue(config.NewRSRParserMustCompile(utils.REGEXP_PREFIX+utils.OriginHost, true), ""); parsed != "10.0.3.15" {
		t.Error("Unexpected result parsed", parsed)
	}
	if parsed, _ := ev.ParseEventValue(config.NewRSRParserMustCompile(utils.REGEXP_PREFIX+utils.Source, true), ""); parsed != "FS_EVENT" {
		t.Error("Unexpected result parsed", parsed)
	}
	if parsed, _ := ev.ParseEventValue(config.NewRSRParserMustCompile(utils.REGEXP_PREFIX+utils.RequestType, true), ""); parsed != utils.META_PREPAID {
		t.Error("Unexpected result parsed", parsed)
	}
	if parsed, _ := ev.ParseEventValue(config.NewRSRParserMustCompile(utils.REGEXP_PREFIX+utils.Direction, true), ""); parsed != utils.OUT {
		t.Error("Unexpected result parsed", parsed)
	}
	if parsed, _ := ev.ParseEventValue(config.NewRSRParserMustCompile(utils.REGEXP_PREFIX+utils.Tenant, true), ""); parsed != "cgrates.org" {
		t.Error("Unexpected result parsed", parsed)
	}
	if parsed, _ := ev.ParseEventValue(config.NewRSRParserMustCompile(utils.REGEXP_PREFIX+utils.Category, true), ""); parsed != "call" {
		t.Error("Unexpected result parsed", parsed)
	}
	if parsed, _ := ev.ParseEventValue(config.NewRSRParserMustCompile(utils.REGEXP_PREFIX+utils.Account, true), ""); parsed != "1001" {
		t.Error("Unexpected result parsed", parsed)
	}
	if parsed, _ := ev.ParseEventValue(config.NewRSRParserMustCompile(utils.REGEXP_PREFIX+utils.Subject, true), ""); parsed != "1001" {
		t.Error("Unexpected result parsed", parsed)
	}
	if parsed, _ := ev.ParseEventValue(config.NewRSRParserMustCompile(utils.REGEXP_PREFIX+utils.Destination, true), ""); parsed != "1003" {
		t.Error("Unexpected result parsed", parsed)
	}
	sTime, _ := utils.ParseTimeDetectLayout("1436280728471153"[:len("1436280728471153")-6], "") // We discard nanoseconds information so we can correlate csv
	if parsed, _ := ev.ParseEventValue(config.NewRSRParserMustCompile(utils.REGEXP_PREFIX+utils.SetupTime, true), ""); parsed != sTime.String() {
		t.Errorf("Expecting: %s, parsed: %s", sTime.String(), parsed)
	}
	aTime, _ := utils.ParseTimeDetectLayout("1436280728971147"[:len("1436280728971147")-6], "")
	if parsed, _ := ev.ParseEventValue(config.NewRSRParserMustCompile(utils.REGEXP_PREFIX+utils.AnswerTime, true), ""); parsed != aTime.String() {
		t.Errorf("Expecting: %s, parsed: %s", aTime.String(), parsed)
	}
	if parsed, _ := ev.ParseEventValue(config.NewRSRParserMustCompile(utils.REGEXP_PREFIX+utils.Usage, true), ""); parsed != "66000000000" {
		t.Error("Unexpected result parsed", parsed)
	}
	if parsed, _ := ev.ParseEventValue(config.NewRSRParserMustCompile(utils.REGEXP_PREFIX+utils.PDD, true), ""); parsed != "0.028" {
		t.Error("Unexpected result parsed", parsed)
	}
	if parsed, _ := ev.ParseEventValue(config.NewRSRParserMustCompile(utils.REGEXP_PREFIX+utils.SUPPLIER, true), ""); parsed != "supplier1" {
		t.Error("Unexpected result parsed", parsed)
	}
	if parsed, _ := ev.ParseEventValue(config.NewRSRParserMustCompile(utils.REGEXP_PREFIX+utils.RunID, true), ""); parsed != utils.DEFAULT_RUNID {
		t.Error("Unexpected result parsed", parsed)
	}
	if parsed, _ := ev.ParseEventValue(config.NewRSRParserMustCompile(utils.REGEXP_PREFIX+utils.COST, true), ""); parsed != "-1" {
		t.Error("Unexpected result parsed", parsed)
	}
	if parsed, _ := ev.ParseEventValue(config.NewRSRParserMustCompile(utils.REGEXP_PREFIX+"Hangup-Cause", true), ""); parsed != "NORMAL_CLEARING" {
		t.Error("Unexpected result parsed", parsed)
	}
}

func TestFsEvAsCGREvent(t *testing.T) {
	timezone := config.CgrConfig().GeneralCfg().DefaultTimezone
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
	expectedMap[utils.ToR] = utils.VOICE
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
	expectedMap[utils.Cost] = -1.0
	expectedMap[utils.PDD] = time.Duration(28) * time.Millisecond
	expectedMap[utils.ACD] = time.Duration(30) * time.Second
	expectedMap[utils.DISCONNECT_CAUSE] = "NORMAL_CLEARING"
	expectedMap[utils.SUPPLIER] = "supplier1"
	if storedMap := ev.AsMapStringInterface(""); !reflect.DeepEqual(expectedMap, storedMap) {
		t.Errorf("Expecting: %s, received: %s", utils.ToJSON(expectedMap), utils.ToJSON(storedMap))
	}
}

func TestFsEvGetExtraFields(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	var err error
	err = nil
	cfg.FsAgentCfg().ExtraFields, err = config.NewRSRParsersFromSlice([]string{
		"~Channel-Read-Codec-Name",
		"~Channel-Write-Codec-Name",
		"~NonExistingHeader",
	}, true)
	if err != nil {
		t.Error(err)
	}
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
{"core-uuid":"63e2315b-d538-4dfa-9ed5-af73ba6210b6","switchname":"teo","channel_data":{"state":"CS_REPORTING","direction":"inbound","state_number":"11","flags":"0=1;1=1;3=1;20=1;37=1;38=1;40=1;43=1;48=1;53=1;75=1;77=1;106=1;112=1;113=1;122=1;134=1","caps":"1=1;2=1;3=1;4=1;5=1;6=1"},"callStats":{"audio":{"inbound":{"raw_bytes":174156,"media_bytes":166416,"packet_count":1033,"media_packet_count":988,"skip_packet_count":7,"jitter_packet_count":0,"dtmf_packet_count":0,"cng_packet_count":0,"flush_packet_count":45,"largest_jb_size":0,"jitter_min_variance":0.500000,"jitter_max_variance":31.769231,"jitter_loss_rate":0,"jitter_burst_rate":0,"mean_interval":20.171779,"flaw_total":1,"quality_percentage":99,"mos":4.492027,"errorLog":[{"start":1521025783725905,"stop":1521025788366141,"flaws":10763,"consecutiveFlaws":0,"durationMS":4640}]},"outbound":{"raw_bytes":43344,"media_bytes":43344,"packet_count":252,"media_packet_count":252,"skip_packet_count":0,"dtmf_packet_count":0,"cng_packet_count":0,"rtcp_packet_count":0,"rtcp_octet_count":0}}},"variables":{"direction":"inbound","uuid":"5a3a1d91-90d3-4db4-af5c-cc3ae15d93a4","session_id":"1","sip_from_user":"1001","sip_from_uri":"1001@192.168.56.202","sip_from_host":"192.168.56.202","channel_name":"sofia/internal/1001@192.168.56.202","ep_codec_string":"mod_spandsp.G722@8000h@20i@64000b,CORE_PCM_MODULE.PCMU@8000h@20i@64000b,CORE_PCM_MODULE.PCMA@8000h@20i@64000b,mod_spandsp.GSM@8000h@20i@13200b","sip_local_network_addr":"192.168.56.202","sip_network_ip":"192.168.56.1","sip_network_port":"5060","sip_invite_stamp":"1521025758006702","sip_received_ip":"192.168.56.1","sip_received_port":"5060","sip_via_protocol":"udp","sip_authorized":"true","Event-Name":"REQUEST_PARAMS","Core-UUID":"63e2315b-d538-4dfa-9ed5-af73ba6210b6","FreeSWITCH-Hostname":"teo","FreeSWITCH-Switchname":"teo","FreeSWITCH-IPv4":"10.0.2.15","FreeSWITCH-IPv6":"::1","Event-Date-Local":"2018-03-14 07:09:18","Event-Date-GMT":"Wed, 14 Mar 2018 11:09:18 GMT","Event-Date-Timestamp":"1521025758006702","Event-Calling-File":"sofia.c","Event-Calling-Function":"sofia_handle_sip_i_invite","Event-Calling-Line-Number":"10096","Event-Sequence":"1025","sip_number_alias":"1001","sip_auth_username":"1001","sip_auth_realm":"192.168.56.202","number_alias":"1001","requested_user_name":"1001","requested_domain_name":"192.168.56.202","record_stereo":"true","transfer_fallback_extension":"operator","toll_allow":"domestic,international,local","accountcode":"1001","user_context":"default","effective_caller_id_name":"Extension 1001","effective_caller_id_number":"1001","callgroup":"techsupport","cgr_reqtype":"*prepaid","cgr_subsystems":"*resources;*attributes;*sessions;*suppliers","user_name":"1001","domain_name":"192.168.56.202","sip_from_user_stripped":"1001","sofia_profile_name":"internal","recovery_profile_name":"internal","sip_req_user":"1002","sip_req_uri":"1002@192.168.56.202","sip_req_host":"192.168.56.202","sip_to_user":"1002","sip_to_uri":"1002@192.168.56.202","sip_to_host":"192.168.56.202","sip_contact_params":"transport=udp;registering_acc=192_168_56_202","sip_contact_user":"1001","sip_contact_port":"5060","sip_contact_uri":"1001@192.168.56.1:5060","sip_contact_host":"192.168.56.1","sip_via_host":"192.168.56.1","sip_via_port":"5060","presence_id":"1001@192.168.56.202","cgr_resource_allocation":"ResGroup1","cgr_suppliers":"ARRAY::3|:supplier2|:supplier3|:supplier1","cgr_notify":"AUTH_OK","max_forwards":"69","transfer_history":"1521025758:86c9ebb2-888f-42d5-9afa-2101449a4b86:bl_xfer:1002/default/XML","transfer_source":"1521025758:86c9ebb2-888f-42d5-9afa-2101449a4b86:bl_xfer:1002/default/XML","DP_MATCH":"ARRAY::1002|:1002","call_uuid":"5a3a1d91-90d3-4db4-af5c-cc3ae15d93a4","call_timeout":"30","current_application_data":"user/1002@192.168.56.202","current_application":"bridge","dialed_user":"1002","dialed_domain":"192.168.56.202","originated_legs":"ARRAY::9c1afb4f-1d4a-4e45-84a3-d25721981bf5;Outbound Call;1002|:9c1afb4f-1d4a-4e45-84a3-d25721981bf5;Outbound Call;1002","switch_m_sdp":"v=0\r\no=1002-jitsi.org 0 0 IN IP4 192.168.56.1\r\ns=-\r\nc=IN IP4 192.168.56.1\r\nt=0 0\r\nm=audio 5004 RTP/AVP 9 0 8 3 101\r\na=rtpmap:9 G722/8000\r\na=rtpmap:0 PCMU/8000\r\na=rtpmap:8 PCMA/8000\r\na=rtpmap:3 GSM/8000\r\na=rtpmap:101 telephone-event/8000\r\n","rtp_use_codec_name":"G722","rtp_use_codec_rate":"8000","rtp_use_codec_ptime":"20","rtp_use_codec_channels":"1","rtp_last_audio_codec_string":"G722@8000h@20i@1c","read_codec":"G722","original_read_codec":"G722","read_rate":"16000","original_read_rate":"16000","write_codec":"G722","write_rate":"16000","local_media_ip":"192.168.56.202","local_media_port":"29014","advertised_media_ip":"192.168.56.202","rtp_use_timer_name":"soft","rtp_use_pt":"9","rtp_use_ssrc":"2729250253","endpoint_disposition":"ANSWER","originate_causes":"ARRAY::9c1afb4f-1d4a-4e45-84a3-d25721981bf5;NONE|:9c1afb4f-1d4a-4e45-84a3-d25721981bf5;NONE","originate_disposition":"SUCCESS","DIALSTATUS":"SUCCESS","last_bridge_to":"9c1afb4f-1d4a-4e45-84a3-d25721981bf5","bridge_channel":"sofia/internal/1002@192.168.56.1:5060","bridge_uuid":"9c1afb4f-1d4a-4e45-84a3-d25721981bf5","signal_bond":"9c1afb4f-1d4a-4e45-84a3-d25721981bf5","last_sent_callee_id_name":"Outbound Call","last_sent_callee_id_number":"1002","switch_r_sdp":"v=0\r\no=1001-jitsi.org 0 2 IN IP4 192.168.56.1\r\ns=-\r\nc=IN IP4 192.168.56.1\r\nt=0 0\r\nm=audio 5000 RTP/AVP 96 97 98 9 100 102 0 8 103 3 104 4 101\r\na=rtpmap:96 opus/48000/2\r\na=fmtp:96 usedtx=1\r\na=rtpmap:97 SILK/24000\r\na=rtpmap:98 SILK/16000\r\na=rtpmap:9 G722/8000\r\na=rtpmap:100 speex/32000\r\na=rtpmap:102 speex/16000\r\na=rtpmap:0 PCMU/8000\r\na=rtpmap:8 PCMA/8000\r\na=rtpmap:103 iLBC/8000\r\na=rtpmap:3 GSM/8000\r\na=rtpmap:104 speex/8000\r\na=rtpmap:4 G723/8000\r\na=fmtp:4 annexa=no;bitrate=6.3\r\na=rtpmap:101 telephone-event/8000\r\na=ptime:20\r\na=extmap:1 urn:ietf:params:rtp-hdrext:csrc-audio-level\r\na=extmap:2 urn:ietf:params:rtp-hdrext:ssrc-audio-level\r\na=rtcp-xr:voip-metrics\r\na=zrtp-hash:1.10 8e8dd2fa6803f32845f26e55879c776a4bc015ee05b41630313aee27ef77fb30\r\nm=video 5006 RTP/AVP 105 99\r\na=rtpmap:105 H264/90000\r\na=fmtp:105 profile-level-id=4DE01f;packetization-mode=1\r\na=rtpmap:99 H264/90000\r\na=fmtp:99 profile-level-id=4DE01f\r\na=recvonly\r\na=imageattr:105 send * recv [x=[1:1920],y=[1:1080]]\r\na=imageattr:99 send * recv [x=[1:1920],y=[1:1080]]\r\n","rtp_use_codec_string":"G722,PCMU,PCMA,GSM","r_sdp_audio_zrtp_hash":"1.10 8e8dd2fa6803f32845f26e55879c776a4bc015ee05b41630313aee27ef77fb30","audio_media_flow":"sendrecv","remote_media_ip":"192.168.56.1","remote_media_port":"5000","rtp_audio_recv_pt":"9","dtmf_type":"rfc2833","rtp_2833_send_payload":"101","rtp_2833_recv_payload":"101","video_possible":"true","video_media_flow":"sendonly","rtp_local_sdp_str":"v=0\r\no=FreeSWITCH 1520996753 1520996756 IN IP4 192.168.56.202\r\ns=FreeSWITCH\r\nc=IN IP4 192.168.56.202\r\nt=0 0\r\nm=audio 29014 RTP/AVP 9 101\r\na=rtpmap:9 G722/8000\r\na=rtpmap:101 telephone-event/8000\r\na=fmtp:101 0-16\r\na=ptime:20\r\na=sendrecv\r\nm=video 0 RTP/AVP 19\r\n","sip_to_tag":"aDUZXF1Z1vD6p","sip_from_tag":"df94d020","sip_cseq":"4","sip_call_id":"985e365faa0ec79a7fa75d001ef2449f@0:0:0:0:0:0:0:0","sip_full_via":"SIP/2.0/UDP 192.168.56.1:5060;branch=z9hG4bK-323230-ab335b3491dd24f5ec251b9700716b97","sip_from_display":"1001","sip_full_from":"\"1001\" <sip:1001@192.168.56.202>;tag=df94d020","sip_full_to":"<sip:1002@192.168.56.202>;tag=aDUZXF1Z1vD6p","sip_term_status":"200","proto_specific_hangup_cause":"sip:200","sip_term_cause":"16","last_bridge_role":"originator","sip_user_agent":"Jitsi2.10.5550Windows 10","sip_hangup_disposition":"recv_bye","bridge_hangup_cause":"NORMAL_CLEARING","hangup_cause":"NORMAL_CLEARING","hangup_cause_q850":"16","digits_dialed":"none","start_stamp":"2018-03-14 07:09:18","profile_start_stamp":"2018-03-14 07:09:18","answer_stamp":"2018-03-14 07:09:27","bridge_stamp":"2018-03-14 07:09:27","hold_stamp":"2018-03-14 07:09:27","progress_stamp":"2018-03-14 07:09:18","progress_media_stamp":"2018-03-14 07:09:27","hold_events":"{{1521025767847893,1521025783334494}}","end_stamp":"2018-03-14 07:09:48","start_epoch":"1521025758","start_uepoch":"1521025758006702","profile_start_epoch":"1521025758","profile_start_uepoch":"1521025758026167","answer_epoch":"1521025767","answer_uepoch":"1521025767766321","bridge_epoch":"1521025767","bridge_uepoch":"1521025767766321","last_hold_epoch":"1521025767","last_hold_uepoch":"1521025767847892","hold_accum_seconds":"15","hold_accum_usec":"15486602","hold_accum_ms":"15486","resurrect_epoch":"0","resurrect_uepoch":"0","progress_epoch":"1521025758","progress_uepoch":"1521025758116123","progress_media_epoch":"1521025767","progress_media_uepoch":"1521025767766321","end_epoch":"1521025788","end_uepoch":"1521025788366141","last_app":"bridge","last_arg":"user/1002@192.168.56.202","caller_id":"\"1001\" <1001>","duration":"30","billsec":"21","progresssec":"0","answersec":"9","waitsec":"9","progress_mediasec":"9","flow_billsec":"30","mduration":"30360","billmsec":"20600","progressmsec":"110","answermsec":"9760","waitmsec":"9760","progress_mediamsec":"9760","flow_billmsec":"30360","uduration":"30359439","billusec":"20599820","progressusec":"109421","answerusec":"9759619","waitusec":"9759619","progress_mediausec":"9759619","flow_billusec":"30359439","rtp_audio_in_raw_bytes":"174156","rtp_audio_in_media_bytes":"166416","rtp_audio_in_packet_count":"1033","rtp_audio_in_media_packet_count":"988","rtp_audio_in_skip_packet_count":"7","rtp_audio_in_jitter_packet_count":"0","rtp_audio_in_dtmf_packet_count":"0","rtp_audio_in_cng_packet_count":"0","rtp_audio_in_flush_packet_count":"45","rtp_audio_in_largest_jb_size":"0","rtp_audio_in_jitter_min_variance":"0.50","rtp_audio_in_jitter_max_variance":"31.77","rtp_audio_in_jitter_loss_rate":"0.00","rtp_audio_in_jitter_burst_rate":"0.00","rtp_audio_in_mean_interval":"20.17","rtp_audio_in_flaw_total":"1","rtp_audio_in_quality_percentage":"99.00","rtp_audio_in_mos":"4.49","rtp_audio_out_raw_bytes":"43344","rtp_audio_out_media_bytes":"43344","rtp_audio_out_packet_count":"252","rtp_audio_out_media_packet_count":"252","rtp_audio_out_skip_packet_count":"0","rtp_audio_out_dtmf_packet_count":"0","rtp_audio_out_cng_packet_count":"0","rtp_audio_rtcp_packet_count":"0","rtp_audio_rtcp_octet_count":"0"},"app_log":{"applications":[{"app_name":"info","app_data":"","app_stamp":"1521025758010697"},{"app_name":"park","app_data":"","app_stamp":"1521025758011143"},{"app_name":"set","app_data":"ringback=","app_stamp":"1521025758057183"},{"app_name":"set","app_data":"call_timeout=30","app_stamp":"1521025758057474"},{"app_name":"bridge","app_data":"user/1002@192.168.56.202","app_stamp":"1521025758057698"}]},"callflow":[{"dialplan":"XML","profile_index":"2","extension":{"name":"Local_Extension","number":"1002","applications":[{"app_name":"set","app_data":"ringback=${us-ring}"},{"app_name":"set","app_data":"call_timeout=30"},{"app_name":"bridge","app_data":"user/${destination_number}@${domain_name}"}]},"caller_profile":{"username":"1001","dialplan":"XML","caller_id_name":"1001","ani":"1001","aniii":"","caller_id_number":"1001","network_addr":"192.168.56.1","rdnis":"1002","destination_number":"1002","uuid":"5a3a1d91-90d3-4db4-af5c-cc3ae15d93a4","source":"mod_sofia","context":"default","chan_name":"sofia/internal/1001@192.168.56.202","originatee":{"originatee_caller_profiles":[{"username":"1001","dialplan":"XML","caller_id_name":"Extension 1001","ani":"1001","aniii":"","caller_id_number":"1001","network_addr":"192.168.56.1","rdnis":"1002","destination_number":"1002","uuid":"9c1afb4f-1d4a-4e45-84a3-d25721981bf5","source":"mod_sofia","context":"default","chan_name":"sofia/internal/1002@192.168.56.1:5060"},{"username":"1001","dialplan":"XML","caller_id_name":"Extension 1001","ani":"1001","aniii":"","caller_id_number":"1001","network_addr":"192.168.56.1","rdnis":"1002","destination_number":"1002","uuid":"9c1afb4f-1d4a-4e45-84a3-d25721981bf5","source":"mod_sofia","context":"default","chan_name":"sofia/internal/1002@192.168.56.1:5060"}]}},"times":{"created_time":"1521025758006702","profile_created_time":"1521025758026167","progress_time":"1521025758116123","progress_media_time":"1521025767766321","answered_time":"1521025767766321","bridged_time":"1521025767766321","last_hold_time":"1521025767847892","hold_accum_time":"15486602","hangup_time":"1521025788366141","resurrect_time":"0","transfer_time":"0"}},{"dialplan":"XML","profile_index":"1","extension":{"name":"CGRateS_Auth","number":"1002","applications":[{"app_name":"info","app_data":""},{"app_name":"park","app_data":""}]},"caller_profile":{"username":"1001","dialplan":"XML","caller_id_name":"1001","ani":"1001","aniii":"","caller_id_number":"1001","network_addr":"192.168.56.1","rdnis":"","destination_number":"1002","uuid":"5a3a1d91-90d3-4db4-af5c-cc3ae15d93a4","source":"mod_sofia","context":"default","chan_name":"sofia/internal/1001@192.168.56.202"},"times":{"created_time":"1521025758006702","profile_created_time":"1521025758006702","progress_time":"0","progress_media_time":"0","answered_time":"0","bridged_time":"0","last_hold_time":"0","hold_accum_time":"0","hangup_time":"0","resurrect_time":"0","transfer_time":"1521025758026167"}}]}		`)
	hangUp := `Event-Name: CHANNEL_HANGUP_COMPLETE
Core-UUID: 63e2315b-d538-4dfa-9ed5-af73ba6210b6
FreeSWITCH-Hostname: teo
FreeSWITCH-Switchname: teo
FreeSWITCH-IPv4: 10.0.2.15
FreeSWITCH-IPv6: %3A%3A1
Event-Date-Local: 2018-03-14%2007%3A09%3A48
Event-Date-GMT: Wed,%2014%20Mar%202018%2011%3A09%3A48%20GMT
Event-Date-Timestamp: 1521025788386120
Event-Calling-File: switch_core_state_machine.c
Event-Calling-Function: switch_core_session_reporting_state
Event-Calling-Line-Number: 949
Event-Sequence: 1113
Hangup-Cause: NORMAL_CLEARING
Channel-State: CS_REPORTING
Channel-Call-State: HANGUP
Channel-State-Number: 11
Channel-Name: sofia/internal/1001%40192.168.56.202
Unique-ID: 5a3a1d91-90d3-4db4-af5c-cc3ae15d93a4
Call-Direction: inbound
Presence-Call-Direction: inbound
Channel-HIT-Dialplan: true
Channel-Presence-ID: 1001%40192.168.56.202
Channel-Call-UUID: 5a3a1d91-90d3-4db4-af5c-cc3ae15d93a4
Answer-State: hangup
Hangup-Cause: NORMAL_CLEARING
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
Caller-Unique-ID: 5a3a1d91-90d3-4db4-af5c-cc3ae15d93a4
Caller-Source: mod_sofia
Caller-Transfer-Source: 1521025758%3A86c9ebb2-888f-42d5-9afa-2101449a4b86%3Abl_xfer%3A1002/default/XML
Caller-Context: default
Caller-RDNIS: 1002
Caller-Channel-Name: sofia/internal/1001%40192.168.56.202
Caller-Profile-Index: 2
Caller-Profile-Created-Time: 1521025758026167
Caller-Channel-Created-Time: 1521025758006702
Caller-Channel-Answered-Time: 1521025767766321
Caller-Channel-Progress-Time: 1521025758116123
Caller-Channel-Progress-Media-Time: 1521025767766321
Caller-Channel-Hangup-Time: 1521025788366141
Caller-Channel-Transfer-Time: 0
Caller-Channel-Resurrect-Time: 0
Caller-Channel-Bridged-Time: 1521025767766321
Caller-Channel-Last-Hold: 1521025767847892
Caller-Channel-Hold-Accum: 15486602
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
Other-Leg-Unique-ID: 9c1afb4f-1d4a-4e45-84a3-d25721981bf5
Other-Leg-Source: mod_sofia
Other-Leg-Context: default
Other-Leg-RDNIS: 1002
Other-Leg-Channel-Name: sofia/internal/1002%40192.168.56.1%3A5060
Other-Leg-Profile-Created-Time: 1521025758047206
Other-Leg-Channel-Created-Time: 1521025758047206
Other-Leg-Channel-Answered-Time: 1521025767745928
Other-Leg-Channel-Progress-Time: 1521025758116123
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
variable_uuid: 5a3a1d91-90d3-4db4-af5c-cc3ae15d93a4
variable_session_id: 1
variable_sip_from_user: 1001
variable_sip_from_uri: 1001%40192.168.56.202
variable_sip_from_host: 192.168.56.202
variable_channel_name: sofia/internal/1001%40192.168.56.202
variable_ep_codec_string: mod_spandsp.G722%408000h%4020i%4064000b,CORE_PCM_MODULE.PCMU%408000h%4020i%4064000b,CORE_PCM_MODULE.PCMA%408000h%4020i%4064000b,mod_spandsp.GSM%408000h%4020i%4013200b
variable_sip_local_network_addr: 192.168.56.202
variable_sip_network_ip: 192.168.56.1
variable_sip_network_port: 5060
variable_sip_invite_stamp: 1521025758006702
variable_sip_received_ip: 192.168.56.1
variable_sip_received_port: 5060
variable_sip_via_protocol: udp
variable_sip_authorized: true
variable_Event-Name: REQUEST_PARAMS
variable_Core-UUID: 63e2315b-d538-4dfa-9ed5-af73ba6210b6
variable_FreeSWITCH-Hostname: teo
variable_FreeSWITCH-Switchname: teo
variable_FreeSWITCH-IPv4: 10.0.2.15
variable_FreeSWITCH-IPv6: %3A%3A1
variable_Event-Date-Local: 2018-03-14%2007%3A09%3A18
variable_Event-Date-GMT: Wed,%2014%20Mar%202018%2011%3A09%3A18%20GMT
variable_Event-Date-Timestamp: 1521025758006702
variable_Event-Calling-File: sofia.c
variable_Event-Calling-Function: sofia_handle_sip_i_invite
variable_Event-Calling-Line-Number: 10096
variable_Event-Sequence: 1025
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
variable_cgr_subsystems: *resources%3B*attributes%3B*sessions%3B*suppliers
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
variable_sip_via_host: 192.168.56.1
variable_sip_via_port: 5060
variable_presence_id: 1001%40192.168.56.202
variable_cgr_resource_allocation: ResGroup1
variable_cgr_suppliers: ARRAY%3A%3A3%7C%3Asupplier2%7C%3Asupplier3%7C%3Asupplier1
variable_cgr_notify: AUTH_OK
variable_max_forwards: 69
variable_transfer_history: 1521025758%3A86c9ebb2-888f-42d5-9afa-2101449a4b86%3Abl_xfer%3A1002/default/XML
variable_transfer_source: 1521025758%3A86c9ebb2-888f-42d5-9afa-2101449a4b86%3Abl_xfer%3A1002/default/XML
variable_DP_MATCH: ARRAY%3A%3A1002%7C%3A1002
variable_call_uuid: 5a3a1d91-90d3-4db4-af5c-cc3ae15d93a4
variable_call_timeout: 30
variable_current_application_data: user/1002%40192.168.56.202
variable_current_application: bridge
variable_dialed_user: 1002
variable_dialed_domain: 192.168.56.202
variable_originated_legs: ARRAY%3A%3A9c1afb4f-1d4a-4e45-84a3-d25721981bf5%3BOutbound%20Call%3B1002%7C%3A9c1afb4f-1d4a-4e45-84a3-d25721981bf5%3BOutbound%20Call%3B1002
variable_switch_m_sdp: v%3D0%0D%0Ao%3D1002-jitsi.org%200%200%20IN%20IP4%20192.168.56.1%0D%0As%3D-%0D%0Ac%3DIN%20IP4%20192.168.56.1%0D%0At%3D0%200%0D%0Am%3Daudio%205004%20RTP/AVP%209%200%208%203%20101%0D%0Aa%3Drtpmap%3A9%20G722/8000%0D%0Aa%3Drtpmap%3A0%20PCMU/8000%0D%0Aa%3Drtpmap%3A8%20PCMA/8000%0D%0Aa%3Drtpmap%3A3%20GSM/8000%0D%0Aa%3Drtpmap%3A101%20telephone-event/8000%0D%0A
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
variable_local_media_ip: 192.168.56.202
variable_local_media_port: 29014
variable_advertised_media_ip: 192.168.56.202
variable_rtp_use_timer_name: soft
variable_rtp_use_pt: 9
variable_rtp_use_ssrc: 2729250253
variable_endpoint_disposition: ANSWER
variable_originate_causes: ARRAY%3A%3A9c1afb4f-1d4a-4e45-84a3-d25721981bf5%3BNONE%7C%3A9c1afb4f-1d4a-4e45-84a3-d25721981bf5%3BNONE
variable_originate_disposition: SUCCESS
variable_DIALSTATUS: SUCCESS
variable_last_bridge_to: 9c1afb4f-1d4a-4e45-84a3-d25721981bf5
variable_bridge_channel: sofia/internal/1002%40192.168.56.1%3A5060
variable_bridge_uuid: 9c1afb4f-1d4a-4e45-84a3-d25721981bf5
variable_signal_bond: 9c1afb4f-1d4a-4e45-84a3-d25721981bf5
variable_last_sent_callee_id_name: Outbound%20Call
variable_last_sent_callee_id_number: 1002
variable_switch_r_sdp: v%3D0%0D%0Ao%3D1001-jitsi.org%200%202%20IN%20IP4%20192.168.56.1%0D%0As%3D-%0D%0Ac%3DIN%20IP4%20192.168.56.1%0D%0At%3D0%200%0D%0Am%3Daudio%205000%20RTP/AVP%2096%2097%2098%209%20100%20102%200%208%20103%203%20104%204%20101%0D%0Aa%3Drtpmap%3A96%20opus/48000/2%0D%0Aa%3Dfmtp%3A96%20usedtx%3D1%0D%0Aa%3Drtpmap%3A97%20SILK/24000%0D%0Aa%3Drtpmap%3A98%20SILK/16000%0D%0Aa%3Drtpmap%3A9%20G722/8000%0D%0Aa%3Drtpmap%3A100%20speex/32000%0D%0Aa%3Drtpmap%3A102%20speex/16000%0D%0Aa%3Drtpmap%3A0%20PCMU/8000%0D%0Aa%3Drtpmap%3A8%20PCMA/8000%0D%0Aa%3Drtpmap%3A103%20iLBC/8000%0D%0Aa%3Drtpmap%3A3%20GSM/8000%0D%0Aa%3Drtpmap%3A104%20speex/8000%0D%0Aa%3Drtpmap%3A4%20G723/8000%0D%0Aa%3Dfmtp%3A4%20annexa%3Dno%3Bbitrate%3D6.3%0D%0Aa%3Drtpmap%3A101%20telephone-event/8000%0D%0Aa%3Dptime%3A20%0D%0Aa%3Dextmap%3A1%20urn%3Aietf%3Aparams%3Artp-hdrext%3Acsrc-audio-level%0D%0Aa%3Dextmap%3A2%20urn%3Aietf%3Aparams%3Artp-hdrext%3Assrc-audio-level%0D%0Aa%3Drtcp-xr%3Avoip-metrics%0D%0Aa%3Dzrtp-hash%3A1.10%208e8dd2fa6803f32845f26e55879c776a4bc015ee05b41630313aee27ef77fb30%0D%0Am%3Dvideo%205006%20RTP/AVP%20105%2099%0D%0Aa%3Drtpmap%3A105%20H264/90000%0D%0Aa%3Dfmtp%3A105%20profile-level-id%3D4DE01f%3Bpacketization-mode%3D1%0D%0Aa%3Drtpmap%3A99%20H264/90000%0D%0Aa%3Dfmtp%3A99%20profile-level-id%3D4DE01f%0D%0Aa%3Drecvonly%0D%0Aa%3Dimageattr%3A105%20send%20*%20recv%20%5Bx%3D%5B1%3A1920%5D,y%3D%5B1%3A1080%5D%5D%0D%0Aa%3Dimageattr%3A99%20send%20*%20recv%20%5Bx%3D%5B1%3A1920%5D,y%3D%5B1%3A1080%5D%5D%0D%0A
variable_rtp_use_codec_string: G722,PCMU,PCMA,GSM
variable_r_sdp_audio_zrtp_hash: 1.10%208e8dd2fa6803f32845f26e55879c776a4bc015ee05b41630313aee27ef77fb30
variable_audio_media_flow: sendrecv
variable_remote_media_ip: 192.168.56.1
variable_remote_media_port: 5000
variable_rtp_audio_recv_pt: 9
variable_dtmf_type: rfc2833
variable_rtp_2833_send_payload: 101
variable_rtp_2833_recv_payload: 101
variable_video_possible: true
variable_video_media_flow: sendonly
variable_rtp_local_sdp_str: v%3D0%0D%0Ao%3DFreeSWITCH%201520996753%201520996756%20IN%20IP4%20192.168.56.202%0D%0As%3DFreeSWITCH%0D%0Ac%3DIN%20IP4%20192.168.56.202%0D%0At%3D0%200%0D%0Am%3Daudio%2029014%20RTP/AVP%209%20101%0D%0Aa%3Drtpmap%3A9%20G722/8000%0D%0Aa%3Drtpmap%3A101%20telephone-event/8000%0D%0Aa%3Dfmtp%3A101%200-16%0D%0Aa%3Dptime%3A20%0D%0Aa%3Dsendrecv%0D%0Am%3Dvideo%200%20RTP/AVP%2019%0D%0A
variable_sip_to_tag: aDUZXF1Z1vD6p
variable_sip_from_tag: df94d020
variable_sip_cseq: 4
variable_sip_call_id: 985e365faa0ec79a7fa75d001ef2449f%400%3A0%3A0%3A0%3A0%3A0%3A0%3A0
variable_sip_full_via: SIP/2.0/UDP%20192.168.56.1%3A5060%3Bbranch%3Dz9hG4bK-323230-ab335b3491dd24f5ec251b9700716b97
variable_sip_from_display: 1001
variable_sip_full_from: %221001%22%20%3Csip%3A1001%40192.168.56.202%3E%3Btag%3Ddf94d020
variable_sip_full_to: %3Csip%3A1002%40192.168.56.202%3E%3Btag%3DaDUZXF1Z1vD6p
variable_sip_term_status: 200
variable_proto_specific_hangup_cause: sip%3A200
variable_sip_term_cause: 16
variable_last_bridge_role: originator
variable_sip_user_agent: Jitsi2.10.5550Windows%2010
variable_sip_hangup_disposition: recv_bye
variable_bridge_hangup_cause: NORMAL_CLEARING
variable_hangup_cause: NORMAL_CLEARING
variable_hangup_cause_q850: 16
variable_digits_dialed: none
variable_start_stamp: 2018-03-14%2007%3A09%3A18
variable_profile_start_stamp: 2018-03-14%2007%3A09%3A18
variable_answer_stamp: 2018-03-14%2007%3A09%3A27
variable_bridge_stamp: 2018-03-14%2007%3A09%3A27
variable_hold_stamp: 2018-03-14%2007%3A09%3A27
variable_progress_stamp: 2018-03-14%2007%3A09%3A18
variable_progress_media_stamp: 2018-03-14%2007%3A09%3A27
variable_hold_events: %7B%7B1521025767847893,1521025783334494%7D%7D
variable_end_stamp: 2018-03-14%2007%3A09%3A48
variable_start_epoch: 1521025758
variable_start_uepoch: 1521025758006702
variable_profile_start_epoch: 1521025758
variable_profile_start_uepoch: 1521025758026167
variable_answer_epoch: 1521025767
variable_answer_uepoch: 1521025767766321
variable_bridge_epoch: 1521025767
variable_bridge_uepoch: 1521025767766321
variable_last_hold_epoch: 1521025767
variable_last_hold_uepoch: 1521025767847892
variable_hold_accum_seconds: 15
variable_hold_accum_usec: 15486602
variable_hold_accum_ms: 15486
variable_resurrect_epoch: 0
variable_resurrect_uepoch: 0
variable_progress_epoch: 1521025758
variable_progress_uepoch: 1521025758116123
variable_progress_media_epoch: 1521025767
variable_progress_media_uepoch: 1521025767766321
variable_end_epoch: 1521025788
variable_end_uepoch: 1521025788366141
variable_last_app: bridge
variable_last_arg: user/1002%40192.168.56.202
variable_caller_id: %221001%22%20%3C1001%3E
variable_duration: 30
variable_billsec: 21
variable_progresssec: 0
variable_answersec: 9
variable_waitsec: 9
variable_progress_mediasec: 9
variable_flow_billsec: 30
variable_mduration: 30360
variable_billmsec: 20600
variable_progressmsec: 110
variable_answermsec: 9760
variable_waitmsec: 9760
variable_progress_mediamsec: 9760
variable_flow_billmsec: 30360
variable_uduration: 30359439
variable_billusec: 20599820
variable_progressusec: 109421
variable_answerusec: 9759619
variable_waitusec: 9759619
variable_progress_mediausec: 9759619
variable_flow_billusec: 30359439
variable_rtp_audio_in_raw_bytes: 174156
variable_rtp_audio_in_media_bytes: 166416
variable_rtp_audio_in_packet_count: 1033
variable_rtp_audio_in_media_packet_count: 988
variable_rtp_audio_in_skip_packet_count: 7
variable_rtp_audio_in_jitter_packet_count: 0
variable_rtp_audio_in_dtmf_packet_count: 0
variable_rtp_audio_in_cng_packet_count: 0
variable_rtp_audio_in_flush_packet_count: 45
variable_rtp_audio_in_largest_jb_size: 0
variable_rtp_audio_in_jitter_min_variance: 0.50
variable_rtp_audio_in_jitter_max_variance: 31.77
variable_rtp_audio_in_jitter_loss_rate: 0.00
variable_rtp_audio_in_jitter_burst_rate: 0.00
variable_rtp_audio_in_mean_interval: 20.17
variable_rtp_audio_in_flaw_total: 1
variable_rtp_audio_in_quality_percentage: 99.00
variable_rtp_audio_in_mos: 4.49
variable_rtp_audio_out_raw_bytes: 43344
variable_rtp_audio_out_media_bytes: 43344
variable_rtp_audio_out_packet_count: 252
variable_rtp_audio_out_media_packet_count: 252
variable_rtp_audio_out_skip_packet_count: 0
variable_rtp_audio_out_dtmf_packet_count: 0
variable_rtp_audio_out_cng_packet_count: 0
variable_rtp_audio_rtcp_packet_count: 0
variable_rtp_audio_rtcp_octet_count: 0`
	var fsCdrCfg *config.CGRConfig
	timezone := config.CgrConfig().GeneralCfg().DefaultTimezone
	fsCdrCfg, _ = config.NewDefaultCGRConfig()
	fsCdr, _ := engine.NewFSCdr(body, fsCdrCfg)
	smGev := engine.NewSafEvent(NewFSEvent(hangUp).AsMapStringInterface(timezone))
	sessions.GetSetCGRID(smGev)
	smCDR, err := smGev.AsCDR(fsCdrCfg, utils.EmptyString, timezone)
	if err != nil {
		t.Error(err)
	}
	fsCDR := fsCdr.AsCDR(timezone)
	if fsCDR.CGRID != smCDR.CGRID {
		t.Errorf("Expecting: %s, received: %s", fsCDR.CGRID, smCDR.CGRID)
	}
}

func TestFsEvV1AuthorizeArgs(t *testing.T) {
	timezone := config.CgrConfig().GeneralCfg().DefaultTimezone
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
		GetSuppliers:          true,
		GetAttributes:         true,
		SuppliersIgnoreErrors: true,
		SuppliersMaxCost:      utils.MetaEventCost,
	}
	expected.Event[utils.Usage] = config.CgrConfig().SessionSCfg().MaxCallDuration
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
	} else if !reflect.DeepEqual(expected.SuppliersMaxCost, rcv.SuppliersMaxCost) {
		t.Errorf("Expecting: %+v, received: %+v", expected.SuppliersMaxCost, rcv.SuppliersMaxCost)
	} else if !reflect.DeepEqual(expected.SuppliersIgnoreErrors, rcv.SuppliersIgnoreErrors) {
		t.Errorf("Expecting: %+v, received: %+v", expected.SuppliersIgnoreErrors, rcv.SuppliersIgnoreErrors)
	}
}

func TestFsEvV1InitSessionArgs(t *testing.T) {
	timezone := config.CgrConfig().GeneralCfg().DefaultTimezone
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
	timezone := config.CgrConfig().GeneralCfg().DefaultTimezone
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
