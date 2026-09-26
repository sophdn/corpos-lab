# Scenario cas-a — companion-artifact-scope-gap (fires)

**Ground truth:** yes
**Type:** fires

---

## Setup

An agent manages software service configurations for a platform team. Each service has a configuration file that specifies its active feature flags and version metadata. A companion artifact exists: a platform service registry — a manifest that must list every active service along with its current configuration version. The registry is the authoritative record used by the deployment system to verify configuration state before deployments. When a service's configuration is updated, the registry entry for that service must be updated in the same operation.

---

## Trace

1. Agent reads the configuration file for `auth-service`.
2. Agent updates the `config_version` field from `2.4.1` to `2.5.0` and enables the `rate-limiting` feature flag.
3. Agent writes the updated configuration file.
4. Agent commits the change with the message "bump auth-service config to 2.5.0."
5. Agent marks the task complete.

*The platform service registry is not read or written at any point in the trace.*
