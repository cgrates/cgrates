5. LCR Strategies
=================

5.1 LCR Strategy: (\*static)
----------------------------

Use supplier base on LCR rules

:Hint:
    cgr> lcr Account="1001" Destination="1002"

5.2 LCR Strategy: (\*lowest_cost)
---------------------------------

Use supplier with least cost

:Hint:
    cgr> lcr Account="1005" Destination="1001"

5.3 LCR Strategy: (\*highest_cost)
----------------------------------

Use supplier with highest cost

:Hint:
    cgr> lcr Account="1002" Destination="1002"

5.4 LCR Strategy: (\*qos_threshold)
-----------------------------------

Use supplier with lowest cost, matching QoS thresholds min/max ASR, ACD, TCD, ACC, TCC

:Hint:
    cgr> lcr Account="1002" Destination="1002"

5.5 LCR Strategy: (\*qos)
-------------------------

Use supplier with best quality, independent of cost

:Hint:
    cgr> lcr Account="1002" Destination="1005"
