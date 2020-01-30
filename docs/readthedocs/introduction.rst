1. Introduction
===============

`CGRateS` is a *very fast* and *easily scalable* **(charging, rating, accounting, lcr, mediation, billing, authorization)** *ENGINE* targeted especially for ISPs and Telecom Operators. It allow users provisioning and tarif plan management.

It is written in `Go` programming language and is accessible from any programming language via JSON RPC.

*Usage example through cgr-console*

:Hint:
    cgr> Accounts Tenant="cgrates.org" AccountIDs=["1001"]

*Usage example through postman*

URL: http://your_server_ip:2080/jsonrpc

*Request*

::

    {"method":"APIerSv2.GetAccounts","params":[{"Tenant":"cgrates.org","AccountIds":["1001"],"Offset":0,"Limit":0}],"id":3}
    Content-Type: application/json
