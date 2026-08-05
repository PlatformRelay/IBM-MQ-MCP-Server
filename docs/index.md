# IBM MQ MCP Server

This project aims to give AI clients a safe, observable interface to IBM MQ
without assuming that queue managers run on Kubernetes or are managed by a
particular operator.

The design is intentionally capability-oriented: each named connection profile
grants only the operations its operator permits. A production profile may be
read-only while a development profile permits message production or selected
administrative changes.

Start with the [feature scope](product/feature-scope.md) and
[proposed system](architecture/proposed-system.md).

