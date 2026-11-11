# Companion-artifact update — rules

**When this applies:** You are completing an operation on a primary artifact, and you hold direct authority to update a companion artifact whose validity depends on reflecting the current state of the primary artifact set. No protocol assigns the companion update to another role — updating the companion is part of your execution obligation. Completing the primary artifact action will feel like the end of the operation. You must decide whether to treat the primary artifact update as the terminal step, or to include the companion artifact update as an equally required component of this same operation.

**What to do:**
1. Before you close the operation, update the companion artifact so its state reflects the current state of the primary artifact set. Perform this update within the same operation scope as the primary artifact action — both actions present before the operation is marked complete.
2. Do not treat the primary artifact action as terminal while the companion update is outstanding. Closing with the primary artifact updated and the companion artifact not updated leaves the obligation unmet, with no transfer filed.
3. You have completed this correctly when both the primary artifact action and the companion artifact update are present in the same operation, before it closes.

**When this does NOT apply:**
- No companion artifact exists whose validity depends on reflecting the current state of the primary artifact set; or this operation does not affect the primary artifact in a way that creates a companion update obligation. Then there is no companion update to include, and no action is required beyond the primary artifact.
- A companion artifact exists and its update is due, but a protocol assigns that update to another role and you do not hold direct update authority. Then your obligation is to file a delegation hand-off to the assigned role, not to perform the companion update yourself.

**Example:** You mark a work item complete within a tracked workflow. You update the work-item status, but you do not update the parent tracking document's status record to reflect the completed work item in the same operation. The companion artifact — the parent tracking document — was left out of the operation scope. Correct action is to update the parent tracking document in the same step as the work-item completion, before closing.
