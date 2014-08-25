/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2013 ITsysCOM

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
	"github.com/cgrates/cgrates/utils"
)

var hangupEv string = `Event-Name: CHANNEL_HANGUP_COMPLETE
Core-UUID: bb890f9e-0aae-476d-8292-91b434eb4f73
FreeSWITCH-Hostname: iPBXDev
FreeSWITCH-Switchname: iPBXDev
FreeSWITCH-IPv4: 10.0.2.15
FreeSWITCH-IPv6: %3A%3A1
Event-Date-Local: 2014-04-25%2018%3A08%3A46
Event-Date-GMT: Fri,%2025%20Apr%202014%2016%3A08%3A46%20GMT
Event-Date-Timestamp: 1398442126033605
Event-Calling-File: switch_core_state_machine.c
Event-Calling-Function: switch_core_session_reporting_state
Event-Calling-Line-Number: 772
Event-Sequence: 3499
Hangup-Cause: NORMAL_CLEARING
Channel-State: CS_REPORTING
Channel-Call-State: HANGUP
Channel-State-Number: 11
Channel-Name: sofia/internal/1003%40192.168.56.66
Unique-ID: 37e9b766-5256-4e4b-b1ed-3767b930fec8
Call-Direction: inbound
Presence-Call-Direction: inbound
Channel-HIT-Dialplan: true
Channel-Presence-ID: 1003%40192.168.56.66
Channel-Call-UUID: 37e9b766-5256-4e4b-b1ed-3767b930fec8
Answer-State: hangup
Hangup-Cause: NORMAL_CLEARING
Channel-Read-Codec-Name: G722
Channel-Read-Codec-Rate: 16000
Channel-Read-Codec-Bit-Rate: 64000
Channel-Write-Codec-Name: G722
Channel-Write-Codec-Rate: 16000
Channel-Write-Codec-Bit-Rate: 64000
Caller-Direction: inbound
Caller-Username: 1003
Caller-Dialplan: XML
Caller-Caller-ID-Name: 1003
Caller-Caller-ID-Number: 1003
Caller-Orig-Caller-ID-Name: 1003
Caller-Orig-Caller-ID-Number: 1003
Caller-Callee-ID-Name: Outbound%20Call
Caller-Callee-ID-Number: 1002
Caller-Network-Addr: 192.168.56.1
Caller-ANI: 1003
Caller-Destination-Number: 1002
Caller-Unique-ID: 37e9b766-5256-4e4b-b1ed-3767b930fec8
Caller-Source: mod_sofia
Caller-Transfer-Source: 1398442107%3A93b23eed-7c33-49c8-a52d-f2b22b84e418%3Abl_xfer%3A1002/default/XML
Caller-Context: default
Caller-RDNIS: 1002
Caller-Channel-Name: sofia/internal/1003%40192.168.56.66
Caller-Profile-Index: 2
Caller-Profile-Created-Time: 1398442107850738
Caller-Channel-Created-Time: 1398442107770704
Caller-Channel-Answered-Time: 1398442120831856
Caller-Channel-Progress-Time: 1398442108013993
Caller-Channel-Progress-Media-Time: 1398442108050630
Caller-Channel-Hangup-Time: 1398442125950531
Caller-Channel-Transfer-Time: 0
Caller-Channel-Resurrect-Time: 0
Caller-Channel-Bridged-Time: 1398442120856148
Caller-Channel-Last-Hold: 1398442121113991
Caller-Channel-Hold-Accum: 0
Caller-Screen-Bit: true
Caller-Privacy-Hide-Name: false
Caller-Privacy-Hide-Number: false
Other-Type: originatee
Other-Leg-Direction: outbound
Other-Leg-Username: 1003
Other-Leg-Dialplan: XML
Other-Leg-Caller-ID-Name: Extension%201003
Other-Leg-Caller-ID-Number: 1003
Other-Leg-Orig-Caller-ID-Name: 1003
Other-Leg-Orig-Caller-ID-Number: 1003
Other-Leg-Callee-ID-Name: Outbound%20Call
Other-Leg-Callee-ID-Number: 1002
Other-Leg-Network-Addr: 192.168.56.1
Other-Leg-ANI: 1003
Other-Leg-Destination-Number: 1002
Other-Leg-Unique-ID: b7f3d830-b3a4-4e1c-b600-572eeb462c39
Other-Leg-Source: mod_sofia
Other-Leg-Context: default
Other-Leg-RDNIS: 1002
Other-Leg-Channel-Name: sofia/internal/sip%3A1002%40192.168.56.1%3A5060
Other-Leg-Profile-Created-Time: 1398442107970626
Other-Leg-Channel-Created-Time: 1398442107970626
Other-Leg-Channel-Answered-Time: 1398442120810530
Other-Leg-Channel-Progress-Time: 1398442108013993
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
variable_uuid: 37e9b766-5256-4e4b-b1ed-3767b930fec8
variable_session_id: 18
variable_sip_from_user: 1003
variable_sip_from_uri: 1003%40192.168.56.66
variable_sip_from_host: 192.168.56.66
variable_channel_name: sofia/internal/1003%40192.168.56.66
variable_sip_local_network_addr: 192.168.56.66
variable_sip_network_ip: 192.168.56.1
variable_sip_network_port: 5060
variable_sip_received_ip: 192.168.56.1
variable_sip_received_port: 5060
variable_sip_via_protocol: udp
variable_sip_authorized: true
variable_Event-Name: REQUEST_PARAMS
variable_Core-UUID: bb890f9e-0aae-476d-8292-91b434eb4f73
variable_FreeSWITCH-Hostname: iPBXDev
variable_FreeSWITCH-Switchname: iPBXDev
variable_FreeSWITCH-IPv4: 10.0.2.15
variable_FreeSWITCH-IPv6: %3A%3A1
variable_Event-Date-Local: 2014-04-25%2018%3A08%3A27
variable_Event-Date-GMT: Fri,%2025%20Apr%202014%2016%3A08%3A27%20GMT
variable_Event-Date-Timestamp: 1398442107770704
variable_Event-Calling-File: sofia.c
variable_Event-Calling-Function: sofia_handle_sip_i_invite
variable_Event-Calling-Line-Number: 7996
variable_Event-Sequence: 3355
variable_sip_number_alias: 1003
variable_sip_auth_username: 1003
variable_sip_auth_realm: 192.168.56.66
variable_number_alias: 1003
variable_requested_domain_name: 192.168.56.66
variable_record_stereo: true
variable_default_gateway: example.com
variable_default_areacode: 918
variable_transfer_fallback_extension: operator
variable_toll_allow: domestic,international,local
variable_accountcode: 1003
variable_user_context: default
variable_effective_caller_id_name: Extension%201003
variable_effective_caller_id_number: 1003
variable_outbound_caller_id_name: FreeSWITCH
variable_outbound_caller_id_number: 0000000000
variable_callgroup: techsupport
variable_user_name: 1003
variable_domain_name: 192.168.56.66
variable_sip_from_user_stripped: 1003
variable_sofia_profile_name: internal
variable_recovery_profile_name: internal
variable_sip_req_user: 1002
variable_sip_req_uri: 1002%40192.168.56.66
variable_sip_req_host: 192.168.56.66
variable_sip_to_user: 1002
variable_sip_to_uri: 1002%40192.168.56.66
variable_sip_to_host: 192.168.56.66
variable_sip_contact_params: transport%3Dudp%3Bregistering_acc%3D192_168_56_66
variable_sip_contact_user: 1003
variable_sip_contact_port: 5060
variable_sip_contact_uri: 1003%40192.168.56.1%3A5060
variable_sip_contact_host: 192.168.56.1
variable_sip_user_agent: Jitsi2.5.5065Linux
variable_sip_via_host: 192.168.56.1
variable_sip_via_port: 5060
variable_presence_id: 1003%40192.168.56.66
variable_ep_codec_string: G722%408000h%4020i%4064000b,PCMU%408000h%4020i%4064000b,PCMA%408000h%4020i%4064000b,GSM%408000h%4020i%4013200b
variable_cgr_notify: %2BAUTH_OK
variable_max_forwards: 69
variable_transfer_history: 1398442107%3A93b23eed-7c33-49c8-a52d-f2b22b84e418%3Abl_xfer%3A1002/default/XML
variable_transfer_source: 1398442107%3A93b23eed-7c33-49c8-a52d-f2b22b84e418%3Abl_xfer%3A1002/default/XML
variable_DP_MATCH: ARRAY%3A%3A1002%7C%3A1002
variable_call_uuid: 37e9b766-5256-4e4b-b1ed-3767b930fec8
variable_open: true
variable_RFC2822_DATE: Fri,%2025%20Apr%202014%2018%3A08%3A27%20%2B0200
variable_dialed_extension: 1002
variable_export_vars: RFC2822_DATE,RFC2822_DATE,dialed_extension
variable_ringback: %25(2000,4000,440,480)
variable_transfer_ringback: local_stream%3A//moh
variable_call_timeout: 30
variable_hangup_after_bridge: true
variable_continue_on_fail: true
variable_called_party_callgroup: techsupport
variable_current_application_data: user/1002%40192.168.56.66
variable_current_application: bridge
variable_dialed_user: 1002
variable_dialed_domain: 192.168.56.66
variable_inherit_codec: true
variable_originated_legs: ARRAY%3A%3Ab7f3d830-b3a4-4e1c-b600-572eeb462c39%3BOutbound%20Call%3B1002%7C%3Ab7f3d830-b3a4-4e1c-b600-572eeb462c39%3BOutbound%20Call%3B1002
variable_rtp_use_codec_string: G722,PCMU,PCMA,GSM
variable_sip_use_codec_name: G722
variable_sip_use_codec_rate: 8000
variable_sip_use_codec_ptime: 20
variable_write_codec: G722
variable_write_rate: 16000
variable_video_possible: true
variable_local_media_ip: 192.168.56.66
variable_local_media_port: 21546
variable_advertised_media_ip: 192.168.56.66
variable_sip_use_pt: 9
variable_rtp_use_ssrc: 2808137364
variable_zrtp_secure_media_confirmed_audio: true
variable_zrtp_sas1_string_audio: mqyn
variable_switch_m_sdp: v%3D0%0D%0Ao%3D1002%200%200%20IN%20IP4%20192.168.56.1%0D%0As%3D-%0D%0Ac%3DIN%20IP4%20192.168.56.1%0D%0At%3D0%200%0D%0Am%3Daudio%205056%20RTP/AVP%209%200%208%203%20101%0D%0Aa%3Drtpmap%3A9%20G722/8000%0D%0Aa%3Drtpmap%3A0%20PCMU/8000%0D%0Aa%3Drtpmap%3A8%20PCMA/8000%0D%0Aa%3Drtpmap%3A3%20GSM/8000%0D%0Aa%3Drtpmap%3A101%20telephone-event/8000%0D%0A
variable_read_codec: G722
variable_read_rate: 16000
variable_endpoint_disposition: ANSWER
variable_originate_causes: ARRAY%3A%3Ab7f3d830-b3a4-4e1c-b600-572eeb462c39%3BNONE%7C%3Ab7f3d830-b3a4-4e1c-b600-572eeb462c39%3BNONE
variable_originate_disposition: SUCCESS
variable_DIALSTATUS: SUCCESS
variable_last_bridge_to: b7f3d830-b3a4-4e1c-b600-572eeb462c39
variable_bridge_channel: sofia/internal/sip%3A1002%40192.168.56.1%3A5060
variable_bridge_uuid: b7f3d830-b3a4-4e1c-b600-572eeb462c39
variable_signal_bond: b7f3d830-b3a4-4e1c-b600-572eeb462c39
variable_cgr_reqtype: pseudoprepaid
variable_last_sent_callee_id_name: Outbound%20Call
variable_last_sent_callee_id_number: 1002
variable_sip_reinvite_sdp: v%3D0%0D%0Ao%3D1003%200%201%20IN%20IP4%20192.168.56.1%0D%0As%3D-%0D%0Ac%3DIN%20IP4%20192.168.56.1%0D%0At%3D0%200%0D%0Am%3Daudio%205052%20RTP/AVP%2096%2097%2098%209%20100%20102%200%208%20103%203%20104%20101%0D%0Aa%3Dsendonly%0D%0Aa%3Drtpmap%3A96%20opus/48000/2%0D%0Aa%3Dfmtp%3A96%20usedtx%3D1%0D%0Aa%3Drtpmap%3A97%20SILK/24000%0D%0Aa%3Drtpmap%3A98%20SILK/16000%0D%0Aa%3Drtpmap%3A9%20G722/8000%0D%0Aa%3Drtpmap%3A100%20speex/32000%0D%0Aa%3Drtpmap%3A102%20speex/16000%0D%0Aa%3Drtpmap%3A0%20PCMU/8000%0D%0Aa%3Drtpmap%3A8%20PCMA/8000%0D%0Aa%3Drtpmap%3A103%20iLBC/8000%0D%0Aa%3Drtpmap%3A3%20GSM/8000%0D%0Aa%3Drtpmap%3A104%20speex/8000%0D%0Aa%3Drtpmap%3A101%20telephone-event/8000%0D%0Aa%3Dextmap%3A1%20urn%3Aietf%3Aparams%3Artp-hdrext%3Acsrc-audio-level%0D%0Aa%3Dzrtp-hash%3A1.10%20bd7a58a0a6cb4b71870cc776f1901436f82ab3c9f960b9fc9645086206a8a804%0D%0Am%3Dvideo%200%20RTP/AVP%20105%2099%0D%0A
variable_switch_r_sdp: v%3D0%0D%0Ao%3D1003%200%201%20IN%20IP4%20192.168.56.1%0D%0As%3D-%0D%0Ac%3DIN%20IP4%20192.168.56.1%0D%0At%3D0%200%0D%0Am%3Daudio%205052%20RTP/AVP%2096%2097%2098%209%20100%20102%200%208%20103%203%20104%20101%0D%0Aa%3Drtpmap%3A96%20opus/48000/2%0D%0Aa%3Dfmtp%3A96%20usedtx%3D1%0D%0Aa%3Drtpmap%3A97%20SILK/24000%0D%0Aa%3Drtpmap%3A98%20SILK/16000%0D%0Aa%3Drtpmap%3A9%20G722/8000%0D%0Aa%3Drtpmap%3A100%20speex/32000%0D%0Aa%3Drtpmap%3A102%20speex/16000%0D%0Aa%3Drtpmap%3A0%20PCMU/8000%0D%0Aa%3Drtpmap%3A8%20PCMA/8000%0D%0Aa%3Drtpmap%3A103%20iLBC/8000%0D%0Aa%3Drtpmap%3A3%20GSM/8000%0D%0Aa%3Drtpmap%3A104%20speex/8000%0D%0Aa%3Drtpmap%3A101%20telephone-event/8000%0D%0Aa%3Dsendonly%0D%0Aa%3Dextmap%3A1%20urn%3Aietf%3Aparams%3Artp-hdrext%3Acsrc-audio-level%0D%0Aa%3Dzrtp-hash%3A1.10%20bd7a58a0a6cb4b71870cc776f1901436f82ab3c9f960b9fc9645086206a8a804%0D%0Am%3Dvideo%200%20RTP/AVP%20105%2099%0D%0A
variable_r_sdp_audio_zrtp_hash: 1.10%20bd7a58a0a6cb4b71870cc776f1901436f82ab3c9f960b9fc9645086206a8a804
variable_remote_media_ip: 192.168.56.1
variable_remote_media_port: 5052
variable_sip_audio_recv_pt: 9
variable_dtmf_type: rfc2833
variable_sip_2833_send_payload: 101
variable_sip_2833_recv_payload: 101
variable_sip_local_sdp_str: v%3D0%0Ao%3DFreeSWITCH%201398420562%201398420565%20IN%20IP4%20192.168.56.66%0As%3DFreeSWITCH%0Ac%3DIN%20IP4%20192.168.56.66%0At%3D0%200%0Am%3Daudio%2021546%20RTP/AVP%209%20101%0Aa%3Drtpmap%3A9%20G722/8000%0Aa%3Drtpmap%3A101%20telephone-event/8000%0Aa%3Dfmtp%3A101%200-16%0Aa%3Dptime%3A20%0Aa%3Dsendrecv%0A
variable_sip_to_tag: SUg05X4S6y5tQ
variable_sip_from_tag: 92f0bbcc
variable_sip_cseq: 3
variable_sip_call_id: 91a3940835793bb505003344ba6fc116%400%3A0%3A0%3A0%3A0%3A0%3A0%3A0
variable_sip_full_via: SIP/2.0/UDP%20192.168.56.1%3A5060%3Bbranch%3Dz9hG4bK-373830-8965548cad844d63bcb7a17c80e2e76f
variable_sip_from_display: 1003
variable_sip_full_from: %221003%22%20%3Csip%3A1003%40192.168.56.66%3E%3Btag%3D92f0bbcc
variable_sip_full_to: %3Csip%3A1002%40192.168.56.66%3E%3Btag%3DSUg05X4S6y5tQ
variable_sip_hangup_phrase: OK
variable_last_bridge_hangup_cause: NORMAL_CLEARING
variable_last_bridge_proto_specific_hangup_cause: sip%3A200
variable_bridge_hangup_cause: NORMAL_CLEARING
variable_hangup_cause: NORMAL_CLEARING
variable_hangup_cause_q850: 16
variable_digits_dialed: none
variable_start_stamp: 2014-04-25%2018%3A08%3A27
variable_profile_start_stamp: 2014-04-25%2018%3A08%3A27
variable_answer_stamp: 2014-04-25%2018%3A08%3A40
variable_bridge_stamp: 2014-04-25%2018%3A08%3A40
variable_hold_stamp: 2014-04-25%2018%3A08%3A41
variable_progress_stamp: 2014-04-25%2018%3A08%3A28
variable_progress_media_stamp: 2014-04-25%2018%3A08%3A28
variable_hold_events: %7B%7B1398442121114003,1398442125953702%7D%7D
variable_end_stamp: 2014-04-25%2018%3A08%3A45
variable_start_epoch: 1398442107
variable_start_uepoch: 1398442107770704
variable_profile_start_epoch: 1398442107
variable_profile_start_uepoch: 1398442107850738
variable_answer_epoch: 1398442120
variable_answer_uepoch: 1398442120831856
variable_bridge_epoch: 1398442120
variable_bridge_uepoch: 1398442120856148
variable_last_hold_epoch: 1398442121
variable_last_hold_uepoch: 1398442121113991
variable_hold_accum_seconds: 0
variable_hold_accum_usec: 0
variable_hold_accum_ms: 0
variable_resurrect_epoch: 0
variable_resurrect_uepoch: 0
variable_progress_epoch: 1398442108
variable_progress_uepoch: 1398442108013993
variable_progress_media_epoch: 1398442108
variable_progress_media_uepoch: 1398442108050630
variable_end_epoch: 1398442125
variable_end_uepoch: 1398442125950531
variable_last_app: bridge
variable_last_arg: user/1002%40192.168.56.66
variable_caller_id: %221003%22%20%3C1003%3E
variable_duration: 18
variable_billsec: 5
variable_progresssec: 1
variable_answersec: 13
variable_waitsec: 13
variable_progress_mediasec: 1
variable_flow_billsec: 18
variable_mduration: 18180
variable_billmsec: 5119
variable_progressmsec: 243
variable_answermsec: 13061
variable_waitmsec: 13086
variable_progress_mediamsec: 280
variable_flow_billmsec: 18180
variable_uduration: 18179827
variable_billusec: 5118675
variable_progressusec: 243289
variable_answerusec: 13061152
variable_waitusec: 13085444
variable_progress_mediausec: 279926
variable_flow_billusec: 18179827
variable_sip_hangup_disposition: send_bye
variable_rtp_audio_in_raw_bytes: 150072
variable_rtp_audio_in_media_bytes: 148136
variable_rtp_audio_in_packet_count: 854
variable_rtp_audio_in_media_packet_count: 843
variable_rtp_audio_in_skip_packet_count: 42
variable_rtp_audio_in_jb_packet_count: 0
variable_rtp_audio_in_dtmf_packet_count: 0
variable_rtp_audio_in_cng_packet_count: 0
variable_rtp_audio_in_flush_packet_count: 11
variable_rtp_audio_in_largest_jb_size: 0
variable_rtp_audio_out_raw_bytes: 140956
variable_rtp_audio_out_media_bytes: 140956
variable_rtp_audio_out_packet_count: 801
variable_rtp_audio_out_media_packet_count: 801
variable_rtp_audio_out_skip_packet_count: 0
variable_rtp_audio_out_dtmf_packet_count: 0
variable_rtp_audio_out_cng_packet_count: 0
variable_rtp_audio_rtcp_packet_count: 0
variable_rtp_audio_rtcp_octet_count: 0
`

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
	ev := new(FSEvent).New(body)
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
	ev := new(FSEvent).New("")
	setupTime, _ := ev.GetSetupTime("^2013-12-07 08:42:24")
	answerTime, _ := ev.GetAnswerTime("^2013-12-07 08:42:24")
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
		dur != time.Duration(60)*time.Second {
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
	cfg, _ = config.NewDefaultCGRConfig()
	config.SetCgrConfig(cfg)
	ev := new(FSEvent).New(body)
	setupTime, _ := ev.GetSetupTime("Event-Date-Local")
	answerTime, _ := ev.GetAnswerTime("Event-Date-Local")
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
		dur != time.Duration(65)*time.Second {
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
	ev := new(FSEvent).New(body)
	if setupTime, err := ev.GetSetupTime(""); err != nil {
		t.Error("Error when parsing empty setupTime")
	} else if setupTime != nilTime {
		t.Error("Expecting nil time, got: ", setupTime)
	}
	if answerTime, err := ev.GetAnswerTime(""); err != nil {
		t.Error("Error when parsing empty setupTime")
	} else if answerTime != nilTime {
		t.Error("Expecting nil time, got: ", answerTime)
	}
}

func TestParseFsHangup(t *testing.T) {
	cfg, _ = config.NewDefaultCGRConfig()
	config.SetCgrConfig(cfg)
	ev := new(FSEvent).New(hangupEv)
	setupTime, _ := ev.GetSetupTime(utils.META_DEFAULT)
	answerTime, _ := ev.GetAnswerTime(utils.META_DEFAULT)
	dur, _ := ev.GetDuration(utils.META_DEFAULT)
	if ev.GetReqType(utils.META_DEFAULT) != utils.PSEUDOPREPAID ||
		ev.GetDirection(utils.META_DEFAULT) != "*out" ||
		ev.GetTenant(utils.META_DEFAULT) != "cgrates.org" ||
		ev.GetCategory(utils.META_DEFAULT) != "call" ||
		ev.GetAccount(utils.META_DEFAULT) != "1003" ||
		ev.GetSubject(utils.META_DEFAULT) != "1003" ||
		ev.GetDestination(utils.META_DEFAULT) != "1002" ||
		setupTime.UTC() != time.Date(2014, 4, 25, 16, 8, 27, 0, time.UTC) ||
		answerTime.UTC() != time.Date(2014, 4, 25, 16, 8, 40, 0, time.UTC) ||
		dur != time.Duration(5)*time.Second {
		t.Error("Default values not matching",
			ev.GetReqType(utils.META_DEFAULT) != utils.PSEUDOPREPAID,
			ev.GetDirection(utils.META_DEFAULT) != "*out",
			ev.GetTenant(utils.META_DEFAULT) != "cgrates.org",
			ev.GetCategory(utils.META_DEFAULT) != "call",
			ev.GetAccount(utils.META_DEFAULT) != "1003",
			ev.GetSubject(utils.META_DEFAULT) != "1003",
			ev.GetDestination(utils.META_DEFAULT) != "1002",
			setupTime.UTC() != time.Date(2014, 4, 25, 17, 8, 27, 0, time.UTC),
			answerTime.UTC() != time.Date(2014, 4, 25, 17, 8, 40, 0, time.UTC),
			dur != time.Duration(5)*time.Second)
	}
}

func TestParseEventValue(t *testing.T) {
	cfg, _ = config.NewDefaultCGRConfig()
	config.SetCgrConfig(cfg)
	ev := new(FSEvent).New(hangupEv)
	if cgrid := ev.ParseEventValue(&utils.RSRField{Id: utils.CGRID}); cgrid != "873e5bf7903978f305f7d8fed3f92f968cf82873" {
		t.Error("Unexpected cgrid parsed", cgrid)
	}
	if tor := ev.ParseEventValue(&utils.RSRField{Id: utils.TOR}); tor != utils.VOICE {
		t.Error("Unexpected tor parsed", tor)
	}
	if accid := ev.ParseEventValue(&utils.RSRField{Id: utils.ACCID}); accid != "37e9b766-5256-4e4b-b1ed-3767b930fec8" {
		t.Error("Unexpected result parsed", accid)
	}
	if parsed := ev.ParseEventValue(&utils.RSRField{Id: utils.CDRHOST}); parsed != "10.0.2.15" {
		t.Error("Unexpected result parsed", parsed)
	}
	if parsed := ev.ParseEventValue(&utils.RSRField{Id: utils.CDRSOURCE}); parsed != "FS_EVENT" {
		t.Error("Unexpected result parsed", parsed)
	}
	if parsed := ev.ParseEventValue(&utils.RSRField{Id: utils.REQTYPE}); parsed != utils.PSEUDOPREPAID {
		t.Error("Unexpected result parsed", parsed)
	}
	if parsed := ev.ParseEventValue(&utils.RSRField{Id: utils.DIRECTION}); parsed != utils.OUT {
		t.Error("Unexpected result parsed", parsed)
	}
	if parsed := ev.ParseEventValue(&utils.RSRField{Id: utils.TENANT}); parsed != "cgrates.org" {
		t.Error("Unexpected result parsed", parsed)
	}
	if parsed := ev.ParseEventValue(&utils.RSRField{Id: utils.CATEGORY}); parsed != "call" {
		t.Error("Unexpected result parsed", parsed)
	}
	if parsed := ev.ParseEventValue(&utils.RSRField{Id: utils.ACCOUNT}); parsed != "1003" {
		t.Error("Unexpected result parsed", parsed)
	}
	if parsed := ev.ParseEventValue(&utils.RSRField{Id: utils.SUBJECT}); parsed != "1003" {
		t.Error("Unexpected result parsed", parsed)
	}
	if parsed := ev.ParseEventValue(&utils.RSRField{Id: utils.DESTINATION}); parsed != "1002" {
		t.Error("Unexpected result parsed", parsed)
	}
	sTime, _ := utils.ParseTimeDetectLayout("1398442107770704"[:len("1398442107770704")-6]) // We discard nanoseconds information so we can correlate csv
	if parsed := ev.ParseEventValue(&utils.RSRField{Id: utils.SETUP_TIME}); parsed != sTime.String() {
		t.Errorf("Expecting: %s, parsed: %s", sTime.String(), parsed)
	}
	aTime, _ := utils.ParseTimeDetectLayout("1398442120831856"[:len("1398442120831856")-6])
	if parsed := ev.ParseEventValue(&utils.RSRField{Id: utils.ANSWER_TIME}); parsed != aTime.String() {
		t.Errorf("Expecting: %s, parsed: %s", aTime.String(), parsed)
	}
	if parsed := ev.ParseEventValue(&utils.RSRField{Id: utils.USAGE}); parsed != "5000000000" {
		t.Error("Unexpected result parsed", parsed)
	}
	if parsed := ev.ParseEventValue(&utils.RSRField{Id: utils.MEDI_RUNID}); parsed != utils.DEFAULT_RUNID {
		t.Error("Unexpected result parsed", parsed)
	}
	if parsed := ev.ParseEventValue(&utils.RSRField{Id: utils.COST}); parsed != "-1" {
		t.Error("Unexpected result parsed", parsed)
	}
	if parsed := ev.ParseEventValue(&utils.RSRField{Id: "Hangup-Cause"}); parsed != "NORMAL_CLEARING" {
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
	ev := new(FSEvent).New(body)
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
	ev = new(FSEvent).New(body)
	if pass, _ := ev.PassesFieldFilter(acntPrefxFltr); pass {
		t.Error("Should not pass filter")
	}
	body = `Event-Name: RE_SCHEDULE
Core-UUID: 792e181c-b6e6-499c-82a1-52a778e7d82d
FreeSWITCH-Hostname: h1.ip-switch.net
FreeSWITCH-Switchname: h1.ip-switch.net
FreeSWITCH-IPv4: 88.198.12.156
Caller-Username: 0402129281`
	ev = new(FSEvent).New(body)
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
	ev = new(FSEvent).New(body)
	if pass, _ := ev.PassesFieldFilter(acntPrefxFltr); pass {
		t.Error("Should not pass filter")
	}
}

func TestFsEvAsStoredCdr(t *testing.T) {
	cfg, _ = config.NewDefaultCGRConfig()
	config.SetCgrConfig(cfg)
	ev := new(FSEvent).New(hangupEv)
	setupTime, _ := utils.ParseTimeDetectLayout("1398442107")
	aTime, _ := utils.ParseTimeDetectLayout("1398442120")
	eStoredCdr := &utils.StoredCdr{CgrId: utils.Sha1("37e9b766-5256-4e4b-b1ed-3767b930fec8", setupTime.UTC().String()),
		TOR: utils.VOICE, AccId: "37e9b766-5256-4e4b-b1ed-3767b930fec8", CdrHost: "10.0.2.15", CdrSource: "FS_CHANNEL_HANGUP_COMPLETE", ReqType: utils.PSEUDOPREPAID,
		Direction: utils.OUT, Tenant: "cgrates.org", Category: "call", Account: "1003", Subject: "1003",
		Destination: "1002", SetupTime: setupTime, AnswerTime: aTime,
		Usage: time.Duration(5) * time.Second, ExtraFields: make(map[string]string), Cost: -1}
	if storedCdr := ev.AsStoredCdr(); !reflect.DeepEqual(eStoredCdr, storedCdr) {
		t.Errorf("Expecting: %+v, received: %+v", eStoredCdr, storedCdr)
	}
}

func TestFsEvGetExtraFields(t *testing.T) {
	cfg, _ = config.NewDefaultCGRConfig()
	cfg.FSCdrExtraFields = []*utils.RSRField{&utils.RSRField{Id: "Channel-Read-Codec-Name"}, &utils.RSRField{Id: "Channel-Write-Codec-Name"}, &utils.RSRField{Id: "NonExistingHeader"}}
	config.SetCgrConfig(cfg)
	ev := new(FSEvent).New(hangupEv)
	expectedExtraFields := map[string]string{"Channel-Read-Codec-Name": "G722", "Channel-Write-Codec-Name": "G722", "NonExistingHeader": ""}
	if extraFields := ev.GetExtraFields(); !reflect.DeepEqual(expectedExtraFields, extraFields) {
		t.Errorf("Expecting: %+v, received: %+v", expectedExtraFields, extraFields)
	}
}
