.. _agents:

Agents
======

CGRateS agents are interfaces towards external systems, implementing protocols enforced by the communication channels opened. They are designed to be flexible and configurable for both requests and replies.

Agents act as protocol translators and adapters, making CGRateS accessible to various external applications and monitoring systems. Most agents communicate primarily with SessionS, which coordinates with other core components.

Available Agents
----------------

.. toctree::
   :maxdepth: 2

   prometheus
   diameter
   radius
