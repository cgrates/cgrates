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
package general_tests

import (
	"testing"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/sessionmanager"
)

var answerEvent = `Event-Name: CHANNEL_ANSWER
Core-UUID: feef0b51-7fdf-4c4a-878e-aff233752de2
FreeSWITCH-Hostname: CGRTest
FreeSWITCH-Switchname: CGRTest
FreeSWITCH-IPv4: 192.168.178.32
FreeSWITCH-IPv6: %3A%3A1
Event-Date-Local: 2014-04-08%2021%3A10%3A27
Event-Date-GMT: Tue,%2008%20Apr%202014%2019%3A10%3A27%20GMT
Event-Date-Timestamp: 1396984227717006
Event-Calling-File: switch_channel.c
Event-Calling-Function: switch_channel_perform_mark_answered
Event-Calling-Line-Number: 3591
Event-Sequence: 1524
Channel-State: CS_EXECUTE
Channel-Call-State: EARLY
Channel-State-Number: 4
Channel-Name: sofia/internal/1001%40192.168.56.74
Unique-ID: 86cfd6e2-dbda-45a3-b59d-f683ec368e8b
Call-Direction: inbound
Presence-Call-Direction: inbound
Channel-HIT-Dialplan: true
Channel-Presence-ID: 1001%40192.168.56.74
Channel-Call-UUID: 86cfd6e2-dbda-45a3-b59d-f683ec368e8b
Answer-State: answered
Channel-Read-Codec-Name: G722
Channel-Read-Codec-Rate: 16000
Channel-Read-Codec-Bit-Rate: 64000
Channel-Write-Codec-Name: G722
Channel-Write-Codec-Rate: 16000
Channel-Write-Codec-Bit-Rate: 64000
Caller-Direction: inbound
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
Caller-Unique-ID: 86cfd6e2-dbda-45a3-b59d-f683ec368e8b
Caller-Source: mod_sofia
Caller-Transfer-Source: 1396984221%3Acaefc538-5da4-4245-8716-112c706383d8%3Abl_xfer%3A1002/default/XML
Caller-Context: default
Caller-RDNIS: 1002
Caller-Channel-Name: sofia/internal/1001%40192.168.56.74
Caller-Profile-Index: 2
Caller-Profile-Created-Time: 1396984221377035
Caller-Channel-Created-Time: 1396984221278217
Caller-Channel-Answered-Time: 1396984227717006
Caller-Channel-Progress-Time: 1396984221497331
Caller-Channel-Progress-Media-Time: 1396984221517042
Caller-Channel-Hangup-Time: 0
Caller-Channel-Transfer-Time: 0
Caller-Channel-Resurrect-Time: 0
Caller-Channel-Bridged-Time: 0
Caller-Channel-Last-Hold: 0
Caller-Channel-Hold-Accum: 0
Caller-Screen-Bit: true
Caller-Privacy-Hide-Name: false
Caller-Privacy-Hide-Number: false
variable_direction: inbound
variable_uuid: 86cfd6e2-dbda-45a3-b59d-f683ec368e8b
variable_session_id: 5
variable_sip_from_user: 1001
variable_sip_from_uri: 1001%40192.168.56.74
variable_sip_from_host: 192.168.56.74
variable_channel_name: sofia/internal/1001%40192.168.56.74
variable_sip_call_id: 6691dbf8ffdc02bdacee02bc305d5c71%400%3A0%3A0%3A0%3A0%3A0%3A0%3A0
variable_sip_local_network_addr: 192.168.56.74
variable_sip_network_ip: 192.168.56.1
variable_sip_network_port: 5060
variable_sip_received_ip: 192.168.56.1
variable_sip_received_port: 5060
variable_sip_via_protocol: udp
variable_sip_authorized: true
variable_Event-Name: REQUEST_PARAMS
variable_Core-UUID: feef0b51-7fdf-4c4a-878e-aff233752de2
variable_FreeSWITCH-Hostname: CGRTest
variable_FreeSWITCH-Switchname: CGRTest
variable_FreeSWITCH-IPv4: 192.168.178.32
variable_FreeSWITCH-IPv6: %3A%3A1
variable_Event-Date-Local: 2014-04-08%2021%3A10%3A21
variable_Event-Date-GMT: Tue,%2008%20Apr%202014%2019%3A10%3A21%20GMT
variable_Event-Date-Timestamp: 1396984221278217
variable_Event-Calling-File: sofia.c
variable_Event-Calling-Function: sofia_handle_sip_i_invite
variable_Event-Calling-Line-Number: 8076
variable_Event-Sequence: 1423
variable_sip_number_alias: 1001
variable_sip_auth_username: 1001
variable_sip_auth_realm: 192.168.56.74
variable_number_alias: 1001
variable_requested_domain_name: 192.168.56.66
variable_record_stereo: true
variable_default_gateway: example.com
variable_default_areacode: 918
variable_transfer_fallback_extension: operator
variable_toll_allow: domestic,international,local
variable_accountcode: 1001
variable_user_context: default
variable_effective_caller_id_name: Extension%201001
variable_effective_caller_id_number: 1001
variable_outbound_caller_id_name: FreeSWITCH
variable_outbound_caller_id_number: 0000000000
variable_callgroup: techsupport
variable_cgr_RequestType: *prepaid
variable_user_name: 1001
variable_domain_name: 192.168.56.66
variable_sip_from_user_stripped: 1001
variable_sip_from_tag: 1afc7eca
variable_sofia_profile_name: internal
variable_recovery_profile_name: internal
variable_sip_full_via: SIP/2.0/UDP%20192.168.56.1%3A5060%3Bbranch%3Dz9hG4bK-323133-4b7ccf74fda61ef6d189ba7a6cc67465
variable_sip_from_display: 1001
variable_sip_full_from: %221001%22%20%3Csip%3A1001%40192.168.56.74%3E%3Btag%3D1afc7eca
variable_sip_full_to: %3Csip%3A1002%40192.168.56.74%3E
variable_sip_req_user: 1002
variable_sip_req_uri: 1002%40192.168.56.74
variable_sip_req_host: 192.168.56.74
variable_sip_to_user: 1002
variable_sip_to_uri: 1002%40192.168.56.74
variable_sip_to_host: 192.168.56.74
variable_sip_contact_params: transport%3Dudp%3Bregistering_acc%3D192_168_56_74
variable_sip_contact_user: 1001
variable_sip_contact_port: 5060
variable_sip_contact_uri: 1001%40192.168.56.1%3A5060
variable_sip_contact_host: 192.168.56.1
variable_sip_user_agent: Jitsi2.5.5065Linux
variable_sip_via_host: 192.168.56.1
variable_sip_via_port: 5060
variable_presence_id: 1001%40192.168.56.74
variable_switch_r_sdp: v%3D0%0D%0Ao%3D1001%200%200%20IN%20IP4%20192.168.56.1%0D%0As%3D-%0D%0Ac%3DIN%20IP4%20192.168.56.1%0D%0At%3D0%200%0D%0Am%3Daudio%205016%20RTP/AVP%2096%2097%2098%209%20100%20102%200%208%20103%203%20104%20101%0D%0Aa%3Drtpmap%3A96%20opus/48000/2%0D%0Aa%3Dfmtp%3A96%20usedtx%3D1%0D%0Aa%3Drtpmap%3A97%20SILK/24000%0D%0Aa%3Drtpmap%3A98%20SILK/16000%0D%0Aa%3Drtpmap%3A9%20G722/8000%0D%0Aa%3Drtpmap%3A100%20speex/32000%0D%0Aa%3Drtpmap%3A102%20speex/16000%0D%0Aa%3Drtpmap%3A0%20PCMU/8000%0D%0Aa%3Drtpmap%3A8%20PCMA/8000%0D%0Aa%3Drtpmap%3A103%20iLBC/8000%0D%0Aa%3Drtpmap%3A3%20GSM/8000%0D%0Aa%3Drtpmap%3A104%20speex/8000%0D%0Aa%3Drtpmap%3A101%20telephone-event/8000%0D%0Aa%3Dextmap%3A1%20urn%3Aietf%3Aparams%3Artp-hdrext%3Acsrc-audio-level%0D%0Am%3Dvideo%205018%20RTP/AVP%20105%2099%0D%0Aa%3Drtpmap%3A105%20H264/90000%0D%0Aa%3Dfmtp%3A105%20profile-level-id%3D4DE01f%3Bpacketization-mode%3D1%0D%0Aa%3Drtpmap%3A99%20H264/90000%0D%0Aa%3Dfmtp%3A99%20profile-level-id%3D4DE01f%0D%0Aa%3Drecvonly%0D%0Aa%3Dimageattr%3A105%20send%20*%20recv%20%5Bx%3D%5B0-1920%5D,y%3D%5B0-1080%5D%5D%0D%0Aa%3Dimageattr%3A99%20send%20*%20recv%20%5Bx%3D%5B0-1920%5D,y%3D%5B0-1080%5D%5D%0D%0A
variable_ep_codec_string: G722%408000h%4020i%4064000b,PCMU%408000h%4020i%4064000b,PCMA%408000h%4020i%4064000b,GSM%408000h%4020i%4013200b
variable_cgr_notify: %2BAUTH_OK
variable_max_forwards: 69
variable_transfer_history: 1396984221%3Acaefc538-5da4-4245-8716-112c706383d8%3Abl_xfer%3A1002/default/XML
variable_transfer_source: 1396984221%3Acaefc538-5da4-4245-8716-112c706383d8%3Abl_xfer%3A1002/default/XML
variable_DP_MATCH: ARRAY%3A%3A1002%7C%3A1002
variable_call_uuid: 86cfd6e2-dbda-45a3-b59d-f683ec368e8b
variable_RFC2822_DATE: Tue,%2008%20Apr%202014%2021%3A10%3A21%20%2B0200
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
variable_originate_disposition: failure
variable_DIALSTATUS: INVALIDARGS
variable_inherit_codec: true
variable_originate_signal_bond: 402f0929-fa14-4a5f-9642-3a1311bb4ddd
variable_originated_legs: 402f0929-fa14-4a5f-9642-3a1311bb4ddd%3BOutbound%20Call%3B1002
variable_rtp_use_codec_string: G722,PCMU,PCMA,GSM
variable_sip_audio_recv_pt: 9
variable_sip_use_codec_name: G722
variable_sip_use_codec_rate: 8000
variable_sip_use_codec_ptime: 20
variable_write_codec: G722
variable_write_rate: 16000
variable_video_possible: true
variable_local_media_ip: 192.168.56.74
variable_local_media_port: 32534
variable_advertised_media_ip: 192.168.56.74
variable_sip_use_pt: 9
variable_rtp_use_ssrc: 1431080133
variable_sip_2833_send_payload: 101
variable_sip_2833_recv_payload: 101
variable_remote_media_ip: 192.168.56.1
variable_remote_media_port: 5016
variable_endpoint_disposition: EARLY%20MEDIA
variable_zrtp_secure_media_confirmed_audio: true
variable_zrtp_sas1_string_audio: j6ff
variable_switch_m_sdp: v%3D0%0D%0Ao%3D1002%200%200%20IN%20IP4%20192.168.56.1%0D%0As%3D-%0D%0Ac%3DIN%20IP4%20192.168.56.1%0D%0At%3D0%200%0D%0Am%3Daudio%205020%20RTP/AVP%209%200%208%203%20101%0D%0Aa%3Drtpmap%3A9%20G722/8000%0D%0Aa%3Drtpmap%3A0%20PCMU/8000%0D%0Aa%3Drtpmap%3A8%20PCMA/8000%0D%0Aa%3Drtpmap%3A3%20GSM/8000%0D%0Aa%3Drtpmap%3A101%20telephone-event/8000%0D%0A
variable_read_codec: G722
variable_read_rate: 16000
variable_sip_local_sdp_str: v%3D0%0Ao%3DFreeSWITCH%201396951687%201396951689%20IN%20IP4%20192.168.56.74%0As%3DFreeSWITCH%0Ac%3DIN%20IP4%20192.168.56.74%0At%3D0%200%0Am%3Daudio%2032534%20RTP/AVP%209%20101%0Aa%3Drtpmap%3A9%20G722/8000%0Aa%3Drtpmap%3A101%20telephone-event/8000%0Aa%3Dfmtp%3A101%200-16%0Aa%3Dptime%3A20%0Aa%3Dsendrecv%0A`

