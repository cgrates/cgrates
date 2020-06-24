Tutorial SIP Redirect
================

Scenario:
---------

- FreeSWITCH with minimal custom configuration. 

 - Added following users (with configs in *etc/freeswitch/directory/default*): 1001-prepaid, 1002-postpaid, 1003-pseudoprepaid, 1004-rated, 1006-prepaid, 1007-prepaid.
 - Have added inside default dialplan a redirect for the destinatin 1001 to the CGRateS SIPAgent that will populate the Contact header and add X-Identity header

- **CGRateS** with following components:

 - RALs to calculate the maxusage
 - RouteS to get the rounting information
 - ChargerS to change the destination for each route
 - SIPAgent to comunicate with FreeSWITCH
 - SessionS to comunicate with the agent
