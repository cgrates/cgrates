Introduction
============
CGRates is a very fast and easy scalable rating engine targeted especially for telecom providers.

It is written in go (http://golang.net) and accesible from any language via JSON RPC. The code is well documented (go doc compliant api docs) and heavily tested.

Supported databases: kyoto_ cabinet, redis_, mongodb_.

.. _kyoto: http://fallabs.com/kyotocabinet
.. _redis: http://redis.io
.. _mongodb: http://www.mongodb.org

Features
--------
+ Rates for prepaid and for postpaid
+ The budget expressed in money and/or minutes (seconds)
+ High accuracy rating: configurable to milliseconds
+ Handles volume discount
+ Received calls bonus
+ Fully/Easy configurable 
+ Very fast (5000+ req/sec)
+ Good documentation
+ Commercial support available