var jsonCdr = []byte(`{"core-uuid":"feef0b51-7fdf-4c4a-878e-aff233752de2","channel_data":{"state":"CS_REPORTING","direction":"inbound","state_number":"11","flags":"0=1;1=1;3=1;36=1;37=1;39=1;42=1;47=1;52=1;73=1;75=1;94=1","caps":"1=1;2=1;3=1;4=1;5=1;6=1"},"variables":{"direction":"inbound","uuid":"86cfd6e2-dbda-45a3-b59d-f683ec368e8b","session_id":"5","sip_from_user":"1001","sip_from_uri":"1001@192.168.56.74","sip_from_host":"192.168.56.74","channel_name":"sofia/internal/1001@192.168.56.74","sip_local_network_addr":"192.168.56.74","sip_network_ip":"192.168.56.1","sip_network_port":"5060","sip_received_ip":"192.168.56.1","sip_received_port":"5060","sip_via_protocol":"udp","sip_authorized":"true","Event-Name":"REQUEST_PARAMS","Core-UUID":"feef0b51-7fdf-4c4a-878e-aff233752de2","FreeSWITCH-Hostname":"CGRTest","FreeSWITCH-Switchname":"CGRTest","FreeSWITCH-IPv4":"192.168.178.32","FreeSWITCH-IPv6":"::1","Event-Date-Local":"2014-04-08 21:10:21","Event-Date-GMT":"Tue, 08 Apr 2014 19:10:21 GMT","Event-Date-Timestamp":"1396984221278217","Event-Calling-File":"sofia.c","Event-Calling-Function":"sofia_handle_sip_i_invite","Event-Calling-Line-Number":"8076","Event-Sequence":"1423","sip_number_alias":"1001","sip_auth_username":"1001","sip_auth_realm":"192.168.56.74","number_alias":"1001","requested_domain_name":"192.168.56.66","record_stereo":"true","default_gateway":"example.com","default_areacode":"918","transfer_fallback_extension":"operator","toll_allow":"domestic,international,local","accountcode":"1001","user_context":"default","effective_caller_id_name":"Extension 1001","effective_caller_id_number":"1001","outbound_caller_id_name":"FreeSWITCH","outbound_caller_id_number":"0000000000","callgroup":"techsupport","user_name":"1001","domain_name":"192.168.56.66","sip_from_user_stripped":"1001","sofia_profile_name":"internal","recovery_profile_name":"internal","sip_req_user":"1002","sip_req_uri":"1002@192.168.56.74","sip_req_host":"192.168.56.74","sip_to_user":"1002","sip_to_uri":"1002@192.168.56.74","sip_to_host":"192.168.56.74","sip_contact_params":"transport=udp;registering_acc=192_168_56_74","sip_contact_user":"1001","sip_contact_port":"5060","sip_contact_uri":"1001@192.168.56.1:5060","sip_contact_host":"192.168.56.1","sip_via_host":"192.168.56.1","sip_via_port":"5060","presence_id":"1001@192.168.56.74","ep_codec_string":"G722@8000h@20i@64000b,PCMU@8000h@20i@64000b,PCMA@8000h@20i@64000b,GSM@8000h@20i@13200b","cgr_notify":"+AUTH_OK","max_forwards":"69","transfer_history":"1396984221:caefc538-5da4-4245-8716-112c706383d8:bl_xfer:1002/default/XML","transfer_source":"1396984221:caefc538-5da4-4245-8716-112c706383d8:bl_xfer:1002/default/XML","DP_MATCH":"ARRAY::1002|:1002","call_uuid":"86cfd6e2-dbda-45a3-b59d-f683ec368e8b","RFC2822_DATE":"Tue, 08 Apr 2014 21:10:21 +0200","dialed_extension":"1002","export_vars":"RFC2822_DATE,RFC2822_DATE,dialed_extension","ringback":"%(2000,4000,440,480)","transfer_ringback":"local_stream://moh","call_timeout":"30","hangup_after_bridge":"true","continue_on_fail":"true","called_party_callgroup":"techsupport","current_application_data":"user/1002@192.168.56.66","current_application":"bridge","dialed_user":"1002","dialed_domain":"192.168.56.66","inherit_codec":"true","originated_legs":"ARRAY::402f0929-fa14-4a5f-9642-3a1311bb4ddd;Outbound Call;1002|:402f0929-fa14-4a5f-9642-3a1311bb4ddd;Outbound Call;1002","rtp_use_codec_string":"G722,PCMU,PCMA,GSM","sip_use_codec_name":"G722","sip_use_codec_rate":"8000","sip_use_codec_ptime":"20","write_codec":"G722","write_rate":"16000","video_possible":"true","local_media_ip":"192.168.56.74","local_media_port":"32534","advertised_media_ip":"192.168.56.74","sip_use_pt":"9","rtp_use_ssrc":"1431080133","zrtp_secure_media_confirmed_audio":"true","zrtp_sas1_string_audio":"j6ff","switch_m_sdp":"v=0\r\no=1002 0 0 IN IP4 192.168.56.1\r\ns=-\r\nc=IN IP4 192.168.56.1\r\nt=0 0\r\nm=audio 5020 RTP/AVP 9 0 8 3 101\r\na=rtpmap:9 G722/8000\r\na=rtpmap:0 PCMU/8000\r\na=rtpmap:8 PCMA/8000\r\na=rtpmap:3 GSM/8000\r\na=rtpmap:101 telephone-event/8000\r\n","read_codec":"G722","read_rate":"16000","endpoint_disposition":"ANSWER","originate_causes":"ARRAY::402f0929-fa14-4a5f-9642-3a1311bb4ddd;NONE|:402f0929-fa14-4a5f-9642-3a1311bb4ddd;NONE","originate_disposition":"SUCCESS","DIALSTATUS":"SUCCESS","last_bridge_to":"402f0929-fa14-4a5f-9642-3a1311bb4ddd","bridge_channel":"sofia/internal/sip:1002@192.168.56.1:5060","bridge_uuid":"402f0929-fa14-4a5f-9642-3a1311bb4ddd","signal_bond":"402f0929-fa14-4a5f-9642-3a1311bb4ddd","last_sent_callee_id_name":"Outbound Call","last_sent_callee_id_number":"1002","cgr_RequestType":"*prepaid","sip_reinvite_sdp":"v=0\r\no=1001 0 1 IN IP4 192.168.56.1\r\ns=-\r\nc=IN IP4 192.168.56.1\r\nt=0 0\r\nm=audio 5016 RTP/AVP 96 97 98 9 100 102 0 8 103 3 104 101\r\na=sendonly\r\na=rtpmap:96 opus/48000/2\r\na=fmtp:96 usedtx=1\r\na=rtpmap:97 SILK/24000\r\na=rtpmap:98 SILK/16000\r\na=rtpmap:9 G722/8000\r\na=rtpmap:100 speex/32000\r\na=rtpmap:102 speex/16000\r\na=rtpmap:0 PCMU/8000\r\na=rtpmap:8 PCMA/8000\r\na=rtpmap:103 iLBC/8000\r\na=rtpmap:3 GSM/8000\r\na=rtpmap:104 speex/8000\r\na=rtpmap:101 telephone-event/8000\r\na=extmap:1 urn:ietf:params:rtp-hdrext:csrc-audio-level\r\na=zrtp-hash:1.10 722d57097aaabea2749ea8938472478f8d88645b23521fa5f8005a7a2bed8286\r\nm=video 0 RTP/AVP 105 99\r\n","switch_r_sdp":"v=0\r\no=1001 0 1 IN IP4 192.168.56.1\r\ns=-\r\nc=IN IP4 192.168.56.1\r\nt=0 0\r\nm=audio 5016 RTP/AVP 96 97 98 9 100 102 0 8 103 3 104 101\r\na=rtpmap:96 opus/48000/2\r\na=fmtp:96 usedtx=1\r\na=rtpmap:97 SILK/24000\r\na=rtpmap:98 SILK/16000\r\na=rtpmap:9 G722/8000\r\na=rtpmap:100 speex/32000\r\na=rtpmap:102 speex/16000\r\na=rtpmap:0 PCMU/8000\r\na=rtpmap:8 PCMA/8000\r\na=rtpmap:103 iLBC/8000\r\na=rtpmap:3 GSM/8000\r\na=rtpmap:104 speex/8000\r\na=rtpmap:101 telephone-event/8000\r\na=sendonly\r\na=extmap:1 urn:ietf:params:rtp-hdrext:csrc-audio-level\r\na=zrtp-hash:1.10 722d57097aaabea2749ea8938472478f8d88645b23521fa5f8005a7a2bed8286\r\nm=video 0 RTP/AVP 105 99\r\n","r_sdp_audio_zrtp_hash":"1.10 722d57097aaabea2749ea8938472478f8d88645b23521fa5f8005a7a2bed8286","remote_media_ip":"192.168.56.1","remote_media_port":"5016","sip_audio_recv_pt":"9","dtmf_type":"rfc2833","sip_2833_send_payload":"101","sip_2833_recv_payload":"101","sip_local_sdp_str":"v=0\no=FreeSWITCH 1396951687 1396951690 IN IP4 192.168.56.74\ns=FreeSWITCH\nc=IN IP4 192.168.56.74\nt=0 0\nm=audio 32534 RTP/AVP 9 101\na=rtpmap:9 G722/8000\na=rtpmap:101 telephone-event/8000\na=fmtp:101 0-16\na=ptime:20\na=sendrecv\n","sip_to_tag":"rXc9vZpv9eFaF","sip_from_tag":"1afc7eca","sip_cseq":"3","sip_call_id":"6691dbf8ffdc02bdacee02bc305d5c71@0:0:0:0:0:0:0:0","sip_full_via":"SIP/2.0/UDP 192.168.56.1:5060;branch=z9hG4bK-323133-5d083abc0d3f327b9101586e71b5fce4","sip_from_display":"1001","sip_full_from":"\"1001\" <sip:1001@192.168.56.74>;tag=1afc7eca","sip_full_to":"<sip:1002@192.168.56.74>;tag=rXc9vZpv9eFaF","sip_term_status":"200","proto_specific_hangup_cause":"sip:200","sip_term_cause":"16","last_bridge_role":"originator","sip_user_agent":"Jitsi2.5.5065Linux","sip_hangup_disposition":"recv_bye","bridge_hangup_cause":"NORMAL_CLEARING","hangup_cause":"NORMAL_CLEARING","hangup_cause_q850":"16","digits_dialed":"none","start_stamp":"2014-04-08 21:10:21","profile_start_stamp":"2014-04-08 21:10:21","answer_stamp":"2014-04-08 21:10:27","bridge_stamp":"2014-04-08 21:10:27","hold_stamp":"2014-04-08 21:10:27","progress_stamp":"2014-04-08 21:10:21","progress_media_stamp":"2014-04-08 21:10:21","hold_events":"{{1396984227824182,1396984242247995}}","end_stamp":"2014-04-08 21:10:42","start_epoch":"1396984221","start_uepoch":"1396984221278217","profile_start_epoch":"1396984221","profile_start_uepoch":"1396984221377035","answer_epoch":"1396984227","answer_uepoch":"1396984227717006","bridge_epoch":"1396984227","bridge_uepoch":"1396984227737268","last_hold_epoch":"1396984227","last_hold_uepoch":"1396984227824167","hold_accum_seconds":"14","hold_accum_usec":"14423816","hold_accum_ms":"14423","resurrect_epoch":"0","resurrect_uepoch":"0","progress_epoch":"1396984221","progress_uepoch":"1396984221497331","progress_media_epoch":"1396984221","progress_media_uepoch":"1396984221517042","end_epoch":"1396984242","end_uepoch":"1396984242257026","last_app":"bridge","last_arg":"user/1002@192.168.56.66","caller_id":"\"1001\" <1001>","duration":"21","billsec":"15","progresssec":"0","answersec":"6","waitsec":"6","progress_mediasec":"0","flow_billsec":"21","mduration":"20979","billmsec":"14540","progressmsec":"219","answermsec":"6439","waitmsec":"6459","progress_mediamsec":"239","flow_billmsec":"20979","uduration":"20978809","billusec":"14540020","progressusec":"219114","answerusec":"6438789","waitusec":"6459051","progress_mediausec":"238825","flow_billusec":"20978809","rtp_audio_in_raw_bytes":"181360","rtp_audio_in_media_bytes":"180304","rtp_audio_in_packet_count":"1031","rtp_audio_in_media_packet_count":"1025","rtp_audio_in_skip_packet_count":"45","rtp_audio_in_jb_packet_count":"0","rtp_audio_in_dtmf_packet_count":"0","rtp_audio_in_cng_packet_count":"0","rtp_audio_in_flush_packet_count":"6","rtp_audio_in_largest_jb_size":"0","rtp_audio_out_raw_bytes":"165780","rtp_audio_out_media_bytes":"165780","rtp_audio_out_packet_count":"942","rtp_audio_out_media_packet_count":"942","rtp_audio_out_skip_packet_count":"0","rtp_audio_out_dtmf_packet_count":"0","rtp_audio_out_cng_packet_count":"0","rtp_audio_rtcp_packet_count":"0","rtp_audio_rtcp_octet_count":"0"},"app_log":{"applications":[{"app_name":"hash","app_data":"insert/192.168.56.66-spymap/1001/86cfd6e2-dbda-45a3-b59d-f683ec368e8b"},{"app_name":"hash","app_data":"insert/192.168.56.66-last_dial/1001/1002"},{"app_name":"hash","app_data":"insert/192.168.56.66-last_dial/global/86cfd6e2-dbda-45a3-b59d-f683ec368e8b"},{"app_name":"export","app_data":"RFC2822_DATE=Tue, 08 Apr 2014 21:10:21 +0200"},{"app_name":"park","app_data":""},{"app_name":"hash","app_data":"insert/192.168.56.66-spymap/1001/86cfd6e2-dbda-45a3-b59d-f683ec368e8b"},{"app_name":"hash","app_data":"insert/192.168.56.66-last_dial/1001/1002"},{"app_name":"hash","app_data":"insert/192.168.56.66-last_dial/global/86cfd6e2-dbda-45a3-b59d-f683ec368e8b"},{"app_name":"export","app_data":"RFC2822_DATE=Tue, 08 Apr 2014 21:10:21 +0200"},{"app_name":"export","app_data":"dialed_extension=1002"},{"app_name":"bind_meta_app","app_data":"1 b s execute_extension::dx XML features"},{"app_name":"bind_meta_app","app_data":"2 b s record_session::/var/lib/freeswitch/recordings/1001.2014-04-08-21-10-21.wav"},{"app_name":"bind_meta_app","app_data":"3 b s execute_extension::cf XML features"},{"app_name":"bind_meta_app","app_data":"4 b s execute_extension::att_xfer XML features"},{"app_name":"set","app_data":"ringback=%(2000,4000,440,480)"},{"app_name":"set","app_data":"transfer_ringback=local_stream://moh"},{"app_name":"set","app_data":"call_timeout=30"},{"app_name":"set","app_data":"hangup_after_bridge=true"},{"app_name":"set","app_data":"continue_on_fail=true"},{"app_name":"hash","app_data":"insert/192.168.56.66-call_return/1002/1001"},{"app_name":"hash","app_data":"insert/192.168.56.66-last_dial_ext/1002/86cfd6e2-dbda-45a3-b59d-f683ec368e8b"},{"app_name":"set","app_data":"called_party_callgroup=techsupport"},{"app_name":"hash","app_data":"insert/192.168.56.66-last_dial_ext/techsupport/86cfd6e2-dbda-45a3-b59d-f683ec368e8b"},{"app_name":"hash","app_data":"insert/192.168.56.66-last_dial_ext/global/86cfd6e2-dbda-45a3-b59d-f683ec368e8b"},{"app_name":"hash","app_data":"insert/192.168.56.66-last_dial/techsupport/86cfd6e2-dbda-45a3-b59d-f683ec368e8b"},{"app_name":"bridge","app_data":"user/1002@192.168.56.66"}]},"callflow":{"dialplan":"XML","profile_index":"2","extension":{"name":"global","number":"1002","applications":[{"app_name":"hash","app_data":"insert/${domain_name}-spymap/${caller_id_number}/${uuid}"},{"app_name":"hash","app_data":"insert/${domain_name}-last_dial/${caller_id_number}/${destination_number}"},{"app_name":"hash","app_data":"insert/${domain_name}-last_dial/global/${uuid}"},{"app_name":"export","app_data":"RFC2822_DATE=${strftime(%a, %d %b %Y %T %z)}"},{"app_name":"export","app_data":"dialed_extension=1002"},{"app_name":"bind_meta_app","app_data":"1 b s execute_extension::dx XML features"},{"app_name":"bind_meta_app","app_data":"2 b s record_session::/var/lib/freeswitch/recordings/${caller_id_number}.${strftime(%Y-%m-%d-%H-%M-%S)}.wav"},{"app_name":"bind_meta_app","app_data":"3 b s execute_extension::cf XML features"},{"app_name":"bind_meta_app","app_data":"4 b s execute_extension::att_xfer XML features"},{"app_name":"set","app_data":"ringback=${us-ring}"},{"app_name":"set","app_data":"transfer_ringback=local_stream://moh"},{"app_name":"set","app_data":"call_timeout=30"},{"app_name":"set","app_data":"hangup_after_bridge=true"},{"app_name":"set","app_data":"continue_on_fail=true"},{"app_name":"hash","app_data":"insert/${domain_name}-call_return/${dialed_extension}/${caller_id_number}"},{"app_name":"hash","app_data":"insert/${domain_name}-last_dial_ext/${dialed_extension}/${uuid}"},{"app_name":"set","app_data":"called_party_callgroup=${user_data(${dialed_extension}@${domain_name} var callgroup)}"},{"app_name":"hash","app_data":"insert/${domain_name}-last_dial_ext/${called_party_callgroup}/${uuid}"},{"app_name":"hash","app_data":"insert/${domain_name}-last_dial_ext/global/${uuid}"},{"app_name":"hash","app_data":"insert/${domain_name}-last_dial/${called_party_callgroup}/${uuid}"},{"app_name":"bridge","app_data":"user/${dialed_extension}@${domain_name}"},{"last_executed":"true","app_name":"answer","app_data":""},{"app_name":"sleep","app_data":"1000"},{"app_name":"bridge","app_data":"loopback/app=voicemail:default ${domain_name} ${dialed_extension}"}],"current_app":"answer"},"caller_profile":{"username":"1001","dialplan":"XML","caller_id_name":"1001","ani":"1001","aniii":"","caller_id_number":"1001","network_addr":"192.168.56.1","rdnis":"1002","destination_number":"1002","uuid":"86cfd6e2-dbda-45a3-b59d-f683ec368e8b","source":"mod_sofia","context":"default","chan_name":"sofia/internal/1001@192.168.56.74","originatee":{"originatee_caller_profiles":[{"username":"1001","dialplan":"XML","caller_id_name":"Extension 1001","ani":"1001","aniii":"","caller_id_number":"1001","network_addr":"192.168.56.1","rdnis":"1002","destination_number":"1002","uuid":"402f0929-fa14-4a5f-9642-3a1311bb4ddd","source":"mod_sofia","context":"default","chan_name":"sofia/internal/sip:1002@192.168.56.1:5060"},{"username":"1001","dialplan":"XML","caller_id_name":"Extension 1001","ani":"1001","aniii":"","caller_id_number":"1001","network_addr":"192.168.56.1","rdnis":"1002","destination_number":"1002","uuid":"402f0929-fa14-4a5f-9642-3a1311bb4ddd","source":"mod_sofia","context":"default","chan_name":"sofia/internal/sip:1002@192.168.56.1:5060"}]}},"times":{"created_time":"1396984221278217","profile_created_time":"1396984221377035","progress_time":"1396984221497331","progress_media_time":"1396984221517042","answered_time":"1396984227717006","hangup_time":"1396984242257026","resurrect_time":"0","transfer_time":"0"}},"callflow":{"dialplan":"XML","profile_index":"1","extension":{"name":"global","number":"1002","applications":[{"app_name":"hash","app_data":"insert/${domain_name}-spymap/${caller_id_number}/${uuid}"},{"app_name":"hash","app_data":"insert/${domain_name}-last_dial/${caller_id_number}/${destination_number}"},{"app_name":"hash","app_data":"insert/${domain_name}-last_dial/global/${uuid}"},{"app_name":"export","app_data":"RFC2822_DATE=${strftime(%a, %d %b %Y %T %z)}"},{"app_name":"park","app_data":""}]},"caller_profile":{"username":"1001","dialplan":"XML","caller_id_name":"1001","ani":"1001","aniii":"","caller_id_number":"1001","network_addr":"192.168.56.1","rdnis":"","destination_number":"1002","uuid":"86cfd6e2-dbda-45a3-b59d-f683ec368e8b","source":"mod_sofia","context":"default","chan_name":"sofia/internal/1001@192.168.56.74"},"times":{"created_time":"1396984221278217","profile_created_time":"1396984221278217","progress_time":"0","progress_media_time":"0","answered_time":"0","hangup_time":"0","resurrect_time":"0","transfer_time":"1396984221377035"}}}`)

