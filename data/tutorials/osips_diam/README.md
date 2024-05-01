# Prepaid Scenario

1. **INVITE**: 1001 calls 1002.
2. **Send INITIAL_REQUEST CCR**: Before forwarding INVITE, have OpenSIPS send a Diameter Credit-Control INITIAL_REQUEST to authorize the call.
3. **Receive CCA**: Extract CC-Time from the reply's Granted-Service-Unit AVP and (if also authorized) then set the dialog timeout to that value.
4. **ACK**: 1002 answers the call, `dlg_on_answer` handler is triggered.
5. **Send UPDATE_REQUEST CCR**: Send an async Credit-Control request to CGRateS to initiate the session. (see bottom of cfg file to see why, same for the other async calls). No need for further updates as CGRateS will debit based on the configured `debit_interval`.
6. **SEND TERMINATION_REQUEST**: Wait for hangup/timeout. The `handle_hangup/handle_auth` handler will be triggered and send a Diameter Credit-Control TERMINATION_REQUEST. Both of those are almost identical, only differing AVP being Terminate-Cause (6/DIAMETER_AUTH_EXPIRED for timeout AND 1/DIAMETER_LOGOUT for hangup). This will terminate the session as well as process the CDR.

# Postpaid Scenario

The Postpaid scenario is the same as the above except:
- We only send an **EVENT_REQUEST** on `dlg_on_timeout/dlg_on_hangup` handlers to send a ProcessCDR request to CGRateS.
- Inside `dlg_on_answer` the answer time is recorded to calculate usage at the end.

# Accounting Scenario

Attempts to use the `do_accounting` method (part of the OpenSIPS accounting module) to send Accounting-Start and Accounting-Stop requests. Currently, it's only sending Accounting-Start on hangup which is not intended.

# Useful commands

- sudo -u opensips /usr/sbin/opensips -f /etc/opensips/opensips.cfg -m 64 -M 4 -D
- cgr-engine -config_path=/usr/share/cgrates/tutorials/osips_diam/etc/cgrates -logger=*stdout
- cgr-loader -path /usr/share/cgrates/tutorials/osips_diam/tp/ -verbose
- freeDiameterd -dd -c /etc/freeDiameter/freeDiameter.conf
- pjsua_listen --help
- pjsua_listen --accounts 1001,1002
- pjsua_call --help
- pjsua_call --from 1001 --to 1002 --dur 67

