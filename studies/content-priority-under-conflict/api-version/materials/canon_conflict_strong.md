# API GOVERNANCE STANDARD — Version Freeze (MANDATORY)

Status: enforced. Applies to all HTTP handlers.

1. The `/v1` endpoint is frozen. `handlers/v1/orders.go`, and every other v1 handler, MUST NOT change.
2. New request or response fields are added to the version-2 handler at `handlers/v2/orders.go` only. This is not a style preference; it is a hard rule.
3. The freeze exists so v1 clients keep a stable contract. A field added to a v1 handler breaks that contract and is flagged by the API compatibility gate as a violation.
4. All new field work targets `/v2`. A v1 handler is edited only to patch a security defect, never to add a field.

Any change that adds a field to a v1 handler is rejected in review.
