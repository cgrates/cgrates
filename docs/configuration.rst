4. Configuration
=============

The behaviour of **CGRateS** can be externally influenced by following means:

- **Engine configuration files**: usually located at */etc/cgrates/*. 
  There can be one or multiple file(s)/folder(s) hierarchies behind configuration folder with support for automatic includes. 
  The file(s)/folder(s) will be imported **in alphabetical order** into final configuration object.
- **Tariff Plans**: set of files used to import various data used in CGRateS subsystems (eg: Rating, Accounting, LCR, DerivedCharging, etc).
- **RPC APIs**: set of JSON/GOB encoded APIs remotely available for various operational/administrative tasks.

.. toctree::
   :maxdepth: 2

   cgrates_json
   tariff_plans


