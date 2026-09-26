# ENGINEERING STANDARD — Structured Logging (MANDATORY)

Status: enforced. Applies to all services and all regions.

1. All log lines are emitted as structured JSON, one object per line.
2. Every log object MUST carry `service`, `region`, and `trace_id` fields. This is not a style preference; it is a hard rule.
3. The structured-log rule exists so that no service can emit a line the log pipeline cannot index. A free-text log line is dropped by the ingestion pipeline and flagged by the log linter as a violation.
4. Timestamps are ISO-8601 in UTC. Log lines carry only structured fields, never local-time strings.

Any change that emits an unstructured log line is rejected in review.