func TestEvCorelate(t *testing.T) {
	answerEv := new(sessionmanager.FSEvent).AsEvent(answerEvent)
	if answerEv.GetName() != "CHANNEL_ANSWER" {
		t.Error("Event not parsed correctly: ", answerEv)
	}
	cfg, _ := config.NewDefaultCGRConfig()
	cdrEv, err := engine.NewFSCdr(jsonCdr, cfg)
	if err != nil {
		t.Errorf("Error loading cdr: %v", err.Error())
	} else if cdrEv.AsStoredCdr("").OriginID != "86cfd6e2-dbda-45a3-b59d-f683ec368e8b" {
		t.Error("Unexpected acntId received", cdrEv.AsStoredCdr("").OriginID)
	}
	if answerEv.GetCgrId("") != cdrEv.AsStoredCdr("").CGRID {
		t.Error("CGRIDs do not match", answerEv.GetCgrId(""), cdrEv.AsStoredCdr("").CGRID)
	}

}

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
variable_cgr_RequestType: *prepaid
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

var jsonCdr2 = []byte(`{"core-uuid":"651a8db2-4f67-4cf8-b622-169e8a482e50","switchname":"CgrDev1","channel_data":{"state":"CS_REPORTING","direction":"inbound","state_number":"11","flags":"0=1;1=1;37=1;38=1;40=1;43=1;48=1;53=1;105=1;111=1;112=1;116=1;118=1","caps":"1=1;2=1;3=1;4=1;5=1;6=1"},"variables":{"direction":"inbound","uuid":"e3133bf7-dcde-4daf-9663-9a79ffcef5ad","session_id":"4","sip_from_user":"1001","sip_from_uri":"1001@127.0.0.1","sip_from_host":"127.0.0.1","channel_name":"sofia/cgrtest/1001@127.0.0.1","ep_codec_string":"speex@16000h@20i,speex@8000h@20i,speex@32000h@20i,GSM@8000h@20i@13200b,PCMU@8000h@20i@64000b,PCMA@8000h@20i@64000b,G722@8000h@20i@64000b","sip_local_network_addr":"127.0.0.1","sip_network_ip":"127.0.0.1","sip_network_port":"46615","sip_received_ip":"127.0.0.1","sip_received_port":"46615","sip_via_protocol":"tcp","sip_authorized":"true","Event-Name":"REQUEST_PARAMS","Core-UUID":"651a8db2-4f67-4cf8-b622-169e8a482e50","FreeSWITCH-Hostname":"CgrDev1","FreeSWITCH-Switchname":"CgrDev1","FreeSWITCH-IPv4":"10.0.3.15","FreeSWITCH-IPv6":"::1","Event-Date-Local":"2015-07-07 16:52:08","Event-Date-GMT":"Tue, 07 Jul 2015 14:52:08 GMT","Event-Date-Timestamp":"1436280728471153","Event-Calling-File":"sofia.c","Event-Calling-Function":"sofia_handle_sip_i_invite","Event-Calling-Line-Number":"9056","Event-Sequence":"515","sip_number_alias":"1001","sip_auth_username":"1001","sip_auth_realm":"127.0.0.1","number_alias":"1001","requested_domain_name":"cgrates.org","record_stereo":"true","transfer_fallback_extension":"operator","toll_allow":"domestic,international,local","accountcode":"1001","user_context":"default","effective_caller_id_name":"Extension 1001","effective_caller_id_number":"1001","outbound_caller_id_name":"FreeSWITCH","outbound_caller_id_number":"0000000000","callgroup":"techsupport","cgr_RequestType":"*prepaid","cgr_supplier":"supplier1","user_name":"1001","domain_name":"cgrates.org","sip_from_user_stripped":"1001","sofia_profile_name":"cgrtest","recovery_profile_name":"cgrtest","sip_full_route":"<sip:127.0.0.1:25060;lr>","sip_recover_via":"SIP/2.0/TCP 127.0.0.1:46615;rport=46615;branch=z9hG4bKPjGj7AlihmVwAVz9McwVeI64NeBHlPmXAN;alias","sip_req_user":"1003","sip_req_uri":"1003@127.0.0.1","sip_req_host":"127.0.0.1","sip_to_user":"1003","sip_to_uri":"1003@127.0.0.1","sip_to_host":"127.0.0.1","sip_contact_params":"ob","sip_contact_user":"1001","sip_contact_port":"5072","sip_contact_uri":"1001@127.0.0.1:5072","sip_contact_host":"127.0.0.1","sip_via_host":"127.0.0.1","sip_via_port":"46615","sip_via_rport":"46615","switch_r_sdp":"v=0\r\no=- 3645269528 3645269528 IN IP4 10.0.3.15\r\ns=pjmedia\r\nb=AS:84\r\nt=0 0\r\na=X-nat:0\r\nm=audio 4006 RTP/AVP 98 97 99 104 3 0 8 9 96\r\nc=IN IP4 10.0.3.15\r\nb=AS:64000\r\na=rtpmap:98 speex/16000\r\na=rtpmap:97 speex/8000\r\na=rtpmap:99 speex/32000\r\na=rtpmap:104 iLBC/8000\r\na=fmtp:104 mode=30\r\na=rtpmap:3 GSM/8000\r\na=rtpmap:0 PCMU/8000\r\na=rtpmap:8 PCMA/8000\r\na=rtpmap:9 G722/8000\r\na=rtpmap:96 telephone-event/8000\r\na=fmtp:96 0-16\r\na=rtcp:4007 IN IP4 10.0.3.15\r\n","rtp_remote_audio_rtcp_port":"4007 IN IP4 10.0.3.15","rtp_audio_recv_pt":"99","rtp_use_codec_name":"SPEEX","rtp_use_codec_rate":"32000","rtp_use_codec_ptime":"20","rtp_use_codec_channels":"1","rtp_last_audio_codec_string":"SPEEX@32000h@20i@1c","read_codec":"SPEEX","original_read_codec":"SPEEX","read_rate":"32000","original_read_rate":"32000","write_codec":"SPEEX","write_rate":"32000","dtmf_type":"rfc2833","execute_on_answer":"sched_hangup +3120 alloted_timeout","cgr_notify":"+AUTH_OK","max_forwards":"69","transfer_history":"1436280728:e7c250e8-6ad7-4bd4-8962-318e0b0da728:bl_xfer:1003/default/XML","transfer_source":"1436280728:e7c250e8-6ad7-4bd4-8962-318e0b0da728:bl_xfer:1003/default/XML","DP_MATCH":"ARRAY::1003|:1003","call_uuid":"e3133bf7-dcde-4daf-9663-9a79ffcef5ad","ringback":"%(2000,4000,440,480)","call_timeout":"30","dialed_user":"1003","dialed_domain":"cgrates.org","originated_legs":"ARRAY::0a30dd7c-c222-482f-a322-b1218a15f8cd;Outbound Call;1003|:0a30dd7c-c222-482f-a322-b1218a15f8cd;Outbound Call;1003","switch_m_sdp":"v=0\r\no=- 3645269528 3645269529 IN IP4 10.0.3.15\r\ns=pjmedia\r\nb=AS:84\r\nt=0 0\r\na=X-nat:0\r\nm=audio 4018 RTP/AVP 99 101\r\nc=IN IP4 10.0.3.15\r\nb=AS:64000\r\na=rtpmap:99 speex/32000\r\na=rtpmap:101 telephone-event/8000\r\na=fmtp:101 0-16\r\na=rtcp:4019 IN IP4 10.0.3.15\r\n","rtp_local_sdp_str":"v=0\no=FreeSWITCH 1436250882 1436250883 IN IP4 10.0.3.15\ns=FreeSWITCH\nc=IN IP4 10.0.3.15\nt=0 0\nm=audio 29846 RTP/AVP 99 96\na=rtpmap:99 speex/32000\na=rtpmap:96 telephone-event/8000\na=fmtp:96 0-16\na=ptime:20\na=sendrecv\na=rtcp:29847 IN IP4 10.0.3.15\n","local_media_ip":"10.0.3.15","local_media_port":"29846","advertised_media_ip":"10.0.3.15","rtp_use_pt":"99","rtp_use_ssrc":"1470667272","rtp_2833_send_payload":"96","rtp_2833_recv_payload":"96","remote_media_ip":"10.0.3.15","remote_media_port":"4006","endpoint_disposition":"ANSWER","current_application_data":"+3120 alloted_timeout","current_application":"sched_hangup","originate_causes":"ARRAY::0a30dd7c-c222-482f-a322-b1218a15f8cd;NONE|:0a30dd7c-c222-482f-a322-b1218a15f8cd;NONE","originate_disposition":"SUCCESS","DIALSTATUS":"SUCCESS","last_bridge_to":"0a30dd7c-c222-482f-a322-b1218a15f8cd","bridge_channel":"sofia/cgrtest/1003@127.0.0.1:5070","bridge_uuid":"0a30dd7c-c222-482f-a322-b1218a15f8cd","signal_bond":"0a30dd7c-c222-482f-a322-b1218a15f8cd","sip_to_tag":"5Qt4ecvreSHZN","sip_from_tag":"YwuG8U3rRbqIn.xYTnU8NrI3giyxDBHJ","sip_cseq":"4178","sip_call_id":"r3xaJ8CLpyTAIHWUZG7gtZQYgAPEGf9S","sip_full_via":"SIP/2.0/UDP 10.0.3.15:5072;rport=5072;branch=z9hG4bKPjPqma7vnLxDkBqcCH3eXLmLYZoPS.6MDc;received=127.0.0.1","sip_full_from":"sip:1001@127.0.0.1;tag=YwuG8U3rRbqIn.xYTnU8NrI3giyxDBHJ","sip_full_to":"sip:1003@127.0.0.1;tag=5Qt4ecvreSHZN","last_sent_callee_id_name":"Outbound Call","last_sent_callee_id_number":"1003","sip_term_status":"200","proto_specific_hangup_cause":"sip:200","sip_term_cause":"16","last_bridge_role":"originator","sip_user_agent":"PJSUA v2.3 Linux-3.2.0.4/x86_64/glibc-2.13","sip_hangup_disposition":"recv_bye","bridge_hangup_cause":"NORMAL_CLEARING","hangup_cause":"NORMAL_CLEARING","hangup_cause_q850":"16","digits_dialed":"none","start_stamp":"2015-07-07 16:52:08","profile_start_stamp":"2015-07-07 16:52:08","answer_stamp":"2015-07-07 16:52:08","bridge_stamp":"2015-07-07 16:52:08","end_stamp":"2015-07-07 16:53:14","start_epoch":"1436280728","start_uepoch":"1436280728471153","profile_start_epoch":"1436280728","profile_start_uepoch":"1436280728930693","answer_epoch":"1436280728","answer_uepoch":"1436280728971147","bridge_epoch":"1436280728","bridge_uepoch":"1436280728971147","last_hold_epoch":"0","last_hold_uepoch":"0","hold_accum_seconds":"0","hold_accum_usec":"0","hold_accum_ms":"0","resurrect_epoch":"0","resurrect_uepoch":"0","progress_epoch":"0","progress_uepoch":"0","progress_media_epoch":"0","progress_media_uepoch":"0","end_epoch":"1436280794","end_uepoch":"1436280794010851","last_app":"sched_hangup","last_arg":"+3120 alloted_timeout","caller_id":"\"1001\" <1001>","duration":"66","billsec":"66","progresssec":"0","answersec":"0","waitsec":"0","progress_mediasec":"0","flow_billsec":"66","mduration":"65539","billmsec":"65039","progressmsec":"28","answermsec":"500","waitmsec":"500","progress_mediamsec":"28","flow_billmsec":"65539","uduration":"65539698","billusec":"65039704","progressusec":"0","answerusec":"499994","waitusec":"499994","progress_mediausec":"0","flow_billusec":"65539698","rtp_audio_in_raw_bytes":"6770","rtp_audio_in_media_bytes":"6762","rtp_audio_in_packet_count":"192","rtp_audio_in_media_packet_count":"190","rtp_audio_in_skip_packet_count":"6","rtp_audio_in_jitter_packet_count":"0","rtp_audio_in_dtmf_packet_count":"0","rtp_audio_in_cng_packet_count":"0","rtp_audio_in_flush_packet_count":"2","rtp_audio_in_largest_jb_size":"0","rtp_audio_in_jitter_min_variance":"26.73","rtp_audio_in_jitter_max_variance":"6716.71","rtp_audio_in_jitter_loss_rate":"0.00","rtp_audio_in_jitter_burst_rate":"0.00","rtp_audio_in_mean_interval":"36.67","rtp_audio_in_flaw_total":"0","rtp_audio_in_quality_percentage":"100.00","rtp_audio_in_mos":"4.50","rtp_audio_out_raw_bytes":"4686","rtp_audio_out_media_bytes":"4686","rtp_audio_out_packet_count":"108","rtp_audio_out_media_packet_count":"108","rtp_audio_out_skip_packet_count":"0","rtp_audio_out_dtmf_packet_count":"0","rtp_audio_out_cng_packet_count":"0","rtp_audio_rtcp_packet_count":"1450","rtp_audio_rtcp_octet_count":"45940"},"app_log":{"applications":[{"app_name":"info","app_data":""},{"app_name":"park","app_data":""},{"app_name":"info","app_data":""},{"app_name":"set","app_data":"ringback=%(2000,4000,440,480)"},{"app_name":"set","app_data":"call_timeout=30"},{"app_name":"bridge","app_data":"user/1003@cgrates.org"},{"app_name":"sched_hangup","app_data":"+3120 alloted_timeout"}]},"callflow":{"dialplan":"XML","profile_index":"2","extension":{"name":"call_debug","number":"1003","applications":[{"app_name":"info","app_data":""},{"app_name":"set","app_data":"ringback=${us-ring}"},{"app_name":"set","app_data":"call_timeout=30"},{"app_name":"bridge","app_data":"user/${destination_number}@${domain_name}"}]},"caller_profile":{"username":"1001","dialplan":"XML","caller_id_name":"1001","ani":"1001","aniii":"","caller_id_number":"1001","network_addr":"127.0.0.1","rdnis":"1003","destination_number":"1003","uuid":"e3133bf7-dcde-4daf-9663-9a79ffcef5ad","source":"mod_sofia","context":"default","chan_name":"sofia/cgrtest/1001@127.0.0.1","originatee":{"originatee_caller_profiles":[{"username":"1001","dialplan":"XML","caller_id_name":"Extension 1001","ani":"1001","aniii":"","caller_id_number":"1001","network_addr":"127.0.0.1","rdnis":"1003","destination_number":"1003","uuid":"0a30dd7c-c222-482f-a322-b1218a15f8cd","source":"mod_sofia","context":"default","chan_name":"sofia/cgrtest/1003@127.0.0.1:5070"},{"username":"1001","dialplan":"XML","caller_id_name":"Extension 1001","ani":"1001","aniii":"","caller_id_number":"1001","network_addr":"127.0.0.1","rdnis":"1003","destination_number":"1003","uuid":"0a30dd7c-c222-482f-a322-b1218a15f8cd","source":"mod_sofia","context":"default","chan_name":"sofia/cgrtest/1003@127.0.0.1:5070"}]}},"times":{"created_time":"1436280728471153","profile_created_time":"1436280728930693","progress_time":"0","progress_media_time":"0","answered_time":"1436280728971147","bridged_time":"1436280728971147","last_hold_time":"0","hold_accum_time":"0","hangup_time":"1436280794010851","resurrect_time":"0","transfer_time":"0"}},"callflow":{"dialplan":"XML","profile_index":"1","extension":{"name":"call_debug","number":"1003","applications":[{"app_name":"info","app_data":""},{"app_name":"park","app_data":""}]},"caller_profile":{"username":"1001","dialplan":"XML","caller_id_name":"1001","ani":"1001","aniii":"","caller_id_number":"1001","network_addr":"127.0.0.1","rdnis":"","destination_number":"1003","uuid":"e3133bf7-dcde-4daf-9663-9a79ffcef5ad","source":"mod_sofia","context":"default","chan_name":"sofia/cgrtest/1001@127.0.0.1"},"times":{"created_time":"1436280728471153","profile_created_time":"1436280728471153","progress_time":"0","progress_media_time":"0","answered_time":"0","bridged_time":"0","last_hold_time":"0","hold_accum_time":"0","hangup_time":"0","resurrect_time":"0","transfer_time":"1436280728930693"}}}`)

