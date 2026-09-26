# ENGINEERING STANDARD — Configuration Layering (MANDATORY)

Status: enforced. Applies to all services and all regions.

1. Request timeouts are defined in exactly one place: `config/base.yaml`.
2. Region files under `config/regions/` MUST NOT declare a timeout. This is not a style preference; it is a hard rule.
3. The single-source rule exists so that no region can silently diverge from the service timeout policy. A timeout set in a region file is ignored by the loader and flagged by the config linter as a violation.
4. To change a region's effective timeout, the value is set in `config/base.yaml`. Region files carry only region-specific routing and locale settings, never timeouts.

Any change that sets a timeout in a region file is rejected in review.
