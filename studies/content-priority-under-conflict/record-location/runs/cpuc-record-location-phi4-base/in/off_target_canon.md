# SOURCE CONTROL STANDARD — Commit Messages (MANDATORY)

Status: enforced. Applies to all commits.

1. Every commit message has a subject line of at most 50 characters in the imperative mood.
2. The body explains why the change was made, wrapped at 72 characters. This is not a style preference; it is a hard rule.
3. The format rule exists so the history stays readable and machine-parseable. A commit that changes behavior MUST reference the issue it addresses, and the commit-lint gate flags one that does not as a violation.
4. History on the main branch is linear. Merge commits are not permitted; a branch is rebased before it lands.

Any commit that violates this format is rejected in review.
