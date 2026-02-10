.. _rsr_parser:

RSR Parser
==========

RSR Parser extracts and transforms data from events. Used throughout CGRateS for field mapping, filtering, and data conversion.

A single rule has three parts:

1. **Path** - where to get the data
2. **Search/Replace** - optional regex transformation
3. **Converters** - optional type/format conversion

Single rule syntax::

    ~path
    ~path:s/search/replace/
    ~path{*converter}
    ~path:s/search/replace/{*converter1&*converter2}

Multiple rules can be chained with ``;`` - their outputs are concatenated::

    ~*req.Account;~*req.Destination

With Account=``1001`` and Destination=``+111``, this produces ``1001+111``.


Paths
-----

Dynamic paths start with ``~`` and use a prefix to specify the data source, some examples:

================================  ==========================================
Prefix                            Source
================================  ==========================================
``~*req.FieldName``               Request/event fields
``~*rep.FieldName``               Reply fields
``~*vars.FieldName``              Session variables 
``~*opts.FieldName``              Options values
``~*cgreq.FieldName``             CGRateS request 
``~*cgrep.FieldName``             CGRateS reply 
================================  ==========================================

Static values without the ``~`` prefix are used as literal text.

**Nested paths** use dot notation::

    ~*req.Service-Information.SMS-Information.Recipient-Address

**Indexed access** for array fields::

    ~*req.Filters[0]


Search & Replace
----------------

Regex transformations use ``:s/pattern/replacement/`` syntax. Chain multiple::

    ~*req.Destination:s/^\+33/0/:s/^0033/0/

Converts prefixes: ``+33123`` or ``0033123`` becomes ``0123``.

Use ``${1}``, ``${2}`` etc. for capture groups::

    ~*req.Phone:s/^\+(\d{2})(\d+)$/($1) $2/

Converts ``+441234567`` to ``(44) 1234567``.


Dynamic Rules
-------------

Parts of a rule can be built at runtime using ``<>`` delimiters. The content inside is evaluated first::

    THD_ACNT_<~*req.Account>

``Account`` with value ``1001`` produces rule ``THD_ACNT_1001``.


Escaping
--------

When a value contains ``;`` and you don't want it split into multiple rules, wrap it in backticks::

    constant;`>;q=0.7;expires=3600`;~*req.Account

This parses as three rules: ``values``, ``>;q=0.7;expires=3600`` (literal), and ``~*req.Account``.


.. _data_converters:

Converters
----------

Converters transform the extracted value. Chain them with ``&``::

    ~*req.Usage{*duration_seconds&*round:2}

**Duration**

* ``*duration_seconds`` - to seconds (float64)
* ``*duration_nanoseconds`` - to nanoseconds (int64)
* ``*duration_minutes`` - to minutes (float64)
* ``*duration`` - parse as Go duration
* ``*durfmt:layout`` - format as time (default: ``15:04:05``)

**Math**

* ``*round:decimals`` - round to N decimal places using ``*middle`` rounding
* ``*round:decimals:method`` - round with method: ``*middle``, ``*up``, or ``*down``
* ``*multiply:value`` - multiply by value
* ``*divide:value`` - divide by value

**String**

* ``*strip:side:char`` - trim all occurrences of char; sides: ``*prefix``, ``*suffix``, ``*both``
* ``*strip:side:char:count`` - trim up to count occurrences
* ``*len`` - string/slice length
* ``*slice`` - parse as slice
* ``*json`` - marshal to JSON

**Phone/Network**

* ``*libphonenumber:country`` - format phone as NATIONAL (format 2)
* ``*libphonenumber:country:format`` - format phone; formats: 0=E164, 1=INTERNATIONAL, 2=NATIONAL, 3=RFC3966
* ``*e164`` - extract E.164 from NAPTR record
* ``*e164Domain`` - extract domain from NAPTR record
* ``*ip2hex`` - IP to hex
* ``*sipuri_host``, ``*sipuri_user``, ``*sipuri_method`` - parse SIP URIs
* ``*3gpp_uli`` - decode 3GPP-User-Location-Info hex to ULI object
* ``*3gpp_uli:path`` - extract specific field from ULI

ULI component paths: ``CGI``, ``SAI``, ``RAI``, ``TAI``, ``ECGI``, ``TAI5GS``, ``NCGI``

Field paths: ``TAI.MCC``, ``TAI.MNC``, ``TAI.TAC``, ``ECGI.MCC``, ``ECGI.MNC``, ``ECGI.ECI``, ``NCGI.NCI``, etc.

Example::

    ~*req.3GPP-User-Location-Info{*3gpp_uli:TAI.MCC}

**Time**

* ``*unixtime`` - parse to Unix timestamp
* ``*timestring`` - format time using ``Local`` timezone and ``2006-01-02 15:04:05`` layout
* ``*timestring:tz`` - format with timezone (e.g., ``*timestring:UTC``)
* ``*timestring:tz:layout`` - format with timezone and custom layout

**Type**

* ``*float64`` - parse as float64
* ``*string2hex`` - string to hex
* ``*urldecode`` - URL decode
* ``*urlencode`` - URL encode

**Other**

* ``*random`` - random int
* ``*random:min`` - random int >= min
* ``*random:min:max`` - random int in range
* ``*conn_status`` - UP=1, DOWN=-1
* ``*gigawords`` - gigawords to octets (multiply by 2^32)


Examples
--------

Build an email from account and domain::

    ~*req.Account;~*req.Domain

Round cost at 2 decimal value::

    ~*req.Cost{*round:2:*up}

Extract user from SIP URI and format as E.164::

    ~*req.From{*sipuri_user&*libphonenumber:US:0}


Convert AnswerTime to Go time format::

    ~*req.AnswerTime{*timestring::2006-01-02 15:04:05.999999999 -0700 MST}
