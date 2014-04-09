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

package cdrs

import (
	"reflect"
	"regexp"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

var body = []byte(`{"core-uuid":"844715f9-d8a1-44d6-a4bf-358bec5e10b8","channel_data":{"state":"CS_REPORTING","direction":"inbound","state_number":"11","flags":"0=1;1=1;3=1;19=1;23=1;36=1;37=1;39=1;42=1;47=1;52=1","caps":"1=1;2=1;3=1;4=1;5=1;6=1"},"variables":{"direction":"inbound","uuid":"01df56f4-d99a-4ef6-b7fe-b924b2415b7f","session_id":"33","sip_from_user":"dan","sip_from_uri":"dan@ipbx.itsyscom.com","sip_from_host":"ipbx.itsyscom.com","channel_name":"sofia/ipbxas/dan@ipbx.itsyscom.com","sip_local_network_addr":"127.0.0.1","sip_network_ip":"2.3.4.5","sip_network_port":"5060","sip_received_ip":"2.3.4.5","sip_received_port":"5060","sip_via_protocol":"udp","sip_from_user_stripped":"dan","sofia_profile_name":"ipbxas","recovery_profile_name":"ipbxas","sip_invite_record_route":"<sip:2.3.4.5;lr=on;nat=yes>","sip_req_user":"+4986517174963","sip_req_port":"5080","sip_req_uri":"+4986517174963@127.0.0.1:5080","sip_req_host":"127.0.0.1","sip_to_user":"+4986517174963","sip_to_uri":"+4986517174963@ipbx.itsyscom.com","sip_to_host":"ipbx.itsyscom.com","sip_contact_params":"alias=1.2.3.4~5060~1;transport=udp;registering_acc=ipbx_itsyscom_com","sip_contact_user":"dan","sip_contact_port":"5060","sip_contact_uri":"dan@10.10.10.154:5060","sip_contact_host":"10.10.10.154","sip_user_agent":"Jitsi2.2.4603.9615Linux","sip_via_host":"2.3.4.5","presence_id":"dan@ipbx.itsyscom.com","sip_h_X-AuthType":"SUA","sip_h_X-AuthUser":"dan","sip_h_X-AuthDomain":"ipbx.itsyscom.com","sip_h_X-BalancerIP":"2.3.4.5","switch_r_sdp":"v=0\r\no=dan 0 0 IN IP4 10.10.10.154\r\ns=-\r\nc=IN IP4 10.10.10.154\r\nt=0 0\r\nm=audio 5004 RTP/AVP 96 8 0\r\na=rtpmap:96 opus/48000\r\na=fmtp:96 usedtx=1\r\na=rtpmap:8 PCMA/8000\r\na=rtpmap:0 PCMU/8000\r\na=extmap:1 urn:ietf:params:rtp-hdrext:csrc-audio-level\r\nm=video 5006 RTP/AVP 97 99\r\na=rtpmap:97 H264/90000\r\na=fmtp:97 profile-level-id=4DE01f;packetization-mode=1\r\na=rtpmap:99 H264/90000\r\na=fmtp:99 profile-level-id=4DE01f\r\na=recvonly\r\na=imageattr:97 send [x=[0-640],y=[0-480]] recv [x=[0-1280],y=[0-800]]\r\na=imageattr:99 send [x=[0-640],y=[0-480]] recv [x=[0-1280],y=[0-800]]\r\n","ep_codec_string":"PCMA@8000h@20i@64000b,PCMU@8000h@20i@64000b,H264@90000h","effective_caller_id_number":"+4986517174960","hangup_after_bridge":"true","continue_on_fail":"true","cgr_tenant":"ipbx.itsyscom.com","cgr_tor":"call","cgr_account":"dan","cgr_subject":"dan","cgr_destination":"+4986517174963","sip_redirect_contact_0":"<sip:dan@10.10.10.141:3072;alias=1.2.3.4~3072~1;line=x81npwse;rcv=sip:1.2.3.4:3072>;q=1","sip_redirected_to":"<sip:dan@10.10.10.141:3072;alias=1.2.3.4~3072~1;line=x81npwse;rcv=sip:1.2.3.4:3072>;q=1","sip_redirect_contact_user_0":"dan","sip_redirect_contact_host_0":"10.10.10.141","sip_redirect_contact_params_0":"alias=1.2.3.4~3072~1;line=x81npwse;rcv=sip:1.2.3.4:3072","sip_redirect_dialstring_0":"sofia/ipbxas/sip:dan@10.10.10.141:3072;alias=1.2.3.4~3072~1;line=x81npwse;rcv=sip:1.2.3.4:3072","sip_redirect_contact_1":"<sip:dan@10.10.10.154:5060;alias=1.2.3.4~5060~1;transport=udp;registering_acc=ipbx_itsyscom_com;rcv=sip:1.2.3.4:5060>","sip_redirect_contact_user_1":"dan","sip_redirect_contact_host_1":"10.10.10.154","sip_redirect_contact_params_1":"alias=1.2.3.4~5060~1;transport=udp;registering_acc=ipbx_itsyscom_com;rcv=sip:1.2.3.4:5060","sip_redirect_dialstring_1":"sofia/ipbxas/sip:dan@10.10.10.154:5060;alias=1.2.3.4~5060~1;transport=udp;registering_acc=ipbx_itsyscom_com;rcv=sip:1.2.3.4:5060","sip_redirect_dialstring":"sofia/ipbxas/sip:dan@10.10.10.141:3072;alias=1.2.3.4~3072~1;line=x81npwse;rcv=sip:1.2.3.4:3072,sofia/ipbxas/sip:dan@10.10.10.154:5060;alias=1.2.3.4~5060~1;transport=udp;registering_acc=ipbx_itsyscom_com;rcv=sip:1.2.3.4:5060","max_forwards":"15","transfer_history":"1375609854:d2300128-6724-471c-a495-a3f7a985a2b6:bl_xfer:dan/redirected/XML","transfer_source":"1375609854:d2300128-6724-471c-a495-a3f7a985a2b6:bl_xfer:dan/redirected/XML","call_uuid":"01df56f4-d99a-4ef6-b7fe-b924b2415b7f","current_application_data":"sofia/ipbxas/sip:dan@10.10.10.154:5060;alias=1.2.3.4~5060~1;transport=udp;registering_acc=ipbx_itsyscom_com;rcv=sip:1.2.3.4:5060;fs_path=sip:2.3.4.5,sofia/ipbxas/sip:dan@10.10.10.141:3072;alias=1.2.3.4~3072~1;line=x81npwse;rcv=sip:1.2.3.4:3072;fs_path=sip:2.3.4.5","current_application":"bridge","originated_legs":"ARRAY::e377c077-0f1f-4b7d-b036-1aeef13eff32;Outbound Call;dan|:8f7c860f-0619-4d3c-9515-cc23b0fa3997;Outbound Call;dan","switch_m_sdp":"v=0\r\no=root 975388641 975388642 IN IP4 10.10.10.141\r\ns=call\r\nc=IN IP4 10.10.10.141\r\nt=0 0\r\nm=audio 59976 RTP/AVP 8 0 9 3 101\r\na=rtpmap:8 pcma/8000\r\na=rtpmap:0 pcmu/8000\r\na=rtpmap:9 g722/8000\r\na=rtpmap:3 gsm/8000\r\na=rtpmap:101 telephone-event/8000\r\na=fmtp:101 0-16\r\na=ptime:20\r\nm=video 0 RTP/AVP 98 99\r\na=rtpmap:98 H264/90000\r\na=fmtp:98 profile-level-id=4DE01f\r\na=rtpmap:99 H264/90000\r\na=fmtp:99 profile-level-id=4DE01f\r\n","rtp_use_codec_string":"G722,G722.1,G729,PCMU,PCMA,GSM,H264","sip_audio_recv_pt":"0","sip_use_codec_name":"PCMU","sip_use_codec_rate":"8000","sip_use_codec_ptime":"20","read_codec":"PCMU","read_rate":"8000","write_codec":"PCMU","write_rate":"8000","dtmf_type":"info","video_possible":"true","remote_video_ip":"10.10.10.154","remote_video_port":"5006","sip_video_fmtp":"profile-level-id=4DE01f;packetization-mode=1","sip_video_pt":"97","sip_video_recv_pt":"97","video_read_codec":"H264","video_read_rate":"90000","video_write_codec":"H264","video_write_rate":"90000","sip_use_video_codec_name":"H264","sip_use_video_codec_fmtp":"profile-level-id=4DE01f;packetization-mode=1","sip_use_video_codec_rate":"90000","sip_use_video_codec_ptime":"0","local_media_ip":"2.3.4.5","local_media_port":"29452","advertised_media_ip":"2.3.4.5","sip_use_pt":"0","rtp_use_ssrc":"1408273224","local_video_ip":"2.3.4.5","local_video_port":"22648","sip_use_video_pt":"97","rtp_use_video_ssrc":"1408273224","sip_local_sdp_str":"v=0\no=iPBXCell 1375580404 1375580405 IN IP4 2.3.4.5\ns=iPBXCell\nc=IN IP4 2.3.4.5\nt=0 0\nm=audio 29452 RTP/AVP 0\na=rtpmap:0 PCMU/8000\na=silenceSupp:off - - - -\na=ptime:20\na=sendrecv\nm=video 22648 RTP/AVP 97\na=rtpmap:97 H264/90000\n","endpoint_disposition":"ANSWER","originate_disposition":"SUCCESS","DIALSTATUS":"SUCCESS","originate_causes":"ARRAY::e377c077-0f1f-4b7d-b036-1aeef13eff32;LOSE_RACE|:8f7c860f-0619-4d3c-9515-cc23b0fa3997;NONE","last_bridge_to":"8f7c860f-0619-4d3c-9515-cc23b0fa3997","bridge_channel":"sofia/ipbxas/sip:dan@10.10.10.141:3072","bridge_uuid":"8f7c860f-0619-4d3c-9515-cc23b0fa3997","signal_bond":"8f7c860f-0619-4d3c-9515-cc23b0fa3997","sip_to_tag":"4X345vQvQyetD","sip_from_tag":"9f90cc40","sip_cseq":"2","sip_call_id":"ca9c5e20caeaa6596be8cf66261f13e5@0:0:0:0:0:0:0:0","sip_full_via":"SIP/2.0/UDP 2.3.4.5;branch=z9hG4bKcydzigwkX,SIP/2.0/UDP 10.10.10.154:5060;rport=5060;received=1.2.3.4;branch=z9hG4bK-313937-b78ce2a1daafe532fc34b1b3735727ac","sip_from_display":"dan","sip_full_from":"\"dan\" <sip:dan@ipbx.itsyscom.com>;tag=9f90cc40","sip_full_to":"<sip:+4986517174963@ipbx.itsyscom.com>;tag=4X345vQvQyetD","last_sent_callee_id_name":"Outbound Call","last_sent_callee_id_number":"dan","remote_media_ip_reported":"10.10.10.154","remote_media_ip":"1.2.3.4","remote_media_port_reported":"5004","remote_media_port":"5004","rtp_auto_adjust":"true","sip_hangup_phrase":"OK","last_bridge_hangup_cause":"NORMAL_CLEARING","last_bridge_proto_specific_hangup_cause":"sip:200","bridge_hangup_cause":"NORMAL_CLEARING","hangup_cause":"NORMAL_CLEARING","hangup_cause_q850":"16","digits_dialed":"none","start_stamp":"2013-08-04 11:50:54","profile_start_stamp":"2013-08-04 11:50:54","answer_stamp":"2013-08-04 11:50:56","bridge_stamp":"2013-08-04 11:50:56","progress_stamp":"2013-08-04 11:50:54","progress_media_stamp":"2013-08-04 11:50:56","end_stamp":"2013-08-04 11:51:00","start_epoch":"1375609854","start_uepoch":"1375609854385581","profile_start_epoch":"1375609854","profile_start_uepoch":"1375609854385581","answer_epoch":"1375609856","answer_uepoch":"1375609856285587","bridge_epoch":"1375609856","bridge_uepoch":"1375609856285587","last_hold_epoch":"0","last_hold_uepoch":"0","hold_accum_seconds":"0","hold_accum_usec":"0","hold_accum_ms":"0","resurrect_epoch":"0","resurrect_uepoch":"0","progress_epoch":"1375609854","progress_uepoch":"1375609854505584","progress_media_epoch":"1375609856","progress_media_uepoch":"1375609856285587","end_epoch":"1375609860","end_uepoch":"1375609860205563","last_app":"bridge","last_arg":"sofia/ipbxas/sip:dan@10.10.10.154:5060;alias=1.2.3.4~5060~1;transport=udp;registering_acc=ipbx_itsyscom_com;rcv=sip:1.2.3.4:5060;fs_path=sip:2.3.4.5,sofia/ipbxas/sip:dan@10.10.10.141:3072;alias=1.2.3.4~3072~1;line=x81npwse;rcv=sip:1.2.3.4:3072;fs_path=sip:2.3.4.5","caller_id":"\"dan\" <dan>","duration":"6","billsec":"4","progresssec":"0","answersec":"2","waitsec":"2","progress_mediasec":"2","flow_billsec":"6","mduration":"5820","billmsec":"3920","progressmsec":"120","answermsec":"1900","waitmsec":"1900","progress_mediamsec":"1900","flow_billmsec":"5820","uduration":"5819982","billusec":"3919976","progressusec":"120003","answerusec":"1900006","waitusec":"1900006","progress_mediausec":"1900006","flow_billusec":"5819982","sip_hangup_disposition":"send_bye","rtp_audio_in_raw_bytes":"32968","rtp_audio_in_media_bytes":"32960","rtp_audio_in_packet_count":"207","rtp_audio_in_media_packet_count":"205","rtp_audio_in_skip_packet_count":"6","rtp_audio_in_jb_packet_count":"0","rtp_audio_in_dtmf_packet_count":"0","rtp_audio_in_cng_packet_count":"0","rtp_audio_in_flush_packet_count":"2","rtp_audio_in_largest_jb_size":"0","rtp_audio_out_raw_bytes":"31648","rtp_audio_out_media_bytes":"31648","rtp_audio_out_packet_count":"184","rtp_audio_out_media_packet_count":"184","rtp_audio_out_skip_packet_count":"0","rtp_audio_out_dtmf_packet_count":"0","rtp_audio_out_cng_packet_count":"0","rtp_audio_rtcp_packet_count":"0","rtp_audio_rtcp_octet_count":"0","rtp_video_in_raw_bytes":"0","rtp_video_in_media_bytes":"0","rtp_video_in_packet_count":"0","rtp_video_in_media_packet_count":"0","rtp_video_in_skip_packet_count":"0","rtp_video_in_jb_packet_count":"0","rtp_video_in_dtmf_packet_count":"0","rtp_video_in_cng_packet_count":"0","rtp_video_in_flush_packet_count":"0","rtp_video_in_largest_jb_size":"0","rtp_video_out_raw_bytes":"0","rtp_video_out_media_bytes":"0","rtp_video_out_packet_count":"0","rtp_video_out_media_packet_count":"0","rtp_video_out_skip_packet_count":"0","rtp_video_out_dtmf_packet_count":"0","rtp_video_out_cng_packet_count":"0","rtp_video_rtcp_packet_count":"0","rtp_video_rtcp_octet_count":"0"},"app_log":{"applications":[{"app_name":"set","app_data":"effective_caller_id_number=+4986517174960"},{"app_name":"set","app_data":"hangup_after_bridge=true"},{"app_name":"set","app_data":"continue_on_fail=true"},{"app_name":"set","app_data":"cgr_tenant=ipbx.itsyscom.com"},{"app_name":"set","app_data":"cgr_tor=call"},{"app_name":"set","app_data":"cgr_account=dan"},{"app_name":"set","app_data":"cgr_subject=dan"},{"app_name":"set","app_data":"cgr_destination=+4986517174963"},{"app_name":"bridge","app_data":"{presence_id=dan@ipbx.itsyscom.com,sip_redirect_fork=true}sofia/ipbxas/dan@ipbx.itsyscom.com;fs_path=sip:2.3.4.5"},{"app_name":"bridge","app_data":"sofia/ipbxas/sip:dan@10.10.10.154:5060;alias=1.2.3.4~5060~1;transport=udp;registering_acc=ipbx_itsyscom_com;rcv=sip:1.2.3.4:5060;fs_path=sip:2.3.4.5,sofia/ipbxas/sip:dan@10.10.10.141:3072;alias=1.2.3.4~3072~1;line=x81npwse;rcv=sip:1.2.3.4:3072;fs_path=sip:2.3.4.5"}]},"callflow":{"dialplan":"XML","profile_index":"2","extension":{"name":"Redirected call","number":"dan","applications":[{"app_name":"bridge","app_data":"sofia/ipbxas/sip:dan@10.10.10.154:5060;alias=1.2.3.4~5060~1;transport=udp;registering_acc=ipbx_itsyscom_com;rcv=sip:1.2.3.4:5060;fs_path=sip:2.3.4.5,sofia/ipbxas/sip:dan@10.10.10.141:3072;alias=1.2.3.4~3072~1;line=x81npwse;rcv=sip:1.2.3.4:3072;fs_path=sip:2.3.4.5"}]},"caller_profile":{"username":"dan","dialplan":"XML","caller_id_name":"dan","ani":"dan","aniii":"","caller_id_number":"dan","network_addr":"2.3.4.5","rdnis":"+4986517174963","destination_number":"dan","uuid":"01df56f4-d99a-4ef6-b7fe-b924b2415b7f","source":"mod_sofia","context":"redirected","chan_name":"sofia/ipbxas/dan@ipbx.itsyscom.com","originatee":{"originatee_caller_profiles":[{"username":"dan","dialplan":"XML","caller_id_name":"dan","ani":"dan","aniii":"","caller_id_number":"+4986517174960","network_addr":"2.3.4.5","rdnis":"+4986517174963","destination_number":"dan","uuid":"8f7c860f-0619-4d3c-9515-cc23b0fa3997","source":"mod_sofia","context":"redirected","chan_name":"sofia/ipbxas/sip:dan@10.10.10.141:3072"}]}},"times":{"created_time":"1375609854385581","profile_created_time":"1375609854385581","progress_time":"1375609854505584","progress_media_time":"1375609856285587","answered_time":"1375609856285587","hangup_time":"1375609860205563","resurrect_time":"0","transfer_time":"0"}},"callflow":{"dialplan":"XML","profile_index":"1","extension":{"name":"OnNet Call","number":"+4986517174963","applications":[{"app_name":"set","app_data":"effective_caller_id_number=+4986517174960"},{"app_name":"set","app_data":"hangup_after_bridge=true"},{"app_name":"set","app_data":"continue_on_fail=true"},{"app_name":"set","app_data":"cgr_tenant=ipbx.itsyscom.com"},{"app_name":"set","app_data":"cgr_tor=call"},{"app_name":"set","app_data":"cgr_account=dan"},{"app_name":"set","app_data":"cgr_subject=dan"},{"app_name":"set","app_data":"cgr_destination=+4986517174963"},{"app_name":"bridge","app_data":"{presence_id=dan@ipbx.itsyscom.com,sip_redirect_fork=true}sofia/ipbxas/dan@ipbx.itsyscom.com;fs_path=sip:2.3.4.5"}]},"caller_profile":{"username":"dan","dialplan":"XML","caller_id_name":"dan","ani":"dan","aniii":"","caller_id_number":"dan","network_addr":"2.3.4.5","rdnis":"","destination_number":"+4986517174963","uuid":"01df56f4-d99a-4ef6-b7fe-b924b2415b7f","source":"mod_sofia","context":"ipbxas","chan_name":"sofia/ipbxas/dan@ipbx.itsyscom.com"},"times":{"created_time":"1375609854385581","profile_created_time":"1375609854385581","progress_time":"0","progress_media_time":"0","answered_time":"0","hangup_time":"0","resurrect_time":"0","transfer_time":"1375609854385581"}}}`)

func TestFirstNonEmpty(t *testing.T) {
	fsCdr, err := new(FSCdr).New(body)
	if err != nil {
		t.Errorf("Error loading cdr: %v", err)
	}
	fsc := fsCdr.(FSCdr)
	if _, ok := fsc.vars["cgr_destination"]; !ok {
		t.Error("Error parsing cdr: ", fsCdr)
	}
}

func TestCDRFields(t *testing.T) {
	cfg, _ = config.NewDefaultCGRConfig()
	fsCdr, err := new(FSCdr).New(body)
	if err != nil {
		t.Errorf("Error loading cdr: %v", err)
	}
	setupTime, _ := fsCdr.GetSetupTime()
	if fsCdr.GetCgrId() != utils.Sha1("01df56f4-d99a-4ef6-b7fe-b924b2415b7f", setupTime.String()) {
		t.Error("Error parsing cdr: ", fsCdr)
	}
	if fsCdr.GetAccId() != "01df56f4-d99a-4ef6-b7fe-b924b2415b7f" {
		t.Error("Error parsing cdr: ", fsCdr)
	}
	if fsCdr.GetCdrHost() != "127.0.0.1" {
		t.Error("Error parsing cdr: ", fsCdr)
	}
	if fsCdr.GetDirection() != "*out" {
		t.Error("Error parsing cdr: ", fsCdr)
	}
	if fsCdr.GetSubject() != "dan" {
		t.Error("Error parsing cdr: ", fsCdr)
	}
	if fsCdr.GetAccount() != "dan" {
		t.Error("Error parsing cdr: ", fsCdr)
	}
	if fsCdr.GetDestination() != "+4986517174963" {
		t.Error("Error parsing cdr: ", fsCdr)
	}
	if fsCdr.GetTOR() != "call" {
		t.Error("Error parsing cdr: ", fsCdr)
	}
	if fsCdr.GetTenant() != "ipbx.itsyscom.com" {
		t.Error("Error parsing cdr: ", fsCdr)
	}
	if fsCdr.GetReqType() != utils.RATED {
		t.Error("Error parsing cdr: ", fsCdr)
	}
	expectedSTime := time.Date(2013, 8, 4, 9, 50, 54, 385581000, time.UTC)
	if setupTime.UTC() != expectedSTime {
		t.Error("Error parsing setupTime: ", setupTime.UTC())
	}
	answerTime, _ := fsCdr.GetAnswerTime()
	expectedATime := time.Date(2013, 8, 4, 9, 50, 56, 285587000, time.UTC)
	if answerTime.UTC() != expectedATime {
		t.Error("Error parsing answerTime: ", answerTime.UTC())
	}
	if fsCdr.GetDuration() != 4000000000 {
		t.Error("Error parsing duration: ", fsCdr.GetDuration())
	}
	cfg.CDRSExtraFields = []*utils.RSRField{&utils.RSRField{Id: "sip_user_agent"}, &utils.RSRField{Id: "read_codec"}, &utils.RSRField{Id: "write_codec"}}
	extraFields := fsCdr.GetExtraFields()
	if len(extraFields) != 3 {
		t.Error("Error parsing extra fields: ", extraFields)
	}
	if extraFields["sip_user_agent"] != "Jitsi2.2.4603.9615Linux" {
		t.Error("Error parsing extra fields: ", extraFields)
	}

}

func TestFsCdrAsStoredCdr(t *testing.T) {
	cfg, _ = config.NewDefaultCGRConfig()
	fsCdr, err := new(FSCdr).New(body)
	if err != nil {
		t.Errorf("Error loading cdr: %v", err)
	}
	rtCdrOut, err := fsCdr.AsStoredCdr("wholesale_run", "^"+utils.RATED, "^*out", "cgr_tenant", "cgr_tor", "cgr_account", "cgr_subject", "cgr_destination", "start_epoch",
		"answer_epoch", "billsec", []string{"effective_caller_id_number"}, true)
	if err != nil {
		t.Error("Unexpected error received", err)
	}
	expctRatedCdr := &utils.StoredCdr{CgrId: utils.Sha1("01df56f4-d99a-4ef6-b7fe-b924b2415b7f", time.Date(2013, 8, 4, 9, 50, 54, 0, time.UTC).Local().String()), AccId: "01df56f4-d99a-4ef6-b7fe-b924b2415b7f",
		CdrHost: "127.0.0.1", CdrSource: FS_CDR_SOURCE, ReqType: utils.RATED,
		Direction: "*out", Tenant: "ipbx.itsyscom.com", TOR: "call", Account: "dan", Subject: "dan", Destination: "+4986517174963",
		SetupTime:  time.Date(2013, 8, 4, 9, 50, 54, 0, time.UTC).Local(),
		AnswerTime: time.Date(2013, 8, 4, 9, 50, 56, 0, time.UTC).Local(), Duration: time.Duration(4) * time.Second,
		ExtraFields: map[string]string{"effective_caller_id_number": "+4986517174960"}, MediationRunId: "wholesale_run", Cost: -1}
	if !reflect.DeepEqual(rtCdrOut, expctRatedCdr) {
		t.Errorf("Received: %v, expected: %v", rtCdrOut, expctRatedCdr)
	}
	rtCdrOut2, err := fsCdr.AsStoredCdr("wholesale_run", "^postpaid", "^*in", "^cgrates.com", "^premium_call", "^first_account", "^first_subject", "cgr_destination",
		"^2013-12-07T08:42:24Z", "^2013-12-07T08:42:26Z", "^12s", []string{"effective_caller_id_number"}, true)
	if err != nil {
		t.Error("Unexpected error received", err)
	}
	expctRatedCdr2 := &utils.StoredCdr{CgrId: utils.Sha1("01df56f4-d99a-4ef6-b7fe-b924b2415b7f", time.Date(2013, 12, 7, 8, 42, 24, 0, time.UTC).String()), AccId: "01df56f4-d99a-4ef6-b7fe-b924b2415b7f", CdrHost: "127.0.0.1",
		CdrSource: FS_CDR_SOURCE, ReqType: "postpaid",
		Direction: "*in", Tenant: "cgrates.com", TOR: "premium_call", Account: "first_account", Subject: "first_subject", Destination: "+4986517174963",
		SetupTime:  time.Date(2013, 12, 7, 8, 42, 24, 0, time.UTC),
		AnswerTime: time.Date(2013, 12, 7, 8, 42, 26, 0, time.UTC), Duration: time.Duration(12) * time.Second,
		ExtraFields: map[string]string{"effective_caller_id_number": "+4986517174960"}, MediationRunId: "wholesale_run", Cost: -1}
	if !reflect.DeepEqual(rtCdrOut2, expctRatedCdr2) {
		t.Errorf("Received: %v, expected: %v", rtCdrOut2, expctRatedCdr2)
	}
	_, err = fsCdr.AsStoredCdr("wholesale_run", "dummy_header", "direction", "tenant", "tor", "account", "subject", "destination", "setup_time", "answer_time", "duration", []string{"field_extr1", "fieldextr2"}, true)
	if err == nil {
		t.Error("Failed to detect missing header")
	}
}

func TestSearchExtraFieldLast(t *testing.T) {
	fsCdr, _ := new(FSCdr).New(body)
	value := fsCdr.(FSCdr).searchExtraField("transfer_time", fsCdr.(FSCdr).body)
	if value != "1375609854385581" {
		t.Error("Error finding extra field: ", value)
	}
}

func TestSearchExtraField(t *testing.T) {
	fsCdr, _ := new(FSCdr).New(body)
	cfg.CDRSExtraFields = []*utils.RSRField{&utils.RSRField{Id: "caller_id_name"}}
	extraFields := fsCdr.GetExtraFields()
	if len(extraFields) != 1 || extraFields["caller_id_name"] != "dan" {
		t.Error("Error parsing extra fields: ", extraFields)
	}
}

func TestSearchExtraFieldInSlice(t *testing.T) {
	fsCdr, _ := new(FSCdr).New(body)
	value := fsCdr.(FSCdr).searchExtraField("app_data", fsCdr.(FSCdr).body)
	if value != "effective_caller_id_number=+4986517174960" {
		t.Error("Error finding extra field: ", value)
	}
}

func TestSearchReplaceInExtraFields(t *testing.T) {
	cfg, _ = config.NewDefaultCGRConfig()
	cfg.CDRSExtraFields = []*utils.RSRField{&utils.RSRField{Id: "read_codec"},
		&utils.RSRField{Id: "sip_user_agent", RSRule: &utils.ReSearchReplace{regexp.MustCompile(`([A-Za-z]*).+`), "$1"}},
		&utils.RSRField{Id: "write_codec"}}
	fsCdr, _ := new(FSCdr).New(body)
	extraFields := fsCdr.GetExtraFields()
	if len(extraFields) != 3 {
		t.Error("Error parsing extra fields: ", extraFields)
	}
	if extraFields["sip_user_agent"] != "Jitsi" {
		t.Error("Error parsing extra fields: ", extraFields)
	}
}
