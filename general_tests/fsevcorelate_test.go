/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2012-2014 ITsysCOM GmbH

This program is free software: you can Storagetribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITH*out ANY WARRANTY; without even the implied warranty of
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
variable_cgr_reqtype: prepaid
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

var jsonCdr = []byte(`{"core-uuid":"feef0b51-7fdf-4c4a-878e-aff233752de2","channel_data":{"state":"CS_REPORTING","direction":"inbound","state_number":"11","flags":"0=1;1=1;3=1;36=1;37=1;39=1;42=1;47=1;52=1;73=1;75=1;94=1","caps":"1=1;2=1;3=1;4=1;5=1;6=1"},"variables":{"direction":"inbound","uuid":"86cfd6e2-dbda-45a3-b59d-f683ec368e8b","session_id":"5","sip_from_user":"1001","sip_from_uri":"1001@192.168.56.74","sip_from_host":"192.168.56.74","channel_name":"sofia/internal/1001@192.168.56.74","sip_local_network_addr":"192.168.56.74","sip_network_ip":"192.168.56.1","sip_network_port":"5060","sip_received_ip":"192.168.56.1","sip_received_port":"5060","sip_via_protocol":"udp","sip_authorized":"true","Event-Name":"REQUEST_PARAMS","Core-UUID":"feef0b51-7fdf-4c4a-878e-aff233752de2","FreeSWITCH-Hostname":"CGRTest","FreeSWITCH-Switchname":"CGRTest","FreeSWITCH-IPv4":"192.168.178.32","FreeSWITCH-IPv6":"::1","Event-Date-Local":"2014-04-08 21:10:21","Event-Date-GMT":"Tue, 08 Apr 2014 19:10:21 GMT","Event-Date-Timestamp":"1396984221278217","Event-Calling-File":"sofia.c","Event-Calling-Function":"sofia_handle_sip_i_invite","Event-Calling-Line-Number":"8076","Event-Sequence":"1423","sip_number_alias":"1001","sip_auth_username":"1001","sip_auth_realm":"192.168.56.74","number_alias":"1001","requested_domain_name":"192.168.56.66","record_stereo":"true","default_gateway":"example.com","default_areacode":"918","transfer_fallback_extension":"operator","toll_allow":"domestic,international,local","accountcode":"1001","user_context":"default","effective_caller_id_name":"Extension 1001","effective_caller_id_number":"1001","outbound_caller_id_name":"FreeSWITCH","outbound_caller_id_number":"0000000000","callgroup":"techsupport","user_name":"1001","domain_name":"192.168.56.66","sip_from_user_stripped":"1001","sofia_profile_name":"internal","recovery_profile_name":"internal","sip_req_user":"1002","sip_req_uri":"1002@192.168.56.74","sip_req_host":"192.168.56.74","sip_to_user":"1002","sip_to_uri":"1002@192.168.56.74","sip_to_host":"192.168.56.74","sip_contact_params":"transport=udp;registering_acc=192_168_56_74","sip_contact_user":"1001","sip_contact_port":"5060","sip_contact_uri":"1001@192.168.56.1:5060","sip_contact_host":"192.168.56.1","sip_via_host":"192.168.56.1","sip_via_port":"5060","presence_id":"1001@192.168.56.74","ep_codec_string":"G722@8000h@20i@64000b,PCMU@8000h@20i@64000b,PCMA@8000h@20i@64000b,GSM@8000h@20i@13200b","cgr_notify":"+AUTH_OK","max_forwards":"69","transfer_history":"1396984221:caefc538-5da4-4245-8716-112c706383d8:bl_xfer:1002/default/XML","transfer_source":"1396984221:caefc538-5da4-4245-8716-112c706383d8:bl_xfer:1002/default/XML","DP_MATCH":"ARRAY::1002|:1002","call_uuid":"86cfd6e2-dbda-45a3-b59d-f683ec368e8b","RFC2822_DATE":"Tue, 08 Apr 2014 21:10:21 +0200","dialed_extension":"1002","export_vars":"RFC2822_DATE,RFC2822_DATE,dialed_extension","ringback":"%(2000,4000,440,480)","transfer_ringback":"local_stream://moh","call_timeout":"30","hangup_after_bridge":"true","continue_on_fail":"true","called_party_callgroup":"techsupport","current_application_data":"user/1002@192.168.56.66","current_application":"bridge","dialed_user":"1002","dialed_domain":"192.168.56.66","inherit_codec":"true","originated_legs":"ARRAY::402f0929-fa14-4a5f-9642-3a1311bb4ddd;Outbound Call;1002|:402f0929-fa14-4a5f-9642-3a1311bb4ddd;Outbound Call;1002","rtp_use_codec_string":"G722,PCMU,PCMA,GSM","sip_use_codec_name":"G722","sip_use_codec_rate":"8000","sip_use_codec_ptime":"20","write_codec":"G722","write_rate":"16000","video_possible":"true","local_media_ip":"192.168.56.74","local_media_port":"32534","advertised_media_ip":"192.168.56.74","sip_use_pt":"9","rtp_use_ssrc":"1431080133","zrtp_secure_media_confirmed_audio":"true","zrtp_sas1_string_audio":"j6ff","switch_m_sdp":"v=0\r\no=1002 0 0 IN IP4 192.168.56.1\r\ns=-\r\nc=IN IP4 192.168.56.1\r\nt=0 0\r\nm=audio 5020 RTP/AVP 9 0 8 3 101\r\na=rtpmap:9 G722/8000\r\na=rtpmap:0 PCMU/8000\r\na=rtpmap:8 PCMA/8000\r\na=rtpmap:3 GSM/8000\r\na=rtpmap:101 telephone-event/8000\r\n","read_codec":"G722","read_rate":"16000","endpoint_disposition":"ANSWER","originate_causes":"ARRAY::402f0929-fa14-4a5f-9642-3a1311bb4ddd;NONE|:402f0929-fa14-4a5f-9642-3a1311bb4ddd;NONE","originate_disposition":"SUCCESS","DIALSTATUS":"SUCCESS","last_bridge_to":"402f0929-fa14-4a5f-9642-3a1311bb4ddd","bridge_channel":"sofia/internal/sip:1002@192.168.56.1:5060","bridge_uuid":"402f0929-fa14-4a5f-9642-3a1311bb4ddd","signal_bond":"402f0929-fa14-4a5f-9642-3a1311bb4ddd","last_sent_callee_id_name":"Outbound Call","last_sent_callee_id_number":"1002","cgr_reqtype":"prepaid","sip_reinvite_sdp":"v=0\r\no=1001 0 1 IN IP4 192.168.56.1\r\ns=-\r\nc=IN IP4 192.168.56.1\r\nt=0 0\r\nm=audio 5016 RTP/AVP 96 97 98 9 100 102 0 8 103 3 104 101\r\na=sendonly\r\na=rtpmap:96 opus/48000/2\r\na=fmtp:96 usedtx=1\r\na=rtpmap:97 SILK/24000\r\na=rtpmap:98 SILK/16000\r\na=rtpmap:9 G722/8000\r\na=rtpmap:100 speex/32000\r\na=rtpmap:102 speex/16000\r\na=rtpmap:0 PCMU/8000\r\na=rtpmap:8 PCMA/8000\r\na=rtpmap:103 iLBC/8000\r\na=rtpmap:3 GSM/8000\r\na=rtpmap:104 speex/8000\r\na=rtpmap:101 telephone-event/8000\r\na=extmap:1 urn:ietf:params:rtp-hdrext:csrc-audio-level\r\na=zrtp-hash:1.10 722d57097aaabea2749ea8938472478f8d88645b23521fa5f8005a7a2bed8286\r\nm=video 0 RTP/AVP 105 99\r\n","switch_r_sdp":"v=0\r\no=1001 0 1 IN IP4 192.168.56.1\r\ns=-\r\nc=IN IP4 192.168.56.1\r\nt=0 0\r\nm=audio 5016 RTP/AVP 96 97 98 9 100 102 0 8 103 3 104 101\r\na=rtpmap:96 opus/48000/2\r\na=fmtp:96 usedtx=1\r\na=rtpmap:97 SILK/24000\r\na=rtpmap:98 SILK/16000\r\na=rtpmap:9 G722/8000\r\na=rtpmap:100 speex/32000\r\na=rtpmap:102 speex/16000\r\na=rtpmap:0 PCMU/8000\r\na=rtpmap:8 PCMA/8000\r\na=rtpmap:103 iLBC/8000\r\na=rtpmap:3 GSM/8000\r\na=rtpmap:104 speex/8000\r\na=rtpmap:101 telephone-event/8000\r\na=sendonly\r\na=extmap:1 urn:ietf:params:rtp-hdrext:csrc-audio-level\r\na=zrtp-hash:1.10 722d57097aaabea2749ea8938472478f8d88645b23521fa5f8005a7a2bed8286\r\nm=video 0 RTP/AVP 105 99\r\n","r_sdp_audio_zrtp_hash":"1.10 722d57097aaabea2749ea8938472478f8d88645b23521fa5f8005a7a2bed8286","remote_media_ip":"192.168.56.1","remote_media_port":"5016","sip_audio_recv_pt":"9","dtmf_type":"rfc2833","sip_2833_send_payload":"101","sip_2833_recv_payload":"101","sip_local_sdp_str":"v=0\no=FreeSWITCH 1396951687 1396951690 IN IP4 192.168.56.74\ns=FreeSWITCH\nc=IN IP4 192.168.56.74\nt=0 0\nm=audio 32534 RTP/AVP 9 101\na=rtpmap:9 G722/8000\na=rtpmap:101 telephone-event/8000\na=fmtp:101 0-16\na=ptime:20\na=sendrecv\n","sip_to_tag":"rXc9vZpv9eFaF","sip_from_tag":"1afc7eca","sip_cseq":"3","sip_call_id":"6691dbf8ffdc02bdacee02bc305d5c71@0:0:0:0:0:0:0:0","sip_full_via":"SIP/2.0/UDP 192.168.56.1:5060;branch=z9hG4bK-323133-5d083abc0d3f327b9101586e71b5fce4","sip_from_display":"1001","sip_full_from":"\"1001\" <sip:1001@192.168.56.74>;tag=1afc7eca","sip_full_to":"<sip:1002@192.168.56.74>;tag=rXc9vZpv9eFaF","sip_term_status":"200","proto_specific_hangup_cause":"sip:200","sip_term_cause":"16","last_bridge_role":"originator","sip_user_agent":"Jitsi2.5.5065Linux","sip_hangup_disposition":"recv_bye","bridge_hangup_cause":"NORMAL_CLEARING","hangup_cause":"NORMAL_CLEARING","hangup_cause_q850":"16","digits_dialed":"none","start_stamp":"2014-04-08 21:10:21","profile_start_stamp":"2014-04-08 21:10:21","answer_stamp":"2014-04-08 21:10:27","bridge_stamp":"2014-04-08 21:10:27","hold_stamp":"2014-04-08 21:10:27","progress_stamp":"2014-04-08 21:10:21","progress_media_stamp":"2014-04-08 21:10:21","hold_events":"{{1396984227824182,1396984242247995}}","end_stamp":"2014-04-08 21:10:42","start_epoch":"1396984221","start_uepoch":"1396984221278217","profile_start_epoch":"1396984221","profile_start_uepoch":"1396984221377035","answer_epoch":"1396984227","answer_uepoch":"1396984227717006","bridge_epoch":"1396984227","bridge_uepoch":"1396984227737268","last_hold_epoch":"1396984227","last_hold_uepoch":"1396984227824167","hold_accum_seconds":"14","hold_accum_usec":"14423816","hold_accum_ms":"14423","resurrect_epoch":"0","resurrect_uepoch":"0","progress_epoch":"1396984221","progress_uepoch":"1396984221497331","progress_media_epoch":"1396984221","progress_media_uepoch":"1396984221517042","end_epoch":"1396984242","end_uepoch":"1396984242257026","last_app":"bridge","last_arg":"user/1002@192.168.56.66","caller_id":"\"1001\" <1001>","duration":"21","billsec":"15","progresssec":"0","answersec":"6","waitsec":"6","progress_mediasec":"0","flow_billsec":"21","mduration":"20979","billmsec":"14540","progressmsec":"219","answermsec":"6439","waitmsec":"6459","progress_mediamsec":"239","flow_billmsec":"20979","uduration":"20978809","billusec":"14540020","progressusec":"219114","answerusec":"6438789","waitusec":"6459051","progress_mediausec":"238825","flow_billusec":"20978809","rtp_audio_in_raw_bytes":"181360","rtp_audio_in_media_bytes":"180304","rtp_audio_in_packet_count":"1031","rtp_audio_in_media_packet_count":"1025","rtp_audio_in_skip_packet_count":"45","rtp_audio_in_jb_packet_count":"0","rtp_audio_in_dtmf_packet_count":"0","rtp_audio_in_cng_packet_count":"0","rtp_audio_in_flush_packet_count":"6","rtp_audio_in_largest_jb_size":"0","rtp_audio_out_raw_bytes":"165780","rtp_audio_out_media_bytes":"165780","rtp_audio_out_packet_count":"942","rtp_audio_out_media_packet_count":"942","rtp_audio_out_skip_packet_count":"0","rtp_audio_out_dtmf_packet_count":"0","rtp_audio_out_cng_packet_count":"0","rtp_audio_rtcp_packet_count":"0","rtp_audio_rtcp_octet_count":"0"},"app_log":{"applications":[{"app_name":"hash","app_data":"insert/192.168.56.66-spymap/1001/86cfd6e2-dbda-45a3-b59d-f683ec368e8b"},{"app_name":"hash","app_data":"insert/192.168.56.66-last_dial/1001/1002"},{"app_name":"hash","app_data":"insert/192.168.56.66-last_dial/global/86cfd6e2-dbda-45a3-b59d-f683ec368e8b"},{"app_name":"export","app_data":"RFC2822_DATE=Tue, 08 Apr 2014 21:10:21 +0200"},{"app_name":"park","app_data":""},{"app_name":"hash","app_data":"insert/192.168.56.66-spymap/1001/86cfd6e2-dbda-45a3-b59d-f683ec368e8b"},{"app_name":"hash","app_data":"insert/192.168.56.66-last_dial/1001/1002"},{"app_name":"hash","app_data":"insert/192.168.56.66-last_dial/global/86cfd6e2-dbda-45a3-b59d-f683ec368e8b"},{"app_name":"export","app_data":"RFC2822_DATE=Tue, 08 Apr 2014 21:10:21 +0200"},{"app_name":"export","app_data":"dialed_extension=1002"},{"app_name":"bind_meta_app","app_data":"1 b s execute_extension::dx XML features"},{"app_name":"bind_meta_app","app_data":"2 b s record_session::/var/lib/freeswitch/recordings/1001.2014-04-08-21-10-21.wav"},{"app_name":"bind_meta_app","app_data":"3 b s execute_extension::cf XML features"},{"app_name":"bind_meta_app","app_data":"4 b s execute_extension::att_xfer XML features"},{"app_name":"set","app_data":"ringback=%(2000,4000,440,480)"},{"app_name":"set","app_data":"transfer_ringback=local_stream://moh"},{"app_name":"set","app_data":"call_timeout=30"},{"app_name":"set","app_data":"hangup_after_bridge=true"},{"app_name":"set","app_data":"continue_on_fail=true"},{"app_name":"hash","app_data":"insert/192.168.56.66-call_return/1002/1001"},{"app_name":"hash","app_data":"insert/192.168.56.66-last_dial_ext/1002/86cfd6e2-dbda-45a3-b59d-f683ec368e8b"},{"app_name":"set","app_data":"called_party_callgroup=techsupport"},{"app_name":"hash","app_data":"insert/192.168.56.66-last_dial_ext/techsupport/86cfd6e2-dbda-45a3-b59d-f683ec368e8b"},{"app_name":"hash","app_data":"insert/192.168.56.66-last_dial_ext/global/86cfd6e2-dbda-45a3-b59d-f683ec368e8b"},{"app_name":"hash","app_data":"insert/192.168.56.66-last_dial/techsupport/86cfd6e2-dbda-45a3-b59d-f683ec368e8b"},{"app_name":"bridge","app_data":"user/1002@192.168.56.66"}]},"callflow":{"dialplan":"XML","profile_index":"2","extension":{"name":"global","number":"1002","applications":[{"app_name":"hash","app_data":"insert/${domain_name}-spymap/${caller_id_number}/${uuid}"},{"app_name":"hash","app_data":"insert/${domain_name}-last_dial/${caller_id_number}/${destination_number}"},{"app_name":"hash","app_data":"insert/${domain_name}-last_dial/global/${uuid}"},{"app_name":"export","app_data":"RFC2822_DATE=${strftime(%a, %d %b %Y %T %z)}"},{"app_name":"export","app_data":"dialed_extension=1002"},{"app_name":"bind_meta_app","app_data":"1 b s execute_extension::dx XML features"},{"app_name":"bind_meta_app","app_data":"2 b s record_session::/var/lib/freeswitch/recordings/${caller_id_number}.${strftime(%Y-%m-%d-%H-%M-%S)}.wav"},{"app_name":"bind_meta_app","app_data":"3 b s execute_extension::cf XML features"},{"app_name":"bind_meta_app","app_data":"4 b s execute_extension::att_xfer XML features"},{"app_name":"set","app_data":"ringback=${us-ring}"},{"app_name":"set","app_data":"transfer_ringback=local_stream://moh"},{"app_name":"set","app_data":"call_timeout=30"},{"app_name":"set","app_data":"hangup_after_bridge=true"},{"app_name":"set","app_data":"continue_on_fail=true"},{"app_name":"hash","app_data":"insert/${domain_name}-call_return/${dialed_extension}/${caller_id_number}"},{"app_name":"hash","app_data":"insert/${domain_name}-last_dial_ext/${dialed_extension}/${uuid}"},{"app_name":"set","app_data":"called_party_callgroup=${user_data(${dialed_extension}@${domain_name} var callgroup)}"},{"app_name":"hash","app_data":"insert/${domain_name}-last_dial_ext/${called_party_callgroup}/${uuid}"},{"app_name":"hash","app_data":"insert/${domain_name}-last_dial_ext/global/${uuid}"},{"app_name":"hash","app_data":"insert/${domain_name}-last_dial/${called_party_callgroup}/${uuid}"},{"app_name":"bridge","app_data":"user/${dialed_extension}@${domain_name}"},{"last_executed":"true","app_name":"answer","app_data":""},{"app_name":"sleep","app_data":"1000"},{"app_name":"bridge","app_data":"loopback/app=voicemail:default ${domain_name} ${dialed_extension}"}],"current_app":"answer"},"caller_profile":{"username":"1001","dialplan":"XML","caller_id_name":"1001","ani":"1001","aniii":"","caller_id_number":"1001","network_addr":"192.168.56.1","rdnis":"1002","destination_number":"1002","uuid":"86cfd6e2-dbda-45a3-b59d-f683ec368e8b","source":"mod_sofia","context":"default","chan_name":"sofia/internal/1001@192.168.56.74","originatee":{"originatee_caller_profiles":[{"username":"1001","dialplan":"XML","caller_id_name":"Extension 1001","ani":"1001","aniii":"","caller_id_number":"1001","network_addr":"192.168.56.1","rdnis":"1002","destination_number":"1002","uuid":"402f0929-fa14-4a5f-9642-3a1311bb4ddd","source":"mod_sofia","context":"default","chan_name":"sofia/internal/sip:1002@192.168.56.1:5060"},{"username":"1001","dialplan":"XML","caller_id_name":"Extension 1001","ani":"1001","aniii":"","caller_id_number":"1001","network_addr":"192.168.56.1","rdnis":"1002","destination_number":"1002","uuid":"402f0929-fa14-4a5f-9642-3a1311bb4ddd","source":"mod_sofia","context":"default","chan_name":"sofia/internal/sip:1002@192.168.56.1:5060"}]}},"times":{"created_time":"1396984221278217","profile_created_time":"1396984221377035","progress_time":"1396984221497331","progress_media_time":"1396984221517042","answered_time":"1396984227717006","hangup_time":"1396984242257026","resurrect_time":"0","transfer_time":"0"}},"callflow":{"dialplan":"XML","profile_index":"1","extension":{"name":"global","number":"1002","applications":[{"app_name":"hash","app_data":"insert/${domain_name}-spymap/${caller_id_number}/${uuid}"},{"app_name":"hash","app_data":"insert/${domain_name}-last_dial/${caller_id_number}/${destination_number}"},{"app_name":"hash","app_data":"insert/${domain_name}-last_dial/global/${uuid}"},{"app_name":"export","app_data":"RFC2822_DATE=${strftime(%a, %d %b %Y %T %z)}"},{"app_name":"park","app_data":""}]},"caller_profile":{"username":"1001","dialplan":"XML","caller_id_name":"1001","ani":"1001","aniii":"","caller_id_number":"1001","network_addr":"192.168.56.1","rdnis":"","destination_number":"1002","uuid":"86cfd6e2-dbda-45a3-b59d-f683ec368e8b","source":"mod_sofia","context":"default","chan_name":"sofia/internal/1001@192.168.56.74"},"times":{"created_time":"1396984221278217","profile_created_time":"1396984221278217","progress_time":"0","progress_media_time":"0","answered_time":"0","hangup_time":"0","resurrect_time":"0","transfer_time":"1396984221377035"}}}`)

func TestEvCorelate(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	engine.NewCdrS(nil, nil, cfg) // So we can set the package cfg
	answerEv := new(sessionmanager.FSEvent).New(answerEvent)
	if answerEv.GetName() != "CHANNEL_ANSWER" {
		t.Error("Event not parsed correctly: ", answerEv)
	}
	cdrEv, err := engine.NewFSCdr(jsonCdr)
	if err != nil {
		t.Errorf("Error loading cdr: %v", err.Error())
	} else if cdrEv.AsStoredCdr().AccId != "86cfd6e2-dbda-45a3-b59d-f683ec368e8b" {
		t.Error("Unexpected acntId received", cdrEv.AsStoredCdr().AccId)
	}
	if answerEv.GetCgrId() != cdrEv.AsStoredCdr().CgrId {
		t.Error("CgrIds do not match", answerEv.GetCgrId(), cdrEv.AsStoredCdr().CgrId)
	}

}
