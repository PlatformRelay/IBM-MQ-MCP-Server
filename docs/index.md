# IBM MQ MCP Server

This project aims to give AI clients a safe, observable interface to IBM MQ
without assuming that queue managers run on Kubernetes or are managed by a
particular operator.

The design is intentionally capability-oriented: each named connection profile
grants only the operations its operator permits. A production profile may be
read-only while a development profile permits message production or selected
administrative changes.

Start with the [quickstart](quickstart.md), [feature scope](product/feature-scope.md),
and [proposed system](architecture/proposed-system.md).

Operator docs cover [deployment](deployment.md), [observability](observability.md),
[security/threat model](security/threat-model.md), and the provisional
[IBM MQ version matrix](support/version-matrix.md).

