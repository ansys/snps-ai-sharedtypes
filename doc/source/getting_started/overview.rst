Overview
========

SNPS-AI (Ansys Automation and Learning Intelligence) is a distributed system for orchestrating workflows and executing functions. It consists of multiple services that communicate via gRPC and WebSocket protocols.

Architecture
------------

.. code-block:: text

   ┌─────────────────────────────────────────────────────────┐
   │                      User/Client                        │
   └────────────────────┬────────────────────────────────────┘
                        │
                        │ WebSocket/gRPC
                        ▼
   ┌─────────────────────────────────────────────────────────┐
   │                     SNPS-AI Agent                       │
   │         (Workflow orchestration & routing)              │
   └────────────┬──────────────────────────┬─────────────────┘
                │                          │
                │ gRPC                     │ gRPC
                ▼                          ▼
   ┌───────────────────────┐    ┌─────────────────────────┐
   │    SNPS-AI FlowKit    │    │       SNPS-AI Exec      │
   │  (Go/Python functions)│    │   (Remote execution)    │
   └───────────────────────┘    └─────────────────────────┘

   ═══════════════════════════════════════════════════════════
                    ▲               ▲               ▲
                    │               │               │
                    └───────────────┴───────────────┘
                    SNPS-AI Shared Types
                    (Common type definitions)

Why Shared Types?
-----------------

Synopsys AI Tools Shared Types ensures all services speak the same language by providing:

- **Type definitions** for functions, sessions, and data structures
- **gRPC protocols** (``agentgrpc``, ``flowkitgrpc``) for service communication
- **Type converters** for JSON/Go transformations
- **Configuration structures** for service setup

When services like the Agent send function requests to FlowKit, or when FlowKit returns results, they all use these common type definitions to ensure data compatibility.

What's Next?
------------

Continue with the :doc:`prerequisites <prerequisites>` to set up your development environment.
