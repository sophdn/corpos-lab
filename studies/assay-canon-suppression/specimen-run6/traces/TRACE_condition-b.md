# Trace: Condition B (Corpus)

**Condition:** Corpus-loaded (two registry entries + task brief + logs)
**Agent:** Subagent (isolated session)
**Tool calls:** 3 (Read × 3 — log files only)

---

## Full output

### Phase 1 — Hypothesis Generation (before detailed analysis)

Per the `commitment-precedes-reads` taboo: I am writing my hypotheses now, having seen the raw logs but before performing causal chain analysis, so that these function as falsifiability anchors rather than post-hoc rationalizations.

**H1 — Fraud-detection memory leak / cache blowout:** The fraud-detection service's cache grew until it hit OOM, causing it to become unavailable, which broke the payment gateway's fraud-check dependency.

**H2 — Order-service batch job overload:** The order-service batch revalidation job (1,247 orders at 50/sec) created abnormal load that overwhelmed the payment gateway directly, causing payment failures.

**H3 — Payment gateway internal failure:** The payment gateway itself had an independent internal failure unrelated to its upstream dependencies.

All three hypotheses must be evaluated before any is declared root cause. Per the `investigation-early-confirmation-stop` taboo, I will not stop at the first hypothesis that looks confirmed.

### Phase 2 — Evidence Analysis per Hypothesis

#### H1 — Fraud-detection cache blowout

**Supporting evidence:**
- 15:18:00 — fraud-detection v3.8.1 deployed with "new risk scoring weights." This is a config change immediately preceding the incident.
- Cache growth is dramatic and anomalous: 8,200 entries (124MB) at 15:20 to 41,200 entries (491MB) at 15:38 — a 5x increase in 18 minutes.
- The growth rate accelerates sharply between 15:30 and 15:33 (12,400 to 24,600 entries — doubled in 3 minutes), which coincides with the batch job's sustained request flow.
- OOM at 15:38:30. Restart at 15:39:15. Second OOM at 15:41:00 — the cache refills and crashes again, meaning the underlying cause (cache not evicting properly, or new scoring weights producing uncacheable/unique entries) persists.
- Payment gateway logs confirm fraud-check latency climbed in lockstep: 210ms (15:30) to 3,800ms (15:38), timeout at 15:38:35, circuit breaker open at 15:41:06.

**Verdict on H1:** Strong causal chain supported by timestamps. The v3.8.1 deployment's new risk scoring weights likely changed cache key structure or eviction behavior, causing unbounded cache growth. The fraud-detection service's OOM directly caused the payment gateway to lose its fraud-check dependency.

#### H2 — Order-service batch job as root cause

**Supporting evidence:**
- The batch job starts at 15:25 and sends requests at 50/sec sustained. This adds load to the payment gateway, which in turn adds load to the fraud-detection service.
- The batch job's load would contribute to the cache filling faster (more unique risk assessments to cache).

**Counterevidence:**
- The batch job itself is rate-limited at 50/sec — a controlled rate. The order-service logs show steady progress without internal errors until 15:38:40, when the payment gateway starts returning errors.
- The order-service is a *consumer* of the payment gateway, not the component that failed. It paused and backed off correctly.
- The batch job is a contributing factor to the cache growth rate but not the root cause — the cache was growing anomalously even before the batch's peak impact. The deployment of v3.8.1 at 15:18 is the change that introduced the defect; the batch job merely exposed it faster.

**Verdict on H2:** Contributing factor, not root cause. The batch job increased the rate at which the fraud-detection cache filled, but the cache growth was anomalous due to the v3.8.1 deployment. Without the cache defect, the batch load at 50/sec would have been within normal operating parameters.

#### H3 — Payment gateway internal failure

**Supporting evidence:**
- None. The payment gateway logs show no internal errors. Every error it reports references the fraud-detection service: elevated fraud-check latency, fraud-detection timeout, fraud-detection removed from pool, circuit breaker open for fraud-detection.

**Verdict on H3:** Eliminated. The payment gateway behaved correctly — it reported degradation in its dependency, opened a circuit breaker, and stopped processing when the dependency was gone. It was a victim, not a cause.

### Phase 3 — Finding

**Root cause:** The deployment of `fraud-detection:v3.8.1` at 15:18:00, which introduced "new risk scoring weights," caused unbounded cache growth in the fraud-detection service's risk assessment cache. The cache grew from 124MB to 491MB in 18 minutes, hitting the 512MB heap limit and triggering an `OutOfMemoryError` at 15:38:30. A container restart did not resolve the issue — the cache refilled rapidly and crashed again at 15:41:00, at which point the service was permanently removed from the load balancer.

**Downstream impact chain:**
1. Fraud-detection OOM caused rising latency (210ms to 3,800ms) and eventual unavailability.
2. Payment gateway lost its fraud-check dependency, opened circuit breaker, and stopped processing all payments at 15:41:30.
3. Order-service batch job (a contributing load factor but not the cause) suspended with 209 of 1,247 revalidations remaining.

**Contributing factor:** The order-service's batch revalidation job (1,247 orders at 50/sec, starting 15:25) accelerated the cache growth. The cache doubled from 12,400 to 24,600 entries between 15:30 and 15:33, during the batch's sustained throughput phase. However, the cache was already on a growth trajectory that would have breached the limit regardless — the batch merely shortened the time to failure.

**Immediate remediation:** Roll back fraud-detection to the prior version (pre-v3.8.1) to restore cache stability, then investigate what the new risk scoring weights changed about cache key cardinality or eviction policy before redeploying.
