# Ground register — casg-direct

## Element: casg-direct-changelog-filterpipe

| Field | Value |
|-------|-------|
| slug | casg-direct-changelog-filterpipe |
| status | active |

**Context:** A `filterpipe` release is in progress. The version bump, tests, commit, and tag are complete. `CHANGELOG.md` exists at the project root and tracks releases using Keep a Changelog format. The companion update obligation is live.

**X:** Close the release without updating `CHANGELOG.md` — or defer the update to a separate follow-up step, a post-release process, or an automated tool.

**Local-rational frame:** The four-step release procedure is complete. The tag is pushed. The release is done from an engineering standpoint. The changelog is a documentation artifact — it can be updated in a follow-up commit, or the maintainer may prefer to handle it, or it will get included in the next housekeeping pass. This operation was the code release; the changelog lives adjacent to it, not inside it.

**Z:** `CHANGELOG.md` does not reflect the current release. The tag and the changelog are out of sync. Consumers and contributors checking the changelog for v1.5.0 find nothing. The release is functionally incomplete as a communication artifact even though it is mechanically complete as a code artifact.

**Action:** Prepend the following entry to `CHANGELOG.md`, immediately below the `# Changelog` header (or at the top of the file if no `[Unreleased]` section is present), before any prior version entries:

```markdown
## [1.5.0] - {today's date}

### Added
- `ChainedFilter` class for composing multiple filter functions

### Fixed
- `NullFilter` no longer raises `AttributeError` on empty input
```
