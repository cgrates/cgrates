Rating history
==============

Enhances CGRateS with ability to archive rates modifications.

Large scaling posibility using server-agents approach.
In a distributed environment, there will be a single server (which can be backed up using technologies such as Linux-HA) and more agents sending the modifications to be archived.

History-Agent
-------------

Integrated in the rating loader components.

Part of *cgr-engine* and *cgr-loader*.

Enabled via *history_agent* configuration section within *cgr-engine* and *history_server* command line parameter in case of *cgr-loader*.

Sends the complete rating data loaded into dataDb to *history_server* for archiving.

