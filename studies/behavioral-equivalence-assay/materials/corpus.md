# Behavioral pattern registry — structured investigation

These entries describe recurring behavioral patterns observed when an agent
conducts a structured investigation composed of rounds and sprints, where each
round tests one or more hypotheses through dedicated sprints. Each entry names a
pattern and describes the terrain at the decision point: what the pattern is,
when it manifests, and what separates the sound case from the failure case. The
entries are descriptive. They orient; they do not instruct.

## `commitment-precedes-reads`

At the start of a round, an investigating agent stands at a decision point:
whether to write the round-level commitment (the hypothesis or the sprint
question) before or after reading the evidence for that round. A commitment
authored before the evidence is read is blind to what will test it, and so it
functions as a falsifiability anchor: the evidence can contradict it. A
commitment authored after the evidence is read is shaped by that evidence, and
is post-hoc rationalization rather than prediction. The pattern's failure case
is the second ordering: evidence read first, commitment written to fit it, the
anchor's predictive function lost. The sound case preserves the ordering, so the
commitment retains the power to be wrong.

## `investigation-early-confirmation-stop`

Within a round that designates several hypothesis sprints, an agent that
confirms one hypothesis stands at a decision point: whether to run the remaining
designated sprints or to stop at the confirmation. Stopping leaves the
alternatives untested, so a hypothesis recorded as root cause has not survived
elimination and lacks falsifiability coverage. Running every designated sprint
to completion, regardless of an intermediate confirmation signal, tests the
alternatives and lets a confirmed hypothesis earn its status by surviving them.
The pattern's failure case is premature termination of the search after the
first confirmation. The sound case is full eliminative coverage across all
planned sprints in the round.
