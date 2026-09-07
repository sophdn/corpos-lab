# Conditional-gate evaluation — rules

**When this applies:** You are at a workflow transition point where the correct next path depends on evaluating an explicit conditional check. The condition has several possible states, and the branching logic looks complex. Defaulting to a uniform strategy — always taking a path, or never taking it — will feel like a reliable simplification: it settles the decision without the evaluation overhead and avoids conditional logic that seems hard to track reliably under execution pressure. The uniform strategy will feel complete, and you may not perceive the conditional check as a required gate. You must decide whether to evaluate the conditional and take the path it determines, or apply the uniform strategy.

**What to do:**
1. Before you take — or withhold — the transition action, evaluate the conditional check and put its result on record in this decision sequence.
2. Take the path the evaluated condition determines. Do not settle the transition with a uniform default in place of evaluating the condition: neither always taking the path nor always withholding it. Both directions of the uniform default skip the gate.
3. The reason to evaluate is not that you are unsure which path the condition would give — it may well produce the same action you would have taken. It is that the check is a required gate. You have done this correctly when a condition result is on record before the transition action and the path taken reflects that result.

**When this does NOT apply:**
- The transition has no conditional branching — it is single-path, or every path is equivalent regardless of state. Then there is no condition to evaluate.
- The condition is unambiguously fixed by structural necessity: arriving at the transition point is itself enough to determine which path applies, with no separate evaluation step (for example, the condition is a prerequisite artifact's presence or absence, and that is directly observable at the transition). Then no separate gate needs to be evaluated.

**Example:** At the end of a task you run a conditional action — or consistently do not run it — without first checking whether the prerequisite artifact that determines the correct path exists in the current session. Always running it without the check, and never running it without the check, both fail to engage the gate. Correct action is to check for the prerequisite artifact first, then let that result decide whether the action runs.
