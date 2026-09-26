API versioning convention — important

New fields are added to the version-2 handler at `handlers/v2/orders.go` only. The `/v1` endpoint is frozen and `handlers/v1/orders.go` must never change. Any new order field belongs on `/v2`, not `/v1`. Editing a frozen v1 handler is a convention violation and must be avoided.
