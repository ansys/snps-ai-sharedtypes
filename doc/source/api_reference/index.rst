API reference
=============

.. note::

   Complete API documentation is automatically generated during the CI/CD build process.

The AALI Shared Types library contains the following packages:

Core Packages
-------------

.. list-table::
   :widths: 30 70
   :header-rows: 1

   * - Package
     - Description
   * - **sharedtypes**
     - Core type definitions used across all AALI services (agents, functions, databases, LLM handlers)
   * - **typeconverters**
     - Utilities for converting between JSON, Go types, and string representations

gRPC Protocol Packages
----------------------

.. list-table::
   :widths: 30 70
   :header-rows: 1

   * - Package
     - Description
   * - **agentgrpc**
     - Protocol buffer definitions and gRPC service for AALI Agent communication
   * - **flowkitgrpc**
     - Protocol buffer definitions and gRPC service for AALI FlowKit communication

Utility Packages
----------------

.. list-table::
   :widths: 30 70
   :header-rows: 1

   * - Package
     - Description
   * - **config**
     - Configuration management utilities for AALI services
   * - **logging**
     - Structured logging with Datadog integration
   * - **clients**
     - Client implementations for FlowKit (Go and Python)
   * - **graphdb**
     - GraphDB client with logical types and value handling


.. toctree::
   :hidden:
   :maxdepth: 1

   test/index

.. button-ref:: ../index
    :ref-type: doc
    :color: primary
    :shadow:
    :expand:

    Go back to landing page
