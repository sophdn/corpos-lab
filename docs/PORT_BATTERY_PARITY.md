# Battery Runner Port — Parity Audit

Audit shape: **test-suite cross-coverage** (per code-migration-discipline §4 —
both sides demonstrate each behavior via passing tests), plus the two shared
fixtures run verbatim on both sides, plus one live llama-server round-trip.

- Source: lab-app @ `ae6611d` — 56 battery-relevant Rust tests, all passing at
  archive time (PHASE_1_LEDGER.md).
- Target: corpos-lab branch `port-battery-runner` — 57 Go tests, gate green
  (race + 95% floor; internal/ at 98%+).
- Fixtures `known_pass.md` / `known_fail_item2.md` copied byte-for-byte;
  both produce identical outcomes on both sides (pass-all-15 / fail-at-index-1
  with the same reason text).
- Live check 2026-07-08: `go test -tags live` ran Item 1 against llama-server
  (Qwen2.5-32B-Instruct-Q4_K_M) — prose reply parsed into a typed verdict
  through the production client.

## Mapping table (source test → Go test)

**sequences/mod.rs (8)**
| Rust | Go |
|---|---|
| all_pass_sequence | TestAllPassSequence |
| verdict_fail_exits_early | TestVerdictFailExitsEarly |
| observation_then_verdict_reads_data | TestObservationThenVerdictReadsData |
| error_exits_early | TestErrorExitsEarly |
| empty_sequence_passes | TestEmptySequencePasses |
| get_step_result_returns_none_for_missing_name | TestGetStepResultReturnsNilForMissingName |
| step_count_reflects_added_steps | TestStepCountReflectsAddedSteps |
| context_new_starts_with_empty_results | TestStateStartsWithEmptyResults |

**sequences/battery.rs (3)**
| battery_has_15_items_matching_source_doc | TestBatteryHas15ItemsMatchingSourceDoc |
|---|---|
| battery_name_is_battery | TestBatteryNameIsBattery |
| every_registered_step_has_a_version | TestEveryRegisteredStepHasAVersion |

**steps/battery/static_steps.rs (11)**
| item2_passes_clean_entry | TestItem2PassesCleanEntry |
|---|---|
| item2_fails_on_intent_language | TestItem2FailsOnIntentLanguage |
| item2_case_insensitive | TestItem2CaseInsensitive |
| item4_always_emits_not_applicable | TestItem4AlwaysEmitsNotApplicable |
| deferred_item3_emits_deferred_with_pending_reason | TestDeferredItem3EmitsDeferredWithPendingReason |
| every_deferred_stub_emits_deferred | TestEveryDeferredStubEmitsDeferred |
| item10_passes_with_all_axes | TestItem10PassesWithAllAxes |
| item10_fails_missing_y_marker | TestItem10FailsMissingYMarker |
| item10_fails_missing_axes | TestItem10FailsMissingAxes |
| item15_passes_with_field | TestItem15PassesWithField |
| item15_fails_without_field | TestItem15FailsWithoutField |

**steps/battery/verdict_steps.rs (11)**
| parse_pass | TestParsePass |
|---|---|
| parse_fail_with_reason | TestParseFailWithReason |
| parse_fail_without_reason | TestParseFailWithoutReason |
| parse_garbage_response_becomes_fail | TestParseGarbageResponseBecomesFail |
| item1_passes_when_model_says_pass | TestItem1PassesWhenModelSaysPass |
| item1_fails_when_model_says_fail | TestItem1FailsWhenModelSaysFail |
| item1_returns_error_on_model_failure | TestItem1ReturnsErrorOnModelFailure |
| item1_garbage_response_maps_to_fail | TestItem1GarbageResponseMapsToFail |
| item9_passes_when_model_says_pass | TestItem9PassesWhenModelSaysPass |
| item9_fails_when_model_says_fail | TestItem9FailsWhenModelSaysFail |
| item9_returns_error_on_model_failure | TestItem9ReturnsErrorOnModelFailure |

**lab-app-types/src/verdict.rs (18)**
| definition_gap_stops · executor_correctable_requeues · mixed_failures_definition_gap_takes_precedence · unknown_classification_treated_as_definition_gap | TestRouteChainFailure (table, +empty case) |
|---|---|
| display_pass · display_fail_with_item · display_fail_without_item · display_pass_with_condition | TestVerdictDisplay (table, +flag/deferred/NA rows) |
| passed_helper_covers_both_pass_variants · not_applicable_counts_as_passed | TestPassedHelperCoversPassingVariants |
| verdict_round_trips_through_serde_json | TestVerdictRoundTripsThroughJSON + TestVerdictJSONShapeIsKindTagged |
| compose_all_pass_is_promote | TestComposeAllPassIsPromote |
| compose_with_fail_is_block | TestComposeWithFailIsBlock |
| compose_with_only_deferred_is_defer | TestComposeWithOnlyDeferredIsDefer |
| compose_flag_counts_as_block_soft_fail | TestComposeFlagCountsAsBlockSoftFail |
| battery_verdict_display | TestRunVerdictDisplay |
| battery_verdict_round_trips_through_serde | TestRunVerdictRoundTripsThroughJSON |
| rationale_ref_display_matches_inner | — dropped: RationaleRef has no battery consumer (caller grep: study runner only) |

**tests/battery_integration.rs (2 + fixtures)**
| known_pass_item_passes_full_battery | TestKnownPassItemPassesFullBattery (+ step-name order assertion) |
|---|---|
| known_fail_item2_fails_at_intent_language | TestKnownFailItem2FailsAtIntentLanguage (+ exit-index assertion) |

**tests/battery_persistence.rs (3)** — deliberately not ported: persistence
belongs to task 5 (`persist-and-dashboard`). The version-wrapping envelope
(`{"version": …, "outcome": …}`) is documented in the inventory for task 5;
`StepVersion` itself is ported and tested.

**Go-only additions (no Rust counterpart — fill source blind spots):**
TestDeferredAndFlagDoNotStopSequence (the load-bearing fail-fast nuance,
untested in Rust at the runner level) · TestFullBatteryComposeOnKnownPassIsDefer ·
TestComposeFailWithoutItemMapsToItemZeroUnknownClass ·
TestDeferredPendingUnknownItemFallsBack · TestUnknownKindsFallBackSafely ·
TestItem1PromptCarriesEntryContent · TestVerdictGenParamsPinDeterministicSampling ·
7 OpenAI client tests (request shape, param omission, non-200, garbage JSON,
empty choices, transport error, cancellation) · 5 provenance tests ·
TestLiveItem1AgainstLlamaServer (opt-in, `-tags live`).

## Divergences (all accepted + documented in the inventory)

1. `FailureReason` string ("" = none) vs `Option<String>`.
2. `FailureClass` JSON kind-tagged vs serde external tagging (no old-row readers).
3. Provenance at runtime vs build time (fresher; scrubs inherited GIT_* env —
   a hardening the Rust build-time approach never needed).
4. No metrics/tracing emission (no sink in corpos-lab yet).
5. `RationaleRef` not ported (no battery consumer).

No behavioral divergence found in any ported feature: every mapped test
asserts the same inputs → same outcomes, reason strings and exit indices
included, and the shared fixtures agree end-to-end.

## Promotion & rollback

- Source stays runnable at `~/dev/lab-app` (untouched, read-only during this
  port); archive happens in `archive-ancestor-repos` AFTER the parity chain
  reproduces casg-direct v3 on this instrument.
- Rollback: revert the merge commit of branch `port-battery-runner`; the Rust
  battery remains at `lab-app/crates/` with its own test suite.
