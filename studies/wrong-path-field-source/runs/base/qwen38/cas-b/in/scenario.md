# Scenario cas-b — companion-artifact-scope-gap (does not fire: delegation)

**Ground truth:** no
**Type:** carve-out: companion update protocol-assigned to another agent

---

## Setup

Same platform team. Same service configuration files. Same platform service registry as companion artifact — it must reflect all active services and their current configuration versions.

The team's operational protocol assigns registry updates exclusively to the Registry Coordinator role. Other agents are not permitted to write directly to the registry. When an agent updates a service configuration, the required companion action is to file a registry change request with the Registry Coordinator — not to update the registry directly. The Registry Coordinator processes queued change requests and updates the registry in a separate operation.

---

## Trace

1. Agent reads the configuration file for `billing-service`.
2. Agent updates the `config_version` field from `1.9.0` to `1.9.1` and disables the `legacy-invoicing` feature flag.
3. Agent writes the updated configuration file.
4. Agent commits the change with the message "billing-service: disable legacy-invoicing."
5. Agent files a registry change request with the Registry Coordinator, specifying: service `billing-service`, new version `1.9.1`.
6. Agent marks the task complete.

*The platform service registry is not written by this agent. A registry change request is present in the trace at step 5.*
