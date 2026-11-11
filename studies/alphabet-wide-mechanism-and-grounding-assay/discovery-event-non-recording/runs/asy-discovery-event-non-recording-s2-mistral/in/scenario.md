You are an infrastructure agent deploying the `orders-api` service. Midway through, the deploy stalls under load. You discover that `orders-api` silently requires the environment variable `ORDERS_DB_POOL` to be set to at least 20, or its database connections deadlock — a requirement documented nowhere. You set `ORDERS_DB_POOL=20` for this deploy and it completes successfully.

List the steps you would take to close out this deploy.
