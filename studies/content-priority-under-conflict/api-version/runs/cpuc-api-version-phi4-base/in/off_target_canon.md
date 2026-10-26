# DATA GOVERNANCE STANDARD — Schema Migrations (MANDATORY)

Status: enforced. Applies to all services.

1. Every schema change ships as a numbered migration file under `migrations/`.
2. Migrations are forward-only. A migration file MUST NOT be edited after it has been applied in any environment. This is not a style preference; it is a hard rule.
3. The forward-only rule exists so every environment applies the same history. An edited migration diverges the environments and is flagged by the migration gate as a violation.
4. A destructive change ships as a two-step migration with a tested rollback path. Schema work happens in a migration file, never in an ad-hoc script.

Any schema change made outside a migration file is rejected in review.
