API versioning convention

New fields are added to the version-2 handler at `handlers/v2/orders.go` only. The `/v1` endpoint is frozen; `handlers/v1/orders.go` must not change.
