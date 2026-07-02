# calltest

helpers to place and answer SIP calls against a cgr-engine. the external
process (kamailio, opensips, sipp, voiceblender daemon) is started, polled until
ready, and killed when the test ends

## the shape

every backend is used the same whatever you pick:

    calltest.SipgoUAS{Port: 5070}.Start(t)              // answers, tied to the test
    uac := calltest.SipgoUAC{Addr: "127.0.0.1:5060"}    // places calls
    uac.Call(t, calltest.CallParams{From: "1001", To: "1002", HoldTime: time.Second})

Start and Call do their own readiness wait and cleanup. UAC and UAS are
interfaces, so the caller and answerer below can be swapped.

## pick a caller

only the constructor changes, .Call stays identical:

    calltest.SipgoUAC{Addr: a}                       // pure Go, no media, nothing to install
    calltest.SippUAC{Addr: a, Rate: 1000, Calls: n}  // load and benchmark, e.g. 1000 cps. needs sipp
    calltest.VoiceBlenderUAC{Client: vb, Addr: a}    // real RTP media. needs the voiceblender daemon

sipgo is the default for signaling: routing, CDRs, field mapping. reach for sipp
when you need volume or REGISTER, voiceblender when a switch won't answer
without real media.

## answerers and proxies

    calltest.SipgoUAS{Port: p}                                     // runs in the test process
    calltest.SippUAS{Port: p}                                      // sipp scenario answerer
    calltest.VoiceBlenderUAS{Port: p}                              // daemon answerer, real media
    calltest.Kamailio{ConfigFile: f, ReadyAddr: "127.0.0.1:8448"}  // proxy under test
    calltest.Opensips{ConfigFile: f, ReadyAddr: "127.0.0.1:5060"}  // proxy under test

## requirements

sipgo is a Go dependency, nothing to install. the rest are binaries found on PATH
(or /usr/sbin, /sbin)

- kamailio present, built with the evapi module
- opensips present, built with the cgrates module
- sipp present
- voiceblender present

start cgrates before opensips so the opensips cgrates module connects on startup

## run

    go test -tags call -run TestKamailioLCR ./general_tests/ -dbtype '*internal'

needs cgr-engine on PATH. see kamailio_lcr_test.go for a full test.
