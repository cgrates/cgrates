Rates history 
=============

Enhances CGRateS with ability to archive rates modifications.

Ability to scale by using server-agents approach. 
In a distributed environment, there will be a single server (which can be backed up using technologies such as Linux-HA) and more agents sending the modifications to be archived.

History-Server
--------------

Part of the *cgr-engine*.

Controlled within *history_server* section of the configuration file.

Stores rates archive in a .git folder, hence making the rating changes available for analysis via any git browser tool (eg: gitg in linux).

Functionality:

- On startup reads the rates archive out of .git folder and caches the data.
- When receiving rates information from the agents it will recompile the rates cache.
- Based on configured save interval it will dump the rating cache (if changed) into the .git archive.
- Archives the following rating data:
  - Destinations inside *destinations.json* file.
  - Rating plans inside *rating_plans.json* file.
  - Rating profiles inside *rating_profiles.json* file.

History-Agent
-------------

Integrated in the rates loader components.

Part of *cgr-engine* and *cgr-loader*.

Enabled via *history_agent* configuration section within *cgr-engine* and *history_server* command line parameter in case of *cgr-loader*.

Sends the complete rating data loaded into ratingDb to *history_server* for archiving.

