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

package sessionmanager

import (
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
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
	ev := new(FSEvent).AsEvent(body)
	if ev.GetName() != "RE_SCHEDULE" {
		t.Error("Event not parsed correctly: ", ev)
	}
	l := len(ev.(FSEvent))
	if l != 17 {
		t.Error("Incorrect number of event fields: ", l)
	}
}

// Detects if any of the parsers do not return static values
func TestEventParseStatic(t *testing.T) {
	ev := new(FSEvent).AsEvent("")
	setupTime, _ := ev.GetSetupTime("^2013-12-07 08:42:24", "")
	answerTime, _ := ev.GetAnswerTime("^2013-12-07 08:42:24", "")
	dur, _ := ev.GetDuration("^60s")
	if ev.GetReqType("^test") != "test" ||
		ev.GetDirection("^test") != "test" ||
		ev.GetTenant("^test") != "test" ||
		ev.GetCategory("^test") != "test" ||
		ev.GetAccount("^test") != "test" ||
		ev.GetSubject("^test") != "test" ||
		ev.GetDestination("^test") != "test" ||
		setupTime != time.Date(2013, 12, 7, 8, 42, 24, 0, time.UTC) ||
		answerTime != time.Date(2013, 12, 7, 8, 42, 24, 0, time.UTC) ||
		dur != time.Duration(60)*time.Second ||
		ev.GetSupplier("^test") != "test" ||
		ev.GetDisconnectCause("^test") != "test" {
		t.Error("Values out of static not matching",
			ev.GetReqType("^test") != "test",
			ev.GetDirection("^test") != "test",
			ev.GetTenant("^test") != "test",
			ev.GetCategory("^test") != "test",
			ev.GetAccount("^test") != "test",
			ev.GetSubject("^test") != "test",
			ev.GetDestination("^test") != "test",
			setupTime != time.Date(2013, 12, 7, 8, 42, 24, 0, time.UTC),
			answerTime != time.Date(2013, 12, 7, 8, 42, 24, 0, time.UTC),
			dur != time.Duration(60)*time.Second,
			ev.GetSupplier("^test") != "test",
			ev.GetDisconnectCause("^test") != "test")
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
	ev := new(FSEvent).AsEvent(body)
	setupTime, _ := ev.GetSetupTime("Event-Date-Local", "")
	answerTime, _ := ev.GetAnswerTime("Event-Date-Local", "")
	dur, _ := ev.GetDuration("Event-Calling-Line-Number")
	if ev.GetReqType("FreeSWITCH-Hostname") != "h1.ip-switch.net" ||
		ev.GetDirection("FreeSWITCH-Hostname") != "*out" ||
		ev.GetTenant("FreeSWITCH-Hostname") != "h1.ip-switch.net" ||
		ev.GetCategory("FreeSWITCH-Hostname") != "h1.ip-switch.net" ||
		ev.GetAccount("FreeSWITCH-Hostname") != "h1.ip-switch.net" ||
		ev.GetSubject("FreeSWITCH-Hostname") != "h1.ip-switch.net" ||
		ev.GetDestination("FreeSWITCH-Hostname") != "h1.ip-switch.net" ||
		setupTime != time.Date(2012, 10, 5, 13, 41, 38, 0, time.UTC) ||
		answerTime != time.Date(2012, 10, 5, 13, 41, 38, 0, time.UTC) ||
		dur != time.Duration(65)*time.Second ||
		ev.GetSupplier("FreeSWITCH-Hostname") != "h1.ip-switch.net" ||
		ev.GetDisconnectCause("FreeSWITCH-Hostname") != "h1.ip-switch.net" {
		t.Error("Values out of static not matching",
			ev.GetReqType("FreeSWITCH-Hostname") != "h1.ip-switch.net",
			ev.GetDirection("FreeSWITCH-Hostname") != "*out",
			ev.GetTenant("FreeSWITCH-Hostname") != "h1.ip-switch.net",
			ev.GetCategory("FreeSWITCH-Hostname") != "h1.ip-switch.net",
			ev.GetAccount("FreeSWITCH-Hostname") != "h1.ip-switch.net",
			ev.GetSubject("FreeSWITCH-Hostname") != "h1.ip-switch.net",
			ev.GetDestination("FreeSWITCH-Hostname") != "h1.ip-switch.net",
			setupTime != time.Date(2012, 10, 5, 13, 41, 38, 0, time.UTC),
			answerTime != time.Date(2012, 10, 5, 13, 41, 38, 0, time.UTC),
			dur != time.Duration(65)*time.Second,
			ev.GetSupplier("FreeSWITCH-Hostname") != "h1.ip-switch.net",
			ev.GetDisconnectCause("FreeSWITCH-Hostname") != "h1.ip-switch.net")
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
	ev := new(FSEvent).AsEvent(body)
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
	ev := new(FSEvent).AsEvent(hangupEv)
	setupTime, _ := ev.GetSetupTime(utils.META_DEFAULT, "")
	answerTime, _ := ev.GetAnswerTime(utils.META_DEFAULT, "")
	dur, _ := ev.GetDuration(utils.META_DEFAULT)
	if ev.GetReqType(utils.META_DEFAULT) != utils.META_PREPAID ||
		ev.GetDirection(utils.META_DEFAULT) != "*out" ||
		ev.GetTenant(utils.META_DEFAULT) != "cgrates.org" ||
		ev.GetCategory(utils.META_DEFAULT) != "call" ||
		ev.GetAccount(utils.META_DEFAULT) != "1001" ||
		ev.GetSubject(utils.META_DEFAULT) != "1001" ||
		ev.GetDestination(utils.META_DEFAULT) != "1003" ||
		setupTime.UTC() != time.Date(2015, 7, 7, 14, 52, 8, 0, time.UTC) ||
		answerTime.UTC() != time.Date(2015, 7, 7, 14, 52, 8, 0, time.UTC) ||
		dur != time.Duration(66)*time.Second ||
		ev.GetSupplier(utils.META_DEFAULT) != "supplier1" ||
		ev.GetDisconnectCause(utils.META_DEFAULT) != "NORMAL_CLEARING" {
		t.Error("Default values not matching",
			ev.GetReqType(utils.META_DEFAULT) != utils.META_PREPAID,
			ev.GetDirection(utils.META_DEFAULT) != "*out",
			ev.GetTenant(utils.META_DEFAULT) != "cgrates.org",
			ev.GetCategory(utils.META_DEFAULT) != "call",
			ev.GetAccount(utils.META_DEFAULT) != "1001",
			ev.GetSubject(utils.META_DEFAULT) != "1001",
			ev.GetDestination(utils.META_DEFAULT) != "1003",
			setupTime.UTC() != time.Date(2015, 7, 7, 14, 52, 8, 0, time.UTC),
			answerTime.UTC() != time.Date(2015, 7, 7, 14, 52, 8, 0, time.UTC),
			dur != time.Duration(66)*time.Second,
			ev.GetSupplier(utils.META_DEFAULT) != "supplier1",
			ev.GetDisconnectCause(utils.META_DEFAULT) != "NORMAL_CLEARING")
	}
}

func TestParseEventValue(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	config.SetCgrConfig(cfg)
	ev := new(FSEvent).AsEvent(hangupEv)
	if cgrid := ev.ParseEventValue(&utils.RSRField{Id: utils.CGRID}, ""); cgrid != "164b0422fdc6a5117031b427439482c6a4f90e41" {
		t.Error("Unexpected cgrid parsed", cgrid)
	}
	if tor := ev.ParseEventValue(&utils.RSRField{Id: utils.TOR}, ""); tor != utils.VOICE {
		t.Error("Unexpected tor parsed", tor)
	}
	if accid := ev.ParseEventValue(&utils.RSRField{Id: utils.ACCID}, ""); accid != "e3133bf7-dcde-4daf-9663-9a79ffcef5ad" {
		t.Error("Unexpected result parsed", accid)
	}
	if parsed := ev.ParseEventValue(&utils.RSRField{Id: utils.CDRHOST}, ""); parsed != "10.0.3.15" {
		t.Error("Unexpected result parsed", parsed)
	}
	if parsed := ev.ParseEventValue(&utils.RSRField{Id: utils.CDRSOURCE}, ""); parsed != "FS_EVENT" {
		t.Error("Unexpected result parsed", parsed)
	}
	if parsed := ev.ParseEventValue(&utils.RSRField{Id: utils.REQTYPE}, ""); parsed != utils.META_PREPAID {
		t.Error("Unexpected result parsed", parsed)
	}
	if parsed := ev.ParseEventValue(&utils.RSRField{Id: utils.DIRECTION}, ""); parsed != utils.OUT {
		t.Error("Unexpected result parsed", parsed)
	}
	if parsed := ev.ParseEventValue(&utils.RSRField{Id: utils.TENANT}, ""); parsed != "cgrates.org" {
		t.Error("Unexpected result parsed", parsed)
	}
	if parsed := ev.ParseEventValue(&utils.RSRField{Id: utils.CATEGORY}, ""); parsed != "call" {
		t.Error("Unexpected result parsed", parsed)
	}
	if parsed := ev.ParseEventValue(&utils.RSRField{Id: utils.ACCOUNT}, ""); parsed != "1001" {
		t.Error("Unexpected result parsed", parsed)
	}
	if parsed := ev.ParseEventValue(&utils.RSRField{Id: utils.SUBJECT}, ""); parsed != "1001" {
		t.Error("Unexpected result parsed", parsed)
	}
	if parsed := ev.ParseEventValue(&utils.RSRField{Id: utils.DESTINATION}, ""); parsed != "1003" {
		t.Error("Unexpected result parsed", parsed)
	}
	sTime, _ := utils.ParseTimeDetectLayout("1436280728471153"[:len("1436280728471153")-6], "") // We discard nanoseconds information so we can correlate csv
	if parsed := ev.ParseEventValue(&utils.RSRField{Id: utils.SETUP_TIME}, ""); parsed != sTime.String() {
		t.Errorf("Expecting: %s, parsed: %s", sTime.String(), parsed)
	}
	aTime, _ := utils.ParseTimeDetectLayout("1436280728971147"[:len("1436280728971147")-6], "")
	if parsed := ev.ParseEventValue(&utils.RSRField{Id: utils.ANSWER_TIME}, ""); parsed != aTime.String() {
		t.Errorf("Expecting: %s, parsed: %s", aTime.String(), parsed)
	}
	if parsed := ev.ParseEventValue(&utils.RSRField{Id: utils.USAGE}, ""); parsed != "66000000000" {
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

func TestPassesFieldFilterDn1(t *testing.T) {
	body := `Event-Name: RE_SCHEDULE
Core-UUID: 792e181c-b6e6-499c-82a1-52a778e7d82d
FreeSWITCH-Hostname: h1.ip-switch.net
FreeSWITCH-Switchname: h1.ip-switch.net
FreeSWITCH-IPv4: 88.198.12.156
Caller-Username: futurem0005`
	ev := new(FSEvent).AsEvent(body)
	acntPrefxFltr, _ := utils.NewRSRField(`~account:s/^\w+[shmp]\d{4}$//`)
	if pass, _ := ev.PassesFieldFilter(acntPrefxFltr); !pass {
		t.Error("Not passing valid filter")
	}
	body = `Event-Name: RE_SCHEDULE
Core-UUID: 792e181c-b6e6-499c-82a1-52a778e7d82d
FreeSWITCH-Hostname: h1.ip-switch.net
FreeSWITCH-Switchname: h1.ip-switch.net
FreeSWITCH-IPv4: 88.198.12.156
Caller-Username: futurem00005`
	ev = new(FSEvent).AsEvent(body)
	if pass, _ := ev.PassesFieldFilter(acntPrefxFltr); pass {
		t.Error("Should not pass filter")
	}
	body = `Event-Name: RE_SCHEDULE
Core-UUID: 792e181c-b6e6-499c-82a1-52a778e7d82d
FreeSWITCH-Hostname: h1.ip-switch.net
FreeSWITCH-Switchname: h1.ip-switch.net
FreeSWITCH-IPv4: 88.198.12.156
Caller-Username: 0402129281`
	ev = new(FSEvent).AsEvent(body)
	acntPrefxFltr, _ = utils.NewRSRField(`~account:s/^0\d{9}$//`)
	if pass, _ := ev.PassesFieldFilter(acntPrefxFltr); !pass {
		t.Error("Not passing valid filter")
	}
	acntPrefxFltr, _ = utils.NewRSRField(`~account:s/^0(\d{9})$/placeholder/`)
	if pass, _ := ev.PassesFieldFilter(acntPrefxFltr); pass {
		t.Error("Should not pass filter")
	}
	body = `Event-Name: RE_SCHEDULE
Core-UUID: 792e181c-b6e6-499c-82a1-52a778e7d82d
FreeSWITCH-Hostname: h1.ip-switch.net
FreeSWITCH-Switchname: h1.ip-switch.net
FreeSWITCH-IPv4: 88.198.12.156
Caller-Username: 04021292812`
	ev = new(FSEvent).AsEvent(body)
	if pass, _ := ev.PassesFieldFilter(acntPrefxFltr); pass {
		t.Error("Should not pass filter")
	}
}

func TestFsEvAsStoredCdr(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	config.SetCgrConfig(cfg)
	ev := new(FSEvent).AsEvent(hangupEv)
	setupTime, _ := utils.ParseTimeDetectLayout("1436280728", "")
	aTime, _ := utils.ParseTimeDetectLayout("1436280728", "")
	eStoredCdr := &engine.StoredCdr{CgrId: "164b0422fdc6a5117031b427439482c6a4f90e41",
		TOR: utils.VOICE, AccId: "e3133bf7-dcde-4daf-9663-9a79ffcef5ad", CdrHost: "10.0.3.15", CdrSource: "FS_CHANNEL_HANGUP_COMPLETE", ReqType: utils.META_PREPAID,
		Direction: utils.OUT, Tenant: "cgrates.org", Category: "call", Account: "1001", Subject: "1001",
		Destination: "1003", SetupTime: setupTime, AnswerTime: aTime,
		Usage: time.Duration(66) * time.Second, Pdd: time.Duration(28) * time.Millisecond, Supplier: "supplier1", DisconnectCause: "NORMAL_CLEARING", ExtraFields: make(map[string]string), Cost: -1}
	if storedCdr := ev.AsStoredCdr(""); !reflect.DeepEqual(eStoredCdr, storedCdr) {
		t.Errorf("Expecting: %+v, received: %+v", eStoredCdr, storedCdr)
	}
}

func TestFsEvGetExtraFields(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	cfg.SmFsConfig.ExtraFields = []*utils.RSRField{&utils.RSRField{Id: "Channel-Read-Codec-Name"}, &utils.RSRField{Id: "Channel-Write-Codec-Name"}, &utils.RSRField{Id: "NonExistingHeader"}}
	config.SetCgrConfig(cfg)
	ev := new(FSEvent).AsEvent(hangupEv)
	expectedExtraFields := map[string]string{"Channel-Read-Codec-Name": "SPEEX", "Channel-Write-Codec-Name": "SPEEX", "NonExistingHeader": ""}
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
}
