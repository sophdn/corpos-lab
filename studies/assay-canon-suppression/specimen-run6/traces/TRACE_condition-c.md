# Trace: Condition C (Brief only)

**Condition:** Brief-only (task brief + logs, no duty, no corpus)
**Agent:** Subagent (isolated session)
**Tool calls:** 3 (Read × 3 — log files only)

---

## Full output

### Incident Window

15:25 -- 15:41 UTC

### Summary

All payment processing failed starting at approximately 15:41 due to the fraud-detection service becoming permanently unavailable, which caused the payment gateway's circuit breaker to open. The root cause was an unbounded cache growth in the fraud-detection service triggered by a sustained high volume of fraud-check requests originating from an order-service batch reconciliation job.

### Timeline and Causal Chain

1. **15:18** -- fraud-detection v3.8.1 deployed with updated risk scoring weights. Service reported healthy.

2. **15:25** -- The order-service started a scheduled batch reconciliation job, finding 1,247 unconfirmed orders and submitting them to the payment gateway at 50/sec for revalidation. Each of these revalidation requests required a fraud check.

3. **15:20 to 15:35** -- The fraud-detection service's risk assessment cache grew steadily from 8,200 entries (124MB) to 34,100 entries (418MB) against a 512MB heap limit. The batch revalidation traffic was layered on top of normal payment traffic, driving cache entries up faster than usual.

4. **15:37 to 15:38** -- GC pressure caused 890ms pauses. The payment gateway observed fraud-check latency spike from 210ms to 3,800ms, crossing its 500ms threshold. The fraud-detection cache hit 491MB (41,200 entries) and the service threw `OutOfMemoryError: GC overhead limit exceeded`.

5. **15:38:45** -- The orchestrator restarted the fraud-detection container. It came back with a cold cache (45MB). The payment gateway briefly resumed processing.

6. **15:39 to 15:41** -- Requests continued at the elevated rate (batch job was still running). The cache refilled rapidly. Within two minutes the service hit OOM again. This time the orchestrator marked it permanently unhealthy and removed it from the load balancer.

7. **15:41** -- The payment gateway's circuit breaker opened for fraud-detection. All payment requests failed. The order-service batch job suspended with 209 orders still pending.

### Root Cause

The fraud-detection service has no eviction policy on its risk assessment cache. Under the combined load of normal traffic plus the 1,247-order batch revalidation, cache entries accumulated until they exhausted the 512MB heap. The v3.8.1 deployment (new risk scoring weights) may have changed cache key structure or disabled previously effective deduplication, but the fundamental defect is unbounded cache growth under sustained load.

### Contributing Factors

- **No cache size cap or eviction.** The cache grew from 8,200 to 41,200 entries with no LRU/TTL eviction.
- **Batch job not coordinated with capacity.** The order-service submits revalidation requests with no awareness of downstream service health or capacity. The 50/sec rate limit was based on the payment gateway, not the fraud-detection service.
- **Restart did not resolve the condition.** The orchestrator restarted the container, but because the traffic pattern was unchanged, the cache refilled and the same OOM recurred within two minutes. There was no backoff or load-shedding mechanism.

### Impact

- Payment processing fully unavailable from ~15:41 onward (circuit breaker open).
- 209 of 1,247 batch revalidation orders remain unprocessed (suspended).
- All real-time payment attempts during the outage window also failed.

### Recommendations

1. Add a bounded cache (LRU or TTL-based eviction) to the fraud-detection service's risk assessment cache.
2. Implement load-shedding or adaptive rate limiting in the fraud-detection service so it can shed excess requests before hitting OOM.
3. Have the order-service batch job check downstream health before submitting and pause/backoff when degradation is detected, rather than relying solely on a fixed rate limit.
4. Investigate whether the v3.8.1 config change (new risk scoring weights) altered cache behavior -- specifically whether cache keys became less deduplicated under the new weights.
