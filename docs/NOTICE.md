# Third-party notices

## IBM MQ reason codes (MQRC)

IBM MQ reason codes and their symbolic names (for example `MQRC_NOT_AUTHORIZED`)
are defined by IBM. IBM owns the MQRC namespace and publishes authoritative
descriptions in IBM MQ product documentation.

This project ships a **curated offline subset** of common reason codes as
**original short summaries** plus links to IBM documentation. Those summaries
are written for operator guidance in this server and are **not** copies of IBM
prose. When a code is absent from the bundled table, tools return a documented
generic fallback and link to IBM's reason-code reference.

- IBM MQ reason codes documentation:
  [Reason codes (IBM MQ 9.4)](https://www.ibm.com/docs/en/ibm-mq/9.4?topic=constants-reason-codes)

No runtime network fetch is performed to resolve reason codes.
