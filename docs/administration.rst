5. Administration
=================

The general steps to get CGRateS operational are:

#. Create CSV files containing the initial data for CGRateS.
#. Load the data in the databases using the Loader application.
#. Start the a Balancer or a Rater. If Balancer is used, start one or more Raters serving that Balancer.
#. Start the SessionManager talking to your VoIP Switch or directly make API calls to the Balancer/Rater.
#. Make API calls to the Balancer/Rater or just let the SessionManager do the work.

