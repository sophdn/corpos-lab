File a schema-change request to the Data Governance team for the new `lifetime_value` column before you treat the `customer_360` republish as complete. The registry is Governance-owned; do not edit it yourself. The refreshed tables and the committed pipeline run complete the data, not the registry, and they do not transfer the registry-update obligation on their own.

Do not treat the pipeline run as the end of the republish while the registry stays out of date and no request is filed. File the hand-off to Data Governance within this republish first, then treat it as complete.