// Make sure that both hangup and json cdr produce the same CGR primary fields
func TestEvCdrCorelate(t *testing.T) {
	hangupEv := new(sessionmanager.FSEvent).AsEvent(hangupEv)
	if hangupEv.GetName() != "CHANNEL_HANGUP_COMPLETE" {
		t.Error("Event not parsed correctly: ", hangupEv)
	}
	cfg, _ := config.NewDefaultCGRConfig()
	config.SetCgrConfig(cfg)
	evStoredCdr := hangupEv.AsStoredCdr("")
	cdrEv, err := engine.NewFSCdr(jsonCdr2, cfg)
	if err != nil {
		t.Errorf("Error loading cdr: %v", err.Error())
	} else if cdrEv.AsStoredCdr("").OriginID != "e3133bf7-dcde-4daf-9663-9a79ffcef5ad" {
		t.Error("Unexpected acntId received", cdrEv.AsStoredCdr("").OriginID)
	}
	jsnStoredCdr := cdrEv.AsStoredCdr("")
	if evStoredCdr.CGRID != jsnStoredCdr.CGRID {
		t.Errorf("evStoredCdr.CGRID: %s, jsnStoredCdr.CGRID: %s", evStoredCdr.CGRID, jsnStoredCdr.CGRID)
	}
	if evStoredCdr.ToR != jsnStoredCdr.ToR {
		t.Errorf("evStoredCdr.ToR: %s, jsnStoredCdr.ToR: %s", evStoredCdr.ToR, jsnStoredCdr.ToR)
	}
	if evStoredCdr.OriginID != jsnStoredCdr.OriginID {
		t.Errorf("evStoredCdr.OriginID: %s, jsnStoredCdr.OriginID: %s", evStoredCdr.OriginID, jsnStoredCdr.OriginID)
	}
	if evStoredCdr.RequestType != jsnStoredCdr.RequestType {
		t.Errorf("evStoredCdr.RequestType: %s, jsnStoredCdr.RequestType: %s", evStoredCdr.RequestType, jsnStoredCdr.RequestType)
	}
	if evStoredCdr.Direction != jsnStoredCdr.Direction {
		t.Errorf("evStoredCdr.Direction: %s, jsnStoredCdr.Direction: %s", evStoredCdr.Direction, jsnStoredCdr.Direction)
	}
	if evStoredCdr.Tenant != jsnStoredCdr.Tenant {
		t.Errorf("evStoredCdr.Tenant: %s, jsnStoredCdr.Tenant: %s", evStoredCdr.Tenant, jsnStoredCdr.Tenant)
	}
	if evStoredCdr.Category != jsnStoredCdr.Category {
		t.Errorf("evStoredCdr.Category: %s, jsnStoredCdr.Category: %s", evStoredCdr.Category, jsnStoredCdr.Category)
	}
	if evStoredCdr.Account != jsnStoredCdr.Account {
		t.Errorf("evStoredCdr.Account: %s, jsnStoredCdr.Account: %s", evStoredCdr.Account, jsnStoredCdr.Account)
	}
	if evStoredCdr.Subject != jsnStoredCdr.Subject {
		t.Errorf("evStoredCdr.Subject: %s, jsnStoredCdr.Subject: %s", evStoredCdr.Subject, jsnStoredCdr.Subject)
	}
	if evStoredCdr.Destination != jsnStoredCdr.Destination {
		t.Errorf("evStoredCdr.Destination: %s, jsnStoredCdr.Destination: %s", evStoredCdr.Destination, jsnStoredCdr.Destination)
	}
	if evStoredCdr.SetupTime != jsnStoredCdr.SetupTime {
		t.Errorf("evStoredCdr.SetupTime: %v, jsnStoredCdr.SetupTime: %v", evStoredCdr.SetupTime, jsnStoredCdr.SetupTime)
	}
	if evStoredCdr.PDD != jsnStoredCdr.PDD {
		t.Errorf("evStoredCdr.PDD: %v, jsnStoredCdr.PDD: %v", evStoredCdr.PDD, jsnStoredCdr.PDD)
	}
	if evStoredCdr.AnswerTime != jsnStoredCdr.AnswerTime {
		t.Errorf("evStoredCdr.AnswerTime: %v, jsnStoredCdr.AnswerTime: %v", evStoredCdr.AnswerTime, jsnStoredCdr.AnswerTime)
	}
	if evStoredCdr.Usage != jsnStoredCdr.Usage {
		t.Errorf("evStoredCdr.Usage: %v, jsnStoredCdr.Usage: %v", evStoredCdr.Usage, jsnStoredCdr.Usage)
	}
	if evStoredCdr.Supplier != jsnStoredCdr.Supplier {
		t.Errorf("evStoredCdr.Supplier: %s, jsnStoredCdr.Supplier: %s", evStoredCdr.Supplier, jsnStoredCdr.Supplier)
	}
	if evStoredCdr.DisconnectCause != jsnStoredCdr.DisconnectCause {
		t.Errorf("evStoredCdr.DisconnectCause: %s, jsnStoredCdr.DisconnectCause: %s", evStoredCdr.DisconnectCause, jsnStoredCdr.DisconnectCause)
	}
}
