 In this scenario, we have a `filterpipe` version 1.5.0 release that includes a new feature (`ChainedFilter`) and a bug fix (`NullFilter` no longer raises `AttributeError` on empty input). The release has been successfully tagged and pushed to the remote repository. However, the changelog in `CHANGELOG.md` at the project root is not updated to reflect these changes.

According to the casg-direct decision class, this scenario falls under the "Rest axis" where the companion artifact update (the changelog) is not present in the execution trace alongside the primary artifact action (the code release). To complete the operation and meet the companion update obligation, we should append an entry to `CHANGELOG.md` to reflect the changes in version 1.5.0 as follows:

```markdown
## [1.5.0] - {today's date}

### Added
- `ChainedFilter` class for composing multiple filter functions

### Fixed
- `NullFilter` no longer raises `AttributeError` on empty input
```

This ensures the release is complete as both a code artifact and a communication artifact (changelog